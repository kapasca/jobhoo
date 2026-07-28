package middleware

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

const csrfCookieName = "jobhoo_csrf"

type csrfCtxKey string

const csrfTokenCtxKey csrfCtxKey = "jobhoo_csrf_token"

// CSRF implements the double-submit-cookie pattern: a random token is set as
// a cookie, and every state-changing request (POST/PUT/PATCH/DELETE) must
// echo that same value back via either a hidden form field (csrf_token) or
// an X-CSRF-Token header (used by HTMX requests via hx-headers on <body>,
// see web/templates/layouts/base.html). An attacker forging a cross-site
// request can make the browser send the cookie automatically, but can't
// read its value to put in the form field/header — so the two won't match.
//
// This alone (without a server-side session-bound token) is intentionally
// simple: it's the standard baseline CSRF defense and doesn't require
// touching the sessions table, but a stateful per-session token would be a
// reasonable hardening step before a security-sensitive production launch.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ensureCSRFCookie(w, r)
		ctx := context.WithValue(r.Context(), csrfTokenCtxKey, token)
		r = r.WithContext(ctx)

		if isStateChanging(r.Method) {
			submitted := r.Header.Get("X-CSRF-Token")
			if submitted == "" {
				submitted = r.FormValue("csrf_token")
			}
			if subtle.ConstantTimeCompare([]byte(submitted), []byte(token)) != 1 {
				http.Error(w, "invalid or missing CSRF token", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// CSRFToken retrieves the current request's CSRF token for embedding in
// forms and the HTMX global header.
func CSRFToken(r *http.Request) string {
	token, _ := r.Context().Value(csrfTokenCtxKey).(string)
	return token
}

func isStateChanging(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func ensureCSRFCookie(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(csrfCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	buf := make([]byte, 24)
	_, _ = rand.Read(buf)
	token := base64.RawURLEncoding.EncodeToString(buf)

	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		// Deliberately NOT HttpOnly: the double-submit pattern requires the
		// page (or at least this middleware reading r.Cookie server-side)
		// to be able to read it back out to embed in forms/headers.
	})
	return token
}
