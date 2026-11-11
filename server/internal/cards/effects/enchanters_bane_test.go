package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const enchantersBaneOracle = "1f0548f8-5dec-44b5-b5a9-e55fa4366e08"

func TestEnchantersBaneIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(enchantersBaneOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Enchanter's Bane: %+v %v", spec, ok)
	}
}

// enchantersBaneTable puts the Bane on my side and a {2}{W} enchantment on
// the opponent's, and walks to my end step with the enchantment targeted.
func enchantersBaneTable(t *testing.T) (g *game.Game, opp *game.Player, ench uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Enchanter's Bane", "Enchantment", enchantersBaneOracle, false)
	ench = uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: ench, Name: "Test Enchantment", TypeLine: "Enchantment", ManaCost: "{2}{W}",
		Owner: opp.ID, Controller: opp.ID,
	})
	advanceTo(t, g, game.StepEnd)
	pickTriggerTarget(t, g, me.ID, ench)
	passPriorityAroundTable(t, g)
	return g, opp, ench
}

func TestEnchantersBaneDealsItsManaValueUnlessSacrificed(t *testing.T) {
	g, opp, ench := enchantersBaneTable(t)
	before := opp.Life
	answerOptionPick(t, g, opp.ID, 0)
	if opp.Life != before-3 {
		t.Errorf("controller life %d → %d, want -3 (mana value of {2}{W})", before, opp.Life)
	}
	if !onBattlefield(g, ench) {
		t.Error("the enchantment was not supposed to be sacrificed")
	}
}

func TestEnchantersBaneSacrificeAvoidsTheDamage(t *testing.T) {
	g, opp, ench := enchantersBaneTable(t)
	before := opp.Life
	answerOptionPick(t, g, opp.ID, 1)
	if opp.Life != before {
		t.Errorf("life %d → %d, want unchanged after sacrificing", before, opp.Life)
	}
	if onBattlefield(g, ench) {
		t.Error("the enchantment was not sacrificed")
	}
}
