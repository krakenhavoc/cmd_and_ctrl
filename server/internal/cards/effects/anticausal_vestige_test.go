package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const anticausalVestigeOracle = "aceea999-90ab-472c-86a7-48af1542cbcf"

// The whole card, through the path #324 was filed on: Anticausal
// Vestige is cast for its warp cost of {4}, enters, is exiled at the
// next end step — and that exile is a leaves-the-battlefield, so its
// trigger draws a card and then offers a TAPPED put of a permanent
// card whose mana value is at most the number of lands its controller
// controls. The just-drawn card is a candidate; a card over the land
// count and a nonpermanent are not. The Vestige is left in exile,
// castable on a later turn.
func TestAnticausalVestigeWarpsThenPaysOffOnLeaving(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aangPushLand(g, me.ID, "Plains", false)
	aangPushLand(g, me.ID, "Island", false)

	vestige := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: vestige, Name: "Anticausal Vestige", TypeLine: "Creature — Eldrazi",
		ManaCost: "{6}", Power: 7, Toughness: 5, OracleID: anticausalVestigeOracle,
		Owner: me.ID, Controller: me.ID,
	})
	twoDrop := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: twoDrop, Name: "Two-Drop Rock", TypeLine: "Artifact", ManaCost: "{2}",
		Owner: me.ID, Controller: me.ID,
	})
	threeDrop := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: threeDrop, Name: "Three-Drop Rock", TypeLine: "Artifact", ManaCost: "{3}",
		Owner: me.ID, Controller: me.ID,
	})
	instant := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: instant, Name: "Cheap Instant", TypeLine: "Instant", ManaCost: "{U}",
		Owner: me.ID, Controller: me.ID,
	})
	drawn := uuid.New()
	me.Library.PushTop(game.Card{
		InstanceID: drawn, Name: "Drawn Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Owner: me.ID, Controller: me.ID,
	})

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	b06AddMana(me, "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, vestige, game.CastSpellParams{Strict: true, AlternativeCost: "warp"}); err != nil {
		t.Fatalf("warp cast for {4}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(vestige) {
		t.Fatal("the warped Vestige did not enter")
	}

	// The end step: warp exiles it, which is the leave the trigger
	// watches. One pass resolves the warp exile, the next the trigger.
	advanceThroughEndStep(t, g)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(vestige) {
		t.Fatal("the Vestige was not exiled by warp")
	}
	if !me.Hand.Contains(drawn) {
		t.Fatal("the leaves-the-battlefield trigger did not draw a card")
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no put-from-hand prompt after the draw")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 — \"you MAY put a permanent card\"", pick.ChooseMin, pick.ChooseMax)
	}
	if !hasID(pick.ChooseCards, twoDrop) {
		t.Error("a mana value 2 permanent with two lands was not offered")
	}
	if !hasID(pick.ChooseCards, drawn) {
		t.Error("the card just drawn was not offered — the put comes AFTER the draw")
	}
	if hasID(pick.ChooseCards, threeDrop) {
		t.Error("a mana value 3 permanent was offered with only two lands")
	}
	if hasID(pick.ChooseCards, instant) {
		t.Error("an instant was offered; the clause says PERMANENT card")
	}

	answerChooseCards(t, g, me.ID, twoDrop)
	if !g.Battlefield.Contains(twoDrop) {
		t.Fatal("the chosen permanent did not reach the battlefield")
	}
	if !tappedOnBattlefield(t, g, twoDrop) {
		t.Error("the put permanent entered untapped; the clause says TAPPED")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(vestige); perm == nil {
		t.Error("the exiled Vestige carries no warp cast-later grant")
	}
}

// Declining the "you may" still draws: the draw is not conditional on
// the put, and nothing enters.
func TestAnticausalVestigeDeclinedPutStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aangPushLand(g, me.ID, "Plains", false)
	rock := handCardForTest(me, "Rock", "Artifact", "")
	vestige := pushCatalogPermanent(g, me.ID, "Anticausal Vestige", "Creature — Eldrazi", anticausalVestigeOracle, false)
	hand := me.Hand.Size()

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(vestige); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after the trigger, want %d (drew one)", got, hand+1)
	}
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("dying (a leave) did not offer the put")
	}
	answerChooseCards(t, g, me.ID)
	if g.Battlefield.Contains(rock) {
		t.Error("a declined put put a permanent onto the battlefield")
	}
}
