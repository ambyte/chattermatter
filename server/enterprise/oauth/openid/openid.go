// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package openid

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
	"github.com/mattermost/mattermost/server/v8/einterfaces"
)

type OpenIDProvider struct{}

type openIDClaims struct {
	Subject           string `json:"sub"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	PreferredUsername string `json:"preferred_username"`
	Nickname          string `json:"nickname"`
	UserName          string `json:"username"`
	UPN               string `json:"upn"`
	OID               string `json:"oid"`
	UserID            string `json:"user_id"`
	ID                string `json:"id"`
}

type openIDDiscoveryDocument struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
}

func init() {
	einterfaces.RegisterOAuthProvider(model.ServiceOpenid, &OpenIDProvider{})
}

func (p *OpenIDProvider) GetSSOSettings(_ request.CTX, config *model.Config, service string) (*model.SSOSettings, error) {
	sso := config.GetSSOService(service)
	if sso == nil {
		return nil, errors.New("sso settings not found")
	}

	resolved := *sso

	needDiscovery := strings.TrimSpace(model.SafeDereference(resolved.DiscoveryEndpoint)) != "" &&
		(strings.TrimSpace(model.SafeDereference(resolved.AuthEndpoint)) == "" ||
			strings.TrimSpace(model.SafeDereference(resolved.TokenEndpoint)) == "" ||
			strings.TrimSpace(model.SafeDereference(resolved.UserAPIEndpoint)) == "")

	if needDiscovery {
		doc, err := fetchOpenIDDiscovery(model.SafeDereference(resolved.DiscoveryEndpoint))
		if err != nil {
			return nil, err
		}

		if strings.TrimSpace(model.SafeDereference(resolved.AuthEndpoint)) == "" {
			resolved.AuthEndpoint = model.NewPointer(doc.AuthorizationEndpoint)
		}
		if strings.TrimSpace(model.SafeDereference(resolved.TokenEndpoint)) == "" {
			resolved.TokenEndpoint = model.NewPointer(doc.TokenEndpoint)
		}
		if strings.TrimSpace(model.SafeDereference(resolved.UserAPIEndpoint)) == "" {
			resolved.UserAPIEndpoint = model.NewPointer(doc.UserinfoEndpoint)
		}
	}

	if strings.TrimSpace(model.SafeDereference(resolved.AuthEndpoint)) == "" {
		return nil, errors.New("openid auth endpoint is empty")
	}
	if strings.TrimSpace(model.SafeDereference(resolved.TokenEndpoint)) == "" {
		return nil, errors.New("openid token endpoint is empty")
	}
	if strings.TrimSpace(model.SafeDereference(resolved.UserAPIEndpoint)) == "" {
		return nil, errors.New("openid user api endpoint is empty")
	}

	return &resolved, nil
}

func (p *OpenIDProvider) GetUserFromIdToken(_ request.CTX, idToken string) (*model.User, error) {
	if strings.TrimSpace(idToken) == "" {
		return nil, errors.New("id_token is empty")
	}

	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return nil, errors.New("invalid id_token format")
	}

	claims := &openIDClaims{}
	if err := decodeBase64SegmentToJSON(parts[1], claims); err != nil {
		return nil, err
	}

	return buildUserFromClaims(claims, nil)
}

func (p *OpenIDProvider) GetUserFromJSON(rctx request.CTX, data io.Reader, tokenUser *model.User, settings *model.SSOSettings) (*model.User, error) {
	claims := &openIDClaims{}
	if err := json.NewDecoder(data).Decode(claims); err != nil {
		return nil, err
	}

	return buildUserFromClaims(claims, &buildUserOptions{
		tokenUser: tokenUser,
		settings:  settings,
		rctx:      rctx,
	})
}

func (p *OpenIDProvider) IsSameUser(_ request.CTX, dbUser, oAuthUser *model.User) bool {
	if dbUser == nil || oAuthUser == nil {
		return false
	}

	if dbUser.AuthData != nil && oAuthUser.AuthData != nil && *dbUser.AuthData != "" && *oAuthUser.AuthData != "" {
		return *dbUser.AuthData == *oAuthUser.AuthData
	}

	return strings.EqualFold(dbUser.Email, oAuthUser.Email)
}

type buildUserOptions struct {
	tokenUser *model.User
	settings  *model.SSOSettings
	rctx      request.CTX
}

func buildUserFromClaims(claims *openIDClaims, opts *buildUserOptions) (*model.User, error) {
	if claims == nil {
		return nil, errors.New("claims are nil")
	}

	user := &model.User{}

	authData := firstNonEmpty(claims.Subject, claims.OID, claims.UserID, claims.ID)
	if authData == "" && opts != nil && opts.tokenUser != nil && opts.tokenUser.AuthData != nil {
		authData = *opts.tokenUser.AuthData
	}
	if authData == "" {
		return nil, errors.New("auth_data is empty")
	}
	user.AuthData = &authData
	user.AuthService = model.ServiceOpenid

	user.Email = strings.ToLower(strings.TrimSpace(claims.Email))
	if user.Email == "" && opts != nil && opts.tokenUser != nil {
		user.Email = strings.ToLower(strings.TrimSpace(opts.tokenUser.Email))
	}
	if user.Email == "" {
		return nil, errors.New("email is empty")
	}

	user.FirstName = strings.TrimSpace(claims.GivenName)
	user.LastName = strings.TrimSpace(claims.FamilyName)
	if user.FirstName == "" && user.LastName == "" {
		first, last := splitName(strings.TrimSpace(claims.Name))
		user.FirstName = first
		user.LastName = last
	}
	if opts != nil && opts.tokenUser != nil {
		if user.FirstName == "" {
			user.FirstName = opts.tokenUser.FirstName
		}
		if user.LastName == "" {
			user.LastName = opts.tokenUser.LastName
		}
	}

	username := resolveUsername(claims, opts)
	if username == "" {
		return nil, errors.New("username is empty")
	}
	if opts != nil && opts.rctx != nil {
		user.Username = model.CleanUsername(opts.rctx.Logger(), username)
	} else {
		user.Username = model.NormalizeUsername(strings.ReplaceAll(username, " ", "-"))
		if !model.IsValidUsername(user.Username) {
			user.Username = model.NewUsername()
		}
	}

	return user, nil
}

func resolveUsername(claims *openIDClaims, opts *buildUserOptions) string {
	if v := sanitizePreferred(claims.PreferredUsername); v != "" {
		return v
	}

	if v := strings.TrimSpace(claims.UserName); v != "" {
		return v
	}

	for _, candidate := range []string{
		claims.UserName,
		claims.Nickname,
		sanitizePreferred(claims.PreferredUsername),
		sanitizePreferred(claims.UPN),
		sanitizePreferred(claims.Email),
	} {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}

	if opts != nil && opts.tokenUser != nil {
		return strings.TrimSpace(opts.tokenUser.Username)
	}

	return ""
}

func sanitizePreferred(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}

	parts := strings.Split(v, "@")
	return parts[0]
}

func splitName(full string) (string, string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "", ""
	}

	parts := strings.Fields(full)
	if len(parts) == 1 {
		return parts[0], ""
	}

	return parts[0], strings.Join(parts[1:], " ")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}

	return ""
}

func decodeBase64SegmentToJSON(segment string, out any) error {
	b, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(segment))
	if err != nil {
		return err
	}

	return json.Unmarshal(b, out)
}

func fetchOpenIDDiscovery(discoveryURL string) (*openIDDiscoveryDocument, error) {
	req, err := http.NewRequest(http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch openid discovery document")
	}

	var doc openIDDiscoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}

	doc.AuthorizationEndpoint = resolveEndpointURL(discoveryURL, doc.AuthorizationEndpoint)
	doc.TokenEndpoint = resolveEndpointURL(discoveryURL, doc.TokenEndpoint)
	doc.UserinfoEndpoint = resolveEndpointURL(discoveryURL, doc.UserinfoEndpoint)

	if strings.TrimSpace(doc.AuthorizationEndpoint) == "" ||
		strings.TrimSpace(doc.TokenEndpoint) == "" ||
		strings.TrimSpace(doc.UserinfoEndpoint) == "" {
		return nil, errors.New("openid discovery document missing required endpoints")
	}

	return &doc, nil
}

func resolveEndpointURL(discoveryURL, endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ""
	}

	parsedEndpoint, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}

	if parsedEndpoint.IsAbs() {
		return endpoint
	}

	parsedDiscovery, err := url.Parse(discoveryURL)
	if err != nil {
		return endpoint
	}

	return parsedDiscovery.ResolveReference(parsedEndpoint).String()
}
