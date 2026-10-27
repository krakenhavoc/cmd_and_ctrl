package effects

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opponent_target_keyed_test.go — #4361 slice 6: the controller's "which
// opponent chooses" question of an "of an opponent's choice" activation
// is a keyed option pick. Since slice 13 (owner decision D7) its parked
// announcement is written too (pendingChoices[].parked), so the question
// holds no frame the census counts. census_s13_restore_test.go restores it.
func TestTheOpponentChooserQuestionHoldsOnlyTheParkedAnnouncement(t *testing.T) {
	g := newCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	pentacle := pushCatalogPermanent(g, me.ID, "Nova Pentacle", "Artifact", otNovaPentacle, false)
	pr7Creature(g, a.ID, "Their Bear", 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, pentacle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	c := otOptionPick(t, g, me.ID)
	if c.PickThen != "option-pick/opponent-target-chooser" {
		t.Fatalf("the question's continuation key is %q, want the opponent-target chooser", c.PickThen)
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap struct {
		PendingChoices []struct {
			ID           string          `json:"id"`
			ResumeFrames []string        `json:"resumeFrames"`
			Parked       json.RawMessage `json:"parked"`
		} `json:"pendingChoices"`
	}
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, pc := range snap.PendingChoices {
		if pc.ID != c.ID.String() {
			continue
		}
		found = true
		if len(pc.ResumeFrames) != 0 {
			t.Errorf("the question holds %v, want no counted frame", pc.ResumeFrames)
		}
		if len(pc.Parked) == 0 {
			t.Error("the question's parked announcement is not written")
		}
	}
	if !found {
		t.Fatalf("the question is not in the snapshot: %s", raw)
	}
	// And the answer still asks the named opponent.
	otChooseOpponent(t, g, me.ID, a.ID)
	otPickFor(t, g, a.ID, me.ID)
}
