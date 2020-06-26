// Discordgo - Discord bindings for Go
// Available at https://github.com/Bios-Marcel/discordgo

// Copyright 2015-2016 Bruce Marriner <bruce@sqls.net>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains high level helper functions and easy entry points for the
// entire discordgo package.  These functions are being developed and are very
// experimental at this point.  They will most likely change so please use the
// low level functions if that's a problem.

// Package discordgo provides Discord binding for Go
package discordgo

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// VERSION of DiscordGo, follows Semantic Versioning. (http://semver.org/)
const VERSION = "0.21.1"

// ErrMFA will be risen by New when the user has 2FA.
var ErrMFA = errors.New("account has 2FA enabled")

// NewWithToken creates a new Discord session and will use the given token
// for authorization.
func NewWithToken(userAgent, token string) (s *Session, err error) {
	s = createEmptySession(userAgent)
	//Make sure there's no unnecessary spaces / newlines from pasting.
	token = strings.TrimSpace(token)
	if strings.HasPrefix(strings.ToLower(token), "bot") {
		//Cut off "bot", ignoring casing and make sure it's "Bot "
		token = "Bot " + strings.TrimSpace(string([]rune(token)[3:]))
		s.Identify.Intents = MakeIntent(IntentsAllWithoutPrivileged)
	}

	s.Token = token
	s.Identify.Token = token
	s.MFA = strings.HasPrefix(token, "mfa")

	// The Session is now able to have RestAPI methods called on it.
	// It is recommended that you now call Open() so that events will trigger.

	return s, nil
}

func createEmptySession(userAgent string) *Session {
	session := &Session{
		State:                  NewState(),
		Ratelimiter:            NewRatelimiter(),
		StateEnabled:           true,
		Compress:               true,
		ShouldReconnectOnError: true,
		ShardID:                0,
		ShardCount:             1,
		MaxRestRetries:         3,
		UserAgent:              userAgent,
		Client:                 &http.Client{Timeout: (20 * time.Second)},
		sequence:               new(int64),
		LastHeartbeatAck:       time.Now().UTC(),
	}

	// These can be modified prior to calling Open()
	// Initilize the Identify Package with defaults
	session.Identify.Compress = true
	session.Identify.LargeThreshold = 250
	session.Identify.GuildSubscriptions = true
	session.Identify.Properties.OS = runtime.GOOS
	session.Identify.Properties.Browser = "Firefox"
	session.Identify.Intents = MakeIntent(IntentsAll)

	return session
}

// NewWithPassword creates a new Discord session and will sign in with the
// provided credentials.
//
// NOTE: While email/pass authentication is supported by DiscordGo it is
// HIGHLY DISCOURAGED by Discord. Please only use email/pass to obtain a token
// and then use that authentication token for all future connections.
// Also, doing any form of automation with a user (non Bot) account may result
// in that account being permanently banned from Discord.
func NewWithPassword(userAgent, username, password string) (s *Session, err error) {
	s = createEmptySession(userAgent)
	_, err = s.Login(username, password)
	if err != nil || s.Token == "" {
		if s.MFA {
			err = ErrMFA
		} else {
			err = fmt.Errorf("Unable to fetch discord authentication token. %v", err)
		}
	}

	return
}

// The Session is now able to have RestAPI methods called on it.
// It is recommended that you now call Open() so that events will trigger.

// NewWithPasswordAndMFA that also takes a MFA token generated by an authenticator.
//
// NOTE: While email/pass authentication is supported by DiscordGo it is
// HIGHLY DISCOURAGED by Discord. Please only use email/pass to obtain a token
// and then use that authentication token for all future connections.
// Also, doing any form of automation with a user (non Bot) account may result
// in that account being permanently banned from Discord.
func NewWithPasswordAndMFA(userAgent, username, password, mfaToken string) (s *Session, err error) {
	s = createEmptySession(userAgent)
	var loginInfo *LoginInfo
	loginInfo, err = s.Login(username, password)
	if err != nil || s.Token == "" {
		if s.MFA {
			if mfaToken == "" {
				err = ErrMFA
			} else {
				if loginInfo == nil {
					err = ErrMFA
					return
				}

				var token string
				token, err = s.totp(loginInfo.Ticket, mfaToken)
				if err != nil {
					return
				}

				s.Token = token
				s.Identify.Token = token
			}
		} else {
			err = fmt.Errorf("Unable to fetch discord authentication token. %v", err)
		}
		return
	}

	// The Session is now able to have RestAPI methods called on it.
	// It is recommended that you now call Open() so that events will trigger.

	return
}
