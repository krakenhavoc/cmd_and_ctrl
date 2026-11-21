package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b08Tap taps a battlefield permanent through the effect API.
func b08Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
}

// b08Permanent pushes a catalog permanent with a real entry event, so
// the layer engine sees it — pushCatalogPermanent emits none, which
// is fine for triggers and abilities but leaves a static invisible
// until something else invalidates the cache.
func b08Permanent(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	})
}

// b08TypedCreature pushes a live creature with the given type line
// under owner, no summoning sickness, with a real ETB timestamp.
func b08TypedCreature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b08EmptyLibrary pops every card out of a player's library.
func b08EmptyLibrary(p *game.Player) {
	for p.Library.Size() > 0 {
		_, _ = p.Library.PopTop()
	}
}

// b08OpponentCastRefused tries to cast `name` from a non-active
// player's hand at the given targets and reports whether the engine
// refused it — the "wrong kind of target" check for an instant.
func b08OpponentCastRefused(t *testing.T, g *game.Game, opp *game.Player, name, oracle string, targets []game.TargetRef) bool {
	t.Helper()
	id := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle,
		Owner: opp.ID, Controller: opp.ID,
	})
	return g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: targets}) != nil
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Bloodfell
// Caves is a row in the gain-land table, so a transposed row is
// invisible until someone plays that exact card.
func TestBatch08CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B08ResculptOracle:              "Resculpt",
		effectstest.B08WarleadersCallOracle:        "Warleader's Call",
		effectstest.B08ZuranOrbOracle:              "Zuran Orb",
		effectstest.B08SpiritedCompanionOracle:     "Spirited Companion",
		effectstest.B08BlightedWoodlandOracle:      "Blighted Woodland",
		effectstest.B08BristlyBillOracle:           "Bristly Bill, Spine Sower",
		effectstest.B08OldGnawboneOracle:           "Old Gnawbone",
		effectstest.B08DispelOracle:                "Dispel",
		effectstest.B08FountainportOracle:          "Fountainport",
		effectstest.B08SolveTheEquationOracle:      "Solve the Equation",
		effectstest.B08BloodfellCavesOracle:        "Bloodfell Caves",
		effectstest.B08ThopterSpyNetworkOracle:     "Thopter Spy Network",
		effectstest.B08WildernessReclamationOracle: "Wilderness Reclamation",
		effectstest.B08CoatOfArmsOracle:            "Coat of Arms",
		effectstest.B08OmnathLocusOfRageOracle:     "Omnath, Locus of Rage",
		effectstest.B08PhyrexianReclamationOracle:  "Phyrexian Reclamation",
		effectstest.B08CruelCelebrantOracle:        "Cruel Celebrant",
		effectstest.B08ArchaeomancerOracle:         "Archaeomancer",
		effectstest.B08MindcrankOracle:             "Mindcrank",
		effectstest.B08CabarettiCourtyardOracle:    "Cabaretti Courtyard",
		effectstest.B08RazorkinNeedleheadOracle:    "Razorkin Needlehead",
		effectstest.B08KrosanVergeOracle:           "Krosan Verge",
		effectstest.B08MoldervineReclamationOracle: "Moldervine Reclamation",
		effectstest.B08SythisOracle:                "Sythis, Harvest's Hand",
		effectstest.B08EnchantresssPresenceOracle:  "Enchantress's Presence",
		effectstest.B08SunkenRuinsOracle:           "Sunken Ruins",
		effectstest.B08EmeriaOracle:                "Emeria, the Sky Ruin",
		effectstest.B08NoxiousGearhulkOracle:       "Noxious Gearhulk",
		effectstest.B08CutADealOracle:              "Cut a Deal",
		effectstest.B08GavonyTownshipOracle:        "Gavony Township",
		effectstest.B08AllIsDustOracle:             "All Is Dust",
		effectstest.B08MinesOfMoriaOracle:          "Mines of Moria",
		effectstest.B08TorbranOracle:               "Torbran, Thane of Red Fell",
		effectstest.B08ManagorgerHydraOracle:       "Managorger Hydra",
	}
	if len(want) != 34 {
		t.Fatalf("the batch is 34 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness == effects.CompletenessUnreviewed {
			t.Errorf("%s ships without a completeness declaration", name)
		}
	}
}

// --- spells --------------------------------------------------------

func TestB08ResculptExilesAnArtifactAndHandsItsControllerAnElemental(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Resculpt", "Instant", effectstest.B08ResculptOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) {
		t.Error("the artifact should be in exile")
	}
	if effectstest.CountBattlefieldNamed(g, opp.ID, "Elemental") != 1 {
		t.Fatal("the VICTIM should get the 4/4 Elemental")
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Elemental") != 0 {
		t.Error("the caster must not get the token")
	}
	elemental := effectstest.FindBattlefieldByName(g, "Elemental")
	card, _ := effectstest.BattlefieldCard(g, elemental)
	if !card.HasColor("U") || !card.HasColor("R") || card.Power != 4 || card.Toughness != 4 {
		t.Errorf("the token is a 4/4 blue and red Elemental, got %d/%d %v", card.Power, card.Toughness, card.Colors)
	}
	// A land is neither an artifact nor a creature.
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	if !effectstest.B03CastRefused(t, g, "Resculpt", "Instant", effectstest.B08ResculptOracle, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}) {
		t.Error("a land was accepted for 'target artifact or creature'")
	}
}

func TestB08DispelCountersOnlyAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Their Instant", effectstest.LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.CastCatalogSpell(t, g, "Dispel", "Instant", effectstest.B08DispelOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the instant was not countered")
	}

	// A creature spell is not an instant spell.
	bear := effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if !b08OpponentCastRefused(t, g, opp, "Dispel", effectstest.B08DispelOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}) {
		t.Error("a creature spell was accepted for 'target instant spell'")
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestB08SolveTheEquationTutorsAnInstantOrSorceryToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bolt := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant"})
	ponder := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Ponder", TypeLine: "Sorcery"})
	bear := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})

	effectstest.CastCatalogSpell(t, g, "Solve the Equation", "Sorcery", effectstest.B08SolveTheEquationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("want a search prompt")
	}
	for _, id := range c.SearchCards {
		if id == bear {
			t.Error("a creature card was offered for 'an instant or sorcery card'")
		}
	}
	if !effectstest.HasID(c.SearchCards, bolt) || !effectstest.HasID(c.SearchCards, ponder) {
		t.Error("both the instant and the sorcery should be offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ponder)
	if !me.Hand.Contains(ponder) {
		t.Error("the chosen sorcery did not reach the hand")
	}
}

func TestB08CutADealDrawsForEachOpponentWhoDrew(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Hand.Size()
	}
	effectstest.CastCatalogSpell(t, g, "Cut a Deal", "Sorcery", effectstest.B08CutADealOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if got := p.Hand.Size() - before[p.ID]; got != 1 {
			t.Errorf("%s drew %d, want 1", p.Name, got)
		}
	}
	// castCatalogSpell adds the spell to the hand and casting removes
	// it, so the difference is exactly the cards drawn.
	if got := me.Hand.Size() - before[me.ID]; got != 3 {
		t.Errorf("the caster drew %d, want 3 (three opponents drew)", got)
	}

	// An opponent with no library draws nothing and earns nothing.
	b08EmptyLibrary(g.Seats[3])
	mid := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Cut a Deal", "Sorcery", effectstest.B08CutADealOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - mid; got != 2 {
		t.Errorf("the caster drew %d, want 2 (one opponent could not draw)", got)
	}
}

func TestB08AllIsDustSacrificesOnlyColouredPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.SeedPermanentFor(g, me.ID, "Rock", "Artifact")
	eldrazi := effectstest.PushVanillaCreature(g, me.ID, "Eldrazi", 5, 5)
	theirLand := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	glory := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glory", TypeLine: "Enchantment", Colors: []string{"W"},
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.CastCatalogSpell(t, g, "All Is Dust", "Kindred Sorcery — Eldrazi", effectstest.B08AllIsDustOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, tc := range []struct {
		id   uuid.UUID
		want bool
		why  string
	}{
		{rock, true, "a colourless artifact stays"},
		{eldrazi, true, "a colourless creature stays"},
		{theirLand, true, "a land stays"},
		{green, false, "a green creature is sacrificed"},
		{glory, false, "a white enchantment is sacrificed — yours too"},
	} {
		if got := g.Battlefield.Contains(tc.id); got != tc.want {
			t.Errorf("%s: on battlefield = %v, want %v", tc.why, got, tc.want)
		}
	}
	if !opp.Graveyard.Contains(green) || !me.Graveyard.Contains(glory) {
		t.Error("sacrificed permanents go to their owners' graveyards")
	}
	if effectstest.CountBattlefieldNamed(g, opp.ID, "Goblin") != 0 {
		t.Error("a red token is sacrificed")
	}
}

// --- ETB and cast triggers -----------------------------------------

func TestB08SpiritedCompanionDrawsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Spirited Companion", "Enchantment Creature — Dog", effectstest.B08SpiritedCompanionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestB08ArchaeomancerReturnsAnInstantOrSorceryCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bolt := effectstest.Batch01GraveyardCard(me, "Lightning Bolt", "Instant")
	bear := effectstest.PushGraveyardCardForTest(me, "Dead Bear")

	effectstest.CastCatalogSpell(t, g, "Archaeomancer", "Creature — Human Wizard", effectstest.B08ArchaeomancerOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(pick.PickTargetCards, bear) {
		t.Error("a creature card was offered for 'target instant or sorcery card'")
	}
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant did not come back to hand")
	}
}

func TestB08SythisGainsAndDrawsOnAnEnchantmentSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Sythis, Harvest's Hand", "Legendary Enchantment Creature — Nymph", effectstest.B08SythisOracle, false)
	hand, life := me.Hand.Size(), me.Life

	effectstest.CastCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}

	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("a creature spell is not an enchantment spell: drew %d, want still 1", got)
	}
}

func TestB08EnchantresssPresenceDrawsWithoutAsking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Enchantress's Presence", "Enchantment", effectstest.B08EnchantresssPresenceOracle, false)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	if len(g.PendingChoices) != 0 {
		t.Fatal("the Presence is mandatory — no prompt")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestB08ManagorgerHydraGrowsOnEveryPlayersSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := effectstest.PushCatalogPermanent(g, me.ID, "Managorger Hydra", "Creature — Hydra", effectstest.B08ManagorgerHydraOracle, false)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, hydra), "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}

	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, hydra, "+1/+1"); got != 1 {
		t.Fatalf("after my spell: %d counters, want 1", got)
	}
	effectstest.Batch01OpponentCasts(t, g, opp, "Their Instant", effectstest.LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, hydra, "+1/+1"); got != 2 {
		t.Errorf("after an opponent's spell: %d counters, want 2", got)
	}
}

func TestB08WarleadersCallPumpsAndPings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Warleader's Call", "Enchantment", effectstest.B08WarleadersCallOracle, false)
	mine := effectstest.SeedCreature(g, "Mine", me.ID)
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)
	if p, tough := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("my 2/2 = %d/%d, want 3/3", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("their 2/2 = %d, want the printed 2", p)
	}

	before := effectstest.LifeOfOpponents(g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d → %d, want -2 (two creatures entered)", i+1, b, got)
		}
	}
	// An opponent's creature and a noncreature of yours are silent.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1)
		_ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d → %d, want still -2", i+1, b, got)
		}
	}
}

func TestB08NoxiousGearhulkDestroysAnotherCreatureAndGainsItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B04Creature(g, opp.ID, "Their Big", 3, 5)
	before := me.Life

	hulk := effectstest.CastCatalogSpell(t, g, "Noxious Gearhulk", "Artifact Creature — Construct", effectstest.B08NoxiousGearhulkOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(pick.PickTargetCards, hulk) {
		t.Error("the Gearhulk offered itself for 'another target creature'")
	}
	effectstest.PickCard(t, g, me.ID, big)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Fatal("the targeted creature should be destroyed")
	}
	if me.Life != before+5 {
		t.Errorf("life %d → %d, want +5 (its toughness)", before, me.Life)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, hulk), "menace") {
		t.Error("printed menace did not reach the effective abilities")
	}
}

func TestB08NoxiousGearhulkGainsNothingForAnIndestructibleCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tough := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Darksteel", TypeLine: "Artifact Creature — Golem",
		Power: 4, Toughness: 4, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Noxious Gearhulk", "Artifact Creature — Construct", effectstest.B08NoxiousGearhulkOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, tough)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tough) {
		t.Fatal("an indestructible creature survives")
	}
	if me.Life != before {
		t.Errorf("life %d → %d, want unchanged — nothing was destroyed this way", before, me.Life)
	}
}

// --- landfall ------------------------------------------------------

func TestB08BristlyBillGrowsATargetOnLandfallAndDoublesOnDemand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bill := effectstest.PushCatalogPermanent(g, me.ID, "Bristly Bill, Spine Sower", "Legendary Creature — Plant Druid", effectstest.B08BristlyBillOracle, false)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)

	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(pick.PickTargetCards, theirs) {
		t.Error("'target creature' includes an opponent's creature")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 1 {
		t.Fatalf("landfall: %d counters on the bear, want 1", got)
	}

	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bill, "+1/+1", 3)
		_ = g.AddCounterForEffect(theirs, "+1/+1", 2)
	})
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "G"},
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, bill, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
		why  string
	}{
		{bear, 2, "one counter doubles to two"},
		{bill, 6, "three double to six"},
		{theirs, 2, "an opponent's creature is untouched"},
	} {
		if got := effectstest.CounterCount(g, tc.id, "+1/+1"); got != tc.want {
			t.Errorf("%s: %d counters, want %d", tc.why, got, tc.want)
		}
	}
}

func TestB08OmnathMakesElementalsOnLandfallAndBoltsWhenTheyDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	omnath := effectstest.PushCatalogPermanent(g, me.ID, "Omnath, Locus of Rage", "Legendary Creature — Elemental", effectstest.B08OmnathLocusOfRageOracle, false)

	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Elemental") != 1 {
		t.Fatal("a land drop should make one Elemental")
	}
	token := effectstest.FindBattlefieldByName(g, "Elemental")
	if card, _ := effectstest.BattlefieldCard(g, token); card.Power != 5 || !card.HasColor("R") || !card.HasColor("G") {
		t.Errorf("the token is a 5/5 red and green Elemental, got %d/%d %v", card.Power, card.Toughness, card.Colors)
	}

	before := opp.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(token) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("an Elemental dying: %d → %d, want -3", before, opp.Life)
	}

	// A non-Elemental of yours and an opponent's Elemental are silent.
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	theirs := b08TypedCreature(g, opp.ID, "Their Elemental", "Creature — Elemental", 3, 3)
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(bear)
		_ = g.SacrificePermanentForEffect(theirs)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("neither death should trigger Omnath")
	}

	// Omnath himself dying is the first clause.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(omnath) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("Omnath dying: %d → %d, want -6 in total", before, opp.Life)
	}
}

// --- dies and life-loss triggers -----------------------------------

func TestB08CruelCelebrantDrainsOnYourCreaturesPlaneswalkersAndItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	celebrant := effectstest.PushCatalogPermanent(g, me.ID, "Cruel Celebrant", "Creature — Vampire", effectstest.B08CruelCelebrantOracle, false)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)
	bounced := effectstest.SeedCreature(g, "Bounced", me.ID)
	before := effectstest.LifeOfOpponents(g)
	meBefore := me.Life

	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(bear)
		_ = g.SacrificePermanentForEffect(walker)
		_ = g.SacrificePermanentForEffect(theirs)
		_ = g.BounceToHandForEffect(bounced)
	})
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d → %d, want -2 (a creature and a planeswalker of yours died)", i+1, b, got)
		}
	}
	if me.Life != meBefore+2 {
		t.Errorf("controller %d → %d, want +2", meBefore, me.Life)
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(celebrant) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-3 {
			t.Errorf("opponent %d after the Celebrant's own death: %d → %d, want -3", i+1, b, got)
		}
	}
}

func TestB08MoldervineReclamationPaysOutOnYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Moldervine Reclamation", "Enchantment", effectstest.B08MoldervineReclamationOracle, false)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)
	hand, life := me.Hand.Size(), me.Life

	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(bear)
		_ = g.SacrificePermanentForEffect(theirs)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
}

func TestB08MindcrankMillsAnOpponentForTheLifeTheyLose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Mindcrank", "Artifact", effectstest.B08MindcrankOracle, false)
	lib, yard := opp.Library.Size(), opp.Graveyard.Size()

	effectstest.B08CastBolt(t, g, opp.ID, false)
	if opp.Graveyard.Size()-yard != 3 || lib-opp.Library.Size() != 3 {
		t.Fatalf("a Bolt: milled %d, want 3", opp.Graveyard.Size()-yard)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size()-yard != 5 {
		t.Errorf("a 2-life drain: milled %d in total, want 5", opp.Graveyard.Size()-yard)
	}
	// Your own loss is not an opponent's.
	mine := me.Graveyard.Size()
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -4) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != mine {
		t.Error("the controller's own life loss milled them")
	}
}

func TestB08RazorkinNeedleheadStrikesFirstOnYourTurnAndPingsDrawers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	needle := b08Permanent(g, me.ID, "Razorkin Needlehead", "Creature — Human Assassin", effectstest.B08RazorkinNeedleheadOracle)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, needle), "first strike") {
		t.Error("on its controller's turn the Needlehead has first strike")
	}
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("two opposing draws: %d → %d, want -2", before, opp.Life)
	}
	meBefore := me.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != meBefore {
		t.Error("your own draw is not an opponent's")
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, needle), "first strike") {
		t.Error("on an opponent's turn the Needlehead has no first strike")
	}
}

// --- combat damage -------------------------------------------------

func TestB08OldGnawboneMakesATreasurePerPointPerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	gnawbone := effectstest.PushCatalogPermanent(g, me.ID, "Old Gnawbone", "Legendary Creature — Dragon", effectstest.B08OldGnawboneOracle, false)
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, gnawbone), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	three := effectstest.PushVanillaCreature(g, me.ID, "Three", 3, 3)
	two := effectstest.PushVanillaCreature(g, me.ID, "Two", 2, 2)

	effectstest.AttackWith(t, g, victim.ID, three, two)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); got != 5 {
		t.Errorf("%d Treasures, want 5 (3 + 2, one trigger per creature)", got)
	}
}

func TestB08ThopterSpyNetworkDrawsOncePerCombatForArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Thopter Spy Network", "Enchantment", effectstest.B08ThopterSpyNetworkOracle, false)
	myr1 := b08TypedCreature(g, me.ID, "Myr", "Artifact Creature — Myr", 1, 1)
	myr2 := b08TypedCreature(g, me.ID, "Myr", "Artifact Creature — Myr", 1, 1)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := me.Hand.Size()

	effectstest.AttackWith(t, g, victim.ID, myr1, myr2, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want exactly 1 — 'one or more' is one trigger", got)
	}
}

func TestB08ThopterSpyNetworkIgnoresNonArtifactAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Thopter Spy Network", "Enchantment", effectstest.B08ThopterSpyNetworkOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := me.Hand.Size()
	effectstest.AttackWith(t, g, victim.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("drew %d, want 0 — a Bear is not an artifact creature", got)
	}
}

func TestB08ThopterSpyNetworkMakesAThopterEachUpkeepWhileYouControlAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.PushCatalogPermanent(g, owner.ID, "Thopter Spy Network", "Enchantment", effectstest.B08ThopterSpyNetworkOracle, false)

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, owner.ID, "Thopter") != 0 {
		t.Fatal("with no artifact the upkeep trigger does not fire")
	}

	effectstest.SeedPermanentFor(g, owner.ID, "Rock", "Artifact")
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, owner.ID, "Thopter") != 1 {
		t.Fatal("with an artifact the upkeep makes a Thopter")
	}
	thopter := effectstest.FindBattlefieldByName(g, "Thopter")
	if card, _ := effectstest.BattlefieldCard(g, thopter); !card.IsArtifact() || !card.IsCreature() {
		t.Error("the Thopter is an artifact creature")
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, thopter), "flying") {
		t.Error("the Thopter flies")
	}
}

// --- end step and upkeep -------------------------------------------

func TestB08WildernessReclamationUntapsYourLandsAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Wilderness Reclamation", "Enchantment", effectstest.B08WildernessReclamationOracle, false)
	mine := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	mine2 := effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	theirs := effectstest.SeedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	for _, id := range []uuid.UUID{mine, mine2, theirs} {
		b08Tap(g, id)
	}

	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, mine2} {
		if card, _ := effectstest.BattlefieldCard(g, id); card.Tapped {
			t.Errorf("%s should be untapped at your end step", card.Name)
		}
	}
	if card, _ := effectstest.BattlefieldCard(g, theirs); !card.Tapped {
		t.Error("an opponent's land is not 'lands you control'")
	}
}

func TestB08EmeriaReturnsACreatureCardWithSevenPlains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.PushCatalogPermanent(g, owner.ID, "Emeria, the Sky Ruin", "Land", effectstest.B08EmeriaOracle, false)
	for i := 0; i < 6; i++ {
		effectstest.SeedLandOnBattlefield(g, owner.ID, "Plains", "Basic Land — Plains")
	}
	dead := effectstest.PushGraveyardCardForTest(owner, "Dead Bear")
	relic := effectstest.Batch01GraveyardCard(owner, "Sol Ring", "Artifact")

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if len(g.PendingChoices) != 0 {
		t.Fatal("six Plains is not seven — no prompt")
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.SeedLandOnBattlefield(g, owner.ID, "Savannah", "Land — Forest Plains")
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.B04WaitForPick(t, g, owner.ID)
	pick := effectstest.LatestPickTarget(g, owner.ID)
	if effectstest.HasID(pick.PickTargetCards, relic) {
		t.Error("an artifact card was offered for 'target creature card'")
	}
	effectstest.PickCard(t, g, owner.ID, dead)
	effectstest.AnswerLatestTriggerPrompt(t, g, owner.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, dead)
	if !ok {
		t.Fatal("the creature card did not return to the battlefield")
	}
	if card.Controller != owner.ID {
		t.Error("it returns under its owner's control")
	}
}

// --- statics -------------------------------------------------------

func TestB08CoatOfArmsPumpsEveryCreatureForSharedTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warrior := b08TypedCreature(g, me.ID, "Goblin Warrior", "Creature — Goblin Warrior", 2, 2)
	shaman := b08TypedCreature(g, me.ID, "Goblin Shaman", "Creature — Goblin Shaman", 2, 2)
	elf := b08TypedCreature(g, me.ID, "Elf", "Creature — Elf", 2, 2)
	theirs := b08TypedCreature(g, opp.ID, "Their Warrior", "Creature — Human Warrior", 2, 2)
	b08Permanent(g, me.ID, "Coat of Arms", "Artifact", effectstest.B08CoatOfArmsOracle)

	for _, tc := range []struct {
		id   uuid.UUID
		want int
		why  string
	}{
		{warrior, 4, "the Goblin Warrior shares a type with the Shaman and with the Human Warrior"},
		{shaman, 3, "the Goblin Shaman shares only Goblin"},
		{elf, 2, "the Elf shares nothing"},
		{theirs, 3, "an opponent's Warrior is pumped too — the card is symmetrical"},
	} {
		if p, tough := effectstest.EffectivePower(t, g, tc.id), effectstest.EffectiveToughness(t, g, tc.id); p != tc.want || tough != tc.want {
			t.Errorf("%s: %d/%d, want %d/%d", tc.why, p, tough, tc.want, tc.want)
		}
	}

	// A second Coat doubles the bonus.
	b08Permanent(g, opp.ID, "Coat of Arms", "Artifact", effectstest.B08CoatOfArmsOracle)
	if p := effectstest.EffectivePower(t, g, warrior); p != 6 {
		t.Errorf("two Coats: the Goblin Warrior is %d, want 6", p)
	}
}

func TestB08TorbranAddsTwoToRedDamageAtOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Torbran, Thane of Red Fell", "Legendary Creature — Dwarf Noble", effectstest.B08TorbranOracle, false)
	oppBefore, meBefore := opp.Life, me.Life

	effectstest.B08CastBolt(t, g, opp.ID, true)
	if opp.Life != oppBefore-5 {
		t.Errorf("a red Bolt at an opponent: %d → %d, want -5", oppBefore, opp.Life)
	}
	effectstest.B08CastBolt(t, g, me.ID, true)
	if me.Life != meBefore-3 {
		t.Errorf("a red Bolt at yourself: %d → %d, want -3 (not an opponent)", meBefore, me.Life)
	}
	effectstest.B08CastBolt(t, g, opp.ID, false)
	if opp.Life != oppBefore-8 {
		t.Errorf("a colourless source: %d → %d, want -8 in total (no bonus)", oppBefore, opp.Life)
	}

	// A red creature's combat damage is red damage.
	theirs := effectstest.SeedCreature(g, "Blocker", opp.ID)
	goblin := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Creature — Goblin", Colors: []string{"R"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	_ = theirs
	effectstest.AttackWith(t, g, opp.ID, goblin)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppBefore-11 {
		t.Errorf("a red 1/1 connecting: %d → %d, want -11 in total (1 + 2)", oppBefore, opp.Life)
	}
}

// --- activated abilities -------------------------------------------

func TestB08ZuranOrbEatsALandForTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	orb := effectstest.PushCatalogPermanent(g, me.ID, "Zuran Orb", "Artifact", effectstest.B08ZuranOrbOracle, false)
	forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	before := me.Life

	if err := g.ActivateCatalogAbility(me.ID, orb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
	}); err == nil {
		t.Fatal("a creature was accepted for 'sacrifice a land'")
	}
	if err := g.ActivateCatalogAbility(me.ID, orb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{forest},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(forest) || !me.Graveyard.Contains(forest) {
		t.Error("the land is sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d → %d, want +2", before, me.Life)
	}
}

func TestB08PhyrexianReclamationBuysBackACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rec := effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Reclamation", "Enchantment", effectstest.B08PhyrexianReclamationOracle, false)
	dead := effectstest.PushGraveyardCardForTest(me, "Dead Bear")
	relic := effectstest.Batch01GraveyardCard(me, "Sol Ring", "Artifact")
	before := me.Life
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, rec, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}},
	}); err == nil {
		t.Fatal("an artifact card was accepted for 'target creature card'")
	}
	if err := g.ActivateCatalogAbility(me.ID, rec, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != before-2 {
		t.Errorf("life %d → %d, want -2 paid at announce", before, me.Life)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not come back to hand")
	}
}

func TestB08GavonyTownshipGrowsYourTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	township := effectstest.PushCatalogPermanent(g, me.ID, "Gavony Township", "Land", effectstest.B08GavonyTownshipOracle, false)
	mine := effectstest.SeedCreature(g, "Mine", me.ID)
	mine2 := effectstest.SeedCreature(g, "Mine Too", me.ID)
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "W"},
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, township, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
	}{{mine, 1}, {mine2, 1}, {theirs, 0}} {
		if got := effectstest.CounterCount(g, tc.id, "+1/+1"); got != tc.want {
			t.Errorf("%s: %d counters, want %d", tc.id, got, tc.want)
		}
	}
	if card, _ := effectstest.BattlefieldCard(g, township); !card.Tapped {
		t.Error("the ability has a tap cost")
	}
}

func TestB08FountainportThreeAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	port := effectstest.PushCatalogPermanent(g, me.ID, "Fountainport", "Land", effectstest.B08FountainportOracle, false)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)

	// {4}, {T}: a Treasure.
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, port, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the Treasure ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Fatal("want one Treasure")
	}
	treasure := effectstest.FindBattlefieldByName(g, "Treasure")

	// {3}, {T}, Pay 1 life: a Fish.
	effectstest.B08Untap(g, port)
	before := me.Life
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, port, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the Fish ability: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1 paid at announce", before, me.Life)
	}
	effectstest.PassPriorityAroundTable(t, g)
	fish := effectstest.FindBattlefieldByName(g, "Fish")
	if card, ok := effectstest.BattlefieldCard(g, fish); !ok || !card.IsCreature() || !card.HasColor("U") {
		t.Error("want a blue Fish creature token")
	}

	// {2}, {T}, Sacrifice a token: a card. A nontoken creature is refused.
	effectstest.B08Untap(g, port)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, port, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
	}); err == nil {
		t.Fatal("a nontoken creature was accepted for 'sacrifice a token'")
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, port, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{treasure},
	}); err != nil {
		t.Fatalf("the draw ability: %v", err)
	}
	if g.Battlefield.Contains(treasure) {
		t.Error("the token is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestB08BlightedWoodlandFetchesUpToTwoBasicsTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	woodland := effectstest.PushCatalogPermanent(g, me.ID, "Blighted Woodland", "Land", effectstest.B08BlightedWoodlandOracle, false)
	forest := effectstest.B07LibraryLand(me, "Forest", "Basic Land — Forest")
	island := effectstest.B07LibraryLand(me, "Island", "Basic Land — Island")
	effectstest.B07LibraryLand(me, "Swamp", "Basic Land — Swamp")
	coffers := effectstest.B07LibraryLand(me, "Cabal Coffers", "Land")
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, woodland, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(woodland) {
		t.Error("the Woodland is sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("want a search prompt")
	}
	if effectstest.HasID(c.SearchCards, coffers) {
		t.Error("a nonbasic was offered for 'basic land cards'")
	}
	if c.SearchMax != 2 {
		t.Errorf("may take %d, want up to two", c.SearchMax)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, forest, island)
	for _, id := range []uuid.UUID{forest, island} {
		card, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Errorf("%s did not reach the battlefield", id)
			continue
		}
		if !card.Tapped {
			t.Errorf("%s entered untapped, want tapped", card.Name)
		}
	}
}

func TestB08KrosanVergeFetchesAForestThenAPlains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	verge := effectstest.PlayLandFromHand(t, g, "Krosan Verge", effectstest.B08KrosanVergeOracle)
	effectstest.Top100AssertEnteredTapped(t, g, verge, "Krosan Verge")
	effectstest.B08Untap(g, verge)
	savannah := effectstest.B07LibraryLand(me, "Savannah", "Land — Forest Plains")
	effectstest.B07LibraryLand(me, "Forest", "Basic Land — Forest")
	plains := effectstest.B07LibraryLand(me, "Plains", "Basic Land — Plains")
	effectstest.B07LibraryLand(me, "Plains", "Basic Land — Plains")
	island := effectstest.B07LibraryLand(me, "Island", "Basic Land — Island")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, verge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(verge) {
		t.Error("the Verge is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)

	// First the Forest search: a Savannah is a Forest card.
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("want the Forest search prompt")
	}
	if effectstest.HasID(c.SearchCards, island) || !effectstest.HasID(c.SearchCards, savannah) {
		t.Error("the Forest search offers Forests only, dual lands included")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, savannah)

	// Then the Plains search, from the same resolution.
	c = effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("want the Plains search prompt after the Forest one")
	}
	if effectstest.HasID(c.SearchCards, island) || !effectstest.HasID(c.SearchCards, plains) {
		t.Error("the Plains search offers Plains only")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, plains)
	for _, id := range []uuid.UUID{savannah, plains} {
		card, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Errorf("%s did not reach the battlefield", id)
			continue
		}
		if !card.Tapped {
			t.Errorf("%s entered untapped, want tapped", card.Name)
		}
	}
	if effectstest.B07OpenSearchPrompts(g) != 0 {
		t.Error("nothing should be left open")
	}
}

// --- lands ---------------------------------------------------------

func TestB08BloodfellCavesEntersTappedAndGainsALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	caves := effectstest.PlayLandFromHand(t, g, "Bloodfell Caves", effectstest.B08BloodfellCavesOracle)
	effectstest.Top100AssertEnteredTapped(t, g, caves, "Bloodfell Caves")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
	effectstest.B08Untap(g, caves)
	if err := g.ActivateManaAbility(me.ID, caves, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("the Caves ask B or R, got %+v", pick)
	}
}

func TestB08CabarettiCourtyardSacrificesItselfForABasicAndALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mountain := effectstest.B07LibraryLand(me, "Mountain", "Basic Land — Mountain")
	effectstest.B07LibraryLand(me, "Plains", "Basic Land — Plains")
	swamp := effectstest.B07LibraryLand(me, "Swamp", "Basic Land — Swamp")
	before := me.Life

	courtyard := effectstest.PlayLandFromHand(t, g, "Cabaretti Courtyard", effectstest.B08CabarettiCourtyardOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(courtyard) || !me.Graveyard.Contains(courtyard) {
		t.Fatal("the Courtyard sacrifices itself")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("want a search prompt")
	}
	if effectstest.HasID(c.SearchCards, swamp) {
		t.Error("a Swamp was offered for 'a basic Mountain, Forest, or Plains card'")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, mountain)
	card, ok := effectstest.BattlefieldCard(g, mountain)
	if !ok || !card.Tapped {
		t.Error("the basic should be on the battlefield tapped")
	}
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
}

func TestB08MinesOfMoriaEntersUntappedWithALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tapped := effectstest.PlayLandFromHand(t, g, "Mines of Moria", effectstest.B08MinesOfMoriaOracle)
	effectstest.Top100AssertEnteredTapped(t, g, tapped, "Mines of Moria (no legend)")

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	b08TypedCreature(g2, me2.ID, "Legend", "Legendary Creature — Human", 2, 2)
	untapped := effectstest.PlayLandFromHand(t, g2, "Mines of Moria", effectstest.B08MinesOfMoriaOracle)
	effectstest.Top100AssertEnteredUntapped(t, g2, untapped, "Mines of Moria (with a legend)")
	if err := g2.ActivateManaAbility(me2.ID, untapped, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me2); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
	_ = me
}

func TestB08SunkenRuinsFiltersOneHybridIntoTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ruins := effectstest.SeedPermanentWithOracle(g, me.ID, "Sunken Ruins", "Land", effectstest.B08SunkenRuinsOracle)

	if err := g.ActivateManaAbility(me.ID, ruins, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the {C} half: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}

	fresh := effectstest.SeedPermanentWithOracle(g, me.ID, "Sunken Ruins", "Land", effectstest.B08SunkenRuinsOracle)
	me.ManaPool.EmptyPool()
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("{U/B} was paid with {G}")
	}
	if card, _ := effectstest.BattlefieldCard(g, fresh); card.Tapped {
		t.Fatal("a refused activation must not tap the Ruins")
	}

	me.ManaPool.EmptyPool()
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	// #3939: one choice, asked once per pick.
	picks := 0
	for c := effectstest.RiderLatestManaPick(g, me.ID); c != nil && picks < 8; c = effectstest.RiderLatestManaPick(g, me.ID) {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
			picks++
			if len(c.ColorOptions) != 2 {
				t.Errorf("each slot must offer exactly U and B, got %v", c.ColorOptions)
			}
			if err := g.ResolveManaChoice(c.ID, me.ID, "U"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
		}
	}
	if picks != 2 {
		t.Fatalf("want two colour picks, got %d", picks)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "U" || got[1] != "U" {
		t.Errorf("pool %v, want [U U]", got)
	}
}
