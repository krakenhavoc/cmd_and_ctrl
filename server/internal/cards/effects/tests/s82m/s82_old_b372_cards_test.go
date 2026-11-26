package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b372_cards_test.go — Journey into Nyx, batch 372 (#5558, S82 OLD).
// Every test plays the real catalog entry through the engine.

const (
	oldB372LeechesOracle     = "662a2158-4176-4b4a-be6a-13f4501d22f9"
	oldB372ThaumaturgeOracle = "70d2f5f7-f727-48e1-9532-8d365b1c1ae5"
	oldB372RollickOracle     = "ecda721d-ec5d-484b-a834-f4de96083b08"
	oldB372AnimusOracle      = "4c9a38f8-8fcb-45f9-8b70-e54298c7b485"
)

func TestOldB372CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB372LeechesOracle:     "Squelching Leeches",
		oldB372ThaumaturgeOracle: "Battlefield Thaumaturge",
		oldB372RollickOracle:     "Rollick of Abandon",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB372SquelchingLeechesIsAsBigAsYourSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	effectstest.PushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	leeches := effectstest.B12Push(g, me.ID, "Squelching Leeches", "Creature — Leech", oldB372LeechesOracle, 0, 0)
	if p, tough := effectstest.PtOf(t, g, leeches); p != 2 || tough != 2 {
		t.Errorf("with two Swamps (and an opponent's) it is %d/%d, want 2/2", p, tough)
	}
}

func TestOldB372RollickOfAbandonGivesEveryCreaturePlusTwoMinusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := effectstest.B12Creature(g, me.ID, "Small Bear", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Rollick of Abandon", "Sorcery", oldB372RollickOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if effectstest.ZoneOfCard(g, small) != game.ZoneGraveyard {
		t.Errorf("the 2/2 is in %v, want the graveyard (0 toughness)", effectstest.ZoneOfCard(g, small))
	}
	if p, tough := effectstest.PtOf(t, g, big); p != 5 || tough != 1 {
		t.Errorf("the 3/3 is %d/%d, want 5/1", p, tough)
	}
}

// oldB372Cast casts a card of the given name, type line and mana cost from hand
// with the given targets, paying `colored` plus `generic` colourless mana.
func oldB372Cast(t *testing.T, g *game.Game, me *game.Player, oracle, name, manaCost, colored string, generic int, targets []game.TargetRef) error {
	t.Helper()
	card := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Instant", ManaCost: manaCost,
		OracleID: oracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	effectstest.PayMana(t, g, me, colored)
	effectstest.B09AddColorless(me, generic)
	return g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: targets})
}

func TestOldB372ThaumaturgeDiscountsOnePerCreatureTheSpellTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Battlefield Thaumaturge", "Creature — Human Wizard", oldB372ThaumaturgeOracle, 2, 1)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.ToMain(t, g)
	// {3}{G} Ancient Animus targets two creatures: {2} less is {1}{G}.
	if err := oldB372Cast(t, g, me, oldB372AnimusOracle, "Ancient Animus", "{3}{G}", "{G}", 1,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}}); err != nil {
		t.Fatalf("two creature targets, {1}{G} paid for a {3}{G} spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// A Bolt at a player targets no creature: no discount.
	if err := oldB372Cast(t, g, me, effectstest.LightningBoltOracle, "Lightning Bolt", "{2}{R}", "{R}", 1,
		effectstest.B16TargetPlayer(opp.ID)); err == nil {
		t.Error("a spell that targets no creature got the discount")
	}
}

func TestOldB372ThaumaturgeDiscountsOnlyInstantsAndSorceriesYouCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Battlefield Thaumaturge", "Creature — Human Wizard", oldB372ThaumaturgeOracle, 2, 1)
	effectstest.ToMain(t, g)
	bear := game.Card{InstanceID: uuid.New(), Name: "Costly Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(bear)
	effectstest.PayMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a creature spell got the discount")
	}
}

func TestOldB372ThaumaturgeGainsHexproofWhenYouTargetIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	thaum := effectstest.B12Push(g, me.ID, "Battlefield Thaumaturge", "Creature — Human Wizard", oldB372ThaumaturgeOracle, 2, 5)
	effectstest.ToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, thaum, "hexproof") {
		t.Fatal("hexproof before it was targeted")
	}
	if err := oldB372Cast(t, g, me, effectstest.LightningBoltOracle, "Lightning Bolt", "{R}", "{R}", 0,
		[]game.TargetRef{{Kind: game.TargetCard, ID: thaum}}); err != nil {
		t.Fatalf("cast at my own Thaumaturge: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, thaum, "hexproof") {
		t.Error("no hexproof after a spell targeted it")
	}
}
