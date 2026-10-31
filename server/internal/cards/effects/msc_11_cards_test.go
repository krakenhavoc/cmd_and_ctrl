package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_11_cards_test.go — Marvel Super Heroes Commander, eleventh batch
// (#4455). Each test plays the real catalog entry and asserts what its text
// says.

const (
	msc11GertOracle       = "aaded61e-32c2-4e02-8557-3f8cd927a64e"
	msc11HammerheadOracle = "2ff5a5ed-d6db-4dfa-b1fd-38ad872ad30a"
	msc11BowslingerOracle = "b218f1f8-984c-488e-8222-37f05c674a6b"
	msc11HobgoblinOracle  = "1d453331-6b8e-4d3d-b658-683eab2760da"
	msc11CosmoOracle      = "b2aa052c-08b7-4334-9ea7-e78da57ff9ea"
	msc11GangstersOracle  = "40b43763-3fbf-4ecd-8187-5d2c2c9cda82"
	msc11IronLadOracle    = "fedcc722-ac5f-464a-bac9-c5f1d4de28d0"
	msc11IronManOracle    = "188e1516-5766-406b-aacb-468f2d9a530a"
	msc11KingpinOracle    = "062ffa63-1985-4409-b3ad-d42407ee755a"
	msc11LiesOracle       = "cbbf61b0-5dd2-4643-8857-b7a10ec76ec8"
	msc11LizardsOracle    = "616312a9-83c8-4fb7-87ce-079ccd3231fa"
)

func TestMsc11EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc11GertOracle:       "Gert and Old Lace, Runaways",
		msc11HammerheadOracle: "Hammerhead, Maggia Boss",
		msc11BowslingerOracle: "Hawkeye, Bowslinger",
		msc11HobgoblinOracle:  "Hobgoblin, Mantled Marauder",
		msc11CosmoOracle:      "Imperial Cosmographer",
		msc11GangstersOracle:  "Inner Demons Gangsters",
		msc11IronLadOracle:    "Iron Lad, Young Avenger",
		msc11IronManOracle:    "Iron Man, Tony Stark",
		msc11KingpinOracle:    "Kingpin, Wilson Fisk",
		msc11LiesOracle:       "Living Lies of Loki",
		msc11LizardsOracle:    "Lurking Lizards",
	}
	for oracle, name := range want {
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

// msc11Cast casts a spell with the given cost and colours from the active
// seat's hand, leaving it on the stack.
func msc11Cast(t *testing.T, g *game.Game, name, typeLine, manaCost string, colors []string) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := handCardFull(me, name, typeLine, manaCost, "", colors)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// --- Lurking Lizards ---

func TestMsc11LurkingLizardsGrowsOnlyForSpellsOfManaValueFourOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lizards := b12Push(g, me.ID, "Lurking Lizards", "Creature — Lizard Villain", msc11LizardsOracle, 1, 3)
	if !hasEffectiveKeyword(t, g, lizards, "trample") {
		t.Error("Lurking Lizards has no trample")
	}
	msc11Cast(t, g, "Small Spell", "Sorcery", "{1}{R}{R}", nil)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, lizards); got != 0 {
		t.Fatalf("a mana value 3 spell gave %d counters, want 0", got)
	}
	msc11Cast(t, g, "Big Spell", "Sorcery", "{2}{R}{R}", nil)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, lizards); got != 1 {
		t.Errorf("a mana value 4 spell gave %d counters, want 1", got)
	}
}

func TestMsc11LurkingLizardsIgnoresAnOpponentsBigSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lizards := b12Push(g, me.ID, "Lurking Lizards", "Creature — Lizard Villain", msc11LizardsOracle, 1, 3)
	advanceToMain(t, g)
	id := handCardFull(opp, "Their Big Spell", "Instant", "{3}{R}{R}", "", nil)
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, lizards); got != 0 {
		t.Errorf("an opponent's spell gave %d counters, want 0", got)
	}
}

// --- Hammerhead, Maggia Boss ---

func TestMsc11HammerheadSacrificesACreatureOrArtifactForACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hammer := b12Push(g, me.ID, "Hammerhead, Maggia Boss", "Legendary Creature — Human Rogue Villain", msc11HammerheadOracle, 2, 1)
	bear := b12Creature(g, me.ID, "Fodder Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Fodder Rock", "Artifact")
	land := b12Permanent(g, me.ID, "Fodder Land", "Land")
	advanceToMain(t, g)

	for i, fodder := range []uuid.UUID{bear, rock} {
		if err := g.ActivateCatalogAbility(me.ID, hammer, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(fodder) {
			t.Errorf("fodder %d was not sacrificed", i)
		}
		if got := plusOnes(g, hammer); got != i+1 {
			t.Errorf("after %d sacrifices Hammerhead has %d counters", i+1, got)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, hammer, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Error("a land was accepted as the sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, hammer, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{hammer}}); err == nil {
		t.Error("Hammerhead sacrificed himself for his own ability: it says another")
	}
}

// --- Iron Lad, Young Avenger ---

func TestMsc11IronLadDiscountsYourNoncreatureSpellsByOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lad := b12Push(g, me.ID, "Iron Lad, Young Avenger", "Legendary Artifact Creature — Human Hero", msc11IronLadOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, lad, "flying") {
		t.Error("Iron Lad has no flying")
	}
	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("your instant: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Sol Ring", "Artifact", "{3}"); got != 2 {
		t.Errorf("your artifact: %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("your creature: %d, want 2 (noncreature spells only)", got)
	}
	if got := priceInHand(t, g, opp, "Lightning Bolt", "Instant", "{1}{R}"); got != 2 {
		t.Errorf("an opponent's instant: %d, want 2 (your spells only)", got)
	}
}

// --- Imperial Cosmographer ---

func TestMsc11CosmographerGrowsWhenAnotherCreatureLeavesWithoutDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cosmo := b12Push(g, me.ID, "Imperial Cosmographer", "Creature — Kree Scientist", msc11CosmoOracle, 2, 3)
	one := b12Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2)
	two := b12Creature(g, me.ID, "Bear Two", "Creature — Bear", 2, 2)
	three := b12Creature(g, me.ID, "Bear Three", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, cosmo, "vigilance") {
		t.Error("Imperial Cosmographer has no vigilance")
	}

	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(one) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, cosmo); got != 2 {
		t.Fatalf("after a bounce: %d counters, want 2", got)
	}

	b27Kill(g, two)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, cosmo); got != 2 {
		t.Errorf("a creature that died gave counters: %d, want 2", got)
	}

	g.WithWriteLock(func() { _ = g.ExileCardForEffect(three) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, cosmo); got != 4 {
		t.Errorf("after an exile: %d counters, want 4", got)
	}
}

func TestMsc11CosmographerIgnoresItselfAndOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cosmo := b12Push(g, me.ID, "Imperial Cosmographer", "Creature — Kree Scientist", msc11CosmoOracle, 2, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, cosmo); got != 0 {
		t.Errorf("an opponent's creature left: %d counters, want 0", got)
	}
}

// --- Hobgoblin, Mantled Marauder ---

func TestMsc11HobgoblinGainsTwoPowerPerCardYouDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hob := b12Push(g, me.ID, "Hobgoblin, Mantled Marauder", "Legendary Creature — Goblin Human Villain", msc11HobgoblinOracle, 1, 2)
	for _, kw := range []string{"flying", "haste"} {
		if !hasEffectiveKeyword(t, g, hob, kw) {
			t.Errorf("Hobgoblin has no %s", kw)
		}
	}
	handCard(me, "Fodder One", "Instant")
	handCard(me, "Fodder Two", "Instant")
	handCard(opp, "Their Card", "Instant")

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, hob); got != 1 {
		t.Fatalf("an opponent's discard pumped Hobgoblin to %d", got)
	}
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, hob); got != 3 {
		t.Errorf("after one discard Hobgoblin has power %d, want 3", got)
	}
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, hob); got != 5 {
		t.Errorf("after two discards Hobgoblin has power %d, want 5", got)
	}
}

// --- Kingpin, Wilson Fisk ---

func TestMsc11KingpinMakesTwoTreasuresOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kingpin := b12Push(g, me.ID, "Kingpin, Wilson Fisk", "Legendary Creature — Human Villain", msc11KingpinOracle, 3, 6)
	one := b12Creature(g, me.ID, "Fodder One", "Creature — Bear", 2, 2)
	two := b12Creature(g, me.ID, "Fodder Two", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Fodder Rock", "Artifact")
	if !hasEffectiveKeyword(t, g, kingpin, "menace") {
		t.Error("Kingpin has no menace")
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(rock) })
	passPriorityAroundTable(t, g)
	if got := countTreasures(g, me.ID); got != 0 {
		t.Fatalf("sacrificing an artifact made %d Treasures, want 0 (creatures only)", got)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(one) })
	passPriorityAroundTable(t, g)
	if got := countTreasures(g, me.ID); got != 2 {
		t.Fatalf("sacrificing a creature made %d Treasures, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(two) })
	passPriorityAroundTable(t, g)
	if got := countTreasures(g, me.ID); got != 2 {
		t.Errorf("a second sacrifice the same turn made Treasures: %d, want 2", got)
	}
}

func TestMsc11KingpinTriggersWhenHeIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kingpin := b12Push(g, me.ID, "Kingpin, Wilson Fisk", "Legendary Creature — Human Villain", msc11KingpinOracle, 3, 6)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(kingpin) })
	passPriorityAroundTable(t, g)
	if got := countTreasures(g, me.ID); got != 2 {
		t.Errorf("sacrificing Kingpin made %d Treasures, want 2", got)
	}
}

func TestMsc11KingpinIgnoresAnOpponentsSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Kingpin, Wilson Fisk", "Legendary Creature — Human Villain", msc11KingpinOracle, 3, 6)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if got := countTreasures(g, me.ID); got != 0 {
		t.Errorf("an opponent's sacrifice made %d Treasures for Kingpin's controller, want 0", got)
	}
}

// --- Hawkeye, Bowslinger ---

func TestMsc11BowslingerGrowsAndScriesForASpellTargetingAnyCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hawk := b12Push(g, me.ID, "Hawkeye, Bowslinger", "Legendary Creature — Human Archer Hero", msc11BowslingerOracle, 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "Top", "Second")
	for _, kw := range []string{"reach", "vigilance"} {
		if !hasEffectiveKeyword(t, g, hawk, kw) {
			t.Errorf("Hawkeye has no %s", kw)
		}
	}

	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(theirs)})
	passPriorityAroundTable(t, g)

	if got := plusOnes(g, hawk); got != 1 {
		t.Errorf("Hawkeye has %d counters after a spell targeting an opponent's creature, want 1", got)
	}
	if c := scryChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("no scry 1 prompt: %+v", c)
	}
}

func TestMsc11BowslingerIgnoresASpellWithNoCreatureTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hawk := b12Push(g, me.ID, "Hawkeye, Bowslinger", "Legendary Creature — Human Archer Hero", msc11BowslingerOracle, 3, 3)
	msc11Cast(t, g, "Plain Spell", "Sorcery", "{1}{W}", nil)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, hawk); got != 0 {
		t.Errorf("Hawkeye has %d counters after an untargeted spell, want 0", got)
	}
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("a scry was offered")
	}
}

// --- Inner Demons Gangsters ---

func TestMsc11GangstersDiscardForPowerAndMenaceAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gang := b12Push(g, me.ID, "Inner Demons Gangsters", "Creature — Human Rogue Villain", msc11GangstersOracle, 3, 4)
	card := handCard(me, "Fodder", "Instant")
	advanceToMain(t, g)

	if hasEffectiveKeyword(t, g, gang, "menace") {
		t.Fatal("the Gangsters start with menace")
	}
	if err := g.ActivateCatalogAbility(me.ID, gang, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{card}}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, gang); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	if !hasEffectiveKeyword(t, g, gang, "menace") {
		t.Error("the Gangsters did not gain menace")
	}
}

func TestMsc11GangstersCannotActivateOutsideAMainPhaseWithAnEmptyStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gang := b12Push(g, me.ID, "Inner Demons Gangsters", "Creature — Human Rogue Villain", msc11GangstersOracle, 3, 4)
	card := handCard(me, "Fodder", "Instant")
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, gang, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{card}}); err == nil {
		t.Error("the Gangsters activated in the beginning of combat step")
	}
	if !me.Hand.Contains(card) {
		t.Error("a refused activation discarded the card")
	}
}

// --- Living Lies of Loki ---

func TestMsc11LivingLiesGetsPowerForEachOtherIllusionYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lies := b12Push(g, me.ID, "Living Lies of Loki", "Creature — Illusion Villain", msc11LiesOracle, 1, 3)
	if got := effectivePower(t, g, lies); got != 1 {
		t.Fatalf("alone: power %d, want 1 (it does not count itself)", got)
	}
	b12Creature(g, me.ID, "Another Illusion", "Creature — Illusion", 1, 1)
	b12Creature(g, me.ID, "A Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Illusion", "Creature — Illusion", 1, 1)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if got := effectivePower(t, g, lies); got != 2 {
		t.Errorf("with one other Illusion: power %d, want 2", got)
	}
	if got := effectiveToughness(t, g, lies); got != 3 {
		t.Errorf("toughness %d, want 3", got)
	}
}

func TestMsc11LivingLiesDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lies := b12Push(g, me.ID, "Living Lies of Loki", "Creature — Illusion Villain", msc11LiesOracle, 1, 3)
	seedLibrary(me, "Top", "Second")
	hand := me.Hand.Size()
	b27Kill(g, lies)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
}

// --- Iron Man, Tony Stark ---

func TestMsc11IronManPumpsOnlyYourAttackingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ironMan := b12Push(g, me.ID, "Iron Man, Tony Stark", "Legendary Artifact Creature — Human Hero", msc11IronManOracle, 3, 3)
	if !hasEffectiveKeyword(t, g, ironMan, "flying") {
		t.Error("Iron Man has no flying")
	}
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	homebody := pushVanillaCreature(g, me.ID, "Homebody", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)

	if got := effectivePower(t, g, attacker); got != 3 {
		t.Errorf("the attacker has power %d, want 3", got)
	}
	if got := effectivePower(t, g, homebody); got != 2 {
		t.Errorf("a creature that stayed home has power %d, want 2", got)
	}
}

func TestMsc11IronManMakesARobotHeroForEachRedSpellYouCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Iron Man, Tony Stark", "Legendary Artifact Creature — Human Hero", msc11IronManOracle, 3, 3)
	robots := func() []game.Card {
		var out []game.Card
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Robot Hero" && c.Controller == me.ID {
					out = append(out, c)
				}
			}
		})
		return out
	}

	msc11Cast(t, g, "Blue Spell", "Sorcery", "{1}{U}", []string{"U"})
	passPriorityAroundTable(t, g)
	if n := len(robots()); n != 0 {
		t.Fatalf("a blue spell made %d Robot Heroes", n)
	}
	advanceToMain(t, g)
	id := handCardFull(opp, "Their Red Spell", "Instant", "{R}", "", []string{"R"})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(robots()); n != 0 {
		t.Fatalf("an opponent's red spell made %d Robot Heroes", n)
	}
	msc11Cast(t, g, "Red Spell", "Sorcery", "{1}{R}", []string{"R"})
	passPriorityAroundTable(t, g)
	got := robots()
	if len(got) != 1 {
		t.Fatalf("a red spell made %d Robot Heroes, want 1", len(got))
	}
	c := got[0]
	if c.Power != 2 || c.Toughness != 1 || !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Hero") || !c.HasSubtype("Robot") || len(c.Colors) != 0 {
		t.Errorf("Robot Hero is %d/%d %q colors %v, want a 2/1 colorless Robot Hero artifact creature", c.Power, c.Toughness, c.TypeLine, c.Colors)
	}
	if !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
		t.Error("the Robot Hero has no flying")
	}
}

// --- Gert and Old Lace, Runaways ---

func TestMsc11GertDiscardsToSearchForABasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	nonbasic := pushLibraryCardForTest(me, game.Card{Name: "Fancy Land", TypeLine: "Land"})
	fodder := handCard(me, "Fodder", "Instant")

	gert := rfdCast(t, g, "Gert and Old Lace, Runaways", "Legendary Creature — Human Dinosaur Hero", msc11GertOracle, 4, 3)
	if !hasEffectiveKeyword(t, g, gert, "trample") {
		t.Error("Gert and Old Lace has no trample")
	}
	if c := discardChoiceFor(g, me.ID); c == nil {
		t.Fatal("no discard prompt")
	}
	answerDiscard(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if !me.Hand.Contains(forest) {
		t.Error("the basic land was not searched into the hand")
	}
	if me.Hand.Contains(nonbasic) {
		t.Error("a nonbasic land was fetched")
	}
}

func TestMsc11GertSearchesNothingWhenYouDeclineToDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	handCard(me, "Fodder", "Instant")

	rfdCast(t, g, "Gert and Old Lace, Runaways", "Legendary Creature — Human Dinosaur Hero", msc11GertOracle, 4, 3)
	if c := discardChoiceFor(g, me.ID); c == nil {
		t.Fatal("no discard prompt")
	}
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if me.Hand.Contains(forest) {
		t.Error("a land was searched although no card was discarded")
	}
}
