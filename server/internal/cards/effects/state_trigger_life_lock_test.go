package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// state_trigger_life_lock_test.go — ADR 0107 §1's state triggers beside
// §5's "players can't gain life" (#1880). A refused gain is no event and
// changes no life total (CR 119.7), so it neither triggers a state
// trigger nor hides one: a state the LOSS half reached still triggers.

const (
	stlOpalAvenger    = "13a85c9f-f653-4d90-bea6-94022ee82527"
	stlTranscendence  = "f9279f8f-3d19-4f5b-b194-cb8c65313b7e"
	stlSulfuricVortex = "7652f328-e142-494b-a869-772ced10c26a"
)

func stlSettle(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for i := 0; i < 6; i++ {
		passPriorityAroundTable(t, g)
		if !answerAnyTriggerOrderPrompt(t, g, me.ID) {
			passPriorityAroundTable(t, g)
			return
		}
	}
}

// Transcendence's "gain 2 for each 1 lost" is refused under Sulfuric
// Vortex, so the life total never reaches 20 and Transcendence's loss
// never triggers; Opal Avenger's "10 or less life" was reached by the
// loss itself, and it triggers.
func TestStateTriggersUnderCantGainLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	apaPush(g, opp.ID, opp.ID, stCard("Sulfuric Vortex", stlSulfuricVortex, "Enchantment", 0, 0))
	avenger := apaPush(g, me.ID, me.ID, stCard("Opal Avenger", stlOpalAvenger, "Enchantment", 0, 0))
	me.Life = 19
	apaPush(g, me.ID, me.ID, stCard("Transcendence", stlTranscendence, "Enchantment", 0, 0))
	passPriorityAroundTable(t, g)

	// Lose 1 at 19: the 2-life gain is refused, so no 20, no loss.
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(avenger, me.ID, -1); err != nil {
			t.Fatal(err)
		}
	})
	stlSettle(t, g, me)
	if g.State != game.StateActive || me.Eliminated {
		t.Fatal("Transcendence's controller lost: a refused gain reached 20 life")
	}
	if me.Life != 18 {
		t.Fatalf("life %d, want 18 — the gain should have been refused", me.Life)
	}

	// Lose 8 to reach 10: Opal Avenger's state was reached by the loss,
	// and it becomes a creature although the gain beside it is refused.
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(avenger, me.ID, -8); err != nil {
			t.Fatal(err)
		}
	})
	stlSettle(t, g, me)
	if me.Life != 10 {
		t.Fatalf("life %d, want 10", me.Life)
	}
	if c := apaLive(g, avenger); c == nil || !c.IsCreature() {
		t.Fatal("Opal Avenger did not become a creature at 10 life")
	}
}
