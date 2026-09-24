package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// split_second_test.go — #1519, CR 702.61. The five proof cards, each
// cast from its catalog entry with no sandbox flag: the keyword the
// entry declares is what turns split second on.

const (
	suddenShockOracle    = "7b139231-dcd2-4b03-bcf2-fcb040617b69"
	suddenDeathOracle    = "a3475ca6-d88e-4855-a554-1338624d1635"
	suddenEdictOracle    = "b95f704d-96b9-437e-8c49-aba874139e12"
	suddenSpoilingOracle = "dce202c7-fe8e-462a-858e-7a5a69bd5b6b"
)

// requireSplitSecondOn fails unless the spell is on the stack stamped
// with split second and the table-wide cache agrees.
func requireSplitSecondOn(t *testing.T, g *game.Game, spell uuid.UUID) {
	t.Helper()
	item := g.StackMeta[spell]
	if item == nil {
		t.Fatal("the spell is not on the stack")
	}
	if !item.SplitSecond || !g.SplitSecondActive {
		t.Fatalf("split second is off (item=%v, table=%v): the catalog declaration was not read",
			item.SplitSecond, g.SplitSecondActive)
	}
}

// Krosan Grip is the whole rule on one board. With the Grip on the
// stack aimed at their Sol Ring, the opponent:
//
//   - can't cast Lightning Bolt (CR 702.61a);
//   - can't activate Goblin Bombardment (CR 702.61a);
//   - CAN tap the Sol Ring for mana (CR 702.61b);
//
// while the caster's prowess creature still triggers off the Grip and
// resolves above it (CR 702.61b — triggered abilities are untouched).
// When the Grip has resolved the Sol Ring is gone, the pump landed,
// and the table can cast again.
func TestKrosanGripCannotBeRespondedTo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: solRingOracle, Owner: opp.ID, Controller: opp.ID,
	})
	bombardment := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Bombardment", TypeLine: "Enchantment",
		OracleID: goblinBombardmentOracle, Owner: opp.ID, Controller: opp.ID,
	})
	fodder := seedPermanentFor(g, opp.ID, "Goblin", "Creature — Goblin")
	monk := pushProwessCreature(g, me.ID, "Swiftspear", 1, 2)

	grip := castCatalogSpell(t, g, "Krosan Grip", "Instant", krosanGripOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ring}})
	requireSplitSecondOn(t, g, grip)
	if n := prowessItemsFrom(g, monk); n != 1 {
		t.Fatalf("CR 702.61b: the prowess trigger did not fire under split second (%d)", n)
	}

	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: lightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
	err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	})
	if !errors.Is(err, game.ErrSplitSecondActive) {
		t.Errorf("Lightning Bolt in response to Krosan Grip: got %v, want ErrSplitSecondActive", err)
	}
	err = g.ActivateCatalogAbility(opp.ID, bombardment, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	})
	if !errors.Is(err, game.ErrSplitSecondActive) {
		t.Errorf("Goblin Bombardment in response to Krosan Grip: got %v, want ErrSplitSecondActive", err)
	}
	if !g.Battlefield.Contains(fodder) {
		t.Error("the refused activation still sacrificed its cost")
	}
	if err := g.ActivateManaAbility(opp.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("CR 702.61b: Sol Ring's mana ability was refused under split second: %v", err)
	}

	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ring) {
		t.Error("Krosan Grip did not destroy the Sol Ring")
	}
	if g.SplitSecondActive {
		t.Error("split second outlived Krosan Grip")
	}
	if got := effectivePower(t, g, monk); got != 2 {
		t.Errorf("the prowess pump did not resolve: power %d, want 2", got)
	}
	err = g.CastSpell(opp.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	})
	if err != nil {
		t.Errorf("Lightning Bolt once the Grip had resolved: %v", err)
	}
}

func TestSuddenShockDealsTwoUnderSplitSecond(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	shock := castCatalogSpell(t, g, "Sudden Shock", "Instant", suddenShockOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	requireSplitSecondOn(t, g, shock)
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("Sudden Shock dealt %d, want 2", got)
	}
	if g.SplitSecondActive {
		t.Error("split second outlived Sudden Shock")
	}
}

func TestSuddenDeathGivesMinusFourUnderSplitSecond(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	giant := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Craw Wurm", TypeLine: "Creature — Wurm",
		Power: 6, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	death := castCatalogSpell(t, g, "Sudden Death", "Instant", suddenDeathOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: giant}})
	requireSplitSecondOn(t, g, death)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, giant), effectiveToughness(t, g, giant); p != 2 || tough != 1 {
		t.Errorf("Sudden Death left a %d/%d, want 2/1", p, tough)
	}
}

func TestSuddenEdictMakesTheTargetPlayerSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keep := seedPermanentFor(g, opp.ID, "Their Keeper", "Creature — Bear")
	lose := seedPermanentFor(g, opp.ID, "Their Fodder", "Creature — Bear")
	mine := seedPermanentFor(g, me.ID, "My Bear", "Creature — Bear")

	edict := castCatalogSpell(t, g, "Sudden Edict", "Instant", suddenEdictOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	requireSplitSecondOn(t, g, edict)
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the caster was asked to sacrifice; only the target player is")
	}
	answerSacrifice(t, g, opp.ID, lose)
	if g.Battlefield.Contains(lose) || !g.Battlefield.Contains(keep) || !g.Battlefield.Contains(mine) {
		t.Errorf("wrong creature left: lose=%v keep=%v mine=%v",
			g.Battlefield.Contains(lose), g.Battlefield.Contains(keep), g.Battlefield.Contains(mine))
	}
}

// Sudden Spoiling: every creature the target player controls as it
// resolves loses all abilities and is a base 0/2 until end of turn. An
// anthem still pumps it on top (7c after 7b), a creature of the
// caster's is untouched, and one the player gets afterwards is not in
// the CR 611.2c set.
func TestSuddenSpoilingSilencesAndShrinksTheTargetPlayersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Serra Angel", TypeLine: "Creature — Angel",
		Power: 4, Toughness: 4, Keywords: []string{"flying", "vigilance"},
		Owner: opp.ID, Controller: opp.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: gloriousAnthemOracle, Owner: opp.ID, Controller: opp.ID,
	})
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Angel", TypeLine: "Creature — Angel",
		Power: 4, Toughness: 4, Keywords: []string{"flying"},
		Owner: me.ID, Controller: me.ID,
	})

	spoil := castCatalogSpell(t, g, "Sudden Spoiling", "Instant", suddenSpoilingOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	requireSplitSecondOn(t, g, spoil)
	passPriorityAroundTable(t, g)

	if p, tough := effectivePower(t, g, flyer), effectiveToughness(t, g, flyer); p != 1 || tough != 3 {
		t.Errorf("the spoiled Angel is %d/%d, want 1/3 (base 0/2 plus the anthem)", p, tough)
	}
	if abilities := effectiveAbilities(t, g, flyer); len(abilities) != 0 {
		t.Errorf("the spoiled Angel kept %v", abilities)
	}
	if !slices.Contains(effectiveAbilities(t, g, mine), "flying") || effectivePower(t, g, mine) != 4 {
		t.Error("Sudden Spoiling touched the caster's creature")
	}
	late := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Late Angel", TypeLine: "Creature — Angel",
		Power: 4, Toughness: 4, Keywords: []string{"flying"},
		Owner: opp.ID, Controller: opp.ID,
	})
	if effectivePower(t, g, late) != 5 || !slices.Contains(effectiveAbilities(t, g, late), "flying") {
		t.Error("a creature that arrived after Sudden Spoiling resolved was spoiled too (CR 611.2c)")
	}
}
