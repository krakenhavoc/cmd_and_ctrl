package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// one_color_source_test.go — #779, the enumerator half.
//
// `canPayExcluding` asks the auto-tapper, so until the planner learned
// to use a "N mana of any one color" source a bot with a Gilded Lotus
// and nothing else was never offered a cast the Lotus could fund. It
// had no way to discover the two-step "float the mana by hand, then
// cast", because the float is a separate move and the cast that
// follows it does not exist yet at the moment the bot has to choose.

// gildedLotus is the printed card as the engine reads it: one tap, one
// colour pick, three tokens of whatever was picked.
func gildedLotus() game.Card {
	return game.Card{
		Name:     "Gilded Lotus",
		TypeLine: "Artifact",
		ManaAbilities: []game.ManaAbilityShape{{
			TapCost:  true,
			Produced: "{W3|U3|B3|R3|G3}",
			Label:    "Add three mana of any one color",
		}},
	}
}

// A seat whose only blue source is a Gilded Lotus is offered the cast.
func TestBotIsOfferedACastOnlyALotusCanFund(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, gildedLotus())
	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Island", "Island"))
	handCard(active, game.Card{Name: "Big Blue Thing", TypeLine: "Creature — Whale", ManaCost: "{3}{U}{U}"})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if !hasLabel(moves, "Cast Big Blue Thing") {
		t.Fatalf("the {3}{U}{U} cast was not offered off Gilded Lotus + 2 Islands: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// And it is NOT offered a cast the one pick cannot fund: two colours
// out of one Lotus is not a payment.
func TestBotIsNotOfferedATwoColourCastOffOneLotus(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, gildedLotus())
	handCard(active, game.Card{Name: "Azorius Thing", TypeLine: "Creature — Bird", ManaCost: "{W}{U}"})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if hasLabel(moves, "Cast Azorius Thing") {
		t.Errorf("a {W}{U} cast was offered off one Gilded Lotus: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}
