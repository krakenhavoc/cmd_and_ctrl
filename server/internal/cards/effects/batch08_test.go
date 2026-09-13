package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch08_test.go — card-level coverage for the card-coverage
// roadmap's batch 08 (#301, `edhrec_rank` 902–1004): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b08-prefixed.

const (
	b08ResculptOracle              = "814e321d-16fd-4f7b-a8d8-2b089be76f2c"
	b08WarleadersCallOracle        = "a751c07b-fc21-4854-96e7-f71abf4e86c9"
	b08ZuranOrbOracle              = "08cb8a30-9cb4-4517-bee5-8848aa60d1a2"
	b08SpiritedCompanionOracle     = "9c5f0d91-9d86-4e66-94fd-4af93ad01838"
	b08BlightedWoodlandOracle      = "02679a2e-303d-412f-87d8-0a37a8ca259c"
	b08BristlyBillOracle           = "d3b2d8a2-d3bc-448c-9cf6-6bead6010c28"
	b08OldGnawboneOracle           = "dff3f4c7-f792-4ca3-9b0a-0738e70664d9"
	b08DispelOracle                = "6d7be242-a072-40ce-b540-95880506cccd"
	b08FountainportOracle          = "94e8b0a9-44a1-4dce-8d44-78681ae638a1"
	b08SolveTheEquationOracle      = "f02682f0-26c0-4032-9df3-273b6a45d0a8"
	b08BloodfellCavesOracle        = "64e29bfc-9313-4e8c-808c-bc27f6b018a6"
	b08ThopterSpyNetworkOracle     = "49be65fd-3755-410d-b0dc-2e5861ea2552"
	b08WildernessReclamationOracle = "6f856f99-4cb4-479d-958d-964220965ed6"
	b08CoatOfArmsOracle            = "5f7f133e-58ea-41ab-b1be-be4b400fac4c"
	b08OmnathLocusOfRageOracle     = "1816eede-c5bd-49df-958f-a3af64cb2932"
	b08PhyrexianReclamationOracle  = "647ca69e-cc01-4b2b-b376-bee2a98331e8"
	b08CruelCelebrantOracle        = "3ee78cfc-0e9e-4737-a7e2-b42f94228040"
	b08ArchaeomancerOracle         = "a91a3266-cadd-47a0-9b20-160307f14c07"
	b08MindcrankOracle             = "c73c1d91-0163-49c6-832a-b9327e7a2c9b"
	b08CabarettiCourtyardOracle    = "65424bea-fd53-4f85-9757-0b91a6d40ba4"
	b08RazorkinNeedleheadOracle    = "a78f981a-bf8a-42a4-b171-d655cc2cc1a2"
	b08KrosanVergeOracle           = "d9a10971-f32b-4978-952d-fed0a5bc9e36"
	b08MoldervineReclamationOracle = "68639a3d-2192-4921-8298-c76bb0cd6b02"
	b08SythisOracle                = "0fc64fd6-f057-4056-9dca-47accb7ff036"
	b08EnchantresssPresenceOracle  = "795b096a-2bce-4588-a2c9-abc5ea40dc0c"
	b08SunkenRuinsOracle           = "e6415ffb-8b7a-41c3-bedf-0d4112b7b795"
	b08EmeriaOracle                = "cc999cf2-c99b-4911-8c52-6cc4a99fcc7b"
	b08NoxiousGearhulkOracle       = "a77b5be2-f361-4135-ba25-670a74d268ac"
	b08CutADealOracle              = "8b5cffd5-5db3-4151-9275-d91387554412"
	b08GavonyTownshipOracle        = "8a44e4e7-dfa2-427b-bbff-11c398fa60bb"
	b08AllIsDustOracle             = "14693689-d087-43b6-9c3f-63ab0648fc20"
	b08MinesOfMoriaOracle          = "583cdebe-0195-45be-bd2e-5765f07cb902"
	b08TorbranOracle               = "8c3495bf-02e7-4ad9-949d-92eb3d2b662a"
	b08ManagorgerHydraOracle       = "b3f2265b-dd65-4b74-8b74-35ee0b147617"
)

// b08CastBolt casts a Lightning Bolt at a player and settles it. With
// `red` the card carries its printed colour, as the deck importer
// stamps it; without, the fixture is colourless — which is what
// separates "a red source" from "a source".
func b08CastBolt(t *testing.T, g *game.Game, target uuid.UUID, red bool) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	card := game.Card{
		InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: lightningBoltOracle,
		Owner: active.ID, Controller: active.ID,
	}
	if red {
		card.Colors = []string{"R"}
		card.ManaCost = "{R}"
	}
	active.Hand.PushTop(card)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}},
	}); err != nil {
		t.Fatalf("CastSpell Lightning Bolt: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// b08Tap taps a battlefield permanent through the effect API.
func b08Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
}

// b08Untap untaps a battlefield permanent through the effect API —
// a land that has used its tap cost once, so a second activation can
// be tested in the same turn.
func b08Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
}

// b08Permanent pushes a catalog permanent with a real entry event, so
// the layer engine sees it — pushCatalogPermanent emits none, which
// is fine for triggers and abilities but leaves a static invisible
// until something else invalidates the cache.
func b08Permanent(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	})
}

// b08TypedCreature pushes a live creature with the given type line
// under owner, no summoning sickness, with a real ETB timestamp.
func b08TypedCreature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
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
		b08ResculptOracle:              "Resculpt",
		b08WarleadersCallOracle:        "Warleader's Call",
		b08ZuranOrbOracle:              "Zuran Orb",
		b08SpiritedCompanionOracle:     "Spirited Companion",
		b08BlightedWoodlandOracle:      "Blighted Woodland",
		b08BristlyBillOracle:           "Bristly Bill, Spine Sower",
		b08OldGnawboneOracle:           "Old Gnawbone",
		b08DispelOracle:                "Dispel",
		b08FountainportOracle:          "Fountainport",
		b08SolveTheEquationOracle:      "Solve the Equation",
		b08BloodfellCavesOracle:        "Bloodfell Caves",
		b08ThopterSpyNetworkOracle:     "Thopter Spy Network",
		b08WildernessReclamationOracle: "Wilderness Reclamation",
		b08CoatOfArmsOracle:            "Coat of Arms",
		b08OmnathLocusOfRageOracle:     "Omnath, Locus of Rage",
		b08PhyrexianReclamationOracle:  "Phyrexian Reclamation",
		b08CruelCelebrantOracle:        "Cruel Celebrant",
		b08ArchaeomancerOracle:         "Archaeomancer",
		b08MindcrankOracle:             "Mindcrank",
		b08CabarettiCourtyardOracle:    "Cabaretti Courtyard",
		b08RazorkinNeedleheadOracle:    "Razorkin Needlehead",
		b08KrosanVergeOracle:           "Krosan Verge",
		b08MoldervineReclamationOracle: "Moldervine Reclamation",
		b08SythisOracle:                "Sythis, Harvest's Hand",
		b08EnchantresssPresenceOracle:  "Enchantress's Presence",
		b08SunkenRuinsOracle:           "Sunken Ruins",
		b08EmeriaOracle:                "Emeria, the Sky Ruin",
		b08NoxiousGearhulkOracle:       "Noxious Gearhulk",
		b08CutADealOracle:              "Cut a Deal",
		b08GavonyTownshipOracle:        "Gavony Township",
		b08AllIsDustOracle:             "All Is Dust",
		b08MinesOfMoriaOracle:          "Mines of Moria",
		b08TorbranOracle:               "Torbran, Thane of Red Fell",
		b08ManagorgerHydraOracle:       "Managorger Hydra",
	}
	if len(want) != 34 {
		t.Fatalf("the batch is 34 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness == CompletenessUnreviewed {
			t.Errorf("%s ships without a completeness declaration", name)
		}
	}
}

// --- spells --------------------------------------------------------

func TestB08ResculptExilesAnArtifactAndHandsItsControllerAnElemental(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")
	castCatalogSpell(t, g, "Resculpt", "Instant", b08ResculptOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) {
		t.Error("the artifact should be in exile")
	}
	if countBattlefieldNamed(g, opp.ID, "Elemental") != 1 {
		t.Fatal("the VICTIM should get the 4/4 Elemental")
	}
	if countBattlefieldNamed(g, me.ID, "Elemental") != 0 {
		t.Error("the caster must not get the token")
	}
	elemental := findBattlefieldByName(g, "Elemental")
	card, _ := battlefieldCard(g, elemental)
	if !card.HasColor("U") || !card.HasColor("R") || card.Power != 4 || card.Toughness != 4 {
		t.Errorf("the token is a 4/4 blue and red Elemental, got %d/%d %v", card.Power, card.Toughness, card.Colors)
	}
	// A land is neither an artifact nor a creature.
	land := seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	if !b03CastRefused(t, g, "Resculpt", "Instant", b08ResculptOracle, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}) {
		t.Error("a land was accepted for 'target artifact or creature'")
	}
}

func TestB08DispelCountersOnlyAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := batch01OpponentCasts(t, g, opp, "Their Instant", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	castCatalogSpell(t, g, "Dispel", "Instant", b08DispelOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the instant was not countered")
	}

	// A creature spell is not an instant spell.
	bear := castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if !b08OpponentCastRefused(t, g, opp, "Dispel", b08DispelOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}) {
		t.Error("a creature spell was accepted for 'target instant spell'")
	}
	passPriorityAroundTable(t, g)
}

func TestB08SolveTheEquationTutorsAnInstantOrSorceryToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bolt := pushLibraryCardForTest(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant"})
	ponder := pushLibraryCardForTest(me, game.Card{Name: "Ponder", TypeLine: "Sorcery"})
	bear := pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})

	castCatalogSpell(t, g, "Solve the Equation", "Sorcery", b08SolveTheEquationOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("want a search prompt")
	}
	for _, id := range c.SearchCards {
		if id == bear {
			t.Error("a creature card was offered for 'an instant or sorcery card'")
		}
	}
	if !hasID(c.SearchCards, bolt) || !hasID(c.SearchCards, ponder) {
		t.Error("both the instant and the sorcery should be offered")
	}
	answerSearchByID(t, g, me.ID, ponder)
	if !me.Hand.Contains(ponder) {
		t.Error("the chosen sorcery did not reach the hand")
	}
}

func TestB08CutADealDrawsForEachOpponentWhoDrew(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Hand.Size()
	}
	castCatalogSpell(t, g, "Cut a Deal", "Sorcery", b08CutADealOracle, nil)
	passPriorityAroundTable(t, g)
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
	castCatalogSpell(t, g, "Cut a Deal", "Sorcery", b08CutADealOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - mid; got != 2 {
		t.Errorf("the caster drew %d, want 2 (one opponent could not draw)", got)
	}
}

func TestB08AllIsDustSacrificesOnlyColouredPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := seedPermanentFor(g, me.ID, "Rock", "Artifact")
	eldrazi := pushVanillaCreature(g, me.ID, "Eldrazi", 5, 5)
	theirLand := seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	green := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	glory := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glory", TypeLine: "Enchantment", Colors: []string{"W"},
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1) })
	passPriorityAroundTable(t, g)

	castCatalogSpell(t, g, "All Is Dust", "Kindred Sorcery — Eldrazi", b08AllIsDustOracle, nil)
	passPriorityAroundTable(t, g)

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
	if countBattlefieldNamed(g, opp.ID, "Goblin") != 0 {
		t.Error("a red token is sacrificed")
	}
}

// --- ETB and cast triggers -----------------------------------------

func TestB08SpiritedCompanionDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Spirited Companion", "Enchantment Creature — Dog", b08SpiritedCompanionOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestB08ArchaeomancerReturnsAnInstantOrSorceryCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bolt := batch01GraveyardCard(me, "Lightning Bolt", "Instant")
	bear := pushGraveyardCardForTest(me, "Dead Bear")

	castCatalogSpell(t, g, "Archaeomancer", "Creature — Human Wizard", b08ArchaeomancerOracle, nil)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if hasID(pick.PickTargetCards, bear) {
		t.Error("a creature card was offered for 'target instant or sorcery card'")
	}
	pickCard(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant did not come back to hand")
	}
}

func TestB08SythisGainsAndDrawsOnAnEnchantmentSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Sythis, Harvest's Hand", "Legendary Enchantment Creature — Nymph", b08SythisOracle, false)
	hand, life := me.Hand.Size(), me.Life

	castCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}

	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("a creature spell is not an enchantment spell: drew %d, want still 1", got)
	}
}

func TestB08EnchantresssPresenceDrawsWithoutAsking(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Enchantress's Presence", "Enchantment", b08EnchantresssPresenceOracle, false)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	if len(g.PendingChoices) != 0 {
		t.Fatal("the Presence is mandatory — no prompt")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestB08ManagorgerHydraGrowsOnEveryPlayersSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := pushCatalogPermanent(g, me.ID, "Managorger Hydra", "Creature — Hydra", b08ManagorgerHydraOracle, false)
	if !eotHasAbility(effectiveAbilities(t, g, hydra), "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}

	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, hydra, "+1/+1"); got != 1 {
		t.Fatalf("after my spell: %d counters, want 1", got)
	}
	batch01OpponentCasts(t, g, opp, "Their Instant", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, hydra, "+1/+1"); got != 2 {
		t.Errorf("after an opponent's spell: %d counters, want 2", got)
	}
}

func TestB08WarleadersCallPumpsAndPings(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Warleader's Call", "Enchantment", b08WarleadersCallOracle, false)
	mine := seedCreature(g, "Mine", me.ID)
	theirs := seedCreature(g, "Theirs", opp.ID)
	if p, tough := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("my 2/2 = %d/%d, want 3/3", p, tough)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("their 2/2 = %d, want the printed 2", p)
	}

	before := lifeOfOpponents(g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 2) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d → %d, want -2 (two creatures entered)", i+1, b, got)
		}
	}
	// An opponent's creature and a noncreature of yours are silent.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1)
		_ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1)
	})
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d → %d, want still -2", i+1, b, got)
		}
	}
}

func TestB08NoxiousGearhulkDestroysAnotherCreatureAndGainsItsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b04Creature(g, opp.ID, "Their Big", 3, 5)
	before := me.Life

	hulk := castCatalogSpell(t, g, "Noxious Gearhulk", "Artifact Creature — Construct", b08NoxiousGearhulkOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if hasID(pick.PickTargetCards, hulk) {
		t.Error("the Gearhulk offered itself for 'another target creature'")
	}
	pickCard(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Fatal("the targeted creature should be destroyed")
	}
	if me.Life != before+5 {
		t.Errorf("life %d → %d, want +5 (its toughness)", before, me.Life)
	}
	if !eotHasAbility(effectiveAbilities(t, g, hulk), "menace") {
		t.Error("printed menace did not reach the effective abilities")
	}
}

func TestB08NoxiousGearhulkGainsNothingForAnIndestructibleCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tough := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Darksteel", TypeLine: "Artifact Creature — Golem",
		Power: 4, Toughness: 4, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	before := me.Life
	castCatalogSpell(t, g, "Noxious Gearhulk", "Artifact Creature — Construct", b08NoxiousGearhulkOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, tough)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tough) {
		t.Fatal("an indestructible creature survives")
	}
	if me.Life != before {
		t.Errorf("life %d → %d, want unchanged — nothing was destroyed this way", before, me.Life)
	}
}

// --- landfall ------------------------------------------------------

func TestB08BristlyBillGrowsATargetOnLandfallAndDoublesOnDemand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bill := pushCatalogPermanent(g, me.ID, "Bristly Bill, Spine Sower", "Legendary Creature — Plant Druid", b08BristlyBillOracle, false)
	bear := seedCreature(g, "Bear", me.ID)
	theirs := seedCreature(g, "Theirs", opp.ID)

	playLandFromHand(t, g, "Forest", "")
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if !hasID(pick.PickTargetCards, theirs) {
		t.Error("'target creature' includes an opponent's creature")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, "+1/+1"); got != 1 {
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
	passPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
		why  string
	}{
		{bear, 2, "one counter doubles to two"},
		{bill, 6, "three double to six"},
		{theirs, 2, "an opponent's creature is untouched"},
	} {
		if got := counterCount(g, tc.id, "+1/+1"); got != tc.want {
			t.Errorf("%s: %d counters, want %d", tc.why, got, tc.want)
		}
	}
}

func TestB08OmnathMakesElementalsOnLandfallAndBoltsWhenTheyDie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	omnath := pushCatalogPermanent(g, me.ID, "Omnath, Locus of Rage", "Legendary Creature — Elemental", b08OmnathLocusOfRageOracle, false)

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Elemental") != 1 {
		t.Fatal("a land drop should make one Elemental")
	}
	token := findBattlefieldByName(g, "Elemental")
	if card, _ := battlefieldCard(g, token); card.Power != 5 || !card.HasColor("R") || !card.HasColor("G") {
		t.Errorf("the token is a 5/5 red and green Elemental, got %d/%d %v", card.Power, card.Toughness, card.Colors)
	}

	before := opp.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(token) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("an Elemental dying: %d → %d, want -3", before, opp.Life)
	}

	// A non-Elemental of yours and an opponent's Elemental are silent.
	bear := seedCreature(g, "Bear", me.ID)
	theirs := b08TypedCreature(g, opp.ID, "Their Elemental", "Creature — Elemental", 3, 3)
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(bear)
		_ = g.SacrificePermanentForEffect(theirs)
	})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("neither death should trigger Omnath")
	}

	// Omnath himself dying is the first clause.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(omnath) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("Omnath dying: %d → %d, want -6 in total", before, opp.Life)
	}
}

// --- dies and life-loss triggers -----------------------------------

func TestB08CruelCelebrantDrainsOnYourCreaturesPlaneswalkersAndItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	celebrant := pushCatalogPermanent(g, me.ID, "Cruel Celebrant", "Creature — Vampire", b08CruelCelebrantOracle, false)
	bear := seedCreature(g, "Bear", me.ID)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID,
	})
	theirs := seedCreature(g, "Theirs", opp.ID)
	bounced := seedCreature(g, "Bounced", me.ID)
	before := lifeOfOpponents(g)
	meBefore := me.Life

	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(bear)
		_ = g.SacrificePermanentForEffect(walker)
		_ = g.SacrificePermanentForEffect(theirs)
		_ = g.BounceToHandForEffect(bounced)
	})
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d → %d, want -2 (a creature and a planeswalker of yours died)", i+1, b, got)
		}
	}
	if me.Life != meBefore+2 {
		t.Errorf("controller %d → %d, want +2", meBefore, me.Life)
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(celebrant) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-3 {
			t.Errorf("opponent %d after the Celebrant's own death: %d → %d, want -3", i+1, b, got)
		}
	}
}

func TestB08MoldervineReclamationPaysOutOnYourCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Moldervine Reclamation", "Enchantment", b08MoldervineReclamationOracle, false)
	bear := seedCreature(g, "Bear", me.ID)
	theirs := seedCreature(g, "Theirs", opp.ID)
	hand, life := me.Hand.Size(), me.Life

	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(bear)
		_ = g.SacrificePermanentForEffect(theirs)
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
}

func TestB08MindcrankMillsAnOpponentForTheLifeTheyLose(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Mindcrank", "Artifact", b08MindcrankOracle, false)
	lib, yard := opp.Library.Size(), opp.Graveyard.Size()

	b08CastBolt(t, g, opp.ID, false)
	if opp.Graveyard.Size()-yard != 3 || lib-opp.Library.Size() != 3 {
		t.Fatalf("a Bolt: milled %d, want 3", opp.Graveyard.Size()-yard)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2) })
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Size()-yard != 5 {
		t.Errorf("a 2-life drain: milled %d in total, want 5", opp.Graveyard.Size()-yard)
	}
	// Your own loss is not an opponent's.
	mine := me.Graveyard.Size()
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -4) })
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != mine {
		t.Error("the controller's own life loss milled them")
	}
}

func TestB08RazorkinNeedleheadStrikesFirstOnYourTurnAndPingsDrawers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	needle := b08Permanent(g, me.ID, "Razorkin Needlehead", "Creature — Human Assassin", b08RazorkinNeedleheadOracle)
	if !eotHasAbility(effectiveAbilities(t, g, needle), "first strike") {
		t.Error("on its controller's turn the Needlehead has first strike")
	}
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 2) })
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("two opposing draws: %d → %d, want -2", before, opp.Life)
	}
	meBefore := me.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if me.Life != meBefore {
		t.Error("your own draw is not an opponent's")
	}

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if eotHasAbility(effectiveAbilities(t, g, needle), "first strike") {
		t.Error("on an opponent's turn the Needlehead has no first strike")
	}
}

// --- combat damage -------------------------------------------------

func TestB08OldGnawboneMakesATreasurePerPointPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	gnawbone := pushCatalogPermanent(g, me.ID, "Old Gnawbone", "Legendary Creature — Dragon", b08OldGnawboneOracle, false)
	if !eotHasAbility(effectiveAbilities(t, g, gnawbone), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	three := pushVanillaCreature(g, me.ID, "Three", 3, 3)
	two := pushVanillaCreature(g, me.ID, "Two", 2, 2)

	attackWith(t, g, victim.ID, three, two)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Treasure"); got != 5 {
		t.Errorf("%d Treasures, want 5 (3 + 2, one trigger per creature)", got)
	}
}

func TestB08ThopterSpyNetworkDrawsOncePerCombatForArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Thopter Spy Network", "Enchantment", b08ThopterSpyNetworkOracle, false)
	myr1 := b08TypedCreature(g, me.ID, "Myr", "Artifact Creature — Myr", 1, 1)
	myr2 := b08TypedCreature(g, me.ID, "Myr", "Artifact Creature — Myr", 1, 1)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := me.Hand.Size()

	attackWith(t, g, victim.ID, myr1, myr2, bear)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want exactly 1 — 'one or more' is one trigger", got)
	}
}

func TestB08ThopterSpyNetworkIgnoresNonArtifactAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Thopter Spy Network", "Enchantment", b08ThopterSpyNetworkOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := me.Hand.Size()
	attackWith(t, g, victim.ID, bear)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("drew %d, want 0 — a Bear is not an artifact creature", got)
	}
}

func TestB08ThopterSpyNetworkMakesAThopterEachUpkeepWhileYouControlAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushCatalogPermanent(g, owner.ID, "Thopter Spy Network", "Enchantment", b08ThopterSpyNetworkOracle, false)

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, owner.ID, "Thopter") != 0 {
		t.Fatal("with no artifact the upkeep trigger does not fire")
	}

	seedPermanentFor(g, owner.ID, "Rock", "Artifact")
	advanceToUpkeepOf(t, g, 2)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, owner.ID, "Thopter") != 1 {
		t.Fatal("with an artifact the upkeep makes a Thopter")
	}
	thopter := findBattlefieldByName(g, "Thopter")
	if card, _ := battlefieldCard(g, thopter); !card.IsArtifact() || !card.IsCreature() {
		t.Error("the Thopter is an artifact creature")
	}
	if !eotHasAbility(effectiveAbilities(t, g, thopter), "flying") {
		t.Error("the Thopter flies")
	}
}

// --- end step and upkeep -------------------------------------------

func TestB08WildernessReclamationUntapsYourLandsAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Wilderness Reclamation", "Enchantment", b08WildernessReclamationOracle, false)
	mine := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	mine2 := seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	theirs := seedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	for _, id := range []uuid.UUID{mine, mine2, theirs} {
		b08Tap(g, id)
	}

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, mine2} {
		if card, _ := battlefieldCard(g, id); card.Tapped {
			t.Errorf("%s should be untapped at your end step", card.Name)
		}
	}
	if card, _ := battlefieldCard(g, theirs); !card.Tapped {
		t.Error("an opponent's land is not 'lands you control'")
	}
}

func TestB08EmeriaReturnsACreatureCardWithSevenPlains(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushCatalogPermanent(g, owner.ID, "Emeria, the Sky Ruin", "Land", b08EmeriaOracle, false)
	for i := 0; i < 6; i++ {
		seedLandOnBattlefield(g, owner.ID, "Plains", "Basic Land — Plains")
	}
	dead := pushGraveyardCardForTest(owner, "Dead Bear")
	relic := batch01GraveyardCard(owner, "Sol Ring", "Artifact")

	advanceToUpkeepOf(t, g, 1)
	if len(g.PendingChoices) != 0 {
		t.Fatal("six Plains is not seven — no prompt")
	}
	passPriorityAroundTable(t, g)

	seedLandOnBattlefield(g, owner.ID, "Savannah", "Land — Forest Plains")
	advanceToUpkeepOf(t, g, 2)
	advanceToUpkeepOf(t, g, 1)
	answerLatestTriggerPrompt(t, g, owner.ID, true)
	b04WaitForPick(t, g, owner.ID)
	pick := latestPickTarget(g, owner.ID)
	if hasID(pick.PickTargetCards, relic) {
		t.Error("an artifact card was offered for 'target creature card'")
	}
	pickCard(t, g, owner.ID, dead)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, dead)
	if !ok {
		t.Fatal("the creature card did not return to the battlefield")
	}
	if card.Controller != owner.ID {
		t.Error("it returns under its owner's control")
	}
}

// --- statics -------------------------------------------------------

func TestB08CoatOfArmsPumpsEveryCreatureForSharedTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warrior := b08TypedCreature(g, me.ID, "Goblin Warrior", "Creature — Goblin Warrior", 2, 2)
	shaman := b08TypedCreature(g, me.ID, "Goblin Shaman", "Creature — Goblin Shaman", 2, 2)
	elf := b08TypedCreature(g, me.ID, "Elf", "Creature — Elf", 2, 2)
	theirs := b08TypedCreature(g, opp.ID, "Their Warrior", "Creature — Human Warrior", 2, 2)
	b08Permanent(g, me.ID, "Coat of Arms", "Artifact", b08CoatOfArmsOracle)

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
		if p, tough := effectivePower(t, g, tc.id), effectiveToughness(t, g, tc.id); p != tc.want || tough != tc.want {
			t.Errorf("%s: %d/%d, want %d/%d", tc.why, p, tough, tc.want, tc.want)
		}
	}

	// A second Coat doubles the bonus.
	b08Permanent(g, opp.ID, "Coat of Arms", "Artifact", b08CoatOfArmsOracle)
	if p := effectivePower(t, g, warrior); p != 6 {
		t.Errorf("two Coats: the Goblin Warrior is %d, want 6", p)
	}
}

func TestB08TorbranAddsTwoToRedDamageAtOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Torbran, Thane of Red Fell", "Legendary Creature — Dwarf Noble", b08TorbranOracle, false)
	oppBefore, meBefore := opp.Life, me.Life

	b08CastBolt(t, g, opp.ID, true)
	if opp.Life != oppBefore-5 {
		t.Errorf("a red Bolt at an opponent: %d → %d, want -5", oppBefore, opp.Life)
	}
	b08CastBolt(t, g, me.ID, true)
	if me.Life != meBefore-3 {
		t.Errorf("a red Bolt at yourself: %d → %d, want -3 (not an opponent)", meBefore, me.Life)
	}
	b08CastBolt(t, g, opp.ID, false)
	if opp.Life != oppBefore-8 {
		t.Errorf("a colourless source: %d → %d, want -8 in total (no bonus)", oppBefore, opp.Life)
	}

	// A red creature's combat damage is red damage.
	theirs := seedCreature(g, "Blocker", opp.ID)
	goblin := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Creature — Goblin", Colors: []string{"R"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	_ = theirs
	attackWith(t, g, opp.ID, goblin)
	passPriorityAroundTable(t, g)
	if opp.Life != oppBefore-11 {
		t.Errorf("a red 1/1 connecting: %d → %d, want -11 in total (1 + 2)", oppBefore, opp.Life)
	}
}

// --- activated abilities -------------------------------------------

func TestB08ZuranOrbEatsALandForTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	orb := pushCatalogPermanent(g, me.ID, "Zuran Orb", "Artifact", b08ZuranOrbOracle, false)
	forest := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	bear := seedCreature(g, "Bear", me.ID)
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
	passPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d → %d, want +2", before, me.Life)
	}
}

func TestB08PhyrexianReclamationBuysBackACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rec := pushCatalogPermanent(g, me.ID, "Phyrexian Reclamation", "Enchantment", b08PhyrexianReclamationOracle, false)
	dead := pushGraveyardCardForTest(me, "Dead Bear")
	relic := batch01GraveyardCard(me, "Sol Ring", "Artifact")
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
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not come back to hand")
	}
}

func TestB08GavonyTownshipGrowsYourTeam(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	township := pushCatalogPermanent(g, me.ID, "Gavony Township", "Land", b08GavonyTownshipOracle, false)
	mine := seedCreature(g, "Mine", me.ID)
	mine2 := seedCreature(g, "Mine Too", me.ID)
	theirs := seedCreature(g, "Theirs", opp.ID)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "W"},
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, township, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
	}{{mine, 1}, {mine2, 1}, {theirs, 0}} {
		if got := counterCount(g, tc.id, "+1/+1"); got != tc.want {
			t.Errorf("%s: %d counters, want %d", tc.id, got, tc.want)
		}
	}
	if card, _ := battlefieldCard(g, township); !card.Tapped {
		t.Error("the ability has a tap cost")
	}
}

func TestB08FountainportThreeAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	port := pushCatalogPermanent(g, me.ID, "Fountainport", "Land", b08FountainportOracle, false)
	bear := seedCreature(g, "Bear", me.ID)

	// {4}, {T}: a Treasure.
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, port, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the Treasure ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Treasure") != 1 {
		t.Fatal("want one Treasure")
	}
	treasure := findBattlefieldByName(g, "Treasure")

	// {3}, {T}, Pay 1 life: a Fish.
	b08Untap(g, port)
	before := me.Life
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, port, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the Fish ability: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1 paid at announce", before, me.Life)
	}
	passPriorityAroundTable(t, g)
	fish := findBattlefieldByName(g, "Fish")
	if card, ok := battlefieldCard(g, fish); !ok || !card.IsCreature() || !card.HasColor("U") {
		t.Error("want a blue Fish creature token")
	}

	// {2}, {T}, Sacrifice a token: a card. A nontoken creature is refused.
	b08Untap(g, port)
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
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestB08BlightedWoodlandFetchesUpToTwoBasicsTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	woodland := pushCatalogPermanent(g, me.ID, "Blighted Woodland", "Land", b08BlightedWoodlandOracle, false)
	forest := b07LibraryLand(me, "Forest", "Basic Land — Forest")
	island := b07LibraryLand(me, "Island", "Basic Land — Island")
	b07LibraryLand(me, "Swamp", "Basic Land — Swamp")
	coffers := b07LibraryLand(me, "Cabal Coffers", "Land")
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, woodland, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(woodland) {
		t.Error("the Woodland is sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("want a search prompt")
	}
	if hasID(c.SearchCards, coffers) {
		t.Error("a nonbasic was offered for 'basic land cards'")
	}
	if c.SearchMax != 2 {
		t.Errorf("may take %d, want up to two", c.SearchMax)
	}
	answerSearchByID(t, g, me.ID, forest, island)
	for _, id := range []uuid.UUID{forest, island} {
		card, ok := battlefieldCard(g, id)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	verge := playLandFromHand(t, g, "Krosan Verge", b08KrosanVergeOracle)
	top100AssertEnteredTapped(t, g, verge, "Krosan Verge")
	b08Untap(g, verge)
	savannah := b07LibraryLand(me, "Savannah", "Land — Forest Plains")
	b07LibraryLand(me, "Forest", "Basic Land — Forest")
	plains := b07LibraryLand(me, "Plains", "Basic Land — Plains")
	b07LibraryLand(me, "Plains", "Basic Land — Plains")
	island := b07LibraryLand(me, "Island", "Basic Land — Island")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, verge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(verge) {
		t.Error("the Verge is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)

	// First the Forest search: a Savannah is a Forest card.
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("want the Forest search prompt")
	}
	if hasID(c.SearchCards, island) || !hasID(c.SearchCards, savannah) {
		t.Error("the Forest search offers Forests only, dual lands included")
	}
	answerSearchByID(t, g, me.ID, savannah)

	// Then the Plains search, from the same resolution.
	c = searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("want the Plains search prompt after the Forest one")
	}
	if hasID(c.SearchCards, island) || !hasID(c.SearchCards, plains) {
		t.Error("the Plains search offers Plains only")
	}
	answerSearchByID(t, g, me.ID, plains)
	for _, id := range []uuid.UUID{savannah, plains} {
		card, ok := battlefieldCard(g, id)
		if !ok {
			t.Errorf("%s did not reach the battlefield", id)
			continue
		}
		if !card.Tapped {
			t.Errorf("%s entered untapped, want tapped", card.Name)
		}
	}
	if b07OpenSearchPrompts(g) != 0 {
		t.Error("nothing should be left open")
	}
}

// --- lands ---------------------------------------------------------

func TestB08BloodfellCavesEntersTappedAndGainsALife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	caves := playLandFromHand(t, g, "Bloodfell Caves", b08BloodfellCavesOracle)
	top100AssertEnteredTapped(t, g, caves, "Bloodfell Caves")
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
	b08Untap(g, caves)
	if err := g.ActivateManaAbility(me.ID, caves, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("the Caves ask B or R, got %+v", pick)
	}
}

func TestB08CabarettiCourtyardSacrificesItselfForABasicAndALife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mountain := b07LibraryLand(me, "Mountain", "Basic Land — Mountain")
	b07LibraryLand(me, "Plains", "Basic Land — Plains")
	swamp := b07LibraryLand(me, "Swamp", "Basic Land — Swamp")
	before := me.Life

	courtyard := playLandFromHand(t, g, "Cabaretti Courtyard", b08CabarettiCourtyardOracle)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(courtyard) || !me.Graveyard.Contains(courtyard) {
		t.Fatal("the Courtyard sacrifices itself")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("want a search prompt")
	}
	if hasID(c.SearchCards, swamp) {
		t.Error("a Swamp was offered for 'a basic Mountain, Forest, or Plains card'")
	}
	answerSearchByID(t, g, me.ID, mountain)
	card, ok := battlefieldCard(g, mountain)
	if !ok || !card.Tapped {
		t.Error("the basic should be on the battlefield tapped")
	}
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
}

func TestB08MinesOfMoriaEntersUntappedWithALegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tapped := playLandFromHand(t, g, "Mines of Moria", b08MinesOfMoriaOracle)
	top100AssertEnteredTapped(t, g, tapped, "Mines of Moria (no legend)")

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	b08TypedCreature(g2, me2.ID, "Legend", "Legendary Creature — Human", 2, 2)
	untapped := playLandFromHand(t, g2, "Mines of Moria", b08MinesOfMoriaOracle)
	top100AssertEnteredUntapped(t, g2, untapped, "Mines of Moria (with a legend)")
	if err := g2.ActivateManaAbility(me2.ID, untapped, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me2); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
	_ = me
}

func TestB08SunkenRuinsFiltersOneHybridIntoTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ruins := seedPermanentWithOracle(g, me.ID, "Sunken Ruins", "Land", b08SunkenRuinsOracle)

	if err := g.ActivateManaAbility(me.ID, ruins, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the {C} half: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}

	fresh := seedPermanentWithOracle(g, me.ID, "Sunken Ruins", "Land", b08SunkenRuinsOracle)
	me.ManaPool.EmptyPool()
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("{U/B} was paid with {G}")
	}
	if card, _ := battlefieldCard(g, fresh); card.Tapped {
		t.Fatal("a refused activation must not tap the Ruins")
	}

	me.ManaPool.EmptyPool()
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	picks := 0
	for _, c := range g.PendingChoices {
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
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "U" || got[1] != "U" {
		t.Errorf("pool %v, want [U U]", got)
	}
}
