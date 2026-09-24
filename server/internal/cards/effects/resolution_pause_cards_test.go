package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_pause_cards_test.go — #1289 on the two other cards whose
// 0/0 lives through a paused counter placement. Earthshape and
// Widespread Brutality are in counter_tail_cards_test.go. Every board
// here is a real Doubling Season beside a real Hardened Scales, which
// makes the counter placement wait on the CR 616 ordering prompt.

// TestEarthbendingLessonOnABareLandOnAPausingBoard: earthbend 4 makes
// the land a 0/0 and its counters wait on the prompt. Before #1289 the
// state-based sweep killed it there and the return brought back a new,
// tapped, plain land.
func TestEarthbendingLessonOnABareLandOnAPausingBoard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	land := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")

	castCatalogSpell(t, g, "Earthbending Lesson", "Sorcery — Lesson", earthbendingLessonOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	passPriorityAroundTable(t, g)

	if findBattlefieldCardByID(g, land) == nil {
		t.Fatal("the earthbent 0/0 land died while its counters were owed to the CR 616 prompt")
	}
	// Doubling Season first: 4 → 8 → 9.
	answerOrderBySource(t, g, ds, hs)

	if got := countersOn(g, land, game.CounterPlusOne); got != 9 {
		t.Fatalf("+1/+1 counters on the land = %d, want 9", got)
	}
	if p, tough, creatureLand, _ := earthbentBody(t, g, land); !creatureLand || p != 9 || tough != 9 {
		t.Errorf("the land is %d/%d (creature land %v), want a 9/9 land creature", p, tough, creatureLand)
	}
}

// TestDreadhordeInvasionsFirstArmySurvivesAPausingBoard: the upkeep
// amass with no Army. Doubling Season doubles the creation, so the
// choose-an-Army prompt comes first and then the counter-order prompt.
// Both 0/0 Armies live until the resolution is over, and then the one
// that got no counter dies to CR 704.5f.
func TestDreadhordeInvasionsFirstArmySurvivesAPausingBoard(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Dreadhorde Invasion",
		TypeLine:   "Enchantment",
		OracleID:   dreadhordeInvasionOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)

	advanceToUpkeepOfSeat(t, g, seat)
	passPriorityAroundTable(t, g)

	var armies []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype(game.ArmySubtype) {
			armies = append(armies, c.InstanceID)
		}
	}
	if len(armies) != 2 {
		t.Fatalf("%d Armies at the choose prompt, want the two fresh 0/0s", len(armies))
	}
	pick := g.PendingChoices[0]
	if pick.Kind != game.PendingChoiceChooseCards {
		t.Fatalf("first prompt = %q, want the choose-an-Army prompt", pick.Kind)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{armies[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	// Doubling Season first: amass 1 → 2 → 3.
	answerOrderBySource(t, g, ds, hs)

	if got := countersOn(g, armies[1], game.CounterPlusOne); got != 3 {
		t.Fatalf("chosen Army counters = %d, want 3", got)
	}
	if findBattlefieldCardByID(g, armies[0]) != nil {
		t.Error("the Army with no counters is still a 0/0 on the battlefield after the resolution")
	}
}
