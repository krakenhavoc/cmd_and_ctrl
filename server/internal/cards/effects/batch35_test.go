package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch35_test.go — card-level coverage for the card-coverage
// roadmap's batch 35 (#398, `edhrec_rank` 3650–3752): the "no new
// machinery" group. One test per observable behaviour, driven through
// a real cast, activation or attack rather than by calling primitives.

const (
	b35SphinxsTutelageOracle    = "158db438-52ba-4c30-b7bf-d1c4f76da917"
	b35RiskyShortcutOracle      = "a7c55495-6aca-4d17-b61a-cab8fa38b9f4"
	b35SealOfPrimordiumOracle   = "f14dbb39-c9f9-4f64-b22a-38dba28f5b1e"
	b35ScaledNurturerOracle     = "13b96709-0e88-476b-9485-956e682bb818"
	b35AetherGaleOracle         = "b21d6482-1b3a-47e6-98a5-3067f5f3818b"
	b35PlagueBelcherOracle      = "f086a82d-8dbe-4d02-a076-03789704e9d4"
	b35MemoryErosionOracle      = "f4a96881-586d-44ad-b427-5cdf8988f9a1"
	b35SmileAtDeathOracle       = "9228f04c-506f-4453-a335-e66876b8ce8d"
	b35LordOfExtinctionOracle   = "ea5e3401-bd6c-47bb-a52a-8eec5f09455d"
	b35TheNecrobloomOracle      = "b981af39-4ee6-4fbc-9a89-618dcad9dfbf"
	b35GarnaOracle              = "3fb83218-e18d-405d-91c8-46b5e4e672a9"
	b35DreadSummonsOracle       = "642fa01d-025e-44dd-8360-5325e5a28282"
	b35AbandonAttachmentsOracle = "82333385-631f-4abf-b159-bb367f1c6fd9"
	b35LuxArtilleryOracle       = "cbd76b22-d04e-48e7-bcab-c7f5466d67d8"
	b35AbsorbOracle             = "132ca99a-a3c7-4ed6-b4d0-0edcd7140ca2"
	b35EdenOracle               = "84856b92-5ce8-47f3-9a1c-78d6a3e26aca"
	b35IkraShidiqiOracle        = "a1a1761c-88e5-40b4-ba4c-60735c054b09"
	b35ScreamingNemesisOracle   = "fcb7c93c-46ab-49b5-a6e0-35d73f3be8f0"
	b35WildwoodScourgeOracle    = "b9dec104-c636-4770-a7fc-7a3331face15"
	b35NekusarOracle            = "8a5e3c8e-8e22-49b9-8ee5-4a36361f0da6"
	b35ThirstForDiscoveryOracle = "1e05e6ef-14af-451d-9d54-e75b1f8871ab"
	b35FinalActOracle           = "335a6c6a-030f-45d4-806c-467a22962eed"
	b35GreenwardenOracle        = "2facb1b3-4522-4610-a0ab-29ac53ca7fcd"

	// Deadly Tempest is the batch's 28th ready card and was already on
	// main from the boardwipe primitives (#382); pinned so the table
	// still accounts for every card on the issue's list.
	b35DeadlyTempestAlreadyOnMainOracle = "b5516bc9-ec8d-4323-8748-96c49d7d0622"
)

// b35Lives is every seat's life total, seat order.
func b35Lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

// b35LibraryCard builds a library card with a colour stamped, the way
// the deck importer would.
func b35LibraryCard(name, typeLine string, colors ...string) game.Card {
	return game.Card{Name: name, TypeLine: typeLine, Colors: colors}
}

// b35Block declares `blocker` as a blocker of `attacker` in the
// declare-blockers step.
func b35Block(t *testing.T, g *game.Game, blocker, attacker uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Deadly Tempest
// was already on main (#382), so the table is the issue's 28 minus
// the four declared skips.
func TestBatch35CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b35SphinxsTutelageOracle:            "Sphinx's Tutelage",
		b35RiskyShortcutOracle:              "Risky Shortcut",
		b35SealOfPrimordiumOracle:           "Seal of Primordium",
		b35ScaledNurturerOracle:             "Scaled Nurturer",
		b35AetherGaleOracle:                 "Aether Gale",
		b35PlagueBelcherOracle:              "Plague Belcher",
		b35MemoryErosionOracle:              "Memory Erosion",
		b35SmileAtDeathOracle:               "Smile at Death",
		b35LordOfExtinctionOracle:           "Lord of Extinction",
		b35TheNecrobloomOracle:              "The Necrobloom",
		b35GarnaOracle:                      "Garna, Bloodfist of Keld",
		b35DreadSummonsOracle:               "Dread Summons",
		b35AbandonAttachmentsOracle:         "Abandon Attachments",
		b35LuxArtilleryOracle:               "Lux Artillery",
		b35AbsorbOracle:                     "Absorb",
		b35EdenOracle:                       "Eden, Seat of the Sanctum",
		b35IkraShidiqiOracle:                "Ikra Shidiqi, the Usurper",
		b35ScreamingNemesisOracle:           "Screaming Nemesis",
		b35WildwoodScourgeOracle:            "Wildwood Scourge",
		b35NekusarOracle:                    "Nekusar, the Mindrazer",
		b35ThirstForDiscoveryOracle:         "Thirst for Discovery",
		b35FinalActOracle:                   "Final Act",
		b35GreenwardenOracle:                "Greenwarden of Murasa",
		b35DeadlyTempestAlreadyOnMainOracle: "Deadly Tempest",
	}
	if len(want) != 24 {
		t.Fatalf("the batch ships 23 cards plus Deadly Tempest, the table lists %d", len(want))
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
			t.Errorf("%s ships unreviewed", name)
		}
	}
}

// --- spells --------------------------------------------------------

func TestB35RiskyShortcutDrawsTwoAndEveryPlayerLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := b35Lives(g)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Risky Shortcut", "Sorcery", b35RiskyShortcutOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("drew %d, want 2", me.Hand.Size()-hand)
	}
	for i, b := range before {
		if got := g.Seats[i].Life; got != b-2 {
			t.Errorf("seat %d: %d → %d, want -2", i, b, got)
		}
	}
}

func TestB35AbsorbCountersAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := b13OpponentCasts(t, g, opp, "Their Opt", "Instant", "", "{U}", nil)
	life := me.Life
	castCatalogSpell(t, g, "Absorb", "Instant", b35AbsorbOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Error("the spell was not countered")
	}
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

func TestB35AetherGaleBouncesExactlySixNonlandPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	a := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	c := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	land := seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	advanceToMain(t, g)
	// Five nonland permanents and a land: not a legal cast.
	if err := b09TryCast(t, g, "Aether Gale", "Sorcery", b35AetherGaleOracle,
		cardRefs(mine, rock, a, b, c, land)); err == nil {
		t.Fatal("a land was accepted among the six")
	}
	d := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	castCatalogSpell(t, g, "Aether Gale", "Sorcery", b35AetherGaleOracle,
		cardRefs(mine, rock, a, b, c, d))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, rock} {
		if !me.Hand.Contains(id) {
			t.Errorf("my permanent %s did not return to my hand", id)
		}
	}
	for _, id := range []uuid.UUID{a, b, c, d} {
		if !opp.Hand.Contains(id) {
			t.Errorf("their permanent %s did not return to their hand", id)
		}
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land is untouched")
	}
}

func TestB35DreadSummonsMillsEveryoneAndZombifiesTheCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep", TypeLine: "Sorcery"},
	)
	seedSearchLibrary(opp,
		game.Card{Name: "Elf", TypeLine: "Creature — Elf"},
		game.Card{Name: "Wurm", TypeLine: "Creature — Wurm"},
		game.Card{Name: "Deep", TypeLine: "Sorcery"},
	)
	// A one-card library mills its one card and nobody loses.
	seedSearchLibrary(other, game.Card{Name: "Lone Bolt", TypeLine: "Instant"})
	castXSpell(t, g, "Dread Summons", "Sorcery", b35DreadSummonsOracle, "{X}{B}{B}", 2, nil)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != 1 || opp.Library.Size() != 1 || other.Library.Size() != 0 {
		t.Errorf("libraries %d/%d/%d, want 1/1/0", me.Library.Size(), opp.Library.Size(), other.Library.Size())
	}
	if other.LosesAtNextSBA {
		t.Error("a mill never loses a player the game")
	}
	zombies := battlefieldIDsNamed(g, "Zombie")
	if len(zombies) != 3 {
		t.Fatalf("three creature cards milled: %d Zombies", len(zombies))
	}
	for _, id := range zombies {
		c, _ := battlefieldCard(g, id)
		if c.Controller != me.ID {
			t.Error("the CASTER creates the Zombies")
		}
		if !c.Tapped {
			t.Error("the Zombies enter tapped")
		}
		if c.Power != 2 || c.Toughness != 2 || !c.HasColor("B") {
			t.Error("a 2/2 black Zombie")
		}
	}
}

func TestB35AbandonAttachmentsDiscardsAsACostAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	_, paid := castWithDiscard(t, g, "Abandon Attachments", b35AbandonAttachmentsOracle, "Sorcery", 1)
	if !me.Graveyard.Contains(paid[0]) {
		t.Fatal("the discard is paid at announce")
	}
	passPriorityAroundTable(t, g)
	// hand + fodder + spell, minus the fodder, minus the spell, plus two.
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want +2 net", hand, me.Hand.Size())
	}
	spec, _ := Lookup(b35AbandonAttachmentsOracle)
	if spec.Completeness != CompletenessCaveats || spec.AdditionalCost == nil || spec.AdditionalCost.DiscardCards != 1 {
		t.Error("the discard-as-cost is a declared gap")
	}
}

func TestB35ThirstForDiscoveryDrawsThreeAndOwesTwoDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Thirst for Discovery", "Instant", b35ThirstForDiscoveryOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("drew %d, want 3", me.Hand.Size()-hand)
	}
	if discardOwed(g, me.ID) != 2 {
		t.Errorf("owes %d discards, want 2", discardOwed(g, me.ID))
	}
	if spec, _ := Lookup(b35ThirstForDiscoveryOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the missing basic-land alternative is a declared gap")
	}
}

func TestB35FinalActSweepsTheChosenModesInPrintedOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	golem := pushIndestructibleWipeCreature(g, opp.ID, "Darksteel Golem")
	walker := pushWalkerForTest(g, opp.ID, "Jace", "", 3)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	dead := b17GraveyardCard(opp, "Dead Elf", "Creature — Elf", "{G}")
	if err := g.AddPlayerCounter(opp.ID, game.CounterPoison, 3); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayerCounter(other.ID, game.CounterEnergy, 2); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayerCounter(me.ID, game.CounterPoison, 1); err != nil {
		t.Fatal(err)
	}
	castModal(t, g, "Final Act", "Sorcery", b35FinalActOracle, []int{0, 3, 4}, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("destroy all creatures")
	}
	if !g.Battlefield.Contains(golem) {
		t.Error("an indestructible creature survives")
	}
	if !g.Battlefield.Contains(walker) || !g.Battlefield.Contains(rock) {
		t.Error("planeswalkers and artifacts are untouched without their mode")
	}
	for _, id := range []uuid.UUID{dead, mine, theirs} {
		if !g.Exile.Contains(id) {
			t.Errorf("exile all graveyards — the destroyed creatures included, since the modes resolve in order: %s", id)
		}
	}
	if opp.Graveyard.Size() != 0 {
		t.Errorf("the opponent's graveyard should be empty, has %d", opp.Graveyard.Size())
	}
	if opp.Counters[game.CounterPoison] != 0 || opp.Poison != 0 || other.Counters[game.CounterEnergy] != 0 {
		t.Errorf("each opponent loses all counters: %v / %v", opp.Counters, other.Counters)
	}
	if me.Counters[game.CounterPoison] != 1 {
		t.Error("the caster keeps their own counters")
	}
	// The planeswalker mode on its own.
	castModal(t, g, "Final Act", "Sorcery", b35FinalActOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("destroy all planeswalkers")
	}
	if !g.Battlefield.Contains(golem) || !g.Battlefield.Contains(rock) {
		t.Error("only the chosen mode happens")
	}
}

// --- permanents ----------------------------------------------------

func TestB35SealOfPrimordiumSacrificesToDestroyAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seal := pushCatalogPermanent(g, me.ID, "Seal of Primordium", "Enchantment", b35SealOfPrimordiumOracle, false)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, seal, 0, game.ActivateAbilityParams{
		Targets: cardRefs(bear),
	}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	b16Activate(t, g, me.ID, seal, 0, game.ActivateAbilityParams{Targets: cardRefs(rock)})
	if g.Battlefield.Contains(seal) || !me.Graveyard.Contains(seal) {
		t.Error("the Seal is sacrificed as a cost")
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
}

func TestB35ScaledNurturerTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	nurturer := pushCatalogPermanent(g, me.ID, "Scaled Nurturer", "Creature — Dragon Druid", b35ScaledNurturerOracle, false)
	if err := g.ActivateManaAbility(me.ID, nurturer, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
	if spec, _ := Lookup(b35ScaledNurturerOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the missing life-gain rider is a declared gap")
	}
}

func TestB35MemoryErosionMillsAnOpponentWhoCasts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Memory Erosion", "Enchantment", b35MemoryErosionOracle, false)
	lib := opp.Library.Size()
	b13OpponentCasts(t, g, opp, "Their Opt", "Instant", "", "{U}", nil)
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-2 {
		t.Errorf("library %d → %d, want -2", lib, opp.Library.Size())
	}
	mine := me.Library.Size()
	castCatalogSpell(t, g, "My Opt", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != mine {
		t.Error("your own spells are silent")
	}
	// A one-card library mills its one card and nobody loses.
	seedSearchLibrary(opp, game.Card{Name: "Last", TypeLine: "Sorcery"})
	b13OpponentCasts(t, g, opp, "Their Opt", "Instant", "", "{U}", nil)
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != 0 || opp.LosesAtNextSBA {
		t.Error("a mill never loses a player the game")
	}
}

func TestB35SphinxsTutelageMillsTwoAndRepeatsWhileTheyShareAColor(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	tutelage := pushCatalogPermanent(g, me.ID, "Sphinx's Tutelage", "Enchantment", b35SphinxsTutelageOracle, false)
	// Top-first: two blue nonland cards (repeat), then a blue card
	// and a land (stop), then filler.
	seedSearchLibrary(opp,
		b35LibraryCard("Opt", "Instant", "U"),
		b35LibraryCard("Ponder", "Sorcery", "U"),
		b35LibraryCard("Brainstorm", "Instant", "U"),
		b35LibraryCard("Island", "Basic Land — Island"),
		b35LibraryCard("Counterspell", "Instant", "U"),
		b35LibraryCard("Negate", "Instant", "U"),
	)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetPlayers, me.ID) {
		t.Error("target OPPONENT")
	}
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != 2 {
		t.Errorf("two blue cards repeat, a land stops: library %d, want 2", opp.Library.Size())
	}
	// A pair sharing no colour stops at once.
	seedSearchLibrary(opp,
		b35LibraryCard("Bolt", "Instant", "R"),
		b35LibraryCard("Opt", "Instant", "U"),
		b35LibraryCard("Ponder", "Sorcery", "U"),
		b35LibraryCard("Preordain", "Sorcery", "U"),
	)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != 2 {
		t.Errorf("red and blue share nothing: library %d, want 2", opp.Library.Size())
	}
	// The loot: draw, then a discard is owed — and the draw fires the
	// trigger again.
	b10AddMana(me, "U", "C", "C", "C", "C", "C")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, tutelage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || discardOwed(g, me.ID) != 1 {
		t.Errorf("hand %d → %d, owes %d discards; want +1 and 1", hand, me.Hand.Size(), discardOwed(g, me.ID))
	}
	discardFromHand(t, g, me.ID)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != 0 {
		t.Errorf("the looted card mills two more: library %d", opp.Library.Size())
	}
	if opp.LosesAtNextSBA {
		t.Error("a mill never loses a player the game")
	}
}

func TestB35PlagueBelcherShrinksACreatureYouControlAndDrainsOnZombieDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	belcher := b20CastCreature(t, g, me, "Plague Belcher", "Creature — Zombie Beast", b35PlagueBelcherOracle, 5, 4)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, belcher) {
		t.Error("target creature YOU control — the Belcher itself is the usual answer")
	}
	pickCard(t, g, me.ID, belcher)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, belcher, game.CounterMinusOne); got != 2 {
		t.Errorf("two -1/-1 counters, got %d", got)
	}
	if currentPower(t, g, belcher) != 3 {
		t.Errorf("a 5/4 with two -1/-1 counters is a 3/2: power %d", currentPower(t, g, belcher))
	}
	assertKeywords(t, g, belcher, "menace")
	// A Zombie of yours dying drains each opponent; a non-Zombie, an
	// opponent's Zombie and the Belcher itself do not.
	before := lifeOfOpponents(g)
	zombie := b12Creature(g, me.ID, "My Zombie", "Creature — Zombie", 2, 2)
	b18Kill(t, g, zombie)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want -1", i+1, b, got)
		}
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b18Kill(t, g, bear)
	theirZombie := b12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	b18Kill(t, g, theirZombie)
	b18Kill(t, g, belcher)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want still -1", i+1, b, got)
		}
	}
}

func TestB35SmileAtDeathReturnsUpToTwoSmallCreaturesWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Smile at Death", "Enchantment", b35SmileAtDeathOracle, false)
	bear := pushGraveyardCardForTest(me, "Dead Bear")
	wurm := b17GraveyardCard(me, "Dead Wurm", "Creature — Wurm", "{4}{G}")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == wurm {
				me.Graveyard.Cards[i].Power, me.Graveyard.Cards[i].Toughness = 5, 5
			}
		}
	})
	elf := b17GraveyardCard(me, "Dead Elf", "Creature — Elf", "{G}")
	b12ToMyNextUpkeep(t, g)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetCards, wurm) {
		t.Error("power 2 or less")
	}
	if !hasID(p.PickTargetCards, bear) || !hasID(p.PickTargetCards, elf) {
		t.Errorf("both small creature cards are offered: %v", p.PickTargetCards)
	}
	b17PickCards(t, g, me.ID, bear, elf)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, elf} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s did not return to the battlefield", id)
			continue
		}
		if got := counterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("each returned creature gets a +1/+1 counter: %d", got)
		}
	}
	if !me.Graveyard.Contains(wurm) {
		t.Error("the Wurm stays")
	}
}

func TestB35LordOfExtinctionIsAsBigAsAllGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushGraveyardCardForTest(me, "Mine")
	pushGraveyardCardForTest(opp, "Theirs")
	pushGraveyardCardForTest(opp, "Theirs Too")
	lord := b12Push(g, me.ID, "Lord of Extinction", "Creature — Elemental", b35LordOfExtinctionOracle, 0, 0)
	if effectivePower(t, g, lord) != 3 || effectiveToughness(t, g, lord) != 3 {
		t.Errorf("three cards in graveyards: %d/%d", effectivePower(t, g, lord), effectiveToughness(t, g, lord))
	}
	// A creature dying is a battlefield exit and shows at once.
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b18Kill(t, g, bear)
	if effectivePower(t, g, lord) != 4 {
		t.Errorf("a dead creature counts: power %d, want 4", effectivePower(t, g, lord))
	}
	if spec, _ := Lookup(b35LordOfExtinctionOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the late read on a mill is a declared gap")
	}
}

func TestB35TheNecrobloomMakesAPlantOrAZombieOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "The Necrobloom", "Legendary Creature — Plant", b35TheNecrobloomOracle, 2, 7)
	for _, name := range []string{"Forest", "Swamp", "Plains", "Island", "Mountain"} {
		seedLandOnBattlefield(g, me.ID, name, "Basic Land — "+name)
	}
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	plants := battlefieldIDsNamed(g, "Plant")
	if len(plants) != 1 {
		t.Fatalf("six lands, five names: %d Plants", len(plants))
	}
	if c, _ := battlefieldCard(g, plants[0]); c.Power != 0 || c.Toughness != 1 || !c.HasColor("G") {
		t.Error("a 0/1 green Plant")
	}
	// The seventh differently-named land makes a Zombie instead.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Wastes", TypeLine: "Token Land — Wastes"}, 1)
	})
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Plant")); n != 2 {
		t.Errorf("six names: still a Plant, %d Plants", n)
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Badlands", TypeLine: "Token Land — Swamp Mountain"}, 1)
	})
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Zombie")); n != 1 {
		t.Errorf("seven names: a Zombie instead, %d Zombies", n)
	}
	if n := len(battlefieldIDsNamed(g, "Plant")); n != 2 {
		t.Errorf("instead, not as well: %d Plants", n)
	}
	if spec, _ := Lookup(b35TheNecrobloomOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the missing dredge is a declared gap")
	}
}

func TestB35GarnaDrawsForAnAttackerDyingAndPingsForAnyOtherDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	garna := b12Push(g, me.ID, "Garna, Bloodfist of Keld", "Legendary Creature — Human Berserker", b35GarnaOracle, 4, 3)
	attacker := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 3, 3)
	hand := me.Hand.Size()
	before := lifeOfOpponents(g)
	declareAttack(t, g, opp.ID, attacker)
	b35Block(t, g, wall, attacker)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Fatal("the Bear should have died to the Wall")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("an attacking creature dying draws: hand %d → %d", hand, me.Hand.Size())
	}
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d: %d → %d, want unchanged for an attacker", i+1, b, got)
		}
	}
	// After combat, a creature that attacked this turn is no longer
	// attacking, and a creature that never attacked never was.
	advanceTo(t, g, game.StepPostcombatMain)
	bystander := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	b18Kill(t, g, bystander)
	if me.Hand.Size() != hand+1 {
		t.Error("a non-attacker dying does not draw")
	}
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want -1 from Garna", i+1, b, got)
		}
	}
	// An opponent's creature, and Garna herself, are silent.
	b18Kill(t, g, wall)
	b18Kill(t, g, garna)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want still -1", i+1, b, got)
		}
	}
}

func TestB35EdenTapsForColorlessMillsTwoAndSacrificesToRegrowAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	eden := pushCatalogPermanent(g, me.ID, "Eden, Seat of the Sanctum", "Land — Town", b35EdenOracle, false)
	if err := g.ActivateManaAbility(me.ID, eden, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	b08Untap(g, eden)
	me.ManaPool = nil
	// The plain mill.
	lib := me.Library.Size()
	b09AddColorless(me, 5)
	b16Activate(t, g, me.ID, eden, 0, game.ActivateAbilityParams{})
	if me.Library.Size() != lib-2 || !g.Battlefield.Contains(eden) {
		t.Errorf("mill two and keep the land: library %d → %d", lib, me.Library.Size())
	}
	b08Untap(g, eden)
	// The sacrifice: a permanent card comes back, an instant is not
	// offered, and Eden goes to the graveyard as a cost.
	rock := b17GraveyardCard(me, "Dead Rock", "Artifact", "{1}")
	bolt := b17GraveyardCard(me, "Dead Bolt", "Instant", "{R}")
	b09AddColorless(me, 5)
	if err := g.ActivateCatalogAbility(me.ID, eden, 1, game.ActivateAbilityParams{Targets: cardRefs(bolt)}); err == nil {
		t.Fatal("an instant is not a permanent card")
	}
	b16Activate(t, g, me.ID, eden, 1, game.ActivateAbilityParams{Targets: cardRefs(rock)})
	if me.Library.Size() != lib-4 {
		t.Errorf("the sacrifice half mills two too: library %d, want %d", me.Library.Size(), lib-4)
	}
	if !me.Hand.Contains(rock) {
		t.Error("the permanent card did not come back to hand")
	}
	if g.Battlefield.Contains(eden) || !me.Graveyard.Contains(eden) {
		t.Error("Eden is sacrificed as a cost")
	}
	if spec, _ := Lookup(b35EdenOracle); spec.Completeness != CompletenessCaveats || len(spec.Activated) != 2 {
		t.Error("the split activation is a declared gap")
	}
}

func TestB35IkraGainsTheConnectingCreaturesToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ikra := b12Push(g, me.ID, "Ikra Shidiqi, the Usurper", "Legendary Creature — Snake Wizard", b35IkraShidiqiOracle, 3, 7)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	blocked := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 4)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 6)
	assertKeywords(t, g, ikra, "menace")
	life := me.Life
	declareAttack(t, g, opp.ID, ikra, bear, blocked)
	b35Block(t, g, wall, blocked)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if me.Life != life+7+2 {
		t.Errorf("Ikra's 7 and the Bear's 2, the blocked Elf nothing: life %d → %d", life, me.Life)
	}
}

func TestB35ScreamingNemesisReflectsDamageToAnyOtherTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nemesis := b12Push(g, me.ID, "Screaming Nemesis", "Creature — Spirit", b35ScreamingNemesisOracle, 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	assertKeywords(t, g, nemesis, "haste")
	life := opp.Life
	b27Damage(g, theirs, nemesis, 1)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("1 damage reflected: %d → %d", life, opp.Life)
	}
	// Choosing the Nemesis itself does nothing — "any OTHER target".
	b27Damage(g, theirs, nemesis, 1)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, nemesis)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, nemesis); got != 2 {
		t.Errorf("damage marked %d, want the 2 it was dealt and none reflected onto itself", got)
	}
	// Lethal damage: the Nemesis dies, and still deals it.
	b27Damage(g, theirs, nemesis, 5)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(nemesis) {
		t.Fatal("seven damage on a 3/3")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the dead Nemesis still deals its 5 to the chosen creature")
	}
	if spec, _ := Lookup(b35ScreamingNemesisOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the missing no-lifegain clause is a declared gap")
	}
}

func TestB35WildwoodScourgeEntersWithXAndGrowsWithOtherNonHydras(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scourge := castXSpell(t, g, "Wildwood Scourge", "Creature — Hydra", b35WildwoodScourgeOracle, "{X}{G}", 3, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, scourge, game.CounterPlusOne); got != 3 {
		t.Fatalf("enters with X = 3 counters, got %d", got)
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b28AddCounter(t, g, bear, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, scourge, game.CounterPlusOne); got != 4 {
		t.Errorf("two counters on another creature is ONE counter on the Scourge: %d", got)
	}
	hydra := b12Creature(g, me.ID, "My Hydra", "Creature — Hydra", 2, 2)
	b28AddCounter(t, g, hydra, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b28AddCounter(t, g, theirs, 1)
	b28AddCounter(t, g, scourge, 1)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, scourge, game.CounterPlusOne); got != 5 {
		t.Errorf("a Hydra, an opponent's creature and itself are silent: %d, want 5", got)
	}
	// A removal is not a placement.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, -1) })
	passPriorityAroundTable(t, g)
	if got := counterCount(g, scourge, game.CounterPlusOne); got != 5 {
		t.Errorf("removing a counter is silent: %d", got)
	}
}

func TestB35NekusarGivesEveryoneAnExtraDrawAndPingsOpponentsPerCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Nekusar, the Mindrazer", "Legendary Creature — Zombie Wizard", b35NekusarOracle, 2, 4)
	hand, life := opp.Hand.Size(), opp.Life
	advanceToStepOf(t, g, 1, game.StepDraw)
	// The turn's draw pings and the draw step's own trigger draws:
	// two triggers from one source at once, so the controller orders
	// them (CR 603.3b); the extra draw then pings on its own.
	b30SettleOrder(t, g)
	if opp.Hand.Size() != hand+2 {
		t.Errorf("the turn's draw and an additional card: hand %d → %d", hand, opp.Hand.Size())
	}
	if opp.Life != life-2 {
		t.Errorf("1 damage per card drawn: %d → %d", life, opp.Life)
	}
	// Every other player's draw step fires the same pair; settle each
	// as it comes, since AdvanceStep never drains the stack.
	for seat := 2; seat < 4; seat++ {
		advanceToStepOf(t, g, seat, game.StepDraw)
		b30SettleOrder(t, g)
	}
	// Your own draw step: the extra card, no damage.
	myHand, myLife := me.Hand.Size(), me.Life
	advanceToStepOf(t, g, 0, game.StepDraw)
	b30SettleOrder(t, g)
	if me.Hand.Size() != myHand+2 || me.Life != myLife {
		t.Errorf("your own draws: hand %d → %d, life %d → %d", myHand, me.Hand.Size(), myLife, me.Life)
	}
}

func TestB35GreenwardenRegrowsOnEntryAndAgainByExilingItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	first := b17GraveyardCard(me, "Dead Rock", "Artifact", "{1}")
	second := b17GraveyardCard(me, "Dead Bolt", "Instant", "{R}")
	warden := castCatalogSpell(t, g, "Greenwarden of Murasa", "Creature — Elemental", b35GreenwardenOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, first)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(first) {
		t.Fatal("the entry did not return the chosen card")
	}
	b18Kill(t, g, warden)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, warden) || !hasID(p.PickTargetCards, second) {
		t.Errorf("any card in your graveyard, the Greenwarden included: %v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, second)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(warden) {
		t.Error("the Greenwarden is exiled from the graveyard")
	}
	if !me.Hand.Contains(second) {
		t.Error("and the chosen card comes back")
	}
}

func TestB35GreenwardenDeclinedStaysInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := b17GraveyardCard(me, "Dead Rock", "Artifact", "{1}")
	warden := castCatalogSpell(t, g, "Greenwarden of Murasa", "Creature — Elemental", b35GreenwardenOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	b18Kill(t, g, warden)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(warden) || !me.Graveyard.Contains(card) {
		t.Error("declining both leaves everything where it was")
	}
}

func TestB35LuxArtilleryFiresAtThirtyCountersOnYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Lux Artillery", "Artifact", b35LuxArtilleryOracle, false)
	pushCounterCreature(g, me.ID, "Big Bear", game.CounterPlusOne, 20)
	rock := b12Permanent(g, me.ID, "Charged Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(rock, "charge", 9) })
	before := lifeOfOpponents(g)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d at 29 counters: %d → %d, want unchanged", i+1, b, got)
		}
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(rock, "charge", 1) })
	advanceToEndStepOf(t, g, 1)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-10 {
			t.Errorf("opponent %d at 30 counters: %d → %d, want -10", i+1, b, got)
		}
	}
	if spec, _ := Lookup(b35LuxArtilleryOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the missing sunburst grant is a declared gap")
	}
}
