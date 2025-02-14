// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package container

import (
	"net/http"

	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/services/auth"
	"code.gitea.io/gitea/services/packages"
)

type Auth struct{}

func (a *Auth) Name() string {
	return "container"
}

// Verify extracts the user from the Bearer token
// If it's an anonymous session a ghost user is returned
func (a *Auth) Verify(req *http.Request, w http.ResponseWriter, store auth.DataStore, sess auth.SessionStore) *user_model.User {
	uid, err := packages.ParseAuthorizationToken(req)
	if err != nil {
		log.Trace("ParseAuthorizationToken: %v", err)
		return nil
	}

	if uid == 0 {
		return nil
	}
	if uid == -1 {
		return user_model.NewGhostUser()
	}

	u, err := user_model.GetUserByID(req.Context(), uid)
	if err != nil {
		log.Error("GetUserByID:  %v", err)
		return nil
	}

	return u
}
