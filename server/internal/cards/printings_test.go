package cards

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestPrintingOptionsOnlyContainTheSamePlayableCard(t *testing.T) {
	idx := NewIndex()
	original := Card{ID: uuid.New(), OracleID: uuid.New(), Name: "Original", SetCode: "one", CollectorNo: "1", Lang: "en"}
	idx.Put(original)
	alternate := original
	alternate.ID, alternate.SetCode, alternate.FlavorName = uuid.New(), "two", "Alternate title"
	idx.Put(alternate)
	token := original
	token.ID, token.Layout = uuid.New(), "token"
	idx.Put(token)
	foreign := original
	foreign.ID, foreign.Lang = uuid.New(), "ja"
	idx.Put(foreign)
	other := original
	other.ID, other.OracleID, other.Name = uuid.New(), uuid.New(), "Other"
	idx.Put(other)
	w := httptest.NewRecorder()
	Handler(idx, nil).ServeHTTP(w, httptest.NewRequest("GET", "/cards/"+original.ID.String()+"/printings", nil))
	var response struct {
		Printings []Card `json:"printings"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || len(response.Printings) != 2 {
		t.Fatalf("unexpected options: %d %s", w.Code, w.Body.String())
	}
	for _, p := range response.Printings {
		if p.ID != original.ID && p.ID != alternate.ID {
			t.Fatalf("wrong identity offered: %+v", p)
		}
	}
}
