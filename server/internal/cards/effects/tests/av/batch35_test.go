package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Deadly Tempest
// was already on main (#382), so the table is the issue's 28 minus
// the four declared skips.
func TestBatch35CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B35SphinxsTutelageOracle:            "Sphinx's Tutelage",
		effectstest.B35RiskyShortcutOracle:              "Risky Shortcut",
		effectstest.B35SealOfPrimordiumOracle:           "Seal of Primordium",
		effectstest.B35ScaledNurturerOracle:             "Scaled Nurturer",
		effectstest.B35AetherGaleOracle:                 "Aether Gale",
		effectstest.B35PlagueBelcherOracle:              "Plague Belcher",
		effectstest.B35MemoryErosionOracle:              "Memory Erosion",
		effectstest.B35SmileAtDeathOracle:               "Smile at Death",
		effectstest.B35LordOfExtinctionOracle:           "Lord of Extinction",
		effectstest.B35TheNecrobloomOracle:              "The Necrobloom",
		effectstest.B35GarnaOracle:                      "Garna, Bloodfist of Keld",
		effectstest.B35DreadSummonsOracle:               "Dread Summons",
		effectstest.B35AbandonAttachmentsOracle:         "Abandon Attachments",
		effectstest.B35LuxArtilleryOracle:               "Lux Artillery",
		effectstest.B35AbsorbOracle:                     "Absorb",
		effectstest.B35EdenOracle:                       "Eden, Seat of the Sanctum",
		effectstest.B35IkraShidiqiOracle:                "Ikra Shidiqi, the Usurper",
		effectstest.B35ScreamingNemesisOracle:           "Screaming Nemesis",
		effectstest.B35WildwoodScourgeOracle:            "Wildwood Scourge",
		effectstest.B35NekusarOracle:                    "Nekusar, the Mindrazer",
		effectstest.B35ThirstForDiscoveryOracle:         "Thirst for Discovery",
		effectstest.B35FinalActOracle:                   "Final Act",
		effectstest.B35GreenwardenOracle:                "Greenwarden of Murasa",
		effectstest.B35DeadlyTempestAlreadyOnMainOracle: "Deadly Tempest",
	}
	if len(want) != 24 {
		t.Fatalf("the batch ships 23 cards plus Deadly Tempest, the table lists %d", len(want))
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
			t.Errorf("%s ships unreviewed", name)
		}
	}
}

// --- spells --------------------------------------------------------

func TestB35RiskyShortcutDrawsTwoAndEveryPlayerLosesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := b35Lives(g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Risky Shortcut", "Sorcery", effectstest.B35RiskyShortcutOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := effectstest.B13OpponentCasts(t, g, opp, "Their Opt", "Instant", "", "{U}", nil)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Absorb", "Instant", effectstest.B35AbsorbOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Error("the spell was not countered")
	}
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

func TestB35AetherGaleBouncesExactlySixNonlandPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	a := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	c := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	// Five nonland permanents and a land: not a legal cast.
	if err := b09TryCast(t, g, "Aether Gale", "Sorcery", effectstest.B35AetherGaleOracle,
		effectstest.CardRefs(mine, rock, a, b, c, land)); err == nil {
		t.Fatal("a land was accepted among the six")
	}
	d := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Aether Gale", "Sorcery", effectstest.B35AetherGaleOracle,
		effectstest.CardRefs(mine, rock, a, b, c, d))
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep", TypeLine: "Sorcery"},
	)
	effectstest.SeedSearchLibrary(opp,
		game.Card{Name: "Elf", TypeLine: "Creature — Elf"},
		game.Card{Name: "Wurm", TypeLine: "Creature — Wurm"},
		game.Card{Name: "Deep", TypeLine: "Sorcery"},
	)
	// A one-card library mills its one card and nobody loses.
	effectstest.SeedSearchLibrary(other, game.Card{Name: "Lone Bolt", TypeLine: "Instant"})
	effectstest.CastXSpell(t, g, "Dread Summons", "Sorcery", effectstest.B35DreadSummonsOracle, "{X}{B}{B}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != 1 || opp.Library.Size() != 1 || other.Library.Size() != 0 {
		t.Errorf("libraries %d/%d/%d, want 1/1/0", me.Library.Size(), opp.Library.Size(), other.Library.Size())
	}
	if other.AttemptedEmptyDraw {
		t.Error("a mill never loses a player the game")
	}
	zombies := effectstest.BattlefieldIDsNamed(g, "Zombie")
	if len(zombies) != 3 {
		t.Fatalf("three creature cards milled: %d Zombies", len(zombies))
	}
	for _, id := range zombies {
		c, _ := effectstest.BattlefieldCard(g, id)
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

func TestB35AbandonAttachmentsMayDiscardAndIfSoDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fodder := effectstest.HandCardFull(me, "Fodder", "Sorcery", "", "", nil)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Abandon Attachments", "Instant — Lesson", effectstest.B35AbandonAttachmentsOracle, nil)
	if !me.Hand.Contains(fodder) {
		t.Fatal("nothing is discarded at cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil || c.ChooseMin != 0 {
		t.Fatalf("a 'may discard' prompt with a floor of zero: %+v", c)
	}
	effectstest.AnswerDiscard(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(fodder) {
		t.Error("the chosen card is discarded")
	}
	// hand before + the spell, minus the spell, minus the fodder, plus two.
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1 net", hand, me.Hand.Size())
	}
}

// "If you do": declining the discard draws nothing, and the card stays.
func TestB35AbandonAttachmentsDeclinedDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fodder := effectstest.HandCardFull(me, "Fodder", "Sorcery", "", "", nil)
	effectstest.CastCatalogSpell(t, g, "Abandon Attachments", "Instant — Lesson", effectstest.B35AbandonAttachmentsOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(fodder) || me.Hand.Size() != hand {
		t.Errorf("declined: hand %d -> %d, fodder kept %v", hand, me.Hand.Size(), me.Hand.Contains(fodder))
	}
	if spec, _ := effects.Lookup(effectstest.B35AbandonAttachmentsOracle); spec.Completeness != effects.CompletenessFull || spec.AdditionalCost != nil {
		t.Error("the discard is an effect now, not a cost")
	}
}

// Castable with no other card in hand, and countering it keeps the card.
func TestB35AbandonAttachmentsCastableWithAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	effectstest.CastCatalogSpell(t, g, "Abandon Attachments", "Instant — Lesson", effectstest.B35AbandonAttachmentsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil || me.Hand.Size() != 0 {
		t.Errorf("an empty hand discards nothing and draws nothing: hand %d", me.Hand.Size())
	}
}

func TestB35ThirstForDiscoveryDrawsThreeAndOwesTwoDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Thirst for Discovery", "Instant", effectstest.B35ThirstForDiscoveryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("drew %d, want 3", me.Hand.Size()-hand)
	}
	if effectstest.DiscardOwed(g, me.ID) != 2 {
		t.Errorf("owes %d discards, want 2", effectstest.DiscardOwed(g, me.ID))
	}
	if spec, _ := effects.Lookup(effectstest.B35ThirstForDiscoveryOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("the basic-land alternative is offered: nothing is left to declare")
	}
}

// "Discard two cards unless you discard a basic land card": one basic
// land card is a complete answer; a nonbasic land is not.
func TestB35ThirstForDiscoveryAcceptsOneBasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.HandCardFull(me, "Forest", "Basic Land — Forest", "", "", nil)
	temple := effectstest.HandCardFull(me, "Temple", "Land", "", "", nil)
	effectstest.CastCatalogSpell(t, g, "Thirst for Discovery", "Instant", effectstest.B35ThirstForDiscoveryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{temple}); err == nil {
		t.Fatal("a nonbasic land alone was accepted")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("one basic land card is a complete discard: %v", err)
	}
	if me.Hand.Contains(forest) || !me.Hand.Contains(temple) {
		t.Error("only the basic land was discarded")
	}
}

func TestB35FinalActSweepsTheChosenModesInPrintedOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	golem := effectstest.PushIndestructibleWipeCreature(g, opp.ID, "Darksteel Golem")
	walker := effectstest.PushWalkerForTest(g, opp.ID, "Jace", "", 3)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	dead := effectstest.B17GraveyardCard(opp, "Dead Elf", "Creature — Elf", "{G}")
	if err := g.AddPlayerCounter(opp.ID, game.CounterPoison, 3); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayerCounter(other.ID, game.CounterEnergy, 2); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayerCounter(me.ID, game.CounterPoison, 1); err != nil {
		t.Fatal(err)
	}
	effectstest.CastModal(t, g, "Final Act", "Sorcery", effectstest.B35FinalActOracle, []int{0, 3, 4}, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.CastModal(t, g, "Final Act", "Sorcery", effectstest.B35FinalActOracle, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("destroy all planeswalkers")
	}
	if !g.Battlefield.Contains(golem) || !g.Battlefield.Contains(rock) {
		t.Error("only the chosen mode happens")
	}
}

// --- permanents ----------------------------------------------------

func TestB35SealOfPrimordiumSacrificesToDestroyAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seal := effectstest.PushCatalogPermanent(g, me.ID, "Seal of Primordium", "Enchantment", effectstest.B35SealOfPrimordiumOracle, false)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, seal, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(bear),
	}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	effectstest.B16Activate(t, g, me.ID, seal, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock)})
	if g.Battlefield.Contains(seal) || !me.Graveyard.Contains(seal) {
		t.Error("the Seal is sacrificed as a cost")
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
}

func TestB35ScaledNurturerTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	nurturer := effectstest.PushCatalogPermanent(g, me.ID, "Scaled Nurturer", "Creature — Dragon Druid", effectstest.B35ScaledNurturerOracle, false)
	if err := g.ActivateManaAbility(me.ID, nurturer, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
	if spec, _ := effects.Lookup(effectstest.B35ScaledNurturerOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("Scaled Nurturer is not full: its life-gain rider fires whenever its mana pays (#3844)")
	}
}

func TestB35MemoryErosionMillsAnOpponentWhoCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Memory Erosion", "Enchantment", effectstest.B35MemoryErosionOracle, false)
	lib := opp.Library.Size()
	effectstest.B13OpponentCasts(t, g, opp, "Their Opt", "Instant", "", "{U}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-2 {
		t.Errorf("library %d → %d, want -2", lib, opp.Library.Size())
	}
	mine := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "My Opt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != mine {
		t.Error("your own spells are silent")
	}
	// A one-card library mills its one card and nobody loses.
	effectstest.SeedSearchLibrary(opp, game.Card{Name: "Last", TypeLine: "Sorcery"})
	effectstest.B13OpponentCasts(t, g, opp, "Their Opt", "Instant", "", "{U}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != 0 || opp.AttemptedEmptyDraw {
		t.Error("a mill never loses a player the game")
	}
}

func TestB35SphinxsTutelageMillsTwoAndRepeatsWhileTheyShareAColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	tutelage := effectstest.PushCatalogPermanent(g, me.ID, "Sphinx's Tutelage", "Enchantment", effectstest.B35SphinxsTutelageOracle, false)
	// Top-first: two blue nonland cards (repeat), then a blue card
	// and a land (stop), then filler.
	effectstest.SeedSearchLibrary(opp,
		b35LibraryCard("Opt", "Instant", "U"),
		b35LibraryCard("Ponder", "Sorcery", "U"),
		b35LibraryCard("Brainstorm", "Instant", "U"),
		b35LibraryCard("Island", "Basic Land — Island"),
		b35LibraryCard("Counterspell", "Instant", "U"),
		b35LibraryCard("Negate", "Instant", "U"),
	)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetPlayers, me.ID) {
		t.Error("target OPPONENT")
	}
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != 2 {
		t.Errorf("two blue cards repeat, a land stops: library %d, want 2", opp.Library.Size())
	}
	// A pair sharing no colour stops at once.
	effectstest.SeedSearchLibrary(opp,
		b35LibraryCard("Bolt", "Instant", "R"),
		b35LibraryCard("Opt", "Instant", "U"),
		b35LibraryCard("Ponder", "Sorcery", "U"),
		b35LibraryCard("Preordain", "Sorcery", "U"),
	)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != 2 {
		t.Errorf("red and blue share nothing: library %d, want 2", opp.Library.Size())
	}
	// The loot: draw, then a discard is owed — and the draw fires the
	// trigger again.
	effectstest.B10AddMana(me, "U", "C", "C", "C", "C", "C")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, tutelage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Errorf("hand %d → %d, owes %d discards; want +1 and 1", hand, me.Hand.Size(), effectstest.DiscardOwed(g, me.ID))
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != 0 {
		t.Errorf("the looted card mills two more: library %d", opp.Library.Size())
	}
	if opp.AttemptedEmptyDraw {
		t.Error("a mill never loses a player the game")
	}
}

func TestB35PlagueBelcherShrinksACreatureYouControlAndDrainsOnZombieDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	belcher := effectstest.B20CastCreature(t, g, me, "Plague Belcher", "Creature — Zombie Beast", effectstest.B35PlagueBelcherOracle, 5, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, theirs) || !effectstest.HasID(p.PickTargetCards, belcher) {
		t.Error("target creature YOU control — the Belcher itself is the usual answer")
	}
	effectstest.PickCard(t, g, me.ID, belcher)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, belcher, game.CounterMinusOne); got != 2 {
		t.Errorf("two -1/-1 counters, got %d", got)
	}
	if effectstest.CurrentPower(t, g, belcher) != 3 {
		t.Errorf("a 5/4 with two -1/-1 counters is a 3/2: power %d", effectstest.CurrentPower(t, g, belcher))
	}
	effectstest.AssertKeywords(t, g, belcher, "menace")
	// A Zombie of yours dying drains each opponent; a non-Zombie, an
	// opponent's Zombie and the Belcher itself do not.
	before := effectstest.LifeOfOpponents(g)
	zombie := effectstest.B12Creature(g, me.ID, "My Zombie", "Creature — Zombie", 2, 2)
	effectstest.B18Kill(t, g, zombie)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want -1", i+1, b, got)
		}
	}
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B18Kill(t, g, bear)
	theirZombie := effectstest.B12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	effectstest.B18Kill(t, g, theirZombie)
	effectstest.B18Kill(t, g, belcher)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want still -1", i+1, b, got)
		}
	}
}

func TestB35SmileAtDeathReturnsUpToTwoSmallCreaturesWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Smile at Death", "Enchantment", effectstest.B35SmileAtDeathOracle, false)
	bear := effectstest.PushGraveyardCardForTest(me, "Dead Bear")
	wurm := effectstest.B17GraveyardCard(me, "Dead Wurm", "Creature — Wurm", "{4}{G}")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == wurm {
				me.Graveyard.Cards[i].Power, me.Graveyard.Cards[i].Toughness = 5, 5
			}
		}
	})
	elf := effectstest.B17GraveyardCard(me, "Dead Elf", "Creature — Elf", "{G}")
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, wurm) {
		t.Error("power 2 or less")
	}
	if !effectstest.HasID(p.PickTargetCards, bear) || !effectstest.HasID(p.PickTargetCards, elf) {
		t.Errorf("both small creature cards are offered: %v", p.PickTargetCards)
	}
	effectstest.B17PickCards(t, g, me.ID, bear, elf)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, elf} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s did not return to the battlefield", id)
			continue
		}
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("each returned creature gets a +1/+1 counter: %d", got)
		}
	}
	if !me.Graveyard.Contains(wurm) {
		t.Error("the Wurm stays")
	}
}

func TestB35LordOfExtinctionIsAsBigAsAllGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushGraveyardCardForTest(me, "Mine")
	effectstest.PushGraveyardCardForTest(opp, "Theirs")
	effectstest.PushGraveyardCardForTest(opp, "Theirs Too")
	lord := effectstest.B12Push(g, me.ID, "Lord of Extinction", "Creature — Elemental", effectstest.B35LordOfExtinctionOracle, 0, 0)
	if effectstest.EffectivePower(t, g, lord) != 3 || effectstest.EffectiveToughness(t, g, lord) != 3 {
		t.Errorf("three cards in graveyards: %d/%d", effectstest.EffectivePower(t, g, lord), effectstest.EffectiveToughness(t, g, lord))
	}
	// A creature dying is a battlefield exit and shows at once.
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B18Kill(t, g, bear)
	if effectstest.EffectivePower(t, g, lord) != 4 {
		t.Errorf("a dead creature counts: power %d, want 4", effectstest.EffectivePower(t, g, lord))
	}
	// The late read on a mill was a declared gap until #1117. A
	// graveyard crossing now bumps the layer version on its own, so
	// the Lord grows the instant a card is milled, with nothing on the
	// battlefield moving in between.
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Milled", TypeLine: "Instant", Owner: me.ID})
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})
	if got := effectstest.EffectivePower(t, g, lord); got != 5 {
		t.Errorf("a milled card counts at once: power %d, want 5", got)
	}
	if spec, _ := effects.Lookup(effectstest.B35LordOfExtinctionOracle); spec.Completeness != effects.CompletenessFull {
		t.Errorf("Completeness = %q, want full now that the late read is gone", spec.Completeness)
	}
}

func TestB35TheNecrobloomMakesAPlantOrAZombieOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "The Necrobloom", "Legendary Creature — Plant", effectstest.B35TheNecrobloomOracle, 2, 7)
	for _, name := range []string{"Forest", "Swamp", "Plains", "Island", "Mountain"} {
		effectstest.SeedLandOnBattlefield(g, me.ID, name, "Basic Land — "+name)
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	plants := effectstest.BattlefieldIDsNamed(g, "Plant")
	if len(plants) != 1 {
		t.Fatalf("six lands, five names: %d Plants", len(plants))
	}
	if c, _ := effectstest.BattlefieldCard(g, plants[0]); c.Power != 0 || c.Toughness != 1 || !c.HasColor("G") {
		t.Error("a 0/1 green Plant")
	}
	// The seventh differently-named land makes a Zombie instead.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Wastes", TypeLine: "Token Land — Wastes"}, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Plant")); n != 2 {
		t.Errorf("six names: still a Plant, %d Plants", n)
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Badlands", TypeLine: "Token Land — Swamp Mountain"}, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Zombie")); n != 1 {
		t.Errorf("seven names: a Zombie instead, %d Zombies", n)
	}
	if n := len(effectstest.BattlefieldIDsNamed(g, "Plant")); n != 2 {
		t.Errorf("instead, not as well: %d Plants", n)
	}
	if spec, _ := effects.Lookup(effectstest.B35TheNecrobloomOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("dredge landed in #2127, so the Necrobloom has no declared gap left")
	}
}

func TestB35GarnaDrawsForAnAttackerDyingAndPingsForAnyOtherDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	garna := effectstest.B12Push(g, me.ID, "Garna, Bloodfist of Keld", "Legendary Creature — Human Berserker", effectstest.B35GarnaOracle, 4, 3)
	attacker := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 3, 3)
	hand := me.Hand.Size()
	before := effectstest.LifeOfOpponents(g)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.B35Block(t, g, wall, attacker)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	bystander := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	effectstest.B18Kill(t, g, bystander)
	if me.Hand.Size() != hand+1 {
		t.Error("a non-attacker dying does not draw")
	}
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want -1 from Garna", i+1, b, got)
		}
	}
	// An opponent's creature, and Garna herself, are silent.
	effectstest.B18Kill(t, g, wall)
	effectstest.B18Kill(t, g, garna)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want still -1", i+1, b, got)
		}
	}
}

// TestB35EdenTapsForColorlessMillsTwoAndSacrificesToRegrowAPermanent
// covers the #796 shape end to end: one activation, the mill, the
// free yes/no asked AFTER the mill, and the CR 603.12 reflexive
// trigger on the "if you do" branch.
//
// The two things that used to be caveats are the two assertions that
// matter here: the decline leaves the land on the battlefield with
// the mana already spent (the decision is made after the cards are
// seen, so declining has to be a real answer and not a no-op), and
// the target is picked after the sacrifice, so a permanent card the
// mill itself just put in the graveyard is a legal pick.
func TestB35EdenTapsForColorlessMillsTwoAndSacrificesToRegrowAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	eden := effectstest.PushCatalogPermanent(g, me.ID, "Eden, Seat of the Sanctum", "Land — Town", effectstest.B35EdenOracle, false)
	if err := g.ActivateManaAbility(me.ID, eden, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	effectstest.B08Untap(g, eden)
	me.ManaPool = nil

	// Declining: the mill happens, the question is asked, and the
	// land stays.
	lib := me.Library.Size()
	effectstest.B09AddColorless(me, 5)
	effectstest.B16Activate(t, g, me.ID, eden, 0, game.ActivateAbilityParams{})
	if me.Library.Size() != lib-2 {
		t.Errorf("mill two: library %d → %d", lib, me.Library.Size())
	}
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatalf("no sacrifice question for the controller: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm(keep): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(eden) {
		t.Error("declining keeps the land")
	}

	// Accepting: the land is sacrificed and the reflexive trigger
	// picks a permanent card from the graveyard.
	effectstest.B08Untap(g, eden)
	rock := effectstest.B17GraveyardCard(me, "Dead Rock", "Artifact", "{1}")
	bolt := effectstest.B17GraveyardCard(me, "Dead Bolt", "Instant", "{R}")
	effectstest.B09AddColorless(me, 5)
	effectstest.B16Activate(t, g, me.ID, eden, 0, game.ActivateAbilityParams{})
	if me.Library.Size() != lib-4 {
		t.Errorf("the second activation mills two too: library %d, want %d", me.Library.Size(), lib-4)
	}
	ask = effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatalf("no sacrifice question: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm(sacrifice): %v", err)
	}
	if g.Battlefield.Contains(eden) || !me.Graveyard.Contains(eden) {
		t.Error("accepting sacrifices the land")
	}
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("the reflexive trigger did not ask for a target: %+v", g.PendingChoices)
	}
	// "another target PERMANENT card": the instant is not offered,
	// and neither is Eden, which is in the graveyard by now.
	if effectstest.HasID(pick.PickTargetCards, bolt) {
		t.Error("an instant is not a permanent card")
	}
	if effectstest.HasID(pick.PickTargetCards, eden) {
		t.Error(`"another" excludes Eden itself`)
	}
	if !effectstest.HasID(pick.PickTargetCards, rock) {
		t.Fatalf("the artifact is not offered: %v", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the permanent card did not come back to hand")
	}
	if spec, _ := effects.Lookup(effectstest.B35EdenOracle); spec.Completeness != effects.CompletenessFull || len(spec.Activated) != 1 {
		t.Error("#796 closed the split: one activation, no caveats")
	}
}

func TestB35IkraGainsTheConnectingCreaturesToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ikra := effectstest.B12Push(g, me.ID, "Ikra Shidiqi, the Usurper", "Legendary Creature — Snake Wizard", effectstest.B35IkraShidiqiOracle, 3, 7)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	blocked := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 4)
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 6)
	effectstest.AssertKeywords(t, g, ikra, "menace")
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, ikra, bear, blocked)
	effectstest.B35Block(t, g, wall, blocked)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+7+2 {
		t.Errorf("Ikra's 7 and the Bear's 2, the blocked Elf nothing: life %d → %d", life, me.Life)
	}
}

func TestB35ScreamingNemesisReflectsDamageToAnyOtherTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nemesis := effectstest.B12Push(g, me.ID, "Screaming Nemesis", "Creature — Spirit", effectstest.B35ScreamingNemesisOracle, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AssertKeywords(t, g, nemesis, "haste")
	life := opp.Life
	effectstest.B27Damage(g, theirs, nemesis, 1)
	g.RunStateChecksForTest() // #3172: "dealt damage" triggers wait for the batch to settle
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("1 damage reflected: %d → %d", life, opp.Life)
	}
	// Choosing the Nemesis itself does nothing — "any OTHER target".
	effectstest.B27Damage(g, theirs, nemesis, 1)
	g.RunStateChecksForTest() // #3172: "dealt damage" triggers wait for the batch to settle
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, nemesis)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, nemesis); got != 2 {
		t.Errorf("damage marked %d, want the 2 it was dealt and none reflected onto itself", got)
	}
	// Lethal damage: the Nemesis dies, and still deals it.
	effectstest.B27Damage(g, theirs, nemesis, 5)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(nemesis) {
		t.Fatal("seven damage on a 3/3")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the dead Nemesis still deals its 5 to the chosen creature")
	}
	if spec, _ := effects.Lookup(effectstest.B35ScreamingNemesisOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the no-lifegain clause and the batched damage both landed (#3172 cav-d)")
	}
}

func TestB35WildwoodScourgeEntersWithXAndGrowsWithOtherNonHydras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scourge := effectstest.CastXSpell(t, g, "Wildwood Scourge", "Creature — Hydra", effectstest.B35WildwoodScourgeOracle, "{X}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, scourge, game.CounterPlusOne); got != 3 {
		t.Fatalf("enters with X = 3 counters, got %d", got)
	}
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B28AddCounter(t, g, bear, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, scourge, game.CounterPlusOne); got != 4 {
		t.Errorf("two counters on another creature is ONE counter on the Scourge: %d", got)
	}
	hydra := effectstest.B12Creature(g, me.ID, "My Hydra", "Creature — Hydra", 2, 2)
	effectstest.B28AddCounter(t, g, hydra, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B28AddCounter(t, g, theirs, 1)
	effectstest.B28AddCounter(t, g, scourge, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, scourge, game.CounterPlusOne); got != 5 {
		t.Errorf("a Hydra, an opponent's creature and itself are silent: %d, want 5", got)
	}
	// A removal is not a placement.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, -1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, scourge, game.CounterPlusOne); got != 5 {
		t.Errorf("removing a counter is silent: %d", got)
	}
}

func TestB35NekusarGivesEveryoneAnExtraDrawAndPingsOpponentsPerCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Nekusar, the Mindrazer", "Legendary Creature — Zombie Wizard", effectstest.B35NekusarOracle, 2, 4)
	hand, life := opp.Hand.Size(), opp.Life
	effectstest.AdvanceToStepOf(t, g, 1, game.StepDraw)
	// The turn's draw pings and the draw step's own trigger draws:
	// two triggers from one source at once, so the controller orders
	// them (CR 603.3b); the extra draw then pings on its own.
	effectstest.B30SettleOrder(t, g)
	if opp.Hand.Size() != hand+2 {
		t.Errorf("the turn's draw and an additional card: hand %d → %d", hand, opp.Hand.Size())
	}
	if opp.Life != life-2 {
		t.Errorf("1 damage per card drawn: %d → %d", life, opp.Life)
	}
	// Every other player's draw step fires the same pair; settle each
	// as it comes, since AdvanceStep never drains the stack.
	for seat := 2; seat < 4; seat++ {
		effectstest.AdvanceToStepOf(t, g, seat, game.StepDraw)
		effectstest.B30SettleOrder(t, g)
	}
	// Your own draw step: the extra card, no damage.
	myHand, myLife := me.Hand.Size(), me.Life
	effectstest.AdvanceToStepOf(t, g, 0, game.StepDraw)
	effectstest.B30SettleOrder(t, g)
	if me.Hand.Size() != myHand+2 || me.Life != myLife {
		t.Errorf("your own draws: hand %d → %d, life %d → %d", myHand, me.Hand.Size(), myLife, me.Life)
	}
}

func TestB35GreenwardenRegrowsOnEntryAndAgainByExilingItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	first := effectstest.B17GraveyardCard(me, "Dead Rock", "Artifact", "{1}")
	second := effectstest.B17GraveyardCard(me, "Dead Bolt", "Instant", "{R}")
	warden := effectstest.CastCatalogSpell(t, g, "Greenwarden of Murasa", "Creature — Elemental", effectstest.B35GreenwardenOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, first)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(first) {
		t.Fatal("the entry did not return the chosen card")
	}
	effectstest.B18Kill(t, g, warden)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, warden) || !effectstest.HasID(p.PickTargetCards, second) {
		t.Errorf("any card in your graveyard, the Greenwarden included: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, second)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(warden) {
		t.Error("the Greenwarden is exiled from the graveyard")
	}
	if !me.Hand.Contains(second) {
		t.Error("and the chosen card comes back")
	}
}

func TestB35GreenwardenDeclinedStaysInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := effectstest.B17GraveyardCard(me, "Dead Rock", "Artifact", "{1}")
	warden := effectstest.CastCatalogSpell(t, g, "Greenwarden of Murasa", "Creature — Elemental", effectstest.B35GreenwardenOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B18Kill(t, g, warden)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(warden) || !me.Graveyard.Contains(card) {
		t.Error("declining both leaves everything where it was")
	}
}

func TestB35LuxArtilleryFiresAtThirtyCountersOnYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Lux Artillery", "Artifact", effectstest.B35LuxArtilleryOracle, false)
	effectstest.PushCounterCreature(g, me.ID, "Big Bear", game.CounterPlusOne, 20)
	rock := effectstest.B12Permanent(g, me.ID, "Charged Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(rock, "charge", 9) })
	before := effectstest.LifeOfOpponents(g)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d at 29 counters: %d → %d, want unchanged", i+1, b, got)
		}
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(rock, "charge", 1) })
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-10 {
			t.Errorf("opponent %d at 30 counters: %d → %d, want -10", i+1, b, got)
		}
	}
	if spec, _ := effects.Lookup(effectstest.B35LuxArtilleryOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("Lux Artillery is complete since a waived payment asks (#3855)")
	}
}
