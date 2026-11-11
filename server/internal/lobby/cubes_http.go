package lobby

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cubes"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draft"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/moderation"
)

// A signed-in person's saved cubes (ADR 0222 §8, #5441; docs/lobby.md):
// the list behind "My cubes", the check of one, and its create, change
// and delete. Every route reads and writes only the caller's own cubes,
// keyed on the account of the session, as the rest of /me/* does: a
// cube that is not theirs is a 404, the answer one that does not exist
// gets.
//
// A cube's list is read by the deck importers (a pasted list, or a
// Moxfield or Archidekt link, fetched under the deck checker's own
// per-address fetch limit) and kept as text. The check is run against
// the catalog on every read, never stored.

// Stable codes of the cube routes' and a cube draft's error bodies.
const (
	CodeCubeNameTaken     = "cube_name_taken"
	CodeCubesFull         = "cubes_full"
	CodeCubeTooLarge      = "cube_too_large"
	CodeCubeNotReady      = "cube_not_ready"
	CodeCubeTooSmall      = "cube_too_small"
	CodeCubeNeedsSignIn   = "cube_needs_sign_in"
	cubeListMaxBytes      = 256 * 1024
	cubeBlockersInMessage = 5
)

// cubeInfo is one cube as the client sees it. List is present on the
// routes that return one cube and absent from the list of them.
type cubeInfo struct {
	ID        uuid.UUID    `json:"id"`
	Name      string       `json:"name"`
	Singleton bool         `json:"singleton"`
	List      string       `json:"list,omitempty"`
	Check     cubes.Report `json:"check"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type cubesResponse struct {
	Cubes []cubeInfo `json:"cubes"`
	// The limits the page states: how many cubes a person may keep, the
	// most cards one may hold, and the pack sizes a cube draft offers.
	MaxCubes      int `json:"max_cubes"`
	MaxCards      int `json:"max_cards"`
	MaxNameLength int `json:"max_name_length"`
	PackSize      int `json:"pack_size"`
	MinPackSize   int `json:"min_pack_size"`
	MaxPackSize   int `json:"max_pack_size"`
	PacksPerSeat  int `json:"packs_per_seat"`
}

type cubeResponse struct {
	Cube cubeInfo `json:"cube"`
}

// saveCubeRequest is the body of POST /me/cubes and PUT /me/cubes/{id}.
// A create needs a name and one of url and text. A change may send any
// of them; what it leaves out stays as it was.
type saveCubeRequest struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
	Text string `json:"text,omitempty"`
	// Singleton is true when left out of a create.
	Singleton *bool `json:"singleton,omitempty"`
}

// checkCube runs the check of a stored cube against the catalog as it
// is now.
func checkCube(c Config, cube cubes.Cube) cubes.Report {
	entries, err := cubes.ParseList(cube.List)
	if err != nil {
		// A stored list was written by Report.Text and reads back; one
		// that does not is a cube with nothing to deal.
		entries = nil
	}
	return cubes.Check(c.Cards, c.DraftSets.IsFull(), entries, cube.Singleton)
}

func cubeInfoOf(c Config, cube cubes.Cube, withList bool) cubeInfo {
	info := cubeInfo{ID: cube.ID, Name: cube.Name, Singleton: cube.Singleton, Check: checkCube(c, cube), UpdatedAt: cube.UpdatedAt}
	if withList {
		info.List = cube.List
	}
	return info
}

// cubeStore answers the caller and the store, or the refusal.
func cubeStore(c Config, r *http.Request) (uuid.UUID, cubes.Store, error) {
	p, err := signedInUser(r)
	if err != nil {
		return uuid.Nil, nil, err
	}
	if c.Cubes == nil {
		return uuid.Nil, nil, httpError(http.StatusServiceUnavailable, "cubes are not available on this server")
	}
	return p.UserID, c.Cubes, nil
}

func cubeIDFromPath(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return uuid.Nil, httpError(http.StatusNotFound, "no such cube")
	}
	return id, nil
}

// writeCubeError answers a store refusal that has its own status.
func writeCubeError(c Config, w http.ResponseWriter, err error) error {
	switch {
	case errors.Is(err, cubes.ErrNotFound):
		return httpError(http.StatusNotFound, "no such cube")
	case errors.Is(err, cubes.ErrNameTaken):
		return writeCoded(w, http.StatusConflict, CodeCubeNameTaken, "You already have a cube with that name.")
	case errors.Is(err, cubes.ErrTooMany):
		return writeCoded(w, http.StatusConflict, CodeCubesFull,
			fmt.Sprintf("You have %d cubes, the most you can keep. Delete one to save this one.", cubes.MaxCubes))
	}
	c.logger().Error("cubes: a store call failed", "err", err)
	return httpError(http.StatusInternalServerError, "could not save your cube; try again")
}

// myCubes is GET /me/cubes: the caller's cubes by name, each with its
// check, without their lists.
func myCubes(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	out := cubesResponse{
		Cubes:    []cubeInfo{},
		MaxCubes: cubes.MaxCubes, MaxCards: cubes.MaxCards, MaxNameLength: cubes.MaxNameLen,
		PackSize: draft.DefaultCubePackSize, MinPackSize: draft.MinCubePackSize, MaxPackSize: draft.MaxCubePackSize,
		PacksPerSeat: draft.Rounds,
	}
	if c.Cubes == nil {
		return writeJSON(w, http.StatusOK, out)
	}
	list, err := c.Cubes.List(r.Context(), p.UserID)
	if err != nil {
		c.logger().Error("listing a user's cubes failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load your cubes; try again")
	}
	for _, cube := range list {
		out.Cubes = append(out.Cubes, cubeInfoOf(c, cube, false))
	}
	return writeJSON(w, http.StatusOK, out)
}

// myCube is GET /me/cubes/{id}: one cube with its list and its check.
func myCube(c Config, w http.ResponseWriter, r *http.Request) error {
	owner, store, err := cubeStore(c, r)
	if err != nil {
		return err
	}
	id, err := cubeIDFromPath(r)
	if err != nil {
		return err
	}
	cube, err := store.Get(r.Context(), owner, id)
	if err != nil {
		return writeCubeError(c, w, err)
	}
	return writeJSON(w, http.StatusOK, cubeResponse{Cube: cubeInfoOf(c, cube, true)})
}

// deleteCube is DELETE /me/cubes/{id}.
func deleteCube(c Config, w http.ResponseWriter, r *http.Request) error {
	owner, store, err := cubeStore(c, r)
	if err != nil {
		return err
	}
	id, err := cubeIDFromPath(r)
	if err != nil {
		return err
	}
	if err := store.Delete(r.Context(), owner, id); err != nil {
		return writeCubeError(c, w, err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// cubeName trims and checks a cube's name: not empty, within
// cubes.MaxNameLen, and through the name filter (ADR 0222 §9).
func cubeName(name string) (string, error) {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		return "", httpError(http.StatusBadRequest, "give the cube a name")
	}
	if utf8.RuneCountInString(name) > cubes.MaxNameLen {
		return "", httpError(http.StatusBadRequest, fmt.Sprintf("a cube name is at most %d characters", cubes.MaxNameLen))
	}
	if err := moderation.Name(name); err != nil {
		return "", err
	}
	return name, nil
}

// cubeEntries reads the list a save sent: a link through the deck
// checker's fetch (its cache, its per-address fetch limit and its
// errors), or pasted text. fetched is the deck site's name for a link.
func (d *deckCheck) cubeEntries(c Config, w http.ResponseWriter, r *http.Request, body saveCubeRequest) (entries []deck.Entry, fetched string, handled bool, err error) {
	if body.URL != "" {
		entries, fetched, err = d.fetchCubeList(r.Context(), c, body.URL, func() bool {
			return d.fetchLimit == nil || d.fetchLimit.AllowRequest(r)
		})
		if errors.Is(err, errFetchLimited) {
			w.Header().Set("Retry-After", strconv.Itoa(fetchLimitRetryAfter))
			return nil, "", true, writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "Too many deck links fetched from here. Wait a few seconds and save again.",
			})
		}
		if err != nil {
			return nil, "", true, writeDeckFetchError(w, err)
		}
		return entries, fetched, false, nil
	}
	if len(body.Text) > cubeListMaxBytes {
		return nil, "", true, writeCoded(w, http.StatusBadRequest, CodeCubeTooLarge,
			fmt.Sprintf("That list is too long. A cube holds at most %d cards.", cubes.MaxCards))
	}
	entries, err = cubes.ParseList(body.Text)
	if err != nil {
		return nil, "", true, httpError(http.StatusBadRequest, "could not read that list: "+strings.TrimPrefix(err.Error(), "deck: "))
	}
	return entries, "", false, nil
}

// fetchCubeList fetches the list a link names, from the deck checker's
// cache when a check or a save of that link filled it. It builds no
// coverage report: a cube is not a deck. The fetch goes to the link's
// canonical URL, so the fetcher never sees a caller's query string.
func (d *deckCheck) fetchCubeList(ctx context.Context, c Config, rawURL string, spend func() bool) ([]deck.Entry, string, error) {
	ref, err := deck.ParseDeckURL(rawURL)
	if err != nil {
		return nil, "", newDeckFetchError(rawURL, err)
	}
	if e, ok := d.cache.get(ref.Key()); ok {
		return e.entries, e.name, nil
	}
	if !spend() {
		return nil, "", errFetchLimited
	}
	name, entries, err := c.deckFetcher()(ctx, ref.URL())
	if err != nil {
		return nil, "", newDeckFetchError(rawURL, err)
	}
	return entries, name, nil
}

// checkedList checks the entries a save sent and refuses a list a cube
// cannot be: one with no card, or one over cubes.MaxCards. A list with
// cards that block is saved, and its check names them.
func checkedList(c Config, w http.ResponseWriter, entries []deck.Entry, singleton bool) (cubes.Report, bool, error) {
	rep := cubes.Check(c.Cards, c.DraftSets.IsFull(), entries, singleton)
	switch {
	case rep.Cards == 0:
		return rep, true, httpError(http.StatusBadRequest, "that list names no cards: paste one card a line, or a deck list")
	case rep.TooLarge:
		return rep, true, writeCoded(w, http.StatusBadRequest, CodeCubeTooLarge,
			fmt.Sprintf("That list has %d cards. A cube holds at most %d.", rep.Cards, cubes.MaxCards))
	}
	return rep, false, nil
}

// createCube handles POST /me/cubes.
func (d *deckCheck) createCube(c Config, w http.ResponseWriter, r *http.Request) error {
	owner, store, err := cubeStore(c, r)
	if err != nil {
		return err
	}
	var body saveCubeRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	body.URL, body.Text = strings.TrimSpace(body.URL), strings.TrimSpace(body.Text)
	switch {
	case body.URL == "" && body.Text == "":
		return httpError(http.StatusBadRequest, "send a deck link as url, or a pasted list as text")
	case body.URL != "" && body.Text != "":
		return httpError(http.StatusBadRequest, "send url or text, not both")
	}
	if c.Cards == nil {
		return httpError(http.StatusServiceUnavailable, "the card index is not loaded on this server")
	}
	// The name first, when one was typed, so a refused name costs no
	// fetch.
	name := strings.TrimSpace(body.Name)
	if name != "" {
		if name, err = cubeName(name); err != nil {
			return err
		}
	}
	entries, fetched, handled, err := d.cubeEntries(c, w, r, body)
	if handled || err != nil {
		return err
	}
	if name == "" {
		// A link's own name stands in for one, cut to fit. It is the
		// caller's text and goes through the filter like a typed one.
		fetched = strings.Join(strings.Fields(fetched), " ")
		if utf8.RuneCountInString(fetched) > cubes.MaxNameLen {
			fetched = strings.TrimSpace(string([]rune(fetched)[:cubes.MaxNameLen]))
		}
		if name, err = cubeName(fetched); err != nil {
			return err
		}
	}
	singleton := body.Singleton == nil || *body.Singleton
	rep, handled, err := checkedList(c, w, entries, singleton)
	if handled || err != nil {
		return err
	}
	cube, err := store.Create(r.Context(), cubes.Cube{
		OwnerID: owner, Name: name, Singleton: singleton, List: rep.Text(), CardCount: rep.Cards,
	})
	if err != nil {
		return writeCubeError(c, w, err)
	}
	return writeJSON(w, http.StatusCreated, cubeResponse{Cube: cubeInfoOf(c, cube, true)})
}

// updateCube handles PUT /me/cubes/{id}: a new name, a new list, the
// singleton setting, or any of them together.
func (d *deckCheck) updateCube(c Config, w http.ResponseWriter, r *http.Request) error {
	owner, store, err := cubeStore(c, r)
	if err != nil {
		return err
	}
	id, err := cubeIDFromPath(r)
	if err != nil {
		return err
	}
	var body saveCubeRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	body.URL, body.Text = strings.TrimSpace(body.URL), strings.TrimSpace(body.Text)
	if body.URL != "" && body.Text != "" {
		return httpError(http.StatusBadRequest, "send url or text, not both")
	}
	cube, err := store.Get(r.Context(), owner, id)
	if err != nil {
		return writeCubeError(c, w, err)
	}
	if strings.TrimSpace(body.Name) != "" {
		if cube.Name, err = cubeName(body.Name); err != nil {
			return err
		}
	}
	if body.Singleton != nil {
		cube.Singleton = *body.Singleton
	}
	if body.URL != "" || body.Text != "" {
		if c.Cards == nil {
			return httpError(http.StatusServiceUnavailable, "the card index is not loaded on this server")
		}
		entries, _, handled, err := d.cubeEntries(c, w, r, body)
		if handled || err != nil {
			return err
		}
		rep, handled, err := checkedList(c, w, entries, cube.Singleton)
		if handled || err != nil {
			return err
		}
		cube.List, cube.CardCount = rep.Text(), rep.Cards
	}
	cube, err = store.Update(r.Context(), cube)
	if err != nil {
		return writeCubeError(c, w, err)
	}
	return writeJSON(w, http.StatusOK, cubeResponse{Cube: cubeInfoOf(c, cube, true)})
}

// cubeDraft is what POST /draft needs of a cube: its name and the pack
// source that deals it.
type cubeDraft struct {
	Name   string
	Source *draft.CubeSource
}

// cubeForDraft resolves the cube a draft was asked to deal: the
// caller's own, checked against the catalog as it is now, and large
// enough for the draft's seats and pack size (ADR 0222 §8). handled is
// true when it wrote the refusal itself.
func cubeForDraft(c Config, w http.ResponseWriter, r *http.Request, owner uuid.UUID, rawID string, seats, packSize int) (out cubeDraft, handled bool, err error) {
	if packSize == 0 {
		packSize = draft.DefaultCubePackSize
	}
	if packSize < draft.MinCubePackSize || packSize > draft.MaxCubePackSize {
		return out, true, httpError(http.StatusBadRequest, fmt.Sprintf("pack_size is %d (want %d to %d)", packSize, draft.MinCubePackSize, draft.MaxCubePackSize))
	}
	if c.Cubes == nil {
		return out, true, httpError(http.StatusServiceUnavailable, "cubes are not available on this server")
	}
	id, perr := uuid.Parse(strings.TrimSpace(rawID))
	if perr != nil {
		return out, true, httpError(http.StatusNotFound, "no such cube")
	}
	cube, gerr := c.Cubes.Get(r.Context(), owner, id)
	if gerr != nil {
		return out, true, writeCubeError(c, w, gerr)
	}
	rep := checkCube(c, cube)
	if !rep.Ready {
		return out, true, writeCoded(w, http.StatusUnprocessableEntity, CodeCubeNotReady, cubeNotReadyMessage(cube.Name, rep))
	}
	if need := cubes.Need(seats, packSize); rep.Cards < need {
		return out, true, writeCoded(w, http.StatusUnprocessableEntity, CodeCubeTooSmall,
			fmt.Sprintf("%q has %d cards. A draft of %d seats with packs of %d needs %d (%d seats x %d packs x %d cards).",
				cube.Name, rep.Cards, seats, packSize, need, seats, draft.Rounds, packSize))
	}
	src, serr := draft.NewCubeSource(rep.Deal(), packSize)
	if serr != nil {
		return out, false, serr
	}
	return cubeDraft{Name: cube.Name, Source: src}, false, nil
}

// cubeNotReadyMessage says why a cube cannot be drafted, naming the
// first few cards that block it.
func cubeNotReadyMessage(name string, rep cubes.Report) string {
	blockers := append(append(append([]string{}, rep.NotFull...), rep.Unknown...), rep.Repeated...)
	switch {
	case rep.Cards == 0:
		return fmt.Sprintf("%q has no cards.", name)
	case len(blockers) == 0:
		return fmt.Sprintf("%q cannot be drafted yet.", name)
	}
	shown := blockers
	more := ""
	if len(shown) > cubeBlockersInMessage {
		more = fmt.Sprintf(" and %d more", len(shown)-cubeBlockersInMessage)
		shown = shown[:cubeBlockersInMessage]
	}
	return fmt.Sprintf("%q cannot be drafted yet. Open it in My cubes and drop or fix: %s%s.", name, strings.Join(shown, ", "), more)
}
