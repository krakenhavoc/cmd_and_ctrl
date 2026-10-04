package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrificed_pool_b_test.go — the second pool of spells that read the
// permanent their additional cost sacrificed (ADR 0113 §1, owner
// decision 2, #2072): the type, colour and mana-value readers, the
// searches and the mana. One test per card.

const (
	foundryHelixOracle           = "f4f66558-3c99-4488-ad2d-90626a922042"
	hellishSideswipeOracle       = "d919d8e9-d1ba-42de-9884-12e38dca78ac"
	splittingThePowerstoneOracle = "178828a5-c202-4503-9755-fc6bb2390209"
	fatalGrudgeOracle            = "37114872-9804-4ee8-a443-21ea6eaddb0c"
	eldritchEvolutionOracle      = "0f77c0c9-4dc4-489a-b547-e93287c4d1a5"
	anchorToRealityOracle        = "6783f559-33d7-4b13-9a91-02b821e97163"
	burntOfferingOracle          = "86eb30a0-0beb-42db-9ddf-cf8be6c99dd3"
	sacrificeSpellOracle         = "068b3692-411b-44d4-a7e9-005262760cfc"
	metamorphosisOracle          = "7140d726-0136-43af-84b5-85005a66a186"
	mindExtractionOracle         = "0077740a-528b-4ee3-b331-fac321b95302"
	endemicPlagueOracle          = "db982577-1c75-4bc9-ab15-1888ea0be16d"
	desperatePleaOracle          = "bc1ea0ba-46bf-49b6-af95-51eaf1ab915e"
)

func pbPool(p *game.Player, color string) (n int, restricted int) {
	for _, tok := range p.ManaPool {
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
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine, theirs := lifeOf(g, me.ID), lifeOf(g, opp.ID)
		paCast(t, g, "Foundry Helix", "Instant", "{1}{R}{W}", foundryHelixOracle, game.CastSpellParams{
			Targets: soPlayer(opp.ID), SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "Thing", tc.typeLine, "", 0, 0)},
		})
		passPriorityAroundTable(t, g)
		if got := lifeOf(g, opp.ID); got != theirs-4 {
			t.Errorf("%s: opponent's life %d → %d, want 4 damage", tc.typeLine, theirs, got)
		}
		if got := lifeOf(g, me.ID) - mine; got != tc.gain {
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
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := soCreature(g, opp.ID, "Victim", 2, 2)
		paCast(t, g, "Hellish Sideswipe", "Sorcery", "{B}", hellishSideswipeOracle, game.CastSpellParams{
			Targets: vsTarget(victim), SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "Fodder", tc.typeLine, "{2}", 3, 3)},
		})
		hand := me.Hand.Size()
		passPriorityAroundTable(t, g)
		if onBattlefield(g, victim) {
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
		g := newCatalogGame(t)
		me := g.Seats[0]
		paCast(t, g, "Splitting the Powerstone", "Sorcery", "{2}{U}", splittingThePowerstoneOracle, game.CastSpellParams{
			SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "Relic", tc.typeLine, "{3}", 0, 0)},
		})
		hand := me.Hand.Size()
		passPriorityAroundTable(t, g)
		if n, tapped := countTokensNamed(g, me.ID, "Powerstone"); n != 2 || tapped != 2 {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	enchantment := paPermanent(g, opp.ID, "Aura of Nothing", "Enchantment", "{1}", 0, 0)
	relic := paPermanent(g, opp.ID, "Their Relic", "Artifact", "{1}", 0, 0)
	paCast(t, g, "Fatal Grudge", "Sorcery", "{B}{R}", fatalGrudgeOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "My Relic", "Artifact", "{2}", 0, 0)},
	})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if c := sacrificeChoiceFor(g, opp.ID); c != nil {
		if err := g.ResolveSacrificeChoice(c.ID, opp.ID, enchantment); err == nil {
			t.Fatal("an enchantment shares no card type with an artifact, but it was accepted")
		}
		answerSacrifice(t, g, opp.ID, relic)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, relic) {
		t.Error("the opponent's artifact survived")
	}
	if !onBattlefield(g, enchantment) {
		t.Error("the opponent's enchantment was sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// Eldritch Evolution: a creature card of mana value up to 2 plus the
// sacrificed creature's, onto the battlefield; the spell is exiled.
func TestEldritchEvolutionFindsUpToTwoMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	big := pushLibraryCardForTest(me, game.Card{Name: "Too Big", TypeLine: "Creature — Giant", ManaCost: "{5}", Power: 5, Toughness: 5})
	fits := pushLibraryCardForTest(me, game.Card{Name: "Just Right", TypeLine: "Creature — Beast", ManaCost: "{4}", Power: 4, Toughness: 4})
	id := paCast(t, g, "Eldritch Evolution", "Sorcery", "{1}{G}{G}", eldritchEvolutionOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)},
	})
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{big}); err == nil {
			t.Fatal("a mana value 5 creature was found for X = 4")
		}
		answerSearchByID(t, g, me.ID, fits)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, big) {
		t.Error("a mana value 5 creature was found for X = 4")
	}
	if !onBattlefield(g, fits) {
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
		g := newCatalogGame(t)
		me := g.Seats[0]
		for i := 0; i < 3; i++ {
			pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
		}
		sword := pushLibraryCardForTest(me, game.Card{Name: "Sword", TypeLine: "Artifact — Equipment", ManaCost: "{3}"})
		paCast(t, g, "Anchor to Reality", "Sorcery", "{2}{U}{U}", anchorToRealityOracle, game.CastSpellParams{
			SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "Relic", "Artifact", tc.fodderCost, 0, 0)},
		})
		passPriorityAroundTable(t, g)
		if searchChoiceFor(g, me.ID) != nil {
			answerSearchByID(t, g, me.ID, sword)
		}
		if !onBattlefield(g, sword) {
			t.Errorf("%s: the Equipment did not enter", tc.fodderCost)
		}
		scry := latestChoiceOfKind(g, game.PendingChoiceScry) != nil
		if scry != tc.scry {
			t.Errorf("sacrificed %s: scry prompt open = %v, want %v", tc.fodderCost, scry, tc.scry)
		}
	}
}

// Burnt Offering: X picks of {B} or {R}, X the sacrificed creature's
// mana value.
func TestBurntOfferingAddsTheManaValueInBlackOrRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	paCast(t, g, "Burnt Offering", "Instant", "{B}", burntOfferingOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "Ogre", "Creature — Ogre", "{2}{R}", 3, 3)},
	})
	passPriorityAroundTable(t, g)
	for _, want := range []string{"R", "B", "R"} {
		c := latestChoiceOfKind(g, game.PendingChoiceMana)
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
	if latestChoiceOfKind(g, game.PendingChoiceMana) != nil {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	paCast(t, g, "Sacrifice", "Instant", "{B}", sacrificeSpellOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "Ogre", "Creature — Ogre", "{3}{R}", 4, 4)},
	})
	passPriorityAroundTable(t, g)
	if b, _ := pbPool(me, "B"); b != 4 {
		t.Errorf("pool has %d B, want 4", b)
	}
}

// Metamorphosis: one colour, 1 plus the mana value of it, every one
// restricted to creature spells.
func TestMetamorphosisAddsRestrictedManaOfOneColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	paCast(t, g, "Metamorphosis", "Sorcery", "{G}", metamorphosisOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)},
	})
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "U")
	n, restricted := pbPool(me, "U")
	if n != 3 || restricted != 3 {
		t.Errorf("pool has %d U (%d restricted), want 3, all restricted", n, restricted)
	}
}

// Mind Extraction: every card sharing a colour with the sacrificed
// creature is discarded, and only those.
func TestMindExtractionDiscardsTheSacrificedColors(t *testing.T) {
	g := newCatalogGame(t)
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
	paCast(t, g, "Mind Extraction", "Sorcery", "{2}{B}", mindExtractionOracle, game.CastSpellParams{
		Targets: soPlayer(opp.ID), SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "Druid", "Creature — Elf", "{R}{G}", 1, 1)},
	})
	passPriorityAroundTable(t, g)
	c := discardChoiceFor(g, opp.ID)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goblin := paPermanent(g, opp.ID, "Goblin", "Creature — Goblin Warrior", "{R}", 1, 1)
	elf := paPermanent(g, opp.ID, "Elf", "Creature — Elf", "{G}", 1, 1)
	shifter := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shifter", TypeLine: "Creature — Shapeshifter",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"changeling"},
	})
	paCast(t, g, "Endemic Plague", "Sorcery", "{3}{B}", endemicPlagueOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "My Goblin", "Creature — Goblin", "{R}", 1, 1)},
	})
	passPriorityAroundTable(t, g)
	if onBattlefield(g, goblin) || onBattlefield(g, shifter) {
		t.Error("a creature sharing the Goblin type survived")
	}
	if !onBattlefield(g, elf) {
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
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dead := pushCatalogGraveyardCard(me, "Dead", "Creature — Zombie", "", tc.power, tc.power)
		victim := soCreature(g, opp.ID, "Victim", 2, 2)
		paCast(t, g, "Desperate Plea", "Sorcery", "{1}{B}", desperatePleaOracle, game.CastSpellParams{
			Modes:        []int{0, 1},
			Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: dead}, {Kind: game.TargetCard, ID: victim}},
			SacrificeIDs: []uuid.UUID{soCreature(g, me.ID, "Beast", 3, 3)},
		})
		passPriorityAroundTable(t, g)
		if got := onBattlefield(g, dead); got != tc.returns {
			t.Errorf("power %d: returned = %v, want %v", tc.power, got, tc.returns)
		}
		if onBattlefield(g, victim) {
			t.Errorf("power %d: the destroy mode missed", tc.power)
		}
	}
}
