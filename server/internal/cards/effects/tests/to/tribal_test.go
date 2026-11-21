package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- lords --------------------------------------------------------

// TestLordOfAtlantisPumpsEveryMerfolk is the S26 oracle fix: the
// printed card has no "you control" clause, so an opponent's Merfolk
// gets the +1/+1 too.
func TestLordOfAtlantisPumpsEveryMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine, theirs := g.Seats[0].ID, g.Seats[1].ID
	myMerfolk := effectstest.PushTribalCreature(g, mine, "Merfolk Looter", "Creature — Merfolk Rogue", 1, 1)
	theirMerfolk := effectstest.PushTribalCreature(g, theirs, "Lullmage Mentor", "Creature — Merfolk Wizard", 2, 2)
	myBear := effectstest.PushTribalCreature(g, mine, "Grizzly Bears", "Creature — Bear", 2, 2)
	lord := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lord of Atlantis",
		TypeLine: "Creature — Merfolk", Power: 2, Toughness: 2,
		OracleID: effectstest.LordOfAtlantisOracle, Owner: mine, Controller: mine,
	})

	if got := effectstest.EffectivePower(t, g, myMerfolk); got != 2 {
		t.Errorf("own Merfolk power = %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, theirMerfolk); got != 3 {
		t.Errorf("OPPONENT's Merfolk power = %d, want 3 — Lord of Atlantis has no controller clause", got)
	}
	if got := effectstest.EffectivePower(t, g, myBear); got != 2 {
		t.Errorf("Bear power = %d, want 2 (not a Merfolk)", got)
	}
	if got := effectstest.EffectivePower(t, g, lord); got != 2 {
		t.Errorf("Lord's own power = %d, want 2 — \"other\" excludes itself", got)
	}
}

// TestGoblinChieftainIsYoursOnly is the other half of the same
// distinction: this lord DOES say "you control".
func TestGoblinChieftainIsYoursOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine, theirs := g.Seats[0].ID, g.Seats[1].ID
	myGoblin := effectstest.PushTribalCreature(g, mine, "Goblin Piker", "Creature — Goblin Warrior", 2, 1)
	theirGoblin := effectstest.PushTribalCreature(g, theirs, "Mogg Fanatic", "Creature — Goblin", 1, 1)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Chieftain",
		TypeLine: "Creature — Goblin", Power: 2, Toughness: 2,
		OracleID: effectstest.GoblinChieftainOracle, Owner: mine, Controller: mine,
	})

	if got := effectstest.EffectivePower(t, g, myGoblin); got != 3 {
		t.Errorf("own Goblin power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, theirGoblin); got != 1 {
		t.Errorf("opponent's Goblin power = %d, want 1 — Chieftain says \"you control\"", got)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, myGoblin), "haste") {
		t.Error("own Goblin did not gain haste from Goblin Chieftain")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirGoblin), "haste") {
		t.Error("opponent's Goblin gained haste from my Goblin Chieftain")
	}
}

// TestDeathBaronCoversBothNamedTypes — "Skeletons you control and
// other Zombies you control".
func TestDeathBaronCoversBothNamedTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	skeleton := effectstest.PushTribalCreature(g, mine, "Drudge Skeletons", "Creature — Skeleton", 1, 1)
	zombie := effectstest.PushTribalCreature(g, mine, "Walking Corpse", "Creature — Zombie", 2, 2)
	baron := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Death Baron",
		TypeLine: "Creature — Zombie Wizard", Power: 2, Toughness: 2,
		OracleID: effectstest.DeathBaronOracle, Owner: mine, Controller: mine,
	})

	if got := effectstest.EffectivePower(t, g, skeleton); got != 2 {
		t.Errorf("Skeleton power = %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, zombie); got != 3 {
		t.Errorf("Zombie power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, baron); got != 2 {
		t.Errorf("Death Baron's own power = %d, want 2 — \"other Zombies\"", got)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, skeleton), "deathtouch") {
		t.Error("Skeleton did not gain deathtouch")
	}
}

// TestElvishArchdruidCountsElvesIncludingChangelings — the mana half,
// which is the sprint's TypeFilter in a scaled produced string.
func TestElvishArchdruidCountsElvesIncludingChangelings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	effectstest.PushTribalCreature(g, mine, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.PushTribalCreature(g, mine, "Universal Automaton", "Artifact Creature — Shapeshifter", 1, 1,
		game.KeywordChangeling)
	// Not an Elf, and not counted.
	effectstest.PushTribalCreature(g, mine, "Grizzly Bears", "Creature — Bear", 2, 2)
	druid := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elvish Archdruid",
		TypeLine: "Creature — Elf Druid", Power: 2, Toughness: 2,
		OracleID: effectstest.ElvishArchdruidOracle, Owner: mine, Controller: mine,
	})

	// Two Elves plus the Archdruid itself — the mana ability does not
	// say "other".
	var produced string
	g.ReadSnapshot(func() {
		card, ok := effectstest.BattlefieldCard(g, druid)
		if !ok {
			t.Fatal("Elvish Archdruid not on the battlefield")
		}
		abilities := game.ManaAbilitiesForCard(card)
		if len(abilities) != 1 || abilities[0].ProducedFunc == nil {
			t.Fatalf("Elvish Archdruid mana abilities = %+v", abilities)
		}
		produced = effectstest.ProducedOfShape(abilities[0], g, mine, druid)
	})
	if produced != "{G}{G}{G}" {
		t.Errorf("Archdruid produced = %q, want {G}{G}{G} (Llanowar + changeling + itself)", produced)
	}
}

// --- named-tribe permanents --------------------------------------

func TestVanquishersBannerPumpsTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine, theirs := g.Seats[0].ID, g.Seats[1].ID
	myElf := effectstest.PushTribalCreature(g, mine, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	myGoblin := effectstest.PushTribalCreature(g, mine, "Mogg Fanatic", "Creature — Goblin", 1, 1)
	theirElf := effectstest.PushTribalCreature(g, theirs, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	effectstest.PushNamedTribePermanent(t, g, mine, "Vanquisher's Banner", "Artifact", effectstest.VanquishersBannerOracle, "Elf")

	if got := effectstest.EffectivePower(t, g, myElf); got != 2 {
		t.Errorf("my Elf power = %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, myGoblin); got != 1 {
		t.Errorf("my Goblin power = %d, want 1 — wrong type", got)
	}
	if got := effectstest.EffectivePower(t, g, theirElf); got != 1 {
		t.Errorf("opponent's Elf power = %d, want 1 — \"creatures you control\"", got)
	}
}

// TestNamedTribePermanentIsInertUntilAnswered — the window the ETB
// prompt opens must be safe. A static reading an empty NamedTribe
// applies to nothing, never to everything.
func TestNamedTribePermanentIsInertUntilAnswered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	elf := effectstest.PushTribalCreature(g, mine, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vanquisher's Banner",
		TypeLine: "Artifact", OracleID: effectstest.VanquishersBannerOracle,
		Owner: mine, Controller: mine,
	})
	if got := effectstest.EffectivePower(t, g, elf); got != 1 {
		t.Errorf("Elf power with no type named = %d, want 1", got)
	}
}

// TestDoorOfDestiniesScalesWithChargeCounters.
func TestDoorOfDestiniesScalesWithChargeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	sliver := effectstest.PushTribalCreature(g, mine, "Metallic Sliver", "Creature — Sliver", 1, 1)
	door := effectstest.PushNamedTribePermanent(t, g, mine, "Door of Destinies", "Artifact", effectstest.DoorOfDestiniesOracle, "Sliver")

	if got := effectstest.EffectivePower(t, g, sliver); got != 1 {
		t.Errorf("Sliver with no charge counters = %d, want 1", got)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(door, "charge", 3); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	if got := effectstest.EffectivePower(t, g, sliver); got != 4 {
		t.Errorf("Sliver with 3 charge counters = %d, want 4", got)
	}
}

// TestAdaptiveAutomatonBecomesTheChosenTypeButDoesNotPumpItself —
// the layer 4 / layer 7c ordering, and the "other" clause.
func TestAdaptiveAutomatonBecomesTheChosenTypeButDoesNotPumpItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	goblin := effectstest.PushTribalCreature(g, mine, "Mogg Fanatic", "Creature — Goblin", 1, 1)
	automaton := effectstest.PushNamedTribePermanent(t, g, mine, "Adaptive Automaton",
		"Artifact Creature — Construct", effectstest.AdaptiveAutomatonOracle, "Goblin")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == automaton {
				g.Battlefield.Cards[i].Power = 2
				g.Battlefield.Cards[i].Toughness = 2
			}
		}
		g.BumpLayerVersionForTest()
	})

	subs := effectstest.EffectiveSubtypes(t, g, automaton)
	if !effectstest.HasAbility(subs, "Goblin") {
		t.Errorf("Automaton subtypes = %v, want a Goblin among them", subs)
	}
	if !effectstest.HasAbility(subs, "Construct") {
		t.Errorf("Automaton subtypes = %v — the type-ADD must keep Construct", subs)
	}
	if got := effectstest.EffectivePower(t, g, goblin); got != 2 {
		t.Errorf("Goblin power = %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, automaton); got != 2 {
		t.Errorf("Automaton's own power = %d, want 2 — it is a Goblin but the clause says \"other\"", got)
	}
}

// --- Coat of Arms -------------------------------------------------

func TestCoatOfArmsCountsSharedTypesPerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine, theirs := g.Seats[0].ID, g.Seats[1].ID
	// The reminder text's own example: two Goblin Warriors and a
	// Goblin Shaman each get +2/+2.
	warriorA := effectstest.PushTribalCreature(g, mine, "Goblin Piker", "Creature — Goblin Warrior", 2, 1)
	warriorB := effectstest.PushTribalCreature(g, mine, "Goblin Raider", "Creature — Goblin Warrior", 2, 2)
	shaman := effectstest.PushTribalCreature(g, theirs, "Goblin Shaman", "Creature — Goblin Shaman", 2, 2)
	loner := effectstest.PushTribalCreature(g, mine, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Coat of Arms", TypeLine: "Artifact",
		OracleID: effectstest.CoatOfArmsOracle, Owner: mine, Controller: mine,
	})

	if got := effectstest.EffectivePower(t, g, warriorA); got != 4 {
		t.Errorf("Goblin Warrior A power = %d, want 4 (+2 from the other Goblins)", got)
	}
	if got := effectstest.EffectivePower(t, g, warriorB); got != 4 {
		t.Errorf("Goblin Warrior B power = %d, want 4", got)
	}
	if got := effectstest.EffectivePower(t, g, shaman); got != 4 {
		t.Errorf("Goblin Shaman power = %d, want 4 — Coat of Arms is symmetrical", got)
	}
	if got := effectstest.EffectivePower(t, g, loner); got != 2 {
		t.Errorf("Bear power = %d, want 2 — shares nothing", got)
	}
}

// TestCoatOfArmsAndChangelings — a changeling shares a type with
// every creature that has one, so it gets the full count and adds one
// to everybody else's.
func TestCoatOfArmsAndChangelings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	goblin := effectstest.PushTribalCreature(g, mine, "Mogg Fanatic", "Creature — Goblin", 1, 1)
	bear := effectstest.PushTribalCreature(g, mine, "Grizzly Bears", "Creature — Bear", 2, 2)
	changeling := effectstest.PushTribalCreature(g, mine, "Woodland Changeling", "Creature — Shapeshifter", 2, 2,
		game.KeywordChangeling)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Coat of Arms", TypeLine: "Artifact",
		OracleID: effectstest.CoatOfArmsOracle, Owner: mine, Controller: mine,
	})

	if got := effectstest.EffectivePower(t, g, goblin); got != 2 {
		t.Errorf("Goblin power = %d, want 2 (shares Goblin with the changeling)", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("Bear power = %d, want 3 (shares Bear with the changeling)", got)
	}
	if got := effectstest.EffectivePower(t, g, changeling); got != 4 {
		t.Errorf("changeling power = %d, want 4 (shares with both others)", got)
	}
}

// --- Maskwood Nexus ----------------------------------------------

func TestMaskwoodNexusMakesYourCreaturesEveryType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine, theirs := g.Seats[0].ID, g.Seats[1].ID
	myBear := effectstest.PushTribalCreature(g, mine, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirBear := effectstest.PushTribalCreature(g, theirs, "Runeclaw Bear", "Creature — Bear", 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Maskwood Nexus", TypeLine: "Artifact",
		OracleID: effectstest.MaskwoodNexusOracle, Owner: mine, Controller: mine,
	})
	// A lord that would otherwise do nothing now pumps the Bear.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elvish Champion",
		TypeLine: "Creature — Elf", Power: 2, Toughness: 2,
		OracleID: effectstest.ElvishChampionOracle, Owner: mine, Controller: mine,
	})

	if got := effectstest.EffectivePower(t, g, myBear); got != 3 {
		t.Errorf("my Bear under Maskwood Nexus + Elvish Champion = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, theirBear); got != 2 {
		t.Errorf("opponent's Bear = %d, want 2 — Nexus says \"creatures you control\"", got)
	}
}

// --- changeling reaches non-catalog cards -------------------------

// TestVanillaChangelingNeedsNoCatalogEntry is the claim Irregular
// Cohort's comment makes: since #330 the deck importer stamps
// Scryfall's keyword array, and S26 made changeling one the engine
// honours, so Woodland Changeling works with no Spec at all.
func TestVanillaChangelingNeedsNoCatalogEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	changeling := effectstest.PushTribalCreature(g, mine, "Woodland Changeling", "Creature — Shapeshifter", 2, 2,
		game.KeywordChangeling)
	if _, ok := effects.Lookup(""); ok {
		t.Fatal("empty oracle ID resolved to a catalog entry")
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elvish Champion",
		TypeLine: "Creature — Elf", Power: 2, Toughness: 2,
		OracleID: effectstest.ElvishChampionOracle, Owner: mine, Controller: mine,
	})
	if got := effectstest.EffectivePower(t, g, changeling); got != 3 {
		t.Errorf("catalog-less changeling under Elvish Champion = %d, want 3", got)
	}
}

// --- Cavern of Souls ----------------------------------------------

// TestCavernOfSoulsMintsRestrictedManaForTheChosenType is the card
// the per-permanent state exists for: the restriction names a type
// nobody printed.
func TestCavernOfSoulsMintsRestrictedManaForTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	cavern := effectstest.PushNamedTribePermanent(t, g, mine, "Cavern of Souls", "Land", effectstest.CavernOfSoulsOracle, "Elf")

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
	if !effectstest.HasAbility(restrictions, game.ManaRestrictSubtype("Elf")) {
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
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	cavern := effectstest.PushNamedTribePermanent(t, g, mine, "Cavern of Souls", "Land", effectstest.CavernOfSoulsOracle, "Elf")
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
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	cavern := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cavern of Souls", TypeLine: "Land",
		OracleID: effectstest.CavernOfSoulsOracle, Owner: mine, Controller: mine,
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
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Irregular Cohort", "Creature — Shapeshifter", effectstest.IrregularCohortOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1].ID
	bear := effectstest.PushTribalCreature(g, victim, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Shields of Velis Vel", "Kindred Instant — Shapeshifter",
		effectstest.ShieldsOfVelisVelOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.EffectiveToughness(t, g, bear); got != 3 {
		t.Errorf("Bear toughness = %d, want 3 (+0/+1)", got)
	}
	// The layer-4 FACT, not the keyword: since #670 "is every
	// creature type" lives on the Characteristic, and the keyword is
	// only what a printed changeling carries.
	if !effectstest.LayeredCard(t, g, bear).Effective().AllCreatureTypes {
		t.Fatal("Bear did not gain all creature types")
	}
	var isElf bool
	g.ReadSnapshot(func() {
		c, ok := effectstest.BattlefieldCard(g, bear)
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
		OracleID: effectstest.ShieldsOfVelisVelOracle,
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
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	bear := effectstest.PushTribalCreature(g, mine, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Chieftain",
		TypeLine: "Creature — Goblin", Power: 2, Toughness: 2,
		OracleID: effectstest.GoblinChieftainOracle, Owner: mine, Controller: mine,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Maskwood Nexus", TypeLine: "Artifact",
		OracleID: effectstest.MaskwoodNexusOracle, Owner: mine, Controller: mine,
	})

	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("Bear power = %d, want 3 (the Chieftain's layer-7c half)", got)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "haste") {
		t.Error("Bear did not gain haste — the Nexus' type grant must apply in layer 4, " +
			"before any lord's layer-6 keyword grant reads the types")
	}
}
