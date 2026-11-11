package lobby

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/drafthistory"
)

// A signed-in person's finished drafts (ADR 0222 §9, #5440): the list
// behind "My drafts" and the read-only page of one. Both read only the
// caller's own records, keyed on the account of the session, as the
// rest of /me/* does.

type myDraftsResponse struct {
	Drafts []drafthistory.Record `json:"drafts"`
}

type myDraftResponse struct {
	Draft drafthistory.Record `json:"draft"`
}

// myDrafts is GET /me/drafts: the caller's drafts, newest first,
// without their cards.
func myDrafts(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	if c.DraftHistory == nil {
		return writeJSON(w, http.StatusOK, myDraftsResponse{Drafts: []drafthistory.Record{}})
	}
	drafts, err := c.DraftHistory.List(r.Context(), p.UserID)
	if err != nil {
		c.logger().Error("listing a user's drafts failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load your drafts; try again")
	}
	return writeJSON(w, http.StatusOK, myDraftsResponse{Drafts: drafts})
}

// myDraft is GET /me/drafts/{id}: one of the caller's drafts, with the
// cards they picked. 404 for a draft that is not theirs, the same
// answer as one that does not exist.
func myDraft(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil || c.DraftHistory == nil {
		return httpError(http.StatusNotFound, "no such draft")
	}
	rec, err := c.DraftHistory.Get(r.Context(), p.UserID, id)
	if errors.Is(err, drafthistory.ErrNotFound) {
		return httpError(http.StatusNotFound, "no such draft")
	}
	if err != nil {
		c.logger().Error("reading a user's draft failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load that draft; try again")
	}
	return writeJSON(w, http.StatusOK, myDraftResponse{Draft: rec})
}
