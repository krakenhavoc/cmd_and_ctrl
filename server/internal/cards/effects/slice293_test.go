package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// slice293_test.go — card-level coverage for the play-rate slice cut
// from tracker #293's gap lists (rites and rituals, flicker, the
// Phyrexian-mana instants, the pay-unless and counter spells, and three
// counter-themed creatures).

const (
	s293RiteOfFlameOracle      = "8a2e53f9-8100-488f-8504-b59e9bd1cc29"
	s293PyreticRitualOracle    = "86ef6474-613f-41fb-931c-d4279b03ed99"
	s293CullingRitualOracle    = "15e5136e-ed15-49a8-b027-e09436673fb4"
	s293GhostlyFlickerOracle   = "ad0070db-6454-41cb-861f-8f5b8fc2a3b8"
	s293NoxiousRevivalOracle   = "97cabeda-9fe3-490d-99b4-4c8d87c17157"
	s293ArchdruidsCharmOracle  = "3c1ef404-e2c6-486d-a5a2-d5779c71d498"
	s293TirelessTrackerOracle  = "e6555cca-e608-4c87-b835-9cd47fd1f543"
	s293ChasmSkulkerOracle     = "69facbc7-3859-4716-b627-5199571fb3cf"
	s293TyvarsStandOracle      = "4247b667-a0a8-4995-97c7-622d20132f7d"
	s293SpellPierceOracle      = "d64b0848-0193-4025-ba62-63ecd8fb9f50"
	s293DismemberOracle        = "fd74f8eb-0253-42dd-8277-186d4934da38"
	s293MuddleTheMixtureOracle = "0826cf7d-7ccc-459b-a92f-29dc169628f8"
	s293EvolutionWitnessOracle = "0e07f1af-5ff6-4da5-8683-aea4cc390975"
)

func TestSlice293CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		s293RiteOfFlameOracle:      "Rite of Flame",
		s293PyreticRitualOracle:    "Pyretic Ritual",
		s293CullingRitualOracle:    "Culling Ritual",
		s293GhostlyFlickerOracle:   "Ghostly Flicker",
		s293NoxiousRevivalOracle:   "Noxious Revival",
		s293ArchdruidsCharmOracle:  "Archdruid's Charm",
		s293TirelessTrackerOracle:  "Tireless Tracker",
		s293ChasmSkulkerOracle:     "Chasm Skulker",
		s293TyvarsStandOracle:      "Tyvar's Stand",
		s293SpellPierceOracle:      "Spell Pierce",
		s293DismemberOracle:        "Dismember",
		s293MuddleTheMixtureOracle: "Muddle the Mixture",
		s293EvolutionWitnessOracle: "Evolution Witness",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
		} else if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

func s293Count(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

// --- rituals -----------------------------------------------------

func TestSlice293RiteOfFlameCountsEveryGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Rite of Flame", "Sorcery", s293RiteOfFlameOracle, nil)
	passPriorityAroundTable(t, g)
	if got := batch01PoolColors(me); len(got) != 2 {
		t.Fatalf("first Rite of Flame added %v, want {R}{R}", got)
	}
	// One Rite now in my graveyard and one in an opponent's: the next
	// copy adds {R}{R} plus one for each.
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Rite of Flame", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	me.ManaPool = nil
	castCatalogSpell(t, g, "Rite of Flame", "Sorcery", s293RiteOfFlameOracle, nil)
	passPriorityAroundTable(t, g)
	got := batch01PoolColors(me)
	if len(got) != 4 {
		t.Fatalf("with two Rites in graveyards the third added %v, want four mana", got)
	}
	for _, c := range got {
		if c != "R" {
			t.Errorf("Rite of Flame added %q, want only R", c)
		}
	}
}

func TestSlice293PyreticRitualAddsThreeRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Pyretic Ritual", "Instant", s293PyreticRitualOracle, nil)
	passPriorityAroundTable(t, g)
	got := batch01PoolColors(me)
	if len(got) != 3 || got[0] != "R" || got[1] != "R" || got[2] != "R" {
		t.Errorf("pool %v, want [R R R]", got)
	}
}

func TestSlice293CullingRitualDestroysSmallPermanentsAndPays(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Small Bear",
		TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mind Stone",
		TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	big := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Ogre",
		TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")

	castCatalogSpell(t, g, "Culling Ritual", "Sorcery", s293CullingRitualOracle, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(small) || g.Battlefield.Contains(rock) {
		t.Error("a permanent with mana value 2 or less survived")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("a mana value 3 creature was destroyed")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was destroyed")
	}
	// Two permanents destroyed: two {B}-or-{G} slots, one prompt each.
	picks := 0
	for i := 0; i < 4; i++ {
		c := latestChoiceOfKind(g, game.PendingChoiceMana)
		if c == nil {
			break
		}
		picks++
		if err := g.ResolveManaChoice(c.ID, me.ID, "G"); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if picks != 2 {
		t.Errorf("%d mana prompts, want one per destroyed permanent (2)", picks)
	}
	if got := batch01PoolColors(me); len(got) != 2 {
		t.Errorf("pool %v, want two mana", got)
	}
}

// --- flicker -----------------------------------------------------

func TestSlice293GhostlyFlickerBlinksBothTargets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock B",
		TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(a, game.CounterPlusOne, 2) })

	castCatalogSpell(t, g, "Ghostly Flicker", "Instant", s293GhostlyFlickerOracle, cardRefsOf(a, b))
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Fatal("an original object is still on the battlefield — it was not exiled")
	}
	if s293Count(g, "Bear A", me.ID) != 1 || s293Count(g, "Rock B", me.ID) != 1 {
		t.Fatal("both cards should be back under my control")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear A" && c.Counters[game.CounterPlusOne] != 0 {
			t.Error("the returned creature kept its counters; it must be a new object")
		}
	}
}

func TestSlice293GhostlyFlickerNeedsTwoTargets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	if err := castCatalogSpellErr(t, g, "Ghostly Flicker", "Instant", s293GhostlyFlickerOracle, cardRefsOf(a)); err == nil {
		t.Error("Ghostly Flicker was cast with one target")
	}
}

func TestSlice293GhostlyFlickerRefusesOpposingPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Ghostly Flicker", "Instant", s293GhostlyFlickerOracle, cardRefsOf(a, theirs)); err == nil {
		t.Error("Ghostly Flicker accepted a permanent I do not control")
	}
}

// --- Noxious Revival ---------------------------------------------

func TestSlice293NoxiousRevivalTucksOntoItsOwnersLibrary(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	dead := pushGraveyardCardForTest(opp, "Dead Thing")
	castCatalogSpell(t, g, "Noxious Revival", "Instant", s293NoxiousRevivalOracle, cardRefsOf(dead))
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(dead) {
		t.Error("the card is still in the graveyard")
	}
	if got := libraryTopIDs(opp, 1); len(got) != 1 || got[0] != dead {
		t.Errorf("top of the owner's library = %v, want the revived card", got)
	}
}

// --- Archdruid's Charm -------------------------------------------

func s293CastCharm(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) error {
	t.Helper()
	_, err := castModalPaying(t, g, "Archdruid's Charm", "Instant", s293ArchdruidsCharmOracle, modes, targets, nil, nil, nil)
	if err == nil {
		passPriorityAroundTable(t, g)
	}
	return err
}

func TestSlice293ArchdruidsCharmFindsALandOntoTheBattlefieldTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := s293CastCharm(t, g, []int{0}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && c.Controller == me.ID {
			found = true
			if !c.Tapped {
				t.Error("the fetched land should enter tapped")
			}
		}
	}
	if !found {
		t.Error("the land was not put onto the battlefield")
	}
}

func TestSlice293ArchdruidsCharmFindsACreatureIntoHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Library Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if err := s293CastCharm(t, g, []int{0}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !me.Hand.Contains(id) {
		t.Error("the creature card should be in hand")
	}
	if g.Battlefield.Contains(id) {
		t.Error("a creature card must not go onto the battlefield")
	}
}

func TestSlice293ArchdruidsCharmCounterThenBite(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Troll", 2, 3)
	if err := s293CastCharm(t, g, []int{1}, cardRefsOf(mine, theirs)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if c := twCard(g, mine); c == nil || c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("my creature should have its counter: %+v", c)
	}
	if twCard(g, theirs) != nil {
		t.Error("the 2/3 survived 3 damage — the counter must land before the damage is read")
	}
}

func TestSlice293ArchdruidsCharmRefusesYourOwnCreatureAsTheVictim(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	if err := s293CastCharm(t, g, []int{1}, cardRefsOf(a, b)); err == nil {
		t.Error("the bite accepted a creature I control as the second target")
	}
}

func TestSlice293ArchdruidsCharmExilesAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := pushArtifactForTest(g, opp.ID, "Rock")
	if err := s293CastCharm(t, g, []int{2}, cardRefsOf(rock)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is still on the battlefield")
	}
	if opp.Graveyard.Contains(rock) {
		t.Error("the artifact was destroyed, not exiled")
	}
}

// --- Tireless Tracker --------------------------------------------

func TestSlice293TirelessTrackerInvestigatesOnLandfallAndGrowsOnClueSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tracker := b12Push(g, me.ID, "Tireless Tracker", "Creature — Human Scout", s293TirelessTrackerOracle, 3, 2)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if n := s293Count(g, "Clue", me.ID); n != 1 {
		t.Fatalf("%d Clues after a land entered, want 1", n)
	}
	var clue uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Clue" {
			clue = c.InstanceID
		}
	}
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(clue); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if c := twCard(g, tracker); c == nil || c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("Tracker should have one +1/+1 counter after a Clue sacrifice: %+v", c)
	}
}

func TestSlice293TirelessTrackerIgnoresOtherSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tracker := b12Push(g, me.ID, "Tireless Tracker", "Creature — Human Scout", s293TirelessTrackerOracle, 3, 2)
	fodder := seedCreature(g, "Fodder", me.ID)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	passPriorityAroundTable(t, g)
	if c := twCard(g, tracker); c == nil || c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("sacrificing a non-Clue grew the Tracker: %+v", c)
	}
}

// --- Chasm Skulker -----------------------------------------------

func TestSlice293ChasmSkulkerGrowsOnDrawAndLeavesSquidsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	skulker := b12Push(g, me.ID, "Chasm Skulker", "Creature — Squid Horror", s293ChasmSkulkerOracle, 1, 1)
	advanceToMain(t, g)
	for i := 0; i < 3; i++ {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if c := twCard(g, skulker); c == nil || c.Counters[game.CounterPlusOne] != 3 {
		t.Fatalf("Skulker after three draws: %+v, want three counters", c)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(skulker); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if n := s293Count(g, "Squid", me.ID); n != 3 {
		t.Errorf("%d Squid tokens, want one per counter (3)", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Squid" && !slices.Contains(effectiveAbilities(t, g, c.InstanceID), "islandwalk") {
			t.Error("a Squid token lacks islandwalk")
		}
	}
}

func TestSlice293ChasmSkulkerWithNoCountersLeavesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	skulker := b12Push(g, me.ID, "Chasm Skulker", "Creature — Squid Horror", s293ChasmSkulkerOracle, 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(skulker) })
	passPriorityAroundTable(t, g)
	if n := s293Count(g, "Squid", me.ID); n != 0 {
		t.Errorf("%d Squids from a counterless Skulker, want 0", n)
	}
}

// --- Tyvar's Stand / Dismember -----------------------------------

func TestSlice293TyvarsStandGrantsXAndProtection(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Tyvar's Stand", TypeLine: "Instant", ManaCost: "{X}{G}",
		OracleID: s293TyvarsStandOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	// Strict is off, so the X is declared without paying for it.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3, Targets: cardRefsOf(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5", got)
	}
	abilities := effectiveAbilities(t, g, bear)
	has := func(k string) bool {
		for _, a := range abilities {
			if a == k {
				return true
			}
		}
		return false
	}
	if !has("hexproof") || !has("indestructible") {
		t.Errorf("abilities %v, want hexproof and indestructible", abilities)
	}
}

func TestSlice293DismemberShrinksByFive(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	big := pushVanillaCreature(g, opp.ID, "Ogre", 6, 6)
	castCatalogSpell(t, g, "Dismember", "Instant", s293DismemberOracle, cardRefsOf(small))
	passPriorityAroundTable(t, g)
	if twCard(g, small) != nil {
		t.Error("the 2/2 survived -5/-5")
	}
	castCatalogSpell(t, g, "Dismember", "Instant", s293DismemberOracle, cardRefsOf(big))
	passPriorityAroundTable(t, g)
	if c := twCard(g, big); c == nil {
		t.Error("the 6/6 died to -5/-5")
	} else if got := effectivePower(t, g, big); got != 1 {
		t.Errorf("the 6/6 has power %d, want 1", got)
	}
}

// --- counterspells -----------------------------------------------

func TestSlice293SpellPierceTaxesANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castFromSeat(t, g, them, "Spell Pierce", "Instant", s293SpellPierceOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passUntilTaxed(t, g, me.ID)
	if !g.Stack.Contains(victim) {
		t.Fatal("the spell resolved while the {2} was unanswered")
	}
	answerPayUnless(t, g, me.ID, false)
	if g.Stack.Contains(victim) {
		t.Error("declining the tax must counter the spell")
	}
}

func TestSlice293SpellPierceRefusesACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	them := g.Seats[1]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell Pierce", TypeLine: "Instant",
		OracleID: s293SpellPierceOracle, Owner: them.ID, Controller: them.ID})
	if err := g.CastSpell(them.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err == nil {
		t.Error("Spell Pierce targeted a creature spell")
	}
}

func TestSlice293MuddleTheMixtureCountersAnInstantButNotACreature(t *testing.T) {
	g := newCatalogGame(t)
	them := g.Seats[1]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Instant", "Instant", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castFromSeat(t, g, them, "Muddle the Mixture", "Instant", s293MuddleTheMixtureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if g.Seats[0].Graveyard.Contains(victim) == false {
		t.Error("the countered instant should be in its owner's graveyard")
	}

	g2 := newCatalogGame(t)
	them2 := g2.Seats[1]
	toMainForCost(t, g2)
	creature := castCatalogSpell(t, g2, "Their Bear", "Creature — Bear", "", nil)
	if err := g2.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them2.Hand.PushTop(game.Card{InstanceID: id, Name: "Muddle the Mixture", TypeLine: "Instant",
		OracleID: s293MuddleTheMixtureOracle, Owner: them2.ID, Controller: them2.ID})
	if err := g2.CastSpell(them2.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}}}); err == nil {
		t.Error("Muddle the Mixture targeted a creature spell")
	}
}

func TestSlice293MuddleTheMixtureTransmutesForSameManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	muddle := handCardFull(me, "Muddle the Mixture", "Instant", "{U}{U}", s293MuddleTheMixtureOracle, []string{"U"})
	needle := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Two Drop", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	spec, _ := Lookup(s293MuddleTheMixtureOracle)
	idx := -1
	for i, a := range spec.Activated {
		if len(a.Label) >= 9 && a.Label[:9] == "Transmute" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("no Transmute ability declared")
	}
	b16Activate(t, g, me.ID, muddle, idx, game.ActivateAbilityParams{})
	if !me.Hand.Contains(needle) {
		t.Error("transmute should have fetched the mana value 2 card")
	}
	if !me.Graveyard.Contains(muddle) {
		t.Error("transmute should discard the card")
	}
}

// --- Evolution Witness -------------------------------------------

func TestSlice293EvolutionWitnessAdaptsOnceAndReturnsAPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	witness := b12Push(g, me.ID, "Evolution Witness", "Creature — Elf Shaman Mutant", s293EvolutionWitnessOracle, 2, 2)
	creature := pushGraveyardCardWithTypeLineID(me, "Dead Bear", "Creature — Bear")
	spell := pushGraveyardCardWithTypeLineID(me, "Dead Spell", "Instant")
	advanceToMain(t, g)

	b16Activate(t, g, me.ID, witness, 0, game.ActivateAbilityParams{})
	// The counters arrived, so the trigger asks for its target.
	pickTriggerTarget(t, g, me.ID, creature)
	passPriorityAroundTable(t, g)
	if c := twCard(g, witness); c == nil || c.Counters[game.CounterPlusOne] != 2 {
		t.Fatalf("Witness after adapt: %+v, want two counters", c)
	}
	if !me.Hand.Contains(creature) && !me.Hand.Contains(spell) {
		t.Fatalf("nothing came back to hand")
	}
	if me.Hand.Contains(spell) {
		t.Error("an instant card is not a permanent card and must not be returned")
	}
	if !me.Hand.Contains(creature) {
		t.Error("the permanent card should have returned to hand")
	}

	// Already has counters: a second adapt adds none and triggers nothing.
	again := pushGraveyardCardWithTypeLineID(me, "Dead Bear 2", "Creature — Bear")
	b16Activate(t, g, me.ID, witness, 0, game.ActivateAbilityParams{})
	if c := twCard(g, witness); c == nil || c.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("a second adapt changed the counters: %+v", c)
	}
	if me.Hand.Contains(again) {
		t.Error("a no-op adapt still returned a card")
	}
}

func pushGraveyardCardWithTypeLineID(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}
