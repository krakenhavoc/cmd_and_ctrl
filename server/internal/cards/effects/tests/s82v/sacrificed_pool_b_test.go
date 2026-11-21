package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func pbPool(p *game.Player, color string) (n int, restricted int) {
	for _, tok := range p.ManaPool.Units() {
		if tok.Color == color {
			n++
			if len(tok.Restrictions) > 0 {
				restricted++
			}
		}
	}
	return n, restricted
}

// Foundry Helix: 4 to any target, and 4 life only when the sacrificed
// permanent was an artifact. A land pays too.
func TestFoundryHelixGainsOnlyForAnArtifact(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		gain     int
	}{
		{"Artifact", 4},
		{"Land", 0},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine, theirs := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
		effectstest.PaCast(t, g, "Foundry Helix", "Instant", "{1}{R}{W}", effectstest.FoundryHelixOracle, game.CastSpellParams{
			Targets: effectstest.SoPlayer(opp.ID), SacrificeIDs: []uuid.UUID{effectstest.PaPermanent(g, me.ID, "Thing", tc.typeLine, "", 0, 0)},
		})
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.LifeOf(g, opp.ID); got != theirs-4 {
			t.Errorf("%s: opponent's life %d → %d, want 4 damage", tc.typeLine, theirs, got)
		}
		if got := effectstest.LifeOf(g, me.ID) - mine; got != tc.gain {
			t.Errorf("%s: gained %d, want %d", tc.typeLine, got, tc.gain)
		}
	}
}

// Hellish Sideswipe: destroy the target, and draw only for a Vehicle.
func TestHellishSideswipeDrawsForAVehicle(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		draw     int
	}{
		{"Artifact — Vehicle", 1},
		{"Creature — Goblin", 0},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := effectstest.SoCreature(g, opp.ID, "Victim", 2, 2)
		effectstest.PaCast(t, g, "Hellish Sideswipe", "Sorcery", "{B}", effectstest.HellishSideswipeOracle, game.CastSpellParams{
			Targets: effectstest.VsTarget(victim), SacrificeIDs: []uuid.UUID{effectstest.PaPermanent(g, me.ID, "Fodder", tc.typeLine, "{2}", 3, 3)},
		})
		hand := me.Hand.Size()
		effectstest.PassPriorityAroundTable(t, g)
		if effects.OnBattlefield(g, victim) {
			t.Errorf("%s: the target survived", tc.typeLine)
		}
		if got := me.Hand.Size() - hand; got != tc.draw {
			t.Errorf("%s: drew %d, want %d", tc.typeLine, got, tc.draw)
		}
	}
}

// Splitting the Powerstone: two tapped Powerstones, and a card only for
// a legendary artifact.
func TestSplittingThePowerstoneDrawsForALegend(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		draw     int
	}{
		{"Legendary Artifact", 1},
		{"Artifact", 0},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PaCast(t, g, "Splitting the Powerstone", "Sorcery", "{2}{U}", effectstest.SplittingThePowerstoneOracle, game.CastSpellParams{
			SacrificeIDs: []uuid.UUID{effectstest.PaPermanent(g, me.ID, "Relic", tc.typeLine, "{3}", 0, 0)},
		})
		hand := me.Hand.Size()
		effectstest.PassPriorityAroundTable(t, g)
		if n, tapped := effectstest.CountTokensNamed(g, me.ID, "Powerstone"); n != 2 || tapped != 2 {
			t.Errorf("%s: %d Powerstones (%d tapped), want 2 tapped", tc.typeLine, n, tapped)
		}
		if got := me.Hand.Size() - hand; got != tc.draw {
			t.Errorf("%s: drew %d, want %d", tc.typeLine, got, tc.draw)
		}
	}
}

// Fatal Grudge: an opponent may sacrifice only what shares a card type
// with the sacrificed permanent; then the caster draws.
func TestFatalGrudgeEdictsBySharedCardType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	enchantment := effectstest.PaPermanent(g, opp.ID, "Aura of Nothing", "Enchantment", "{1}", 0, 0)
	relic := effectstest.PaPermanent(g, opp.ID, "Their Relic", "Artifact", "{1}", 0, 0)
	effectstest.PaCast(t, g, "Fatal Grudge", "Sorcery", "{B}{R}", effectstest.FatalGrudgeOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{effectstest.PaPermanent(g, me.ID, "My Relic", "Artifact", "{2}", 0, 0)},
	})
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SacrificeChoiceFor(g, opp.ID); c != nil {
		if err := g.ResolveSacrificeChoice(c.ID, opp.ID, enchantment); err == nil {
			t.Fatal("an enchantment shares no card type with an artifact, but it was accepted")
		}
		effectstest.AnswerSacrifice(t, g, opp.ID, relic)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, relic) {
		t.Error("the opponent's artifact survived")
	}
	if !effects.OnBattlefield(g, enchantment) {
		t.Error("the opponent's enchantment was sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// Eldritch Evolution: a creature card of mana value up to 2 plus the
// sacrificed creature's, onto the battlefield; the spell is exiled.
func TestEldritchEvolutionFindsUpToTwoMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	big := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Too Big", TypeLine: "Creature — Giant", ManaCost: "{5}", Power: 5, Toughness: 5})
	fits := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Just Right", TypeLine: "Creature — Beast", ManaCost: "{4}", Power: 4, Toughness: 4})
	id := effectstest.PaCast(t, g, "Eldritch Evolution", "Sorcery", "{1}{G}{G}", effectstest.EldritchEvolutionOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{effectstest.PaPermanent(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{big}); err == nil {
			t.Fatal("a mana value 5 creature was found for X = 4")
		}
		effectstest.AnswerSearchByID(t, g, me.ID, fits)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, big) {
		t.Error("a mana value 5 creature was found for X = 4")
	}
	if !effects.OnBattlefield(g, fits) {
		t.Error("the creature card did not enter")
	}
	if !g.Exile.Contains(id) {
		t.Error("Eldritch Evolution was not exiled")
	}
}

// Anchor to Reality: an Equipment onto the battlefield, and scry 2 when
// it costs less than the sacrificed permanent.
func TestAnchorToRealityScriesForACheaperFind(t *testing.T) {
	for _, tc := range []struct {
		fodderCost string
		scry       bool
	}{
		{"{5}", true},
		{"{1}", false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		for i := 0; i < 3; i++ {
			effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
		}
		sword := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Sword", TypeLine: "Artifact — Equipment", ManaCost: "{3}"})
		effectstest.PaCast(t, g, "Anchor to Reality", "Sorcery", "{2}{U}{U}", effectstest.AnchorToRealityOracle, game.CastSpellParams{
			SacrificeIDs: []uuid.UUID{effectstest.PaPermanent(g, me.ID, "Relic", "Artifact", tc.fodderCost, 0, 0)},
		})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.SearchChoiceFor(g, me.ID) != nil {
			effectstest.AnswerSearchByID(t, g, me.ID, sword)
		}
		if !effects.OnBattlefield(g, sword) {
			t.Errorf("%s: the Equipment did not enter", tc.fodderCost)
		}
		scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry) != nil
		if scry != tc.scry {
			t.Errorf("sacrificed %s: scry prompt open = %v, want %v", tc.fodderCost, scry, tc.scry)
		}
	}
}

// Burnt Offering: X picks of {B} or {R}, X the sacrificed creature's
// mana value.
func TestBurntOfferingAddsTheManaValueInBlackOrRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PaCast(t, g, "Burnt Offering", "Instant", "{B}", effectstest.BurntOfferingOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{effectstest.PaPermanent(g, me.ID, "Ogre", "Creature — Ogre", "{2}{R}", 3, 3)},
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, want := range []string{"R", "B", "R"} {
		c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceMana)
		if c == nil {
			t.Fatal("fewer than three mana picks")
		}
		if len(c.ColorOptions) != 2 {
			t.Fatalf("pick offers %v, want {B} and {R} only", c.ColorOptions)
		}
		if err := g.ResolveManaChoice(c.ID, me.ID, want); err != nil {
			t.Fatalf("ResolveManaChoice(%s): %v", want, err)
		}
	}
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceMana) != nil {
		t.Error("more than three mana picks")
	}
	r, _ := pbPool(me, "R")
	b, _ := pbPool(me, "B")
	if r != 2 || b != 1 {
		t.Errorf("pool has %d R and %d B, want 2 and 1", r, b)
	}
}

// Sacrifice: {B} equal to the mana value.
func TestSacrificeAddsBlackEqualToTheManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PaCast(t, g, "Sacrifice", "Instant", "{B}", effectstest.SacrificeSpellOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{effectstest.PaPermanent(g, me.ID, "Ogre", "Creature — Ogre", "{3}{R}", 4, 4)},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if b, _ := pbPool(me, "B"); b != 4 {
		t.Errorf("pool has %d B, want 4", b)
	}
}

// Metamorphosis: one colour, 1 plus the mana value of it, every one
// restricted to creature spells.
func TestMetamorphosisAddsRestrictedManaOfOneColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PaCast(t, g, "Metamorphosis", "Sorcery", "{G}", effectstest.MetamorphosisOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{effectstest.PaPermanent(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)},
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "U")
	n, restricted := pbPool(me, "U")
	if n != 3 || restricted != 3 {
		t.Errorf("pool has %d U (%d restricted), want 3, all restricted", n, restricted)
	}
}

// Mind Extraction: every card sharing a colour with the sacrificed
// creature is discarded, and only those.
func TestMindExtractionDiscardsTheSacrificedColors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Hand.Cards = nil
	red := uuid.New()
	gold := uuid.New()
	blue := uuid.New()
	for _, c := range []game.Card{
		{InstanceID: red, Name: "Red", TypeLine: "Instant", ManaCost: "{R}"},
		{InstanceID: gold, Name: "Gold", TypeLine: "Instant", ManaCost: "{U}{G}"},
		{InstanceID: blue, Name: "Blue", TypeLine: "Instant", ManaCost: "{U}"},
	} {
		c.Owner, c.Controller = opp.ID, opp.ID
		opp.Hand.PushTop(c)
	}
	effectstest.PaCast(t, g, "Mind Extraction", "Sorcery", "{2}{B}", effectstest.MindExtractionOracle, game.CastSpellParams{
		Targets: effectstest.SoPlayer(opp.ID), SacrificeIDs: []uuid.UUID{effectstest.PaPermanent(g, me.ID, "Druid", "Creature — Elf", "{R}{G}", 1, 1)},
	})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, opp.ID)
	if c == nil || c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Fatalf("discard prompt = %+v, want exactly two", c)
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{red, blue}); err == nil {
		t.Fatal("a blue card was accepted for a red and green creature")
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{red, gold}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !opp.Hand.Contains(blue) || opp.Hand.Contains(red) || opp.Hand.Contains(gold) {
		t.Error("the hand is not left with only the blue card")
	}
}

// Endemic Plague: every creature sharing a creature type with the
// sacrificed one is destroyed, a changeling included; others survive.
func TestEndemicPlagueDestroysTheSharedType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goblin := effectstest.PaPermanent(g, opp.ID, "Goblin", "Creature — Goblin Warrior", "{R}", 1, 1)
	elf := effectstest.PaPermanent(g, opp.ID, "Elf", "Creature — Elf", "{G}", 1, 1)
	shifter := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shifter", TypeLine: "Creature — Shapeshifter",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"changeling"},
	})
	effectstest.PaCast(t, g, "Endemic Plague", "Sorcery", "{3}{B}", effectstest.EndemicPlagueOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{effectstest.PaPermanent(g, me.ID, "My Goblin", "Creature — Goblin", "{R}", 1, 1)},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, goblin) || effects.OnBattlefield(g, shifter) {
		t.Error("a creature sharing the Goblin type survived")
	}
	if !effects.OnBattlefield(g, elf) {
		t.Error("the Elf was destroyed")
	}
}

// Desperate Plea: both modes; the return happens only at or below the
// sacrificed creature's power.
func TestDesperatePleaReturnsOnlyAtOrBelowThePower(t *testing.T) {
	for _, tc := range []struct {
		power   int
		returns bool
	}{
		{3, true},
		{4, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dead := effectstest.PushCatalogGraveyardCard(me, "Dead", "Creature — Zombie", "", tc.power, tc.power)
		victim := effectstest.SoCreature(g, opp.ID, "Victim", 2, 2)
		effectstest.PaCast(t, g, "Desperate Plea", "Sorcery", "{1}{B}", effectstest.DesperatePleaOracle, game.CastSpellParams{
			Modes:        []int{0, 1},
			Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: dead}, {Kind: game.TargetCard, ID: victim}},
			SacrificeIDs: []uuid.UUID{effectstest.SoCreature(g, me.ID, "Beast", 3, 3)},
		})
		effectstest.PassPriorityAroundTable(t, g)
		if got := effects.OnBattlefield(g, dead); got != tc.returns {
			t.Errorf("power %d: returned = %v, want %v", tc.power, got, tc.returns)
		}
		if effects.OnBattlefield(g, victim) {
			t.Errorf("power %d: the destroy mode missed", tc.power)
		}
	}
}
