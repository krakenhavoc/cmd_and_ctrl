package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_counter_ability_test.go — #1211. A countered ABILITY has no card
// of its own: its id names a StackMeta entry the client cannot look
// up, and its source permanent is still standing on the battlefield.
// The log line therefore has to name the ability by its LABEL, which
// is what the stack overlay prints for the same item.
//
// Before this the event put the countered ability's own source card in
// `Source`, which the projection reads as "who countered it", so the
// line came out as "Their Rock countered <uuid>" — the victim in the
// attacker's place and a raw id in the victim's, which is the one
// thing every log test in this package asserts never happens.

func TestLogNamesACounteredAbilityByItsLabel(t *testing.T) {
	g := buildActiveGame(t)
	opp := g.Seats[1]
	rock := battlefieldCard(t, g, "Their Rock", "Artifact", opp.ID)

	const label = "Their Rock — do a thing"
	if err := g.AnnounceTrigger(opp.ID, rock, game.AbilityParams{Label: label}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	var ability uuid.UUID
	g.WithWriteLock(func() {
		for id, it := range g.StackMeta {
			if it != nil && it.Label == label {
				ability = id
			}
		}
	})
	if ability == uuid.Nil {
		t.Fatal("the trigger is not on the stack")
	}
	if err := g.CounterAbility(ability); err != nil {
		t.Fatalf("CounterAbility: %v", err)
	}

	entry := findLog(t, ViewOfGame(g).Log, LogCounter)
	assertNoUUID(t, entry.Text)
	if entry.Label != label {
		t.Errorf("label = %q, want the ability's own %q", entry.Label, label)
	}
	if entry.Target != "" {
		t.Errorf("target = %q; an ability item has no card for the client to name", entry.Target)
	}
	if !strings.Contains(entry.Text, label) {
		t.Errorf("line = %q, want it to name the countered ability", entry.Text)
	}
	if strings.Contains(entry.Text, "Their Rock countered") {
		t.Errorf("line = %q — the ability's SOURCE is not the counter", entry.Text)
	}
}
