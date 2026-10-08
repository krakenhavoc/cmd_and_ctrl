package lobby

// playmat_http.go is ADR 0128's routes: a signed-in person's saved
// playmats (up to three, one of them on show; §11), which the table
// draws behind their battlefield.
//
//	GET    /me/playmats                the slots, which is active, the wash
//	PUT    /me/playmats/{slot}         multipart upload into a slot, part "file"
//	POST   /me/playmats/{slot}/link    {"url": "https://..."}; fetched once, stored
//	POST   /me/playmats/{slot}/fit     {"x": n, "y": n}; crop to the best size
//	DELETE /me/playmats/{slot}         remove a slot's playmat
//	PUT    /me/playmats/active         {"slot": n | null}; which one is on show
//	PATCH  /me/playmats                {"wash": 30..90}; one wash per account
//	GET    /playmats/{id}              the image, to any signed-in session
//
// Caller rule is the rest of /me/*: a signed-in person, so a guest, the
// admin token and a deployment with no database get 403 (never 401,
// #1154). A server with no data directory answers the same people
// {"enabled":false} on GET and 503 on a write.

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/playmat"
)

// maxPlaymatRequestBytes caps a whole upload request: the image's own
// cap plus room for the multipart envelope. Past it the reader stops
// before anything is decoded.
const maxPlaymatRequestBytes = playmat.MaxUploadBytes + 64<<10

// playmatSlotBody is one saved playmat on the wire.
type playmatSlotBody struct {
	Slot   int    `json:"slot"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	// Fits is true when the image needs no fitting. Suggestion is what
	// "fit to best size" would do, present only when it would change
	// something: a wrong-shaped image. A right-shaped image that is
	// small has Fits false and no suggestion, because the fix is a
	// bigger image and the server never enlarges one.
	Fits       bool                `json:"fits"`
	Suggestion *playmat.Suggestion `json:"suggestion,omitempty"`
}

// playmatResponse is the body of every /me/playmats route.
type playmatResponse struct {
	// Enabled is false on a server with nowhere to store a playmat.
	// The client hides the Settings section then.
	Enabled bool `json:"enabled"`
	// MaxSlots, IdealWidth and IdealHeight are the server's constants,
	// so the client prints the ones the server fits to.
	MaxSlots    int `json:"max_slots,omitempty"`
	IdealWidth  int `json:"ideal_width,omitempty"`
	IdealHeight int `json:"ideal_height,omitempty"`
	// Slots holds the occupied slots, in slot order; absent for none.
	Slots []playmatSlotBody `json:"slots,omitempty"`
	// Active is the slot the table shows; absent when none is, which is
	// a legal state for a person with saved mats.
	Active *int `json:"active,omitempty"`
	// Slot names the slot an upload, link or fit just wrote, so the
	// client can open the prompt on its suggestion. Absent otherwise.
	Slot int `json:"slot,omitempty"`
	// Wash is the owner-set darkness over every one of their mats, in
	// percent (ADR 0128 §10): theirs, or playmat.DefaultWash. Sent
	// whenever the feature is enabled.
	Wash int `json:"wash,omitempty"`
}

func (c Config) playmatService() *playmat.Service { return c.Playmats }

// writePlaymats answers a route with the account's current state. The
// same URL stops being the answer the moment they change a slot, so it
// is never cached. written is the slot a write just made, or 0.
func writePlaymats(c Config, w http.ResponseWriter, r *http.Request, user uuid.UUID, written int) error {
	svc := c.playmatService()
	st, err := svc.State(r.Context(), user)
	if err != nil {
		return playmatError(c, w, err)
	}
	body := playmatResponse{
		Enabled:     true,
		MaxSlots:    playmat.MaxSlots,
		IdealWidth:  playmat.IdealWidth,
		IdealHeight: playmat.IdealHeight,
		Wash:        svc.Wash(user),
		Slot:        written,
	}
	for _, s := range st.Slots {
		body.Slots = append(body.Slots, playmatSlotBody{
			Slot: s.Slot, URL: s.URL, Width: s.Width, Height: s.Height,
			Fits: s.Fits, Suggestion: s.Suggestion,
		})
	}
	if st.Active != 0 {
		a := st.Active
		body.Active = &a
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, body)
}

// playmatSlot reads the {slot} path value. A slot outside 1 to 3 is a
// 400, whatever is behind it.
func playmatSlot(r *http.Request) (int, error) {
	n, err := strconv.Atoi(r.PathValue("slot"))
	if err != nil || !playmat.ValidSlot(n) {
		return 0, httpError(http.StatusBadRequest, playmat.ErrBadSlot.Error())
	}
	return n, nil
}

// playmatWriter is the shared start of every write: a signed-in person,
// an enabled feature. It returns the service and what the table shows
// now, which notifyPlaymat compares against once the write is done.
func playmatWriter(c Config, w http.ResponseWriter, r *http.Request) (playmatCaller, error) {
	p, err := signedInUser(r)
	if err != nil {
		return playmatCaller{}, err
	}
	svc := c.playmatService()
	if !svc.Enabled() {
		return playmatCaller{}, playmatError(c, w, playmat.ErrDisabled)
	}
	return playmatCaller{svc: svc, user: p.UserID, before: svc.URL(p.UserID)}, nil
}

type playmatCaller struct {
	svc    *playmat.Service
	user   uuid.UUID
	before string
}

// notify tells the tables this person sits at, when the playmat they
// show is no longer the one it was before the write. Saving into an
// inactive slot changes nothing the table can see, so it costs no push.
func (pc playmatCaller) notify(c Config) {
	if now := pc.svc.URL(pc.user); now != pc.before {
		c.playmatChanged(pc.user, now)
	}
}

// patchPlaymatRequest is the body of PATCH /me/playmats.
type patchPlaymatRequest struct {
	Wash int `json:"wash"`
}

// patchMyPlaymats is PATCH /me/playmats: the owner-set wash (ADR 0128
// §10), 30 to 90, one per account. It may be set before any image is
// saved. The table sees the change at once, through the same push a
// change of mat uses.
func patchMyPlaymats(c Config, w http.ResponseWriter, r *http.Request) error {
	pc, err := playmatWriter(c, w, r)
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
	if err := pc.svc.SetWash(r.Context(), pc.user, req.Wash); err != nil {
		if errors.Is(err, playmat.ErrBadWash) {
			return httpError(http.StatusBadRequest, err.Error())
		}
		return playmatError(c, w, err)
	}
	if url := pc.svc.URL(pc.user); url != "" {
		c.playmatChanged(pc.user, url)
	}
	return writePlaymats(c, w, r, pc.user, 0)
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
	case errors.Is(err, playmat.ErrBadSlot), errors.Is(err, playmat.ErrBadCrop):
		return httpError(http.StatusBadRequest, err.Error())
	case errors.Is(err, playmat.ErrNoSlot):
		return httpError(http.StatusNotFound, err.Error())
	case errors.Is(err, playmat.ErrAlreadyFits), errors.Is(err, playmat.ErrConflict):
		return httpError(http.StatusConflict, err.Error())
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

// myPlaymats is GET /me/playmats.
func myPlaymats(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	if !c.playmatService().Enabled() {
		w.Header().Set("Cache-Control", "no-store")
		return writeJSON(w, http.StatusOK, playmatResponse{})
	}
	return writePlaymats(c, w, r, p.UserID, 0)
}

// putMyPlaymat is PUT /me/playmats/{slot}: a multipart body whose
// "file" part is the image. The part's own Content-Type and file name
// are ignored; the bytes are decoded (playmat.Normalize) and are the
// only evidence.
func putMyPlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	pc, err := playmatWriter(c, w, r)
	if err != nil {
		return err
	}
	slot, err := playmatSlot(r)
	if err != nil {
		return err
	}
	if r.ContentLength > maxPlaymatRequestBytes {
		return playmatError(c, w, playmat.ErrTooLarge)
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxPlaymatRequestBytes)
	data, err := readPlaymatPart(r)
	if err != nil {
		return err
	}
	if _, err := pc.svc.SetFromBytes(r.Context(), pc.user, slot, data); err != nil {
		return playmatError(c, w, err)
	}
	pc.notify(c)
	return writePlaymats(c, w, r, pc.user, slot)
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

// linkMyPlaymat is POST /me/playmats/{slot}/link: the server fetches
// the URL once, behind the SSRF guard, and stores the result exactly
// like an upload. The link is never stored and never reaches another
// player.
func linkMyPlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	pc, err := playmatWriter(c, w, r)
	if err != nil {
		return err
	}
	slot, err := playmatSlot(r)
	if err != nil {
		return err
	}
	var req playmatLinkRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return err
	}
	if _, err := pc.svc.SetFromURL(r.Context(), pc.user, slot, req.URL); err != nil {
		return playmatError(c, w, err)
	}
	pc.notify(c)
	return writePlaymats(c, w, r, pc.user, slot)
}

// playmatFitRequest is the body of POST /me/playmats/{slot}/fit: the
// crop's top-left corner, in the stored image's pixels. The crop's size
// is the server's (the largest ideal-shaped rectangle that fits), so a
// client names only where it sits.
type playmatFitRequest struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// fitMyPlaymat is POST /me/playmats/{slot}/fit: crop the stored image
// to the ideal shape at the given origin, scale it down to the ideal
// size (never up) and store it under a new id. A mat that is already
// the ideal shape is a 409.
func fitMyPlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	pc, err := playmatWriter(c, w, r)
	if err != nil {
		return err
	}
	slot, err := playmatSlot(r)
	if err != nil {
		return err
	}
	var req playmatFitRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return err
	}
	if _, err := pc.svc.Fit(r.Context(), pc.user, slot, req.X, req.Y); err != nil {
		return playmatError(c, w, err)
	}
	pc.notify(c)
	return writePlaymats(c, w, r, pc.user, slot)
}

// deleteMyPlaymat is DELETE /me/playmats/{slot}. Removing an empty slot
// is a success.
func deleteMyPlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	pc, err := playmatWriter(c, w, r)
	if err != nil {
		return err
	}
	slot, err := playmatSlot(r)
	if err != nil {
		return err
	}
	if err := pc.svc.Remove(r.Context(), pc.user, slot); err != nil {
		return playmatError(c, w, err)
	}
	pc.notify(c)
	return writePlaymats(c, w, r, pc.user, 0)
}

// activatePlaymatRequest is the body of PUT /me/playmats/active. A null
// slot shows none and keeps every saved mat.
type activatePlaymatRequest struct {
	Slot *int `json:"slot"`
}

// activateMyPlaymat is PUT /me/playmats/active: which saved playmat the
// table shows. It never touches a file.
func activateMyPlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	pc, err := playmatWriter(c, w, r)
	if err != nil {
		return err
	}
	var req activatePlaymatRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return err
	}
	slot := 0
	if req.Slot != nil {
		if !playmat.ValidSlot(*req.Slot) {
			return playmatError(c, w, playmat.ErrBadSlot)
		}
		slot = *req.Slot
	}
	if err := pc.svc.Activate(r.Context(), pc.user, slot); err != nil {
		return playmatError(c, w, err)
	}
	pc.notify(c)
	return writePlaymats(c, w, r, pc.user, 0)
}

// adminRemovePlaymat is DELETE /admin/users/{id}/playmat: moderation
// that takes away ALL of an account's saved playmats (ADR 0128 §9,
// §11). It does what the person's own removals do, to anyone's account,
// and the table sees the mat go on the next snapshot. The person can
// upload again; this removes images, it does not ban the feature.
// Removing none is a success, so a second click is harmless.
func adminRemovePlaymat(c Config, w http.ResponseWriter, r *http.Request) error {
	return adminPlaymatRemoval(c, w, r, 0)
}

// adminRemovePlaymatSlot is DELETE /admin/users/{id}/playmats/{slot}:
// one saved playmat, the account view's per-thumbnail Remove.
func adminRemovePlaymatSlot(c Config, w http.ResponseWriter, r *http.Request) error {
	slot, err := playmatSlot(r)
	if err != nil {
		return err
	}
	return adminPlaymatRemoval(c, w, r, slot)
}

// adminPlaymatRemoval removes one slot, or every slot when slot is 0.
func adminPlaymatRemoval(c Config, w http.ResponseWriter, r *http.Request, slot int) error {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil || id == uuid.Nil {
		return httpError(http.StatusBadRequest, "invalid user id")
	}
	svc := c.playmatService()
	if !svc.Enabled() {
		return playmatError(c, w, playmat.ErrDisabled)
	}
	before := svc.URL(id)
	if slot == 0 {
		err = svc.RemoveAll(r.Context(), id)
	} else {
		err = svc.Remove(r.Context(), id, slot)
	}
	if err != nil {
		if errors.Is(err, playmat.ErrNoUser) {
			return httpError(http.StatusNotFound, "user not found")
		}
		return playmatError(c, w, err)
	}
	if now := svc.URL(id); now != before {
		c.playmatChanged(id, now)
	}
	c.logger().Info("a playmat was removed by an admin", "user_id", id, "slot", slot)
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
