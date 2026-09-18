package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tribal_test.go — S26 (#78). The lords, the named-tribe permanents,
// Coat of Arms and the changeling interactions.

const (
	cavernOfSoulsOracle     = "89ca686a-7c72-4d8f-9290-e89635624a83"
	doorOfDestiniesOracle   = "9b6d3dcf-aa5a-4516-bd48-a0723e86bfd1"
	vanquishersBannerOracle = "8cf38025-5821-45a0-9483-266353b7e82d"
	adaptiveAutomatonOracle = "53c730c6-2f8c-4af8-b400-b9d573a71e60"
	coatOfArmsOracle        = "5f7f133e-58ea-41ab-b1be-be4b400fac4c"
	maskwoodNexusOracle     = "9b2cdbed-c733-409b-b0e4-2c8960c25111"
	elvishArchdruidOracle   = "6e2c2423-d854-4478-99e6-64f29851f026"
	elvishChampionOracle    = "7e40f37c-9a0c-40e0-b195-7ea94b12f798"
	goblinChieftainOracle   = "368b4052-174e-4458-a6e6-eaf8093aa0fe"
	deathBaronOracle        = "99024aa8-5687-4d38-8a4b-feef42d6c1ff"
	irregularCohortOracle   = "c0636d16-671c-4e80-af8c-67d80d2cd979"
	shieldsOfVelisVelOracle = "7ad6be4e-5c3c-4633-a641-beb06e4129b9"
)

// pushTribalCreature seeds a creature on the battlefield under
// `owner` with the layer listener fired, and returns its ID.
func pushTribalCreature(g *game.Game, owner uuid.UUID, name, typeLine string, p, t int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		Power:      p,
		Toughness:  t,
		Keywords:   keywords,
		Owner:      owner,
		Controller: owner,
	})
}

// pushNamedTribePermanent seeds a catalog permanent and answers its
// "choose a creature type" prompt with `tribe`.
func pushNamedTribePermanent(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracleID, tribe string) uuid.UUID {
	t.Helper()
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      owner,
		Controller: owner,
	})
	// The OnETB hook runs off the real resolution path; seeding the
	// battlefield directly skips it, so the choice is stamped the
	// same way the resolve path would.
	var choiceID uuid.UUID
	g.WithWriteLock(func() {
		choiceID = g.QueueCreatureTypeChoiceForEffect(owner, id, name)
	})
	if err := g.ResolveCreatureTypeChoice(choiceID, owner, tribe); err != nil {
		t.Fatalf("ResolveCreatureTypeChoice(%s, %s): %v", name, tribe, err)
	}
	return id
}

// --- lords --------------------------------------------------------

// TestLordOfAtlantisPumpsEveryMerfolk is the S26 oracle fix: the
// printed card has no "you control" clause, so an opponent's Merfolk
// gets the +1/+1 too.
func TestLordOfAtlantisPumpsEveryMerfolk(t *testing.T) {
	g := newCatalogGame(t)
	mine, theirs := g.Seats[0].ID, g.Seats[1].ID
	myMerfolk := pushTribalCreature(g, mine, "Merfolk Looter", "Creature — Merfolk Rogue", 1, 1)
	theirMerfolk := pushTribalCreature(g, theirs, "Lullmage Mentor", "Creature — Merfolk Wizard", 2, 2)
	myBear := pushTribalCreature(g, mine, "Grizzly Bears", "Creature — Bear", 2, 2)
	lord := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lord of Atlantis",
		TypeLine: "Creature — Merfolk", Power: 2, Toughness: 2,
		OracleID: lordOfAtlantisOracle, Owner: mine, Controller: mine,
	})

	if got := effectivePower(t, g, myMerfolk); got != 2 {
		t.Errorf("own Merfolk power = %d, want 2", got)
	}
	if got := effectivePower(t, g, theirMerfolk); got != 3 {
		t.Errorf("OPPONENT's Merfolk power = %d, want 3 — Lord of Atlantis has no controller clause", got)
	}
	if got := effectivePower(t, g, myBear); got != 2 {
		t.Errorf("Bear power = %d, want 2 (not a Merfolk)", got)
	}
	if got := effectivePower(t, g, lord); got != 2 {
		t.Errorf("Lord's own power = %d, want 2 — \"other\" excludes itself", got)
	}
}

// TestGoblinChieftainIsYoursOnly is the other half of the same
// distinction: this lord DOES say "you control".
func TestGoblinChieftainIsYoursOnly(t *testing.T) {
	g := newCatalogGame(t)
	mine, theirs := g.Seats[0].ID, g.Seats[1].ID
	myGoblin := pushTribalCreature(g, mine, "Goblin Piker", "Creature — Goblin Warrior", 2, 1)
	theirGoblin := pushTribalCreature(g, theirs, "Mogg Fanatic", "Creature — Goblin", 1, 1)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Chieftain",
		TypeLine: "Creature — Goblin", Power: 2, Toughness: 2,
		OracleID: goblinChieftainOracle, Owner: mine, Controller: mine,
	})

	if got := effectivePower(t, g, myGoblin); got != 3 {
		t.Errorf("own Goblin power = %d, want 3", got)
	}
	if got := effectivePower(t, g, theirGoblin); got != 1 {
		t.Errorf("opponent's Goblin power = %d, want 1 — Chieftain says \"you control\"", got)
	}
	if !hasAbility(effectiveAbilities(t, g, myGoblin), "haste") {
		t.Error("own Goblin did not gain haste from Goblin Chieftain")
	}
	if hasAbility(effectiveAbilities(t, g, theirGoblin), "haste") {
		t.Error("opponent's Goblin gained haste from my Goblin Chieftain")
	}
}

// TestDeathBaronCoversBothNamedTypes — "Skeletons you control and
// other Zombies you control".
func TestDeathBaronCoversBothNamedTypes(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	skeleton := pushTribalCreature(g, mine, "Drudge Skeletons", "Creature — Skeleton", 1, 1)
	zombie := pushTribalCreature(g, mine, "Walking Corpse", "Creature — Zombie", 2, 2)
	baron := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Death Baron",
		TypeLine: "Creature — Zombie Wizard", Power: 2, Toughness: 2,
		OracleID: deathBaronOracle, Owner: mine, Controller: mine,
	})

	if got := effectivePower(t, g, skeleton); got != 2 {
		t.Errorf("Skeleton power = %d, want 2", got)
	}
	if got := effectivePower(t, g, zombie); got != 3 {
		t.Errorf("Zombie power = %d, want 3", got)
	}
	if got := effectivePower(t, g, baron); got != 2 {
		t.Errorf("Death Baron's own power = %d, want 2 — \"other Zombies\"", got)
	}
	if !hasAbility(effectiveAbilities(t, g, skeleton), "deathtouch") {
		t.Error("Skeleton did not gain deathtouch")
	}
}

// TestElvishArchdruidCountsElvesIncludingChangelings — the mana half,
// which is the sprint's TypeFilter in a scaled produced string.
func TestElvishArchdruidCountsElvesIncludingChangelings(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	pushTribalCreature(g, mine, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	pushTribalCreature(g, mine, "Universal Automaton", "Artifact Creature — Shapeshifter", 1, 1,
		game.KeywordChangeling)
	// Not an Elf, and not counted.
	pushTribalCreature(g, mine, "Grizzly Bears", "Creature — Bear", 2, 2)
	druid := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elvish Archdruid",
		TypeLine: "Creature — Elf Druid", Power: 2, Toughness: 2,
		OracleID: elvishArchdruidOracle, Owner: mine, Controller: mine,
	})

	// Two Elves plus the Archdruid itself — the mana ability does not
	// say "other".
	var produced string
	g.ReadSnapshot(func() {
		card, ok := battlefieldCard(g, druid)
		if !ok {
			t.Fatal("Elvish Archdruid not on the battlefield")
		}
		abilities := game.ManaAbilitiesForCard(card)
		if len(abilities) != 1 || abilities[0].ProducedFunc == nil {
			t.Fatalf("Elvish Archdruid mana abilities = %+v", abilities)
		}
		produced = abilities[0].ProducedFunc(g, mine, druid)
	})
	if produced != "{G}{G}{G}" {
		t.Errorf("Archdruid produced = %q, want {G}{G}{G} (Llanowar + changeling + itself)", produced)
	}
}

// --- named-tribe permanents --------------------------------------

func TestVanquishersBannerPumpsTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	mine, theirs := g.Seats[0].ID, g.Seats[1].ID
	myElf := pushTribalCreature(g, mine, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	myGoblin := pushTribalCreature(g, mine, "Mogg Fanatic", "Creature — Goblin", 1, 1)
	theirElf := pushTribalCreature(g, theirs, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	pushNamedTribePermanent(t, g, mine, "Vanquisher's Banner", "Artifact", vanquishersBannerOracle, "Elf")

	if got := effectivePower(t, g, myElf); got != 2 {
		t.Errorf("my Elf power = %d, want 2", got)
	}
	if got := effectivePower(t, g, myGoblin); got != 1 {
		t.Errorf("my Goblin power = %d, want 1 — wrong type", got)
	}
	if got := effectivePower(t, g, theirElf); got != 1 {
		t.Errorf("opponent's Elf power = %d, want 1 — \"creatures you control\"", got)
	}
}

// TestNamedTribePermanentIsInertUntilAnswered — the window the ETB
// prompt opens must be safe. A static reading an empty NamedTribe
// applies to nothing, never to everything.
func TestNamedTribePermanentIsInertUntilAnswered(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	elf := pushTribalCreature(g, mine, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vanquisher's Banner",
		TypeLine: "Artifact", OracleID: vanquishersBannerOracle,
		Owner: mine, Controller: mine,
	})
	if got := effectivePower(t, g, elf); got != 1 {
		t.Errorf("Elf power with no type named = %d, want 1", got)
	}
}

// TestDoorOfDestiniesScalesWithChargeCounters.
func TestDoorOfDestiniesScalesWithChargeCounters(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	sliver := pushTribalCreature(g, mine, "Metallic Sliver", "Creature — Sliver", 1, 1)
	door := pushNamedTribePermanent(t, g, mine, "Door of Destinies", "Artifact", doorOfDestiniesOracle, "Sliver")

	if got := effectivePower(t, g, sliver); got != 1 {
		t.Errorf("Sliver with no charge counters = %d, want 1", got)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(door, "charge", 3); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	if got := effectivePower(t, g, sliver); got != 4 {
		t.Errorf("Sliver with 3 charge counters = %d, want 4", got)
	}
}

// TestAdaptiveAutomatonBecomesTheChosenTypeButDoesNotPumpItself —
// the layer 4 / layer 7c ordering, and the "other" clause.
func TestAdaptiveAutomatonBecomesTheChosenTypeButDoesNotPumpItself(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	goblin := pushTribalCreature(g, mine, "Mogg Fanatic", "Creature — Goblin", 1, 1)
	automaton := pushNamedTribePermanent(t, g, mine, "Adaptive Automaton",
		"Artifact Creature — Construct", adaptiveAutomatonOracle, "Goblin")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == automaton {
				g.Battlefield.Cards[i].Power = 2
				g.Battlefield.Cards[i].Toughness = 2
			}
		}
		g.BumpLayerVersionForTest()
	})

	subs := effectiveSubtypes(t, g, automaton)
	if !hasAbility(subs, "Goblin") {
		t.Errorf("Automaton subtypes = %v, want a Goblin among them", subs)
	}
	if !hasAbility(subs, "Construct") {
		t.Errorf("Automaton subtypes = %v — the type-ADD must keep Construct", subs)
	}
	if got := effectivePower(t, g, goblin); got != 2 {
		t.Errorf("Goblin power = %d, want 2", got)
	}
	if got := effectivePower(t, g, automaton); got != 2 {
		t.Errorf("Automaton's own power = %d, want 2 — it is a Goblin but the clause says \"other\"", got)
	}
}

// --- Coat of Arms -------------------------------------------------

func TestCoatOfArmsCountsSharedTypesPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	mine, theirs := g.Seats[0].ID, g.Seats[1].ID
	// The reminder text's own example: two Goblin Warriors and a
	// Goblin Shaman each get +2/+2.
	warriorA := pushTribalCreature(g, mine, "Goblin Piker", "Creature — Goblin Warrior", 2, 1)
	warriorB := pushTribalCreature(g, mine, "Goblin Raider", "Creature — Goblin Warrior", 2, 2)
	shaman := pushTribalCreature(g, theirs, "Goblin Shaman", "Creature — Goblin Shaman", 2, 2)
	loner := pushTribalCreature(g, mine, "Grizzly Bears", "Creature — Bear", 2, 2)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Coat of Arms", TypeLine: "Artifact",
		OracleID: coatOfArmsOracle, Owner: mine, Controller: mine,
	})

	if got := effectivePower(t, g, warriorA); got != 4 {
		t.Errorf("Goblin Warrior A power = %d, want 4 (+2 from the other Goblins)", got)
	}
	if got := effectivePower(t, g, warriorB); got != 4 {
		t.Errorf("Goblin Warrior B power = %d, want 4", got)
	}
	if got := effectivePower(t, g, shaman); got != 4 {
		t.Errorf("Goblin Shaman power = %d, want 4 — Coat of Arms is symmetrical", got)
	}
	if got := effectivePower(t, g, loner); got != 2 {
		t.Errorf("Bear power = %d, want 2 — shares nothing", got)
	}
}

// TestCoatOfArmsAndChangelings — a changeling shares a type with
// every creature that has one, so it gets the full count and adds one
// to everybody else's.
func TestCoatOfArmsAndChangelings(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	goblin := pushTribalCreature(g, mine, "Mogg Fanatic", "Creature — Goblin", 1, 1)
	bear := pushTribalCreature(g, mine, "Grizzly Bears", "Creature — Bear", 2, 2)
	changeling := pushTribalCreature(g, mine, "Woodland Changeling", "Creature — Shapeshifter", 2, 2,
		game.KeywordChangeling)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Coat of Arms", TypeLine: "Artifact",
		OracleID: coatOfArmsOracle, Owner: mine, Controller: mine,
	})

	if got := effectivePower(t, g, goblin); got != 2 {
		t.Errorf("Goblin power = %d, want 2 (shares Goblin with the changeling)", got)
	}
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("Bear power = %d, want 3 (shares Bear with the changeling)", got)
	}
	if got := effectivePower(t, g, changeling); got != 4 {
		t.Errorf("changeling power = %d, want 4 (shares with both others)", got)
	}
}

// --- Maskwood Nexus ----------------------------------------------

func TestMaskwoodNexusMakesYourCreaturesEveryType(t *testing.T) {
	g := newCatalogGame(t)
	mine, theirs := g.Seats[0].ID, g.Seats[1].ID
	myBear := pushTribalCreature(g, mine, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirBear := pushTribalCreature(g, theirs, "Runeclaw Bear", "Creature — Bear", 2, 2)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Maskwood Nexus", TypeLine: "Artifact",
		OracleID: maskwoodNexusOracle, Owner: mine, Controller: mine,
	})
	// A lord that would otherwise do nothing now pumps the Bear.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elvish Champion",
		TypeLine: "Creature — Elf", Power: 2, Toughness: 2,
		OracleID: elvishChampionOracle, Owner: mine, Controller: mine,
	})

	if got := effectivePower(t, g, myBear); got != 3 {
		t.Errorf("my Bear under Maskwood Nexus + Elvish Champion = %d, want 3", got)
	}
	if got := effectivePower(t, g, theirBear); got != 2 {
		t.Errorf("opponent's Bear = %d, want 2 — Nexus says \"creatures you control\"", got)
	}
}

// --- changeling reaches non-catalog cards -------------------------

// TestVanillaChangelingNeedsNoCatalogEntry is the claim Irregular
// Cohort's comment makes: since #330 the deck importer stamps
// Scryfall's keyword array, and S26 made changeling one the engine
// honours, so Woodland Changeling works with no Spec at all.
func TestVanillaChangelingNeedsNoCatalogEntry(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	changeling := pushTribalCreature(g, mine, "Woodland Changeling", "Creature — Shapeshifter", 2, 2,
		game.KeywordChangeling)
	if _, ok := Lookup(""); ok {
		t.Fatal("empty oracle ID resolved to a catalog entry")
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elvish Champion",
		TypeLine: "Creature — Elf", Power: 2, Toughness: 2,
		OracleID: elvishChampionOracle, Owner: mine, Controller: mine,
	})
	if got := effectivePower(t, g, changeling); got != 3 {
		t.Errorf("catalog-less changeling under Elvish Champion = %d, want 3", got)
	}
}

func hasAbility(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

// --- Cavern of Souls ----------------------------------------------

// TestCavernOfSoulsMintsRestrictedManaForTheChosenType is the card
// the per-permanent state exists for: the restriction names a type
// nobody printed.
func TestCavernOfSoulsMintsRestrictedManaForTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	cavern := pushNamedTribePermanent(t, g, mine, "Cavern of Souls", "Land", cavernOfSoulsOracle, "Elf")

	// Ability 1 is the coloured, restricted half; it is a pipe, so it
	// queues a colour pick carrying the restrictions.
	if err := g.ActivateManaAbility(mine, cavern, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	var choiceID uuid.UUID
	var restrictions []string
	g.ReadSnapshot(func() {
		if len(g.PendingChoices) != 1 {
			t.Fatalf("pending choices = %d, want 1 (the colour pick)", len(g.PendingChoices))
		}
		choiceID = g.PendingChoices[0].ID
		restrictions = g.PendingChoices[0].ManaRestrictions
	})
	if !hasAbility(restrictions, game.ManaRestrictSubtype("Elf")) {
		t.Fatalf("colour pick restrictions = %v, want a subtype:Elf tag", restrictions)
	}
	if err := g.ResolveManaChoice(choiceID, mine, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}

	var pool game.ManaPool
	g.ReadSnapshot(func() { pool = g.Seats[0].ManaPool })
	cost, err := game.ParseCost("{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	elf := game.ManaSpendForCast(game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid"})
	if !pool.CanPayFor(cost, 0, elf) {
		t.Error("Cavern mana named Elf could not pay for an Elf creature spell")
	}
	goblin := game.ManaSpendForCast(game.Card{Name: "Mogg Fanatic", TypeLine: "Creature — Goblin"})
	if pool.CanPayFor(cost, 0, goblin) {
		t.Error("Cavern mana named Elf paid for a Goblin — the restriction is not being enforced")
	}
	changeling := game.ManaSpendForCast(game.Card{
		Name: "Universal Automaton", TypeLine: "Artifact Creature — Shapeshifter",
		Keywords: []string{game.KeywordChangeling},
	})
	if !pool.CanPayFor(cost, 0, changeling) {
		t.Error("Cavern mana named Elf refused a changeling — CR 702.73a says it IS an Elf")
	}
	sorcery := game.ManaSpendForCast(game.Card{Name: "Lava Spike", TypeLine: "Sorcery — Arcane"})
	if pool.CanPayFor(cost, 0, sorcery) {
		t.Error("Cavern mana paid for a noncreature spell")
	}
}

// TestCavernOfSoulsColorlessHalfIsUnrestricted — the first ability is
// a plain {C} and must not pick up the second's restriction.
func TestCavernOfSoulsColorlessHalfIsUnrestricted(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	cavern := pushNamedTribePermanent(t, g, mine, "Cavern of Souls", "Land", cavernOfSoulsOracle, "Elf")
	if err := g.ActivateManaAbility(mine, cavern, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	var pool game.ManaPool
	g.ReadSnapshot(func() { pool = g.Seats[0].ManaPool })
	if len(pool) != 1 {
		t.Fatalf("pool = %v, want one colorless token", pool)
	}
	if len(pool[0].Restrictions) != 0 {
		t.Errorf("colorless half minted restrictions %v, want none", pool[0].Restrictions)
	}
}

// TestCavernOfSoulsWithNoTypeNamedIsUnspendable pins the failure
// direction: before the prompt is answered the coloured mana must be
// unusable, never unrestricted.
func TestCavernOfSoulsWithNoTypeNamedIsUnspendable(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	cavern := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cavern of Souls", TypeLine: "Land",
		OracleID: cavernOfSoulsOracle, Owner: mine, Controller: mine,
	})
	if err := g.ActivateManaAbility(mine, cavern, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	var choiceID uuid.UUID
	g.ReadSnapshot(func() { choiceID = g.PendingChoices[0].ID })
	if err := g.ResolveManaChoice(choiceID, mine, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	var pool game.ManaPool
	g.ReadSnapshot(func() { pool = g.Seats[0].ManaPool })
	cost, _ := game.ParseCost("{G}")
	elf := game.ManaSpendForCast(game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid"})
	if pool.CanPayFor(cost, 0, elf) {
		t.Error("a Cavern with no type named produced spendable mana")
	}
}

// --- changeling cards ---------------------------------------------

func TestIrregularCohortMakesAChangelingToken(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Irregular Cohort", "Creature — Shapeshifter", irregularCohortOracle, nil)
	passPriorityAroundTable(t, g)

	var token *game.Card
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Shapeshifter" {
				c := g.Battlefield.Cards[i]
				token = &c
			}
		}
	})
	if token == nil {
		t.Fatal("Irregular Cohort created no Shapeshifter token")
	}
	if !token.HasSubtype("Goblin") {
		t.Error("the token is not every creature type — its changeling keyword did not survive")
	}
}

func TestShieldsOfVelisVelGrantsAllCreatureTypesUntilEOT(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1].ID
	bear := pushTribalCreature(g, victim, "Grizzly Bears", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Shields of Velis Vel", "Kindred Instant — Shapeshifter",
		shieldsOfVelisVelOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	passPriorityAroundTable(t, g)

	if got := effectiveToughness(t, g, bear); got != 3 {
		t.Errorf("Bear toughness = %d, want 3 (+0/+1)", got)
	}
	// The layer-4 FACT, not the keyword: since #670 "is every
	// creature type" lives on the Characteristic, and the keyword is
	// only what a printed changeling carries.
	if !layeredCard(t, g, bear).Effective().AllCreatureTypes {
		t.Fatal("Bear did not gain all creature types")
	}
	var isElf bool
	g.ReadSnapshot(func() {
		c, ok := battlefieldCard(g, bear)
		if !ok {
			t.Fatal("Bear left the battlefield")
		}
		isElf = c.HasSubtype("Elf")
	})
	if !isElf {
		t.Error("Bear with all creature types is not an Elf")
	}
}

// TestShieldsOfVelisVelIsAKindredCardInEveryZone — the spell itself
// prints changeling, so it is every creature type in the graveyard it
// routes to. Nothing in the card file arranges this; the printed
// keyword does.
func TestShieldsOfVelisVelIsAKindredCardInEveryZone(t *testing.T) {
	caster := game.Card{
		Name: "Shields of Velis Vel", TypeLine: "Kindred Instant — Shapeshifter",
		OracleID: shieldsOfVelisVelOracle,
	}
	if !caster.HasSubtype("Elf") {
		t.Error("a Kindred card with changeling is not an Elf outside the battlefield")
	}
}

// TestMaskwoodNexusAppliesBeforeAnyLordRegardlessOfEntryOrder is the
// regression test for the one place the "every creature type is a
// keyword" shortcut could have leaked.
//
// The grant is a TYPE change, so it belongs in CR 613's layer 4 and
// it is declared there — even though what it writes is an ability
// string. Declaring it in layer 6 (where an ability grant would
// normally live) compiles and passes the obvious test, because a
// lord's +1/+1 is layer 7c and runs after ALL of layer 6 either way.
// What breaks is the lord's KEYWORD half, which is layer 6 too and
// would then be ordered against the Nexus by timestamp: a Goblin
// Chieftain that entered first would grant haste before the Bear
// became a Goblin, and the card would look half-working.
//
// So the Nexus enters LAST here on purpose. Both halves must land.
func TestMaskwoodNexusAppliesBeforeAnyLordRegardlessOfEntryOrder(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	bear := pushTribalCreature(g, mine, "Grizzly Bears", "Creature — Bear", 2, 2)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Chieftain",
		TypeLine: "Creature — Goblin", Power: 2, Toughness: 2,
		OracleID: goblinChieftainOracle, Owner: mine, Controller: mine,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Maskwood Nexus", TypeLine: "Artifact",
		OracleID: maskwoodNexusOracle, Owner: mine, Controller: mine,
	})

	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("Bear power = %d, want 3 (the Chieftain's layer-7c half)", got)
	}
	if !hasAbility(effectiveAbilities(t, g, bear), "haste") {
		t.Error("Bear did not gain haste — the Nexus' type grant must apply in layer 4, " +
			"before any lord's layer-6 keyword grant reads the types")
	}
}
