package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_mp_cards_test.go — Foundations, M–P (#4465): the cards from that
// stretch of the alphabet that needed no new seam. Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	fdnMPMaalfeldTwinsOracle    = "dda4b515-49c0-43fe-9f6a-36defa326bb1"
	fdnMPMagnigothSentryOracle  = "b52824f9-7897-46b7-a4e5-a0261861cb5b"
	fdnMPMischievousPupOracle   = "e75b187d-0e3b-4d94-a940-8b7e4e7ed1ca"
	fdnMPMockingSpriteOracle    = "013ca531-1ff7-4efc-a280-4bc3ab574072"
	fdnMPMomentOfCravingOracle  = "1489551f-5714-471b-a0e6-f40bc3085afc"
	fdnMPMomentOfTriumphOracle  = "57579799-2582-417f-9864-4560cac39ac3"
	fdnMPMysticArchaeologistOrc = "b77dfe2a-ebc9-46b0-9134-2ecb2abdd8be"
	fdnMPOfferImmortalityOracle = "be4499e6-616a-4c11-b37e-9201d2492682"
	fdnMPPelakkaWurmOracle      = "d36075c2-de66-4202-9217-b1102a2bc14b"
	fdnMPPrepostProportionsOrc  = "1e03aad2-6426-4d18-a0e5-bc3d8e6b7b11"
	fdnMPPridefulParentOracle   = "a90d1d35-6d67-43b3-8256-b0318b3a4b07"
	fdnMPPulseTrackerOracle     = "f916f689-a314-4952-8107-41c4c2f2e811"
)

func TestFdnMPEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnMPMaalfeldTwinsOracle:    "Maalfeld Twins",
		fdnMPMagnigothSentryOracle:  "Magnigoth Sentry",
		fdnMPMischievousPupOracle:   "Mischievous Pup",
		fdnMPMockingSpriteOracle:    "Mocking Sprite",
		fdnMPMomentOfCravingOracle:  "Moment of Craving",
		fdnMPMomentOfTriumphOracle:  "Moment of Triumph",
		fdnMPMysticArchaeologistOrc: "Mystic Archaeologist",
		fdnMPOfferImmortalityOracle: "Offer Immortality",
		fdnMPPelakkaWurmOracle:      "Pelakka Wurm",
		fdnMPPrepostProportionsOrc:  "Preposterous Proportions",
		fdnMPPridefulParentOracle:   "Prideful Parent",
		fdnMPPulseTrackerOracle:     "Pulse Tracker",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// fdnMPHandSize is the size of p's hand.
func fdnMPHandSize(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = p.Hand.Size() })
	return n
}

// --- Maalfeld Twins -----------------------------------------------------------

func TestFdnMPMaalfeldTwinsMakesTwoZombiesWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	twins := b27Push(g, me.ID, "Maalfeld Twins", "Creature — Zombie", fdnMPMaalfeldTwinsOracle, "{5}{B}", 4, 4, "B")
	b27Kill(g, twins)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Zombie"); got != 2 {
		t.Errorf("the Twins' controller has %d Zombie tokens, want 2", got)
	}
	if got := tmtB1Count(g, opp.ID, "Zombie"); got != 0 {
		t.Errorf("an opponent has %d Zombie tokens, want 0", got)
	}
	var sizes [][2]int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Zombie" {
				eff := c.Effective()
				sizes = append(sizes, [2]int{eff.Power, eff.Toughness})
			}
		}
	})
	for _, s := range sizes {
		if s != [2]int{2, 2} {
			t.Errorf("a Zombie token is %d/%d, want 2/2", s[0], s[1])
		}
	}
}

// --- Magnigoth Sentry ---------------------------------------------------------

func TestFdnMPMagnigothSentryHasReach(t *testing.T) {
	g := newCatalogGame(t)
	sentry := b12Push(g, g.Seats[0].ID, "Magnigoth Sentry", "Creature — Treefolk", fdnMPMagnigothSentryOracle, 4, 4)
	assertKeywords(t, g, sentry, "reach")
}

// --- Mischievous Pup ----------------------------------------------------------

func TestFdnMPMischievousPupReturnsAnotherPermanentYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	pup := castHoldCreature(t, g, "Mischievous Pup", "Creature — Dog", fdnMPMischievousPupOracle, 3, 1)
	assertKeywords(t, g, pup, "flash")
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{theirs, pup} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as a target", bad)
		}
	}
	answerPickTarget(t, g, mine)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("your other permanent was not returned to your hand")
	}
	if !g.Battlefield.Contains(pup) || !g.Battlefield.Contains(theirs) {
		t.Error("the Pup or the opponent's creature left the battlefield")
	}
}

func TestFdnMPMischievousPupMayReturnNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	castHoldCreature(t, g, "Mischievous Pup", "Creature — Dog", fdnMPMischievousPupOracle, 3, 1)
	b04WaitForPick(t, g, me.ID)
	b26DeclinePickTarget(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("choosing no target still returned a permanent")
	}
}

// --- Mocking Sprite -----------------------------------------------------------

func TestFdnMPMockingSpriteFliesAndDiscountsYourInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	sprite := b12Push(g, me.ID, "Mocking Sprite", "Creature — Faerie Rogue", fdnMPMockingSpriteOracle, 2, 1)
	assertKeywords(t, g, sprite, "flying")
	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{2}{U}"); got != 2 {
		t.Errorf("your sorcery costs %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Shock", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("your instant costs %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("your creature costs %d, want 2 (undiscounted)", got)
	}
	if got := priceInHand(t, g, them, "Their Divination", "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("an opponent's sorcery costs %d, want 3 (undiscounted)", got)
	}
	if got := priceInHand(t, g, me, "Counterspell", "Instant", "{U}{U}"); got != 2 {
		t.Errorf("{U}{U} under the reduction costs %d, want 2 (colours untouched)", got)
	}
}

// --- Moment of Craving --------------------------------------------------------

func TestFdnMPMomentOfCravingShrinksAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	before := lifeOf(g, me.ID)
	castCatalogSpell(t, g, "Moment of Craving", "Instant", fdnMPMomentOfCravingOracle, cardRefs(ogre))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, ogre), effectiveToughness(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
	if got := lifeOf(g, me.ID); got != before+2 {
		t.Errorf("life %d after the spell, want %d", got, before+2)
	}
}

func TestFdnMPMomentOfCravingKillsATwoToughnessCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Moment of Craving", "Instant", fdnMPMomentOfCravingOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Error("a 2/2 survived -2/-2")
	}
}

// --- Moment of Triumph --------------------------------------------------------

func TestFdnMPMomentOfTriumphPumpsAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	before := lifeOf(g, me.ID)
	castCatalogSpell(t, g, "Moment of Triumph", "Instant", fdnMPMomentOfTriumphOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
	if got := lifeOf(g, me.ID); got != before+2 {
		t.Errorf("life %d after the spell, want %d", got, before+2)
	}
}

// --- Mystic Archaeologist -----------------------------------------------------

func TestFdnMPMysticArchaeologistDrawsTwoForFiveMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arch := b12Push(g, me.ID, "Mystic Archaeologist", "Creature — Human Wizard", fdnMPMysticArchaeologistOrc, 2, 1)
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	before := fdnMPHandSize(g, me)

	b06AddMana(me, "U", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, arch, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only one blue mana")
	}
	b06AddMana(me, "U")
	b16Activate(t, g, me.ID, arch, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := fdnMPHandSize(g, me); got != before+2 {
		t.Errorf("hand %d -> %d, want +2", before, got)
	}
}

// --- Offer Immortality --------------------------------------------------------

func TestFdnMPOfferImmortalityGrantsDeathtouchAndIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Offer Immortality", "Instant", fdnMPOfferImmortalityOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, bear, "deathtouch", "indestructible")
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the target is %d/%d, want an unchanged 2/2", p, tough)
	}
	b27Kill(g, bear)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(bear) {
		t.Error("the indestructible creature was destroyed")
	}
}

// --- Pelakka Wurm -------------------------------------------------------------

func TestFdnMPPelakkaWurmGainsSevenOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := lifeOf(g, me.ID)
	wurm := castHoldCreature(t, g, "Pelakka Wurm", "Creature — Wurm", fdnMPPelakkaWurmOracle, 7, 7)
	assertKeywords(t, g, wurm, "trample")
	if got := lifeOf(g, me.ID); got != before+7 {
		t.Errorf("life %d after the Wurm entered, want %d", got, before+7)
	}
}

func TestFdnMPPelakkaWurmDrawsOnDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B")
	wurm := b27Push(g, me.ID, "Pelakka Wurm", "Creature — Wurm", fdnMPPelakkaWurmOracle, "{4}{G}{G}{G}", 7, 7, "G")
	before := fdnMPHandSize(g, me)
	b27Kill(g, wurm)
	passPriorityAroundTable(t, g)
	if got := fdnMPHandSize(g, me); got != before+1 {
		t.Errorf("hand %d after the Wurm died, want %d", got, before+1)
	}
}

// --- Preposterous Proportions -------------------------------------------------

func TestFdnMPPreposterousProportionsPumpsYourTeam(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Preposterous Proportions", "Sorcery", fdnMPPrepostProportionsOrc, nil)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 12 || tough != 12 {
		t.Errorf("your creature is %d/%d, want 12/12", p, tough)
	}
	if !effectiveAbilitiesContain(t, g, mine, "vigilance") {
		t.Error("your creature did not gain vigilance")
	}
	if p := effectivePower(t, g, theirs); p != 2 || effectiveAbilitiesContain(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature was affected")
	}
	later := b12Creature(g, me.ID, "Later Bear", "Creature — Bear", 2, 2)
	if p := effectivePower(t, g, later); p != 2 {
		t.Errorf("a creature that entered after the spell has power %d, want 2", p)
	}
}

// --- Prideful Parent ----------------------------------------------------------

func TestFdnMPPridefulParentMakesACat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := tmtB1Count(g, me.ID, "Cat")
	parent := castHoldCreature(t, g, "Prideful Parent", "Creature — Cat", fdnMPPridefulParentOracle, 2, 2)
	assertKeywords(t, g, parent, "vigilance")
	if got := tmtB1Count(g, me.ID, "Cat"); got != before+1 {
		t.Errorf("%d Cat tokens after the Parent entered, want %d", got, before+1)
	}
}

// --- Pulse Tracker ------------------------------------------------------------

func TestFdnMPPulseTrackerDrainsEachOpponentWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tracker := b12Push(g, me.ID, "Pulse Tracker", "Creature — Vampire Rogue", fdnMPPulseTrackerOracle, 1, 1)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(tracker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		switch s.ID {
		case me.ID:
			if s.Life != lives[s.ID] {
				t.Errorf("my life %d, want an unchanged %d", s.Life, lives[s.ID])
			}
		case opp.ID:
			// The attacked player may also have taken the combat damage by
			// now; never less than the drain.
			if s.Life > lives[s.ID]-1 {
				t.Errorf("the attacked opponent is at %d, want at most %d", s.Life, lives[s.ID]-1)
			}
		default:
			if s.Life != lives[s.ID]-1 {
				t.Errorf("seat %s life %d, want %d", s.ID, s.Life, lives[s.ID]-1)
			}
		}
	}
}
