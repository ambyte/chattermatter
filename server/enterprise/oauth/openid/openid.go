// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package openid

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"

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

func init() {
	einterfaces.RegisterOAuthProvider(model.ServiceOpenid, &OpenIDProvider{})
}

func (p *OpenIDProvider) GetSSOSettings(_ request.CTX, config *model.Config, service string) (*model.SSOSettings, error) {
	sso := config.GetSSOService(service)
	if sso == nil {
		return nil, errors.New("sso settings not found")
	}

	return sso, nil
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
