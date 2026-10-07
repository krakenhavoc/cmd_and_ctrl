package lobby

// playmat_http.go is ADR 0128's routes: a signed-in person's playmat,
// the image behind their part of the board that everyone at the table
// sees.
//
//	GET    /me/playmat        — signed in: {path, wash}, or {} for none
//	PUT    /me/playmat        — signed in: multipart "image" (+ "wash")
//	PATCH  /me/playmat        — signed in: {"wash": 30..90}
//	DELETE /me/playmat        — signed in: remove it
//	GET    /playmats/{file}   — any session: the image
//
// Every write puts the result on each seat the person holds at a table
// that is not archived (Lobby.ApplyPlaymat), so the table sees it at
// once.

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/playmats"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/ratelimit"
)

// maxPlaymatRequestBytes caps the whole multipart upload: the image's
// own cap plus room for the form around it.
const maxPlaymatRequestBytes = playmats.MaxImageBytes + 64*1024

// playmatResponse is the body of every /me/playmat answer. A person
// with no playmat gets {}.
type playmatResponse struct {
	Path string `json:"path,omitempty"`
	Wash int    `json:"wash,omitempty"`
}

func playmatResponseOf(p playmats.Playmat) playmatResponse {
	return playmatResponse{Path: p.Path(), Wash: p.Wash}
}

func (c Config) playmatStore() playmats.Store {
	if c.Playmats == nil {
		return playmats.NoStore{}
	}
	return c.Playmats
}

// pushPlaymat puts p on the person's live seats. A lobby that is not
// wired (tests of the store alone) changes no seat.
func (c Config) pushPlaymat(user auth.Principal, p playmats.Playmat) {
	if c.Lobby == nil {
		return
	}
	c.Lobby.ApplyPlaymat(user.UserID, p.Path(), p.Wash)
}

// myPlaymat is GET /me/playmat.
func myPlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	cur, err := c.playmatStore().Get(r.Context(), p.UserID)
	if errors.Is(err, playmats.ErrNotFound) {
		return writePlaymat(w, playmatResponse{})
	}
	if err != nil {
		c.logger().Error("reading a user's playmat failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load your playmat; try again")
	}
	return writePlaymat(w, playmatResponseOf(cur))
}

// putMyPlaymat is PUT /me/playmat: a multipart body with the image as
// "image" and, optionally, "wash" (default playmats.DefaultWash).
//
//   - 400 for a missing image, one that is not a PNG, JPEG, GIF or
//     WebP (sniffed from the bytes), or a wash out of range.
//   - 413 for an image over 4 MiB.
//   - 429 past the per-person bucket.
//   - 503 when this server cannot store images (no data directory or
//     no database).
func putMyPlaymat(limit *ratelimit.Limiter) lobbyHandler {
	return func(c Config, w http.ResponseWriter, r *http.Request) error {
		p, err := signedInUser(r)
		if err != nil {
			return err
		}
		if !c.PlaymatFiles.Enabled() {
			return httpError(http.StatusServiceUnavailable, "playmats cannot be saved on this server")
		}
		if !limit.Allow("user:" + p.UserID.String()) {
			w.Header().Set("Retry-After", "5")
			return httpError(http.StatusTooManyRequests, "too many playmat uploads; try again in a few seconds")
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxPlaymatRequestBytes)
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			var mbe *http.MaxBytesError
			if errors.As(err, &mbe) {
				return httpError(http.StatusRequestEntityTooLarge, "a playmat image must be 4 MiB or smaller")
			}
			return httpError(http.StatusBadRequest, "invalid multipart body: "+err.Error())
		}
		defer func() { _ = r.MultipartForm.RemoveAll() }()
		wash := playmats.DefaultWash
		if v := r.FormValue("wash"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil {
				return httpError(http.StatusBadRequest, "wash must be a whole number")
			}
			wash = n
		}
		if err := playmats.ValidWash(wash); err != nil {
			return httpError(http.StatusBadRequest, validationMessage(err))
		}
		file, _, err := r.FormFile("image")
		if err != nil {
			return httpError(http.StatusBadRequest, "image is required")
		}
		defer func() { _ = file.Close() }()
		name, err := c.PlaymatFiles.Save(file)
		switch {
		case errors.Is(err, playmats.ErrTooLarge):
			return httpError(http.StatusRequestEntityTooLarge, "a playmat image must be 4 MiB or smaller")
		case errors.Is(err, playmats.ErrInvalid):
			return httpError(http.StatusBadRequest, validationMessage(err))
		case err != nil:
			c.logger().Error("storing a playmat image failed", "err", err)
			return httpError(http.StatusInternalServerError, "could not save your playmat; try again")
		}
		old, err := c.playmatStore().Put(r.Context(), p.UserID, name, wash)
		if err != nil {
			_ = c.PlaymatFiles.Remove(name)
			if errors.Is(err, playmats.ErrNoStore) {
				return httpError(http.StatusServiceUnavailable, "playmats cannot be saved on this server")
			}
			c.logger().Error("saving a user's playmat failed", "err", err)
			return httpError(http.StatusInternalServerError, "could not save your playmat; try again")
		}
		saved := playmats.Playmat{File: name, Wash: wash}
		c.pushPlaymat(p, saved)
		if old.File != "" && old.File != name {
			if err := c.PlaymatFiles.Remove(old.File); err != nil {
				c.logger().Warn("removing a replaced playmat image failed", "err", err)
			}
		}
		return writePlaymat(w, playmatResponseOf(saved))
	}
}

// patchPlaymatRequest is the body of PATCH /me/playmat.
type patchPlaymatRequest struct {
	Wash int `json:"wash"`
}

// patchMyPlaymat is PATCH /me/playmat: change the wash of the playmat
// the person already has. 404 when they have none.
func patchMyPlaymat(limit *ratelimit.Limiter) lobbyHandler {
	return func(c Config, w http.ResponseWriter, r *http.Request) error {
		p, err := signedInUser(r)
		if err != nil {
			return err
		}
		var req patchPlaymatRequest
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			return httpError(http.StatusBadRequest, "invalid body: "+err.Error())
		}
		if err := playmats.ValidWash(req.Wash); err != nil {
			return httpError(http.StatusBadRequest, validationMessage(err))
		}
		if !limit.Allow("user:" + p.UserID.String()) {
			w.Header().Set("Retry-After", "1")
			return httpError(http.StatusTooManyRequests, "too many playmat changes; try again in a second")
		}
		saved, err := c.playmatStore().SetWash(r.Context(), p.UserID, req.Wash)
		switch {
		case errors.Is(err, playmats.ErrNotFound):
			return httpError(http.StatusNotFound, "you have no playmat; upload one first")
		case errors.Is(err, playmats.ErrNoStore):
			return httpError(http.StatusServiceUnavailable, "playmats cannot be saved on this server")
		case err != nil:
			c.logger().Error("changing a user's playmat failed", "err", err)
			return httpError(http.StatusInternalServerError, "could not change your playmat; try again")
		}
		c.pushPlaymat(p, saved)
		return writePlaymat(w, playmatResponseOf(saved))
	}
}

// deleteMyPlaymat is DELETE /me/playmat. Removing a playmat that is
// not there is not an error: the answer is {} either way.
func deleteMyPlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	old, err := c.playmatStore().Delete(r.Context(), p.UserID)
	switch {
	case errors.Is(err, playmats.ErrNotFound):
		return writePlaymat(w, playmatResponse{})
	case err != nil:
		c.logger().Error("removing a user's playmat failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not remove your playmat; try again")
	}
	c.pushPlaymat(p, playmats.Playmat{})
	if err := c.PlaymatFiles.Remove(old.File); err != nil {
		c.logger().Warn("removing a playmat image failed", "err", err)
	}
	return writePlaymat(w, playmatResponse{})
}

// playmatImage is GET /playmats/{file}: any session, since every
// player and spectator at a table sees every seat's playmat.
func playmatImage(c Config, w http.ResponseWriter, r *http.Request) error {
	err := c.PlaymatFiles.Serve(w, r, r.PathValue("file"))
	switch {
	case err == nil:
		return nil
	case errors.Is(err, playmats.ErrNoStore):
		return httpError(http.StatusServiceUnavailable, "playmats are not stored on this server")
	case errors.Is(err, playmats.ErrNotFound):
		return httpError(http.StatusNotFound, "no such playmat")
	default:
		return httpError(http.StatusInternalServerError, fmt.Sprintf("could not serve the playmat: %v", err))
	}
}

// writePlaymat is writeJSON for a person's own playmat: never cached,
// since it changes under the same URL.
func writePlaymat(w http.ResponseWriter, body playmatResponse) error {
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, body)
}
