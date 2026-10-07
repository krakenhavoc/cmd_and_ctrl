package lobby

// playmat_http.go is ADR 0128's routes: a signed-in person's playmat,
// one image per account that the table draws behind their battlefield.
//
//	GET    /me/playmat        metadata, or {"enabled":true} with no url
//	PUT    /me/playmat        multipart upload, part "file"
//	POST   /me/playmat/link   {"url": "https://..."}; fetched once, stored
//	DELETE /me/playmat        remove it
//	GET    /playmats/{id}     the image, to any signed-in session
//
// Caller rule is the rest of /me/*: a signed-in person, so a guest, the
// admin token and a deployment with no database get 403 (never 401,
// #1154). A server with no data directory answers the same people
// {"enabled":false} on GET and 503 on a write.

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/playmat"
)

// maxPlaymatRequestBytes caps a whole upload request: the image's own
// cap plus room for the multipart envelope. Past it the reader stops
// before anything is decoded.
const maxPlaymatRequestBytes = playmat.MaxUploadBytes + 64<<10

// playmatResponse is the body of every /me/playmat route.
type playmatResponse struct {
	// Enabled is false on a server with nowhere to store a playmat.
	// The client hides the Settings section then.
	Enabled bool   `json:"enabled"`
	URL     string `json:"url,omitempty"`
	Width   int    `json:"width,omitempty"`
	Height  int    `json:"height,omitempty"`
}

func (c Config) playmatService() *playmat.Service { return c.Playmats }

func playmatBody(info playmat.Info) playmatResponse {
	return playmatResponse{Enabled: true, URL: info.URL, Width: info.Width, Height: info.Height}
}

// writePlaymat is writeJSON for a person's own playmat: the same URL
// stops being the answer the moment they change it.
func writePlaymat(w http.ResponseWriter, body playmatResponse) error {
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, body)
}

// playmatError maps the playmat package's refusals to statuses. Every
// message is written for the person who uploaded the file.
func playmatError(c Config, w http.ResponseWriter, err error) error {
	switch {
	case errors.Is(err, playmat.ErrTooLarge), errors.Is(err, playmat.ErrTooManyPixels):
		return httpError(http.StatusRequestEntityTooLarge, err.Error())
	case errors.Is(err, playmat.ErrNotImage):
		return httpError(http.StatusUnsupportedMediaType, err.Error())
	case errors.Is(err, playmat.ErrFetch):
		return httpError(http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, playmat.ErrBusy):
		w.Header().Set("Retry-After", "5")
		return httpError(http.StatusServiceUnavailable, err.Error())
	case errors.Is(err, playmat.ErrDisabled):
		return httpError(http.StatusServiceUnavailable, "playmats are not available on this server")
	case errors.Is(err, playmat.ErrNoUser):
		return httpError(http.StatusForbidden, "this session is not signed in as a person")
	default:
		c.logger().Error("playmat request failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not save your playmat; try again")
	}
}

// myPlaymat is GET /me/playmat.
func myPlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	svc := c.playmatService()
	if !svc.Enabled() {
		return writePlaymat(w, playmatResponse{})
	}
	info, ok, err := svc.Get(r.Context(), p.UserID)
	if err != nil {
		return playmatError(c, w, err)
	}
	if !ok {
		return writePlaymat(w, playmatResponse{Enabled: true})
	}
	return writePlaymat(w, playmatBody(info))
}

// putMyPlaymat is PUT /me/playmat: a multipart body whose "file" part is
// the image. The part's own Content-Type and file name are ignored; the
// bytes are decoded (playmat.Normalize) and are the only evidence.
func putMyPlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	svc := c.playmatService()
	if !svc.Enabled() {
		return playmatError(c, w, playmat.ErrDisabled)
	}
	if r.ContentLength > maxPlaymatRequestBytes {
		return playmatError(c, w, playmat.ErrTooLarge)
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxPlaymatRequestBytes)
	data, err := readPlaymatPart(r)
	if err != nil {
		return err
	}
	info, err := svc.SetFromBytes(r.Context(), p.UserID, data)
	if err != nil {
		return playmatError(c, w, err)
	}
	c.playmatChanged(p.UserID, info.URL)
	return writePlaymat(w, playmatBody(info))
}

// readPlaymatPart streams the multipart body to its "file" part and
// reads at most the image cap from it. It never spools to a temp file.
func readPlaymatPart(r *http.Request) ([]byte, error) {
	mt, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mt != "multipart/form-data" || params["boundary"] == "" {
		return nil, httpError(http.StatusBadRequest, "send the image as multipart/form-data in a part named file")
	}
	mr := multipart.NewReader(r.Body, params["boundary"])
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			return nil, httpError(http.StatusBadRequest, "no image was sent; the part must be named file")
		}
		if err != nil {
			var mbe *http.MaxBytesError
			if errors.As(err, &mbe) {
				return nil, httpError(http.StatusRequestEntityTooLarge, playmat.ErrTooLarge.Error())
			}
			return nil, httpError(http.StatusBadRequest, "could not read the upload")
		}
		if part.FormName() != "file" {
			_ = part.Close()
			continue
		}
		data, err := playmat.ReadLimited(part)
		_ = part.Close()
		if errors.Is(err, playmat.ErrTooLarge) {
			return nil, httpError(http.StatusRequestEntityTooLarge, err.Error())
		}
		if err != nil {
			var mbe *http.MaxBytesError
			if errors.As(err, &mbe) {
				return nil, httpError(http.StatusRequestEntityTooLarge, playmat.ErrTooLarge.Error())
			}
			return nil, httpError(http.StatusBadRequest, "could not read the upload")
		}
		return data, nil
	}
}

type playmatLinkRequest struct {
	URL string `json:"url"`
}

// linkMyPlaymat is POST /me/playmat/link: the server fetches the URL
// once, behind the SSRF guard, and stores the result exactly like an
// upload. The link is never stored and never reaches another player.
func linkMyPlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	svc := c.playmatService()
	if !svc.Enabled() {
		return playmatError(c, w, playmat.ErrDisabled)
	}
	var req playmatLinkRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return err
	}
	info, err := svc.SetFromURL(r.Context(), p.UserID, req.URL)
	if err != nil {
		return playmatError(c, w, err)
	}
	c.playmatChanged(p.UserID, info.URL)
	return writePlaymat(w, playmatBody(info))
}

// deleteMyPlaymat is DELETE /me/playmat. Removing none is a success.
func deleteMyPlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	svc := c.playmatService()
	if !svc.Enabled() {
		return playmatError(c, w, playmat.ErrDisabled)
	}
	if err := svc.Remove(r.Context(), p.UserID); err != nil {
		return playmatError(c, w, err)
	}
	c.playmatChanged(p.UserID, "")
	return writePlaymat(w, playmatResponse{Enabled: true})
}

// adminRemovePlaymat is DELETE /admin/users/{id}/playmat: moderation
// for a mat that is not fit for a shared table (ADR 0128 section 9).
// It does what the person's own DELETE does, to anyone's account, and
// the table sees the mat go on the next snapshot. The person can
// upload another; this removes an image, it does not ban the feature.
// Removing none is a success, so a second click is harmless.
func adminRemovePlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil || id == uuid.Nil {
		return httpError(http.StatusBadRequest, "invalid user id")
	}
	svc := c.playmatService()
	if !svc.Enabled() {
		return playmatError(c, w, playmat.ErrDisabled)
	}
	if err := svc.Remove(r.Context(), id); err != nil {
		if errors.Is(err, playmat.ErrNoUser) {
			return httpError(http.StatusNotFound, "user not found")
		}
		return playmatError(c, w, err)
	}
	c.playmatChanged(id, "")
	c.logger().Info("a playmat was removed by an admin", "user_id", id)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// playmatChanged tells the tables a person sits at. Nil Lobby (a test
// with no lobby) is a no-op.
func (c Config) playmatChanged(user uuid.UUID, url string) {
	if c.Lobby != nil {
		c.Lobby.PlaymatChanged(user, url)
	}
}

// servePlaymat is GET /playmats/{id}.
//
// It sits behind auth.Middleware, like /avatars and /cards: the id is a
// 122-bit capability, but an image of someone's living room is worth a
// session check, and the browser sends the session cookie with an <img>
// request anyway. Any signed-in session may fetch any id, because every
// player at a table must see every other player's mat; there is no
// per-table check, and no listing route to find an id from.
//
// The bytes are our own JPEG, so the type is fixed, not sniffed or read
// from anything a client sent. The remaining headers keep the response
// inert even if something other than an image ever sat at the path.
func servePlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	svc := c.playmatService()
	if svc == nil {
		return httpError(http.StatusNotFound, "not found")
	}
	f, info, err := svc.Open(r.PathValue("id"))
	if err != nil {
		return httpError(http.StatusNotFound, "not found")
	}
	defer func() { _ = f.Close() }()
	h := w.Header()
	h.Set("Content-Type", "image/jpeg")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Content-Security-Policy", "default-src 'none'; sandbox")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("Cross-Origin-Resource-Policy", "same-origin")
	// A replacement is a new id, so the bytes at a URL never change.
	// private: a shared cache must not keep a session-gated image.
	h.Set("Cache-Control", "private, max-age=31536000, immutable")
	http.ServeContent(w, r, "", info.ModTime(), f)
	return nil
}
