package testss82v

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSlice293CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.S293RiteOfFlameOracle:      "Rite of Flame",
		effectstest.S293PyreticRitualOracle:    "Pyretic Ritual",
		effectstest.S293CullingRitualOracle:    "Culling Ritual",
		effectstest.S293GhostlyFlickerOracle:   "Ghostly Flicker",
		effectstest.S293NoxiousRevivalOracle:   "Noxious Revival",
		effectstest.S293ArchdruidsCharmOracle:  "Archdruid's Charm",
		effectstest.S293TirelessTrackerOracle:  "Tireless Tracker",
		effectstest.S293ChasmSkulkerOracle:     "Chasm Skulker",
		effectstest.S293TyvarsStandOracle:      "Tyvar's Stand",
		effectstest.S293SpellPierceOracle:      "Spell Pierce",
		effectstest.S293DismemberOracle:        "Dismember",
		effectstest.S293MuddleTheMixtureOracle: "Muddle the Mixture",
		effectstest.S293EvolutionWitnessOracle: "Evolution Witness",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Rite of Flame", "Sorcery", effectstest.S293RiteOfFlameOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 {
		t.Fatalf("first Rite of Flame added %v, want {R}{R}", got)
	}
	// One Rite now in my graveyard and one in an opponent's: the next
	// copy adds {R}{R} plus one for each.
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Rite of Flame", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	me.ManaPool = nil
	effectstest.CastCatalogSpell(t, g, "Rite of Flame", "Sorcery", effectstest.S293RiteOfFlameOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.Batch01PoolColors(me)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Pyretic Ritual", "Instant", effectstest.S293PyreticRitualOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.Batch01PoolColors(me)
	if len(got) != 3 || got[0] != "R" || got[1] != "R" || got[2] != "R" {
		t.Errorf("pool %v, want [R R R]", got)
	}
}

func TestSlice293CullingRitualDestroysSmallPermanentsAndPays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Small Bear",
		TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mind Stone",
		TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Ogre",
		TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")

	effectstest.CastCatalogSpell(t, g, "Culling Ritual", "Sorcery", effectstest.S293CullingRitualOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

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
		c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceMana)
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
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 {
		t.Errorf("pool %v, want two mana", got)
	}
}

// --- flicker -----------------------------------------------------

func TestSlice293GhostlyFlickerBlinksBothTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock B",
		TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(a, game.CounterPlusOne, 2) })

	effectstest.CastCatalogSpell(t, g, "Ghostly Flicker", "Instant", effectstest.S293GhostlyFlickerOracle, effectstest.CardRefsOf(a, b))
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Ghostly Flicker", "Instant", effectstest.S293GhostlyFlickerOracle, effectstest.CardRefsOf(a)); err == nil {
		t.Error("Ghostly Flicker was cast with one target")
	}
}

func TestSlice293GhostlyFlickerRefusesOpposingPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Ghostly Flicker", "Instant", effectstest.S293GhostlyFlickerOracle, effectstest.CardRefsOf(a, theirs)); err == nil {
		t.Error("Ghostly Flicker accepted a permanent I do not control")
	}
}

// --- Noxious Revival ---------------------------------------------

func TestSlice293NoxiousRevivalTucksOntoItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	dead := effectstest.PushGraveyardCardForTest(opp, "Dead Thing")
	effectstest.CastCatalogSpell(t, g, "Noxious Revival", "Instant", effectstest.S293NoxiousRevivalOracle, effectstest.CardRefsOf(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(dead) {
		t.Error("the card is still in the graveyard")
	}
	if got := effectstest.LibraryTopIDs(opp, 1); len(got) != 1 || got[0] != dead {
		t.Errorf("top of the owner's library = %v, want the revived card", got)
	}
}

// --- Archdruid's Charm -------------------------------------------

func s293CastCharm(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) error {
	t.Helper()
	_, err := effectstest.CastModalPaying(t, g, "Archdruid's Charm", "Instant", effectstest.S293ArchdruidsCharmOracle, modes, targets, nil, nil, nil)
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestSlice293ArchdruidsCharmFindsALandOntoTheBattlefieldTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := s293CastCharm(t, g, []int{0}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.TakeSearchPrompts(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Library Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if err := s293CastCharm(t, g, []int{0}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.TakeSearchPrompts(t, g)
	if !me.Hand.Contains(id) {
		t.Error("the creature card should be in hand")
	}
	if g.Battlefield.Contains(id) {
		t.Error("a creature card must not go onto the battlefield")
	}
}

func TestSlice293ArchdruidsCharmCounterThenBite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Troll", 2, 3)
	if err := s293CastCharm(t, g, []int{1}, effectstest.CardRefsOf(mine, theirs)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if c := effectstest.TwCard(g, mine); c == nil || c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("my creature should have its counter: %+v", c)
	}
	if effectstest.TwCard(g, theirs) != nil {
		t.Error("the 2/3 survived 3 damage — the counter must land before the damage is read")
	}
}

func TestSlice293ArchdruidsCharmRefusesYourOwnCreatureAsTheVictim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	if err := s293CastCharm(t, g, []int{1}, effectstest.CardRefsOf(a, b)); err == nil {
		t.Error("the bite accepted a creature I control as the second target")
	}
}

func TestSlice293ArchdruidsCharmExilesAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.PushArtifactForTest(g, opp.ID, "Rock")
	if err := s293CastCharm(t, g, []int{2}, effectstest.CardRefsOf(rock)); err != nil {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tracker := effectstest.B12Push(g, me.ID, "Tireless Tracker", "Creature — Human Scout", effectstest.S293TirelessTrackerOracle, 3, 2)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.TwCard(g, tracker); c == nil || c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("Tracker should have one +1/+1 counter after a Clue sacrifice: %+v", c)
	}
}

func TestSlice293TirelessTrackerIgnoresOtherSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tracker := effectstest.B12Push(g, me.ID, "Tireless Tracker", "Creature — Human Scout", effectstest.S293TirelessTrackerOracle, 3, 2)
	fodder := effectstest.SeedCreature(g, "Fodder", me.ID)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.TwCard(g, tracker); c == nil || c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("sacrificing a non-Clue grew the Tracker: %+v", c)
	}
}

// --- Chasm Skulker -----------------------------------------------

func TestSlice293ChasmSkulkerGrowsOnDrawAndLeavesSquidsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skulker := effectstest.B12Push(g, me.ID, "Chasm Skulker", "Creature — Squid Horror", effectstest.S293ChasmSkulkerOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if c := effectstest.TwCard(g, skulker); c == nil || c.Counters[game.CounterPlusOne] != 3 {
		t.Fatalf("Skulker after three draws: %+v, want three counters", c)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(skulker); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := s293Count(g, "Squid", me.ID); n != 3 {
		t.Errorf("%d Squid tokens, want one per counter (3)", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Squid" && !slices.Contains(effectstest.EffectiveAbilities(t, g, c.InstanceID), "islandwalk") {
			t.Error("a Squid token lacks islandwalk")
		}
	}
}

func TestSlice293ChasmSkulkerWithNoCountersLeavesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skulker := effectstest.B12Push(g, me.ID, "Chasm Skulker", "Creature — Squid Horror", effectstest.S293ChasmSkulkerOracle, 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(skulker) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := s293Count(g, "Squid", me.ID); n != 0 {
		t.Errorf("%d Squids from a counterless Skulker, want 0", n)
	}
}

// --- Tyvar's Stand / Dismember -----------------------------------

func TestSlice293TyvarsStandGrantsXAndProtection(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Tyvar's Stand", TypeLine: "Instant", ManaCost: "{X}{G}",
		OracleID: effectstest.S293TyvarsStandOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	// Strict is off, so the X is declared without paying for it.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3, Targets: effectstest.CardRefsOf(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5", got)
	}
	abilities := effectstest.EffectiveAbilities(t, g, bear)
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	small := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Ogre", 6, 6)
	effectstest.CastCatalogSpell(t, g, "Dismember", "Instant", effectstest.S293DismemberOracle, effectstest.CardRefsOf(small))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TwCard(g, small) != nil {
		t.Error("the 2/2 survived -5/-5")
	}
	effectstest.CastCatalogSpell(t, g, "Dismember", "Instant", effectstest.S293DismemberOracle, effectstest.CardRefsOf(big))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.TwCard(g, big); c == nil {
		t.Error("the 6/6 died to -5/-5")
	} else if got := effectstest.EffectivePower(t, g, big); got != 1 {
		t.Errorf("the 6/6 has power %d, want 1", got)
	}
}

// --- counterspells -----------------------------------------------

func TestSlice293SpellPierceTaxesANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastFromSeat(t, g, them, "Spell Pierce", "Instant", effectstest.S293SpellPierceOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassUntilTaxed(t, g, me.ID)
	if !g.Stack.Contains(victim) {
		t.Fatal("the spell resolved while the {2} was unanswered")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if g.Stack.Contains(victim) {
		t.Error("declining the tax must counter the spell")
	}
}

func TestSlice293SpellPierceRefusesACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell Pierce", TypeLine: "Instant",
		OracleID: effectstest.S293SpellPierceOracle, Owner: them.ID, Controller: them.ID})
	if err := g.CastSpell(them.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err == nil {
		t.Error("Spell Pierce targeted a creature spell")
	}
}

func TestSlice293MuddleTheMixtureCountersAnInstantButNotACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Instant", "Instant", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastFromSeat(t, g, them, "Muddle the Mixture", "Instant", effectstest.S293MuddleTheMixtureOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[0].Graveyard.Contains(victim) == false {
		t.Error("the countered instant should be in its owner's graveyard")
	}

	g2 := effectstest.NewCatalogGame(t)
	them2 := g2.Seats[1]
	effectstest.ToMainForCost(t, g2)
	creature := effectstest.CastCatalogSpell(t, g2, "Their Bear", "Creature — Bear", "", nil)
	if err := g2.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them2.Hand.PushTop(game.Card{InstanceID: id, Name: "Muddle the Mixture", TypeLine: "Instant",
		OracleID: effectstest.S293MuddleTheMixtureOracle, Owner: them2.ID, Controller: them2.ID})
	if err := g2.CastSpell(them2.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}}}); err == nil {
		t.Error("Muddle the Mixture targeted a creature spell")
	}
}

func TestSlice293MuddleTheMixtureTransmutesForSameManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	muddle := effectstest.HandCardFull(me, "Muddle the Mixture", "Instant", "{U}{U}", effectstest.S293MuddleTheMixtureOracle, []string{"U"})
	needle := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Two Drop", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	spec, _ := effects.Lookup(effectstest.S293MuddleTheMixtureOracle)
	idx := -1
	for i, a := range spec.Activated {
		if len(a.Label) >= 9 && a.Label[:9] == "Transmute" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("no Transmute ability declared")
	}
	effectstest.B16Activate(t, g, me.ID, muddle, idx, game.ActivateAbilityParams{})
	effectstest.TakeSearchPrompts(t, g)
	if !me.Hand.Contains(needle) {
		t.Error("transmute should have fetched the mana value 2 card")
	}
	if !me.Graveyard.Contains(muddle) {
		t.Error("transmute should discard the card")
	}
}

// --- Evolution Witness -------------------------------------------

func TestSlice293EvolutionWitnessAdaptsOnceAndReturnsAPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	witness := effectstest.B12Push(g, me.ID, "Evolution Witness", "Creature — Elf Shaman Mutant", effectstest.S293EvolutionWitnessOracle, 2, 2)
	creature := effectstest.PushGraveyardCardWithTypeLineID(me, "Dead Bear", "Creature — Bear")
	spell := effectstest.PushGraveyardCardWithTypeLineID(me, "Dead Spell", "Instant")
	effectstest.AdvanceToMain(t, g)

	effectstest.B16Activate(t, g, me.ID, witness, 0, game.ActivateAbilityParams{})
	// The counters arrived, so the trigger asks for its target.
	effectstest.PickTriggerTarget(t, g, me.ID, creature)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.TwCard(g, witness); c == nil || c.Counters[game.CounterPlusOne] != 2 {
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
	again := effectstest.PushGraveyardCardWithTypeLineID(me, "Dead Bear 2", "Creature — Bear")
	effectstest.B16Activate(t, g, me.ID, witness, 0, game.ActivateAbilityParams{})
	if c := effectstest.TwCard(g, witness); c == nil || c.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("a second adapt changed the counters: %+v", c)
	}
	if me.Hand.Contains(again) {
		t.Error("a no-op adapt still returned a card")
	}
}
