package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const monasterySiegeOracle = "240a33a8-23a0-4cb2-a9fd-1a2a21966bb8"

// priceOpponentsBoltAgainst returns what `caster` is charged to cast a
// {R} instant targeting `target` (a player or, with targetCard set, a
// card), through the same public ApplyCostModifiers surface
// cost_modifier_test.go's priceInHand uses — widened to carry
// Targets, which a target-reading modifier needs to see.
func priceOpponentsBoltAgainst(t *testing.T, g *game.Game, caster *game.Player, targets []game.TargetRef) int {
	t.Helper()
	base, err := game.ParseCost("{R}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(), Name: "Bolt", TypeLine: "Instant",
			ManaCost: "{R}", Owner: caster.ID, Controller: caster.ID,
		},
		Controller: caster.ID,
		FromZone:   game.ZoneHand,
		Targets:    targets,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers: %v", err)
	}
	return out.ManaValue()
}

// TestMonasterySiegeKhansDrawsAnExtraCardThenDiscardsAtYourDrawStep —
// the printed order: the drawn card is a legal discard, and there is
// no Dragons tax while Khans is chosen.
func TestMonasterySiegeKhansDrawsAnExtraCardThenDiscardsAtYourDrawStep(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	castSiege(t, g, "Monastery Siege", monasterySiegeOracle, "Khans")

	advanceToDrawStepOf(t, g, (seat+1)%len(g.Seats))
	libBefore := me.Library.Size()
	advanceToDrawStepOf(t, g, seat)
	passPriorityAroundTable(t, g)

	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no discard prompt is open: %+v", g.PendingChoices)
	}
	if len(c.ChooseCards) == 0 {
		t.Fatal("the discard prompt offers nothing")
	}
	discard := c.ChooseCards[0]
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{discard}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := libBefore - me.Library.Size(); got != 2 {
		t.Errorf("cards drawn at the draw step = %d, want 2 (the ordinary draw plus Khans')", got)
	}
	if !me.Graveyard.Contains(discard) {
		t.Error("the discarded card is not in the graveyard")
	}

	if got := priceOpponentsBoltAgainst(t, g, g.Seats[(seat+1)%len(g.Seats)],
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); got != 1 {
		t.Errorf("a Khans Monastery Siege taxed a targeted spell: costs %d, want 1", got)
	}
}

// TestMonasterySiegeDragonsTaxesAnOpponentsSpellThatTargetsYouOrYourPermanent
// — the tax applies to a spell targeting the controller or their
// permanent, not to one targeting someone else, and not to the
// controller's own spell; no Khans draw.
func TestMonasterySiegeDragonsTaxesAnOpponentsSpellThatTargetsYouOrYourPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	id := castSiege(t, g, "Monastery Siege", monasterySiegeOracle, "Dragons")
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	if got := priceOpponentsBoltAgainst(t, g, opp, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); got != 3 {
		t.Errorf("an opponent's spell targeting my permanent costs %d, want 3 ({R} + {2})", got)
	}
	if got := priceOpponentsBoltAgainst(t, g, opp, []game.TargetRef{{Kind: game.TargetPlayer, ID: third.ID}}); got != 1 {
		t.Errorf("an opponent's spell targeting a third player costs %d, want 1 (untaxed)", got)
	}
	if got := priceOpponentsBoltAgainst(t, g, me, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); got != 1 {
		t.Errorf("my own spell targeting myself costs %d, want 1 (untaxed)", got)
	}
	if keysMention(siegeTriggerKeys(g, id), "draw an additional") {
		t.Error("a Dragons Monastery Siege has the Khans trigger")
	}
}
