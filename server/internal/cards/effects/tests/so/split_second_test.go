package testsso

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: effectstest.SolRingOracle, Owner: opp.ID, Controller: opp.ID,
	})
	bombardment := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Bombardment", TypeLine: "Enchantment",
		OracleID: effectstest.GoblinBombardmentOracle, Owner: opp.ID, Controller: opp.ID,
	})
	fodder := effectstest.SeedPermanentFor(g, opp.ID, "Goblin", "Creature — Goblin")
	monk := effectstest.PushProwessCreature(g, me.ID, "Swiftspear", 1, 2)

	grip := effectstest.CastCatalogSpell(t, g, "Krosan Grip", "Instant", effectstest.KrosanGripOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ring}})
	effectstest.RequireSplitSecondOn(t, g, grip)
	if n := effectstest.ProwessItemsFrom(g, monk); n != 1 {
		t.Fatalf("CR 702.61b: the prowess trigger did not fire under split second (%d)", n)
	}

	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: effectstest.LightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
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

	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ring) {
		t.Error("Krosan Grip did not destroy the Sol Ring")
	}
	if g.SplitSecondActive {
		t.Error("split second outlived Krosan Grip")
	}
	if got := effectstest.EffectivePower(t, g, monk); got != 2 {
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	shock := effectstest.CastCatalogSpell(t, g, "Sudden Shock", "Instant", effectstest.SuddenShockOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.RequireSplitSecondOn(t, g, shock)
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("Sudden Shock dealt %d, want 2", got)
	}
	if g.SplitSecondActive {
		t.Error("split second outlived Sudden Shock")
	}
}

func TestSuddenDeathGivesMinusFourUnderSplitSecond(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	giant := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Craw Wurm", TypeLine: "Creature — Wurm",
		Power: 6, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	death := effectstest.CastCatalogSpell(t, g, "Sudden Death", "Instant", effectstest.SuddenDeathOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: giant}})
	effectstest.RequireSplitSecondOn(t, g, death)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, giant), effectstest.EffectiveToughness(t, g, giant); p != 2 || tough != 1 {
		t.Errorf("Sudden Death left a %d/%d, want 2/1", p, tough)
	}
}

func TestSuddenEdictMakesTheTargetPlayerSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keep := effectstest.SeedPermanentFor(g, opp.ID, "Their Keeper", "Creature — Bear")
	lose := effectstest.SeedPermanentFor(g, opp.ID, "Their Fodder", "Creature — Bear")
	mine := effectstest.SeedPermanentFor(g, me.ID, "My Bear", "Creature — Bear")

	edict := effectstest.CastCatalogSpell(t, g, "Sudden Edict", "Instant", effectstest.SuddenEdictOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.RequireSplitSecondOn(t, g, edict)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the caster was asked to sacrifice; only the target player is")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, lose)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Serra Angel", TypeLine: "Creature — Angel",
		Power: 4, Toughness: 4, Keywords: []string{"flying", "vigilance"},
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: effectstest.GloriousAnthemOracle, Owner: opp.ID, Controller: opp.ID,
	})
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Angel", TypeLine: "Creature — Angel",
		Power: 4, Toughness: 4, Keywords: []string{"flying"},
		Owner: me.ID, Controller: me.ID,
	})

	spoil := effectstest.CastCatalogSpell(t, g, "Sudden Spoiling", "Instant", effectstest.SuddenSpoilingOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.RequireSplitSecondOn(t, g, spoil)
	effectstest.PassPriorityAroundTable(t, g)

	if p, tough := effectstest.EffectivePower(t, g, flyer), effectstest.EffectiveToughness(t, g, flyer); p != 1 || tough != 3 {
		t.Errorf("the spoiled Angel is %d/%d, want 1/3 (base 0/2 plus the anthem)", p, tough)
	}
	if abilities := effectstest.EffectiveAbilities(t, g, flyer); len(abilities) != 0 {
		t.Errorf("the spoiled Angel kept %v", abilities)
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, mine), "flying") || effectstest.EffectivePower(t, g, mine) != 4 {
		t.Error("Sudden Spoiling touched the caster's creature")
	}
	late := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Late Angel", TypeLine: "Creature — Angel",
		Power: 4, Toughness: 4, Keywords: []string{"flying"},
		Owner: opp.ID, Controller: opp.ID,
	})
	if effectstest.EffectivePower(t, g, late) != 5 || !slices.Contains(effectstest.EffectiveAbilities(t, g, late), "flying") {
		t.Error("a creature that arrived after Sudden Spoiling resolved was spoiled too (CR 611.2c)")
	}
}

// Angel's Grace (#749) shipped with "Split second isn't enforced"; the
// declaration clears it. The point of the card is that the lethal spell
// it answers can't be followed by another one.
func TestAngelsGraceHasSplitSecond(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grace := effectstest.CastCatalogSpell(t, g, "Angel's Grace", "Instant", effectstest.AngelsGraceOracle, nil)
	effectstest.RequireSplitSecondOn(t, g, grace)
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: effectstest.LightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
	err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	})
	if !errors.Is(err, game.ErrSplitSecondActive) {
		t.Errorf("Lightning Bolt in response to Angel's Grace: got %v, want ErrSplitSecondActive", err)
	}
}
