package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const continueOracle = "175c76c8-f36c-4948-aaa8-0208eac5f0d3"

func continueRefs(ids ...uuid.UUID) []game.TargetRef {
	refs := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return refs
}

func continueCast(t *testing.T, g *game.Game, me *game.Player, pool string, targets ...uuid.UUID) error {
	t.Helper()
	id := effectstest.SocB20HandCard(me, "Continue?", "Instant", "{1}{W}", continueOracle)
	return effectstest.SocB20CastStrict(t, g, me, id, pool, continueRefs(targets...))
}

// Up to four creature cards that died this turn return together, under
// their owner's control.
func TestContinueReturnsUpToFourCreaturesThatDiedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	var dead []uuid.UUID
	for i := 0; i < 4; i++ {
		id := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.DestroyNow(t, g, id)
		dead = append(dead, id)
	}
	if err := continueCast(t, g, me, "CW", dead...); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range dead {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Errorf("creature %d did not return", i)
			continue
		}
		if c.Controller != me.ID {
			t.Errorf("creature %d returned under %v, want its owner", i, c.Controller)
		}
	}
}

// A fifth target is one more than "up to four".
func TestContinueRefusesAFifthTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	var dead []uuid.UUID
	for i := 0; i < 5; i++ {
		id := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.DestroyNow(t, g, id)
		dead = append(dead, id)
	}
	if err := continueCast(t, g, me, "CW", dead...); err == nil {
		t.Fatal("five targets were accepted")
	}
}

// Not legal targets: a creature card discarded this turn, one that died
// last turn, a noncreature card that died this turn, an opponent's
// creature that died this turn, and a creature that died, went back to
// hand and was then discarded (a new object from the hand, CR 400.7).
func TestContinueRefusesCardsNotPutThereFromTheBattlefieldThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	old := effectstest.B12Creature(g, me.ID, "Old Bear", "Creature — Bear", 2, 2)
	effectstest.DestroyNow(t, g, old)
	for i := 0; i < 4; i++ {
		effectstest.EndTurn(t, g)
	}
	if g.Turn.ActiveSeat != me.Seat {
		t.Fatalf("not back to my turn: seat %d", g.Turn.ActiveSeat)
	}
	effectstest.AdvanceToMain(t, g)

	discarded := effectstest.PushGraveyardCardTyped(me, "Discarded Bear", "Creature — Bear")
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.DestroyNow(t, g, rock)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DestroyNow(t, g, theirs)
	again := effectstest.B12Creature(g, me.ID, "Again Bear", "Creature — Bear", 2, 2)
	effectstest.DestroyNow(t, g, again)
	gy := game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}
	hand := game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}
	if err := g.MoveCardByID(gy, hand, again); err != nil {
		t.Fatalf("to hand: %v", err)
	}
	if err := g.MoveCardByID(hand, gy, again); err != nil {
		t.Fatalf("back to the graveyard: %v", err)
	}

	for name, id := range map[string]uuid.UUID{
		"discarded creature card":      discarded,
		"last turn's creature":         old,
		"noncreature card":             rock,
		"opponent's creature":          theirs,
		"creature that came from hand": again,
	} {
		if err := continueCast(t, g, me, "CW", id); err == nil {
			t.Errorf("the %s was accepted as a target", name)
			effectstest.PassPriorityAroundTable(t, g)
		}
	}
	// A creature that did die this turn is still fine beside them.
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DestroyNow(t, g, bear)
	if err := continueCast(t, g, me, "CW", bear); err != nil {
		t.Fatalf("cast at a creature that died this turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("the creature that died this turn did not return")
	}
	if !me.Graveyard.Contains(again) || !me.Graveyard.Contains(old) {
		t.Error("an untargeted card left the graveyard")
	}
}

// {1}{W} is the cost: one white mana alone does not pay it.
func TestContinueRefusesAPoolThatCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DestroyNow(t, g, bear)
	if err := continueCast(t, g, me, "W", bear); err == nil {
		t.Fatal("cast for {W} was accepted")
	}
}
