package testsbr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const channeledForceOracle = "e12ed186-01c1-4da5-b90f-cfbb77c4e8d0"

// channeledForceBoard puts Channeled Force and three other cards in the
// active seat's otherwise empty hand at a main phase with {2}{U}{R}
// floating, and a 3/3 on an opponent's battlefield.
func channeledForceBoard(t *testing.T) (g *game.Game, me, opp *game.Player, spell uuid.UUID, others []uuid.UUID, bear uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me.Hand.Cards = nil
	spell = uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: spell, Name: "Channeled Force", TypeLine: "Instant", ManaCost: "{2}{U}{R}",
		OracleID: channeledForceOracle, Owner: me.ID, Controller: me.ID,
	})
	for range 3 {
		others = append(others, effectstest.HandCard(me, "Filler", "Instant"))
	}
	bear = effectstest.AcPermanent(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{U}{R}")
	return g, me, opp, spell, others, bear
}

// CR 107.3a / 601.2f: X is announced with the cast and is the number of
// cards discarded as the cost, while the spell is on the stack; on
// resolution the target player draws X and the creature is dealt X.
func TestChanneledForceDiscardsXThenDrawsAndDealsX(t *testing.T) {
	g, me, opp, spell, others, bear := channeledForceBoard(t)
	before := opp.Hand.Size()
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Strict: true, XValue: 2, DiscardIDs: others[:2],
		Targets: []game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
			{Kind: game.TargetCard, ID: bear, Slot: 1},
		},
	})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for _, id := range others[:2] {
		if !me.Graveyard.Contains(id) {
			t.Fatalf("discard %s was not paid at announce", id)
		}
	}
	if !me.Hand.Contains(others[2]) {
		t.Fatal("the third card, not named, left the hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - before; got != 2 {
		t.Errorf("target player drew %d, want 2", got)
	}
	if got := effectstest.DamageMarkedOn(g, bear); got != 2 {
		t.Errorf("damage on the creature = %d, want 2", got)
	}
}

// "Up to one" target creature: the spell may be cast at the player alone.
func TestChanneledForceWithNoCreatureTargetStillDraws(t *testing.T) {
	g, me, _, spell, others, _ := channeledForceBoard(t)
	before := me.Hand.Size()
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Strict: true, XValue: 1, DiscardIDs: others[:1],
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID, Slot: 0}},
	})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The hand lost the spell and one discard, then drew one.
	if got := me.Hand.Size(); got != before-2+1 {
		t.Errorf("hand = %d, want %d", got, before-2+1)
	}
}

// The discards must number exactly the announced X, and X can't exceed
// the cards in hand: the spell itself is not one (CR 601.2a).
func TestChanneledForceRefusesADiscardThatIsNotX(t *testing.T) {
	g, me, opp, spell, others, _ := channeledForceBoard(t)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0}}
	for name, params := range map[string]game.CastSpellParams{
		"X=2, one discard":     {XValue: 2, DiscardIDs: others[:1]},
		"X=1, two discards":    {XValue: 1, DiscardIDs: others[:2]},
		"X=4, three in hand":   {XValue: 4, DiscardIDs: others},
		"the spell discarded":  {XValue: 1, DiscardIDs: []uuid.UUID{spell}},
		"a card twice":         {XValue: 2, DiscardIDs: []uuid.UUID{others[0], others[0]}},
		"X=1, nothing named":   {XValue: 1},
		"X=0, a discard named": {XValue: 0, DiscardIDs: others[:1]},
	} {
		params.Strict = true
		params.Targets = target
		err := g.CastSpell(me.ID, spell, params)
		if !errors.Is(err, game.ErrInvalidParam) && !errors.Is(err, game.ErrCardNotFound) {
			t.Errorf("%s: err = %v, want a refusal", name, err)
		}
		if g.Stack.Contains(spell) || me.Hand.Size() != 4 {
			t.Fatalf("%s: a refused cast moved a card", name)
		}
	}
}
