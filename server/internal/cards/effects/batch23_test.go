package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch23_test.go — card-level coverage for the card-coverage
// roadmap's batch 23 (#385, `edhrec_rank` 2436–2535): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, draw or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b23-prefixed.

const (
	b23HordelingOutburstAlreadyOID = "a6450b8e-eb18-431c-9eb7-7daf107978b2"
	b23BorealDruidOracle           = "2fcc69ff-8ab5-4e14-afe3-db892049a872"
	b23DiscipleOfTheVaultOracle    = "c8625113-0ce4-4454-83a1-25c31b8bfb9a"
	b23PrimordialHydraOracle       = "1c36ed3a-c806-47e5-83f9-e44999c67fe5"
	b23IvyLaneDenizenOracle        = "a5da5ad6-4ed2-4041-a983-76a8c87fa109"
	b23EncroachingDragonstormOID   = "1e95c273-7fec-4f8d-8be9-e21ad4b93717"
	b23PyrohemiaOracle             = "9ac57a10-3402-4656-9079-f713884cde35"
	b23InsightEngineOracle         = "83409a22-10de-4d51-90a3-e0579ca8cbea"
	b23RegisaurAlphaOracle         = "0673f4e0-66ff-458c-b4ba-eb067e560cce"
	b23DionusOracle                = "d5e5cf55-eb9e-4f01-9056-791215411b27"
	b23ForcedFruitionOracle        = "448b27a5-7c0c-4ab6-bddf-bd62b920aacc"
	b23FangsOfKaloniaOracle        = "8d9bd4cc-564b-4fdf-9462-b4bb30583642"
	b23NivMizzetFiremindOracle     = "959acb66-84ca-4535-bca2-ad591895735e"
	b23ProteanHulkOracle           = "10180e2f-90c5-4d41-ba44-16b14948f923"
	b23RipApartOracle              = "cbdbf18f-0180-4ade-a79e-1e644dd42d6f"
	// #1210: both were declared skips — the activation gate and the
	// opponent-activation watch — and both landed with that issue.
	b23CollectorOupheOracle = "0c4bc9ea-a5fd-4f44-96a1-5448eee228c4"
	b23RunicArmasaurOracle  = "48e8ae59-a234-498e-9dae-bac8d1424ea5"
)

// b23Cast seeds a fully-specified card — colours and mana cost
// included, for a colour or mana-value check — into the active
// seat's hand and casts it from a main phase.
func b23Cast(t *testing.T, g *game.Game, name, typeLine, manaCost, oracle string, colors []string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := handCardFull(active, name, typeLine, manaCost, oracle, colors)
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// b23Draw draws n cards for a player through the effect API and
// leaves whatever it triggered pending.
func b23Draw(g *game.Game, p uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p, n) })
}

// b23Hands reads every seat's hand size.
func b23Hands(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats))
	for _, p := range g.Seats {
		out = append(out, p.Hand.Size())
	}
	return out
}

// b23TapMana activates a creature's first mana ability.
func b23TapMana(t *testing.T, g *game.Game, controller, card uuid.UUID) {
	t.Helper()
	if err := g.ActivateManaAbility(controller, card, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Hordeling
// Outburst was already on main from the S21 token work and is pinned
// here so the table matches the issue's 21.
func TestBatch23CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b23HordelingOutburstAlreadyOID: "Hordeling Outburst",
		b23BorealDruidOracle:           "Boreal Druid",
		b23DiscipleOfTheVaultOracle:    "Disciple of the Vault",
		b23PrimordialHydraOracle:       "Primordial Hydra",
		b23IvyLaneDenizenOracle:        "Ivy Lane Denizen",
		b23EncroachingDragonstormOID:   "Encroaching Dragonstorm",
		b23PyrohemiaOracle:             "Pyrohemia",
		b23InsightEngineOracle:         "Insight Engine",
		b23RegisaurAlphaOracle:         "Regisaur Alpha",
		b23DionusOracle:                "Dionus, Elvish Archdruid",
		b23ForcedFruitionOracle:        "Forced Fruition",
		b23FangsOfKaloniaOracle:        "Fangs of Kalonia",
		b23NivMizzetFiremindOracle:     "Niv-Mizzet, the Firemind",
		b23ProteanHulkOracle:           "Protean Hulk",
		b23RipApartOracle:              "Rip Apart",
		b23CollectorOupheOracle:        "Collector Ouphe",
		b23RunicArmasaurOracle:         "Runic Armasaur",
	}
	if len(want) != 17 {
		t.Fatalf("the batch registers 16 cards plus Hordeling Outburst, the table lists %d", len(want))
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
	}
	// The four remaining declared skips must stay out until their
	// seam lands: an opponent's choice at resolution, an additional
	// phase, a card put from hand at resolution, and a trigger-
	// replacement ("triggers an additional time"). Collector Ouphe
	// (the activation gate) and Runic Armasaur (the ability-
	// activation trigger event) came off this list with #1210 and
	// are in `want` above.
	for oracle, name := range map[string]string{
		"5bb32efd-9e58-4021-bbda-4ffccfa1d601": "Druid of Purification",
		"516101be-be39-4d84-8fee-d8a79930dd0a": "Sphinx of the Second Sun",
		"8d571129-9030-47e0-9624-a49fb63e5a1b": "Ilharg, the Raze-Boar",
		"95b53836-18aa-451a-992d-2a111deeeab2": "Yarok, the Desecrated",
	} {
		if _, ok := Lookup(oracle); ok {
			t.Errorf("%s is declared skipped on #385 but is registered — update the issue", name)
		}
	}
}

// --- the mana dork and the spells ----------------------------------

func TestB23BorealDruidTapsForColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	druid := b21Push(g, me.ID, "Boreal Druid", "Snow Creature — Elf Druid", b23BorealDruidOracle, 1, 1, "G")
	advanceToMain(t, g)
	b23TapMana(t, g, me.ID, druid)
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want C", got)
	}
	if !b16Tapped(t, g, druid) {
		t.Error("the Druid taps")
	}
	sick := castCatalogSpell(t, g, "Boreal Druid", "Snow Creature — Elf Druid", b23BorealDruidOracle, nil)
	passPriorityAroundTable(t, g)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a creature's tap mana ability waits out summoning sickness")
	}
	if spec, _ := Lookup(b23BorealDruidOracle); spec.Completeness != CompletenessFull {
		t.Error("a mana dork is whole")
	}
}

func TestB23RipApartBurnsOrDestroys(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	big := b16Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 5, 5, "G")
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jace", TypeLine: "Legendary Planeswalker — Jace", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	shrine := b12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	castModal(t, g, "Rip Apart", "Sorcery", b23RipApartOracle, []int{0}, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("3 damage kills the 2/2")
	}
	castModal(t, g, "Rip Apart", "Sorcery", b23RipApartOracle, []int{0}, b16TargetCard(walker))
	passPriorityAroundTable(t, g)
	if counterCount(g, walker, game.CounterLoyalty) != 1 {
		t.Errorf("3 damage to a planeswalker removes 3 loyalty: %d, want 1", counterCount(g, walker, game.CounterLoyalty))
	}
	castModal(t, g, "Rip Apart", "Sorcery", b23RipApartOracle, []int{1}, b16TargetCard(shrine))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shrine) || !g.Battlefield.Contains(rock) {
		t.Error("the chosen enchantment is destroyed, the artifact not")
	}
	// A mode's clause is enforced: the damage mode cannot aim at an
	// artifact, the destroy mode cannot aim at a creature.
	for _, tc := range []struct {
		mode   int
		target uuid.UUID
		what   string
	}{{0, rock, "damage mode at an artifact"}, {1, big, "destroy mode at a creature"}} {
		id := handCardFull(me, "Rip Apart", "Sorcery", "{R}{W}", b23RipApartOracle, nil)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{tc.mode}, Targets: b16TargetCard(tc.target)}); err == nil {
			t.Errorf("%s is refused", tc.what)
		}
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 5/5 was never touched")
	}
}

func TestB23FangsOfKaloniaGrowsThenDoubles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bare := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	laden := b16Creature(g, me.ID, "Hydra", "Creature — Hydra", 0, 0, "G")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(laden, "+1/+1", 3) })
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	castCatalogSpell(t, g, "Fangs of Kalonia", "Sorcery", b23FangsOfKaloniaOracle, b16TargetCard(bare))
	passPriorityAroundTable(t, g)
	if n := counterCount(g, bare, "+1/+1"); n != 2 {
		t.Errorf("a bare creature: one counter, doubled to 2, got %d", n)
	}
	castCatalogSpell(t, g, "Fangs of Kalonia", "Sorcery", b23FangsOfKaloniaOracle, b16TargetCard(laden))
	passPriorityAroundTable(t, g)
	if n := counterCount(g, laden, "+1/+1"); n != 8 {
		t.Errorf("three counters: one more makes 4, doubled to 8, got %d", n)
	}
	// Target creature YOU control.
	id := handCardFull(me, "Fangs of Kalonia", "Sorcery", "{1}{G}", b23FangsOfKaloniaOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(theirs)}); err == nil {
		t.Error("an opponent's creature is not a legal target")
	}
	// Overloaded: each creature you control, no target, opponents'
	// creatures untouched.
	castWithAltCost(t, g, "Fangs of Kalonia", "Sorcery", b23FangsOfKaloniaOracle, "overload")
	passPriorityAroundTable(t, g)
	if n := counterCount(g, bare, "+1/+1"); n != 6 {
		t.Errorf("overload on the 2-counter Bear: 3 doubled to 6, got %d", n)
	}
	if n := counterCount(g, laden, "+1/+1"); n != 18 {
		t.Errorf("overload on the 8-counter Hydra: 9 doubled to 18, got %d", n)
	}
	if n := counterCount(g, theirs, "+1/+1"); n != 0 {
		t.Errorf("an opponent's creature gets nothing: %d", n)
	}
	if spec, _ := Lookup(b23FangsOfKaloniaOracle); spec.Completeness != CompletenessFull {
		t.Error("whole")
	}
}

// --- the creatures: statics and ETB --------------------------------

func TestB23RegisaurAlphaMakesAHastyDinosaur(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dino := b16Creature(g, me.ID, "My Raptor", "Creature — Dinosaur", 2, 2, "R")
	bear := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Raptor", "Creature — Dinosaur", 2, 2, "R")
	alpha := castCatalogSpell(t, g, "Regisaur Alpha", "Creature — Dinosaur", b23RegisaurAlphaOracle, nil)
	passPriorityAroundTable(t, g)
	tokens := battlefieldIDsNamed(g, "Dinosaur")
	if len(tokens) != 1 {
		t.Fatalf("one Dinosaur token, got %d", len(tokens))
	}
	token := tokens[0]
	c := cardByID(g, token)
	if controllerOf(t, g, token) != me.ID || effectivePower(t, g, token) != 3 || effectiveToughness(t, g, token) != 3 || len(c.Colors) != 1 || c.Colors[0] != "G" {
		t.Error("a 3/3 green Dinosaur under your control")
	}
	if !hasEffectiveKeyword(t, g, token, "trample") {
		t.Error("with trample")
	}
	if !hasEffectiveKeyword(t, g, token, "haste") || !hasEffectiveKeyword(t, g, dino, "haste") {
		t.Error("other Dinosaurs you control have haste")
	}
	if hasEffectiveKeyword(t, g, alpha, "haste") || hasEffectiveKeyword(t, g, bear, "haste") || hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("not the Alpha itself, not a Bear, not an opponent's Dinosaur")
	}
}

func TestB23IvyLaneDenizenGrowsATargetPerGreenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	denizen := b21Push(g, me.ID, "Ivy Lane Denizen", "Creature — Elf Warrior", b23IvyLaneDenizenOracle, 2, 3, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	elf := b23Cast(t, g, "Elf", "Creature — Elf", "{G}", "", []string{"G"}, nil)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, elf) || !hasID(p.PickTargetCards, denizen) || !hasID(p.PickTargetCards, theirs) {
		t.Error("target CREATURE — the entering one, the Denizen, an opponent's")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if counterCount(g, theirs, "+1/+1") != 1 {
		t.Error("the chosen creature gets the counter")
	}
	// A green token counts; a colourless artifact creature and a
	// red creature do not; an opponent's green creature is not yours.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 green Dinosaur with trample"), 1) })
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, elf)
	passPriorityAroundTable(t, g)
	if counterCount(g, elf, "+1/+1") != 1 {
		t.Error("a green token entering triggers")
	}
	b23Cast(t, g, "Myr", "Artifact Creature — Myr", "{2}", "", nil, nil)
	passPriorityAroundTable(t, g)
	b23Cast(t, g, "Goblin", "Creature — Goblin", "{R}", "", []string{"R"}, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a colourless or red creature entering triggers nothing")
	}
	advanceToMainOf(t, g, 1)
	b23Cast(t, g, "Their Elf", "Creature — Elf", "{G}", "", []string{"G"}, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's green creature is not one you control")
	}
}

func TestB23EncroachingDragonstormRampsAndReturnsOnADragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Mountain", "Basic Land — Mountain"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Command Tower", "Land"),
	)
	storm := castCatalogSpell(t, g, "Encroaching Dragonstorm", "Enchantment", b23EncroachingDragonstormOID, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("three basics for up to two: the chooser asks")
	}
	if c.SearchMax != 2 || searchOptionNamed(g, c, "Command Tower") != uuid.Nil {
		t.Error("up to two BASIC land cards")
	}
	answerSearchNamed(t, g, me.ID, "Forest", "Mountain")
	for _, name := range []string{"Forest", "Mountain"} {
		id := findBattlefieldByName(g, name)
		if id == uuid.Nil {
			t.Fatalf("%s enters", name)
		}
		top100AssertEnteredTapped(t, g, id, name)
	}
	if !g.Battlefield.Contains(storm) {
		t.Fatal("the enchantment stays")
	}
	// A non-Dragon: nothing. An opponent's Dragon: nothing. Your
	// Dragon: back to hand.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(storm) {
		t.Error("a Bear is not a Dragon")
	}
	advanceToMainOf(t, g, 1)
	castCatalogSpell(t, g, "Their Dragon", "Creature — Dragon", "", nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(storm) {
		t.Error("an opponent's Dragon is not one you control")
	}
	advanceToMainOf(t, g, 0)
	castCatalogSpell(t, g, "Dragon", "Creature — Dragon", "", nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(storm) || !me.Hand.Contains(storm) {
		t.Error("a Dragon you control entering returns the enchantment to your hand")
	}
	_ = opp
}

func TestB23ProteanHulkFetchesCreaturesWithinSixManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hulk := b21Push(g, me.ID, "Protean Hulk", "Creature — Beast", b23ProteanHulkOracle, 6, 6, "G")
	ids := seedSearchLibrary(me,
		game.Card{Name: "Elf", TypeLine: "Creature — Elf", ManaCost: "{G}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Hydra", TypeLine: "Creature — Hydra", ManaCost: "{X}{G}{G}"},
		game.Card{Name: "Titan", TypeLine: "Creature — Giant", ManaCost: "{4}{G}{G}"},
		game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"},
	)
	elf, bear, hydra, titan, bolt := ids[0], ids[1], ids[2], ids[3], ids[4]
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(hulk) })
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		t.Fatal("a bounce is not a death")
	}
	hulk = b21Push(g, me.ID, "Protean Hulk", "Creature — Beast", b23ProteanHulkOracle, 6, 6, "G")
	b18Kill(t, g, hulk)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("dying: the chooser asks")
	}
	if c.SearchMax < 4 {
		t.Errorf("any number: max %d", c.SearchMax)
	}
	for _, id := range []uuid.UUID{elf, bear, hydra, titan} {
		if !hasID(c.SearchCards, id) {
			t.Errorf("%s (a creature card) is offered", cardByID(g, id).Name)
		}
	}
	if hasID(c.SearchCards, bolt) {
		t.Error("an instant is not")
	}
	// Fail to find is legal: nothing taken, the library shuffled.
	answerSearchFailToFind(t, g, me.ID)
	if me.Library.Size() != 5 {
		t.Error("nothing taken")
	}
	// Again. Total mana value 6 or less: Titan (6) plus Bear (2) is
	// refused; Elf (1) + Bear (2) + Hydra (2, X as zero) is taken.
	again := b21Push(g, me.ID, "Protean Hulk", "Creature — Beast", b23ProteanHulkOracle, 6, 6, "G")
	b18Kill(t, g, again)
	c = searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the second Hulk asks too")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{titan, bear}); err == nil {
		t.Error("a set over six mana value is refused")
	}
	answerSearchByID(t, g, me.ID, elf, bear, hydra)
	for _, name := range []string{"Elf", "Bear", "Hydra"} {
		if id := findBattlefieldByName(g, name); id == uuid.Nil || controllerOf(t, g, id) != me.ID {
			t.Errorf("%s enters under your control", name)
		}
	}
	if !me.Library.Contains(titan) || !me.Library.Contains(bolt) {
		t.Error("the rest stays in the library")
	}
	if spec, _ := Lookup(b23ProteanHulkOracle); spec.Completeness != CompletenessFull {
		t.Error("whole")
	}
}

// --- the creatures: X, upkeep and counters -------------------------

func TestB23PrimordialHydraEntersWithXAndDoublesEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	hydra := castXSpell(t, g, "Primordial Hydra", "Creature — Hydra", b23PrimordialHydraOracle, "{X}{G}{G}", 3, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hydra) {
		t.Fatal("X=3: the 0/0 survives on its counters")
	}
	if n := counterCount(g, hydra, "+1/+1"); n != 3 || cardByID(g, hydra).CurrentPower() != 3 {
		t.Errorf("enters with X counters: %d, power %d", n, cardByID(g, hydra).CurrentPower())
	}
	if hasEffectiveKeyword(t, g, hydra, "trample") {
		t.Error("no trample under ten counters")
	}
	// An opponent's upkeep does nothing; yours doubles.
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, hydra, "+1/+1"); n != 3 {
		t.Errorf("an opponent's upkeep: still 3, got %d", n)
	}
	advanceToUpkeepOf(t, g, 0)
	if triggerOnStack(g, hydra) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("your upkeep: the doubling is a trigger")
	}
	passPriorityAroundTable(t, g)
	if n := counterCount(g, hydra, "+1/+1"); n != 6 || cardByID(g, hydra).CurrentPower() != 6 {
		t.Errorf("doubled: 6 counters, power 6 — got %d, power %d", n, cardByID(g, hydra).CurrentPower())
	}
	// Ten or more: trample, the moment the tenth lands.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(hydra, "+1/+1", 4) })
	if !hasEffectiveKeyword(t, g, hydra, "trample") {
		t.Error("ten counters: trample")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(hydra, "+1/+1", -1) })
	if hasEffectiveKeyword(t, g, hydra, "trample") {
		t.Error("nine counters: no trample")
	}
	// #1002: the X counters ride the CR 614 entry pipeline now, so the
	// gap they used to declare is gone and the card is complete.
	if spec, _ := Lookup(b23PrimordialHydraOracle); spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Primordial Hydra is complete: %v %v", spec.Completeness, spec.Caveats)
	}
}

func TestB23InsightEngineChargesAndDrawsPerCharge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	engine := b21Push(g, me.ID, "Insight Engine", "Artifact", b23InsightEngineOracle, 0, 0)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, engine, 0, game.ActivateAbilityParams{})
	if !b16Tapped(t, g, engine) {
		t.Error("the Engine taps")
	}
	passPriorityAroundTable(t, g)
	if counterCount(g, engine, "charge") != 1 || me.Hand.Size() != hand+1 {
		t.Errorf("first activation: 1 charge, 1 card — %d charges, drew %d", counterCount(g, engine, "charge"), me.Hand.Size()-hand)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(engine) })
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, engine, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if counterCount(g, engine, "charge") != 2 || me.Hand.Size() != hand+3 {
		t.Errorf("second activation: 2 charges, 2 more cards — %d charges, drew %d", counterCount(g, engine, "charge"), me.Hand.Size()-hand)
	}
	// Tapped: cannot activate again.
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a tapped Engine cannot activate")
	}
}

// --- the creatures: cast, draw, tap and dies triggers --------------

func TestB23ForcedFruitionFeedsOpponentsSeven(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Forced Fruition", "Enchantment", b23ForcedFruitionOracle, 0, 0, "U")
	before := b23Hands(g)
	castCatalogSpell(t, g, "My Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before[0] {
		t.Error("your own spell draws nothing")
	}
	b13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", b16TargetPlayer(me.ID))
	if triggerOnStack(g, uuid.Nil) == nil && len(g.PendingTriggers) == 0 && len(g.StackMeta) == 0 {
		t.Fatal("an opponent's spell: the draw is a trigger")
	}
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != before[1]+7 {
		t.Errorf("that player draws seven: hand %d, want %d", opp.Hand.Size(), before[1]+7)
	}
	if g.Seats[2].Hand.Size() != before[2] {
		t.Error("only the caster draws")
	}
}

func TestB23NivMizzetPingsPerDrawAndTapsToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	niv := b21Push(g, me.ID, "Niv-Mizzet, the Firemind", "Legendary Creature — Dragon Wizard", b23NivMizzetFiremindOracle, 4, 4, "U", "R")
	if !hasEffectiveKeyword(t, g, niv, "flying") {
		t.Error("flying")
	}
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 1, "G")
	b23Draw(g, me.ID, 2)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetPlayers, opp.ID) || !hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetCards, bear) || !hasID(p.PickTargetCards, niv) {
		t.Error("any target: a player (you included), a creature")
	}
	pickPlayer(t, g, me.ID, other.ID)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if other.Life != 39 {
		t.Errorf("two draws: two triggers; the first pings the chosen player: life %d, want 39", other.Life)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("… the second kills the 1/1")
	}
	// An opponent's draw is not yours.
	b23Draw(g, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an opponent drawing triggers nothing")
	}
	// {T}: draw a card — and the draw pings.
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, niv, 0, game.ActivateAbilityParams{})
	if !b16Tapped(t, g, niv) {
		t.Error("Niv taps")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Error("the activation draws a card")
	}
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Errorf("the drawn card pings: life %d, want 39", opp.Life)
	}
}

func TestB23DiscipleOfTheVaultDrainsPerArtifactDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b21Push(g, me.ID, "Disciple of the Vault", "Creature — Human Cleric", b23DiscipleOfTheVaultOracle, 1, 1, "B")
	theirRock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	myMyr := b16Creature(g, me.ID, "My Myr", "Artifact Creature — Myr", 1, 1)
	shrine := b12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	exiled := b12Permanent(g, opp.ID, "Their Other Rock", "Artifact")
	// An opponent's artifact destroyed: you may, target opponent.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirRock) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) || !hasID(p.PickTargetPlayers, other.ID) {
		t.Error("target OPPONENT")
	}
	pickPlayer(t, g, me.ID, other.ID)
	passPriorityAroundTable(t, g)
	if other.Life != 39 || opp.Life != 40 {
		t.Errorf("the chosen opponent loses 1: %d / %d", other.Life, opp.Life)
	}
	// Your own artifact creature dying, and a Treasure you sacrifice
	// for mana: each asks. Declining does nothing.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(myMyr) })
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if other.Life != 39 {
		t.Error("declined: no loss")
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	treasure := findBattlefieldByName(g, "Treasure")
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, treasure, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack the Treasure: %v", err)
	}
	// #730: the colour pick gates the table. Answer it so the death
	// trigger the sacrifice also queued can be passed around.
	riderAnswerManaPicks(t, g, me.ID, "B")
	if !b19HasTriggerPromptFor(g, me.ID) {
		t.Fatal("a Treasure sacrificed for mana is an artifact put into a graveyard")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Errorf("the Treasure drains: life %d, want 39", opp.Life)
	}
	// An enchantment dying, an artifact exiled: nothing.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(shrine) })
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(exiled) })
	passPriorityAroundTable(t, g)
	if b19HasTriggerPromptFor(g, me.ID) {
		t.Error("an enchantment is not an artifact; exile is not a graveyard")
	}
}

func TestB23PyrohemiaPingsEverythingAndLeavesWithTheCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pyro := b21Push(g, me.ID, "Pyrohemia", "Enchantment", b23PyrohemiaOracle, 0, 0, "R")
	mine := b16Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1, "R")
	theirs := b16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1, "R")
	big := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	before := b17Life(g)
	advanceToMain(t, g)
	b06AddMana(me, "R")
	b16Activate(t, g, me.ID, pyro, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != before[i]-1 {
			t.Errorf("seat %d life %d, want %d — each player, you included", i, p.Life, before[i]-1)
		}
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("each creature takes 1: the 1/1s die, yours included")
	}
	if !g.Battlefield.Contains(big) || damageMarkedOn(g, big) != 1 {
		t.Error("the 2/2 takes 1 and lives")
	}
	// The end step with a creature out: the enchantment stays.
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(pyro) {
		t.Fatal("a creature on the battlefield keeps Pyrohemia")
	}
	// No creatures at the end step, but one arrives in response: the
	// intervening-if is re-checked and it stays.
	advanceToMainOf(t, g, 1)
	b18Kill(t, g, big)
	advanceTo(t, g, game.StepEnd)
	if len(g.PendingTriggers) == 0 && triggerOnStack(g, pyro) == nil {
		t.Fatal("no creatures at the end step: the sacrifice triggers (any player's end step)")
	}
	b16Creature(g, opp.ID, "Late Bear", "Creature — Bear", 2, 2, "G")
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(pyro) {
		t.Fatal("a creature arriving in response keeps Pyrohemia")
	}
	// No creatures, nothing in response: sacrificed.
	advanceToMainOf(t, g, 2)
	b18Kill(t, g, findBattlefieldByName(g, "Late Bear"))
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pyro) || !me.Graveyard.Contains(pyro) {
		t.Error("with no creatures at the end step, Pyrohemia is sacrificed")
	}
}

func TestB23DionusUntapsAndGrowsEachElfOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dionus := b21Push(g, me.ID, "Dionus, Elvish Archdruid", "Legendary Creature — Elf Druid", b23DionusOracle, 3, 3, "G")
	elf := b21Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, 1, 1, "G")
	otherElf := b21Push(g, me.ID, "Fyndhorn Elves", "Creature — Elf Druid", llanowarElvesOracle, 1, 1, "G")
	bird := b21Push(g, me.ID, "Birds", "Creature — Bird", llanowarElvesOracle, 0, 1, "G")
	theirElf := b21Push(g, opp.ID, "Their Elves", "Creature — Elf Druid", llanowarElvesOracle, 1, 1, "G")
	advanceToMain(t, g)
	b23TapMana(t, g, me.ID, elf)
	if len(g.PendingTriggers) == 0 && triggerOnStack(g, dionus) == nil {
		t.Fatal("an Elf you control tapped on your turn: the granted ability triggers")
	}
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, elf) || counterCount(g, elf, "+1/+1") != 1 {
		t.Errorf("untapped with a counter: tapped %v, counters %d", b16Tapped(t, g, elf), counterCount(g, elf, "+1/+1"))
	}
	// Tap it again this turn: once each turn, per Elf — the other Elf
	// still has its own trigger.
	b23TapMana(t, g, me.ID, elf)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, elf) || counterCount(g, elf, "+1/+1") != 1 {
		t.Error("the ability triggers only once each turn")
	}
	b23TapMana(t, g, me.ID, otherElf)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, otherElf) || counterCount(g, otherElf, "+1/+1") != 1 {
		t.Error("each Elf carries its own once-per-turn ability")
	}
	// A non-Elf tapping: nothing.
	b23TapMana(t, g, me.ID, bird)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, bird) || counterCount(g, bird, "+1/+1") != 0 {
		t.Error("a Bird is not an Elf")
	}
	// Dionus himself is an Elf: attacking taps him, and he untaps
	// and grows.
	declareAttack(t, g, opp.ID, dionus)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, dionus) || counterCount(g, dionus, "+1/+1") != 1 {
		t.Error("an attacking Elf becomes tapped: it untaps and grows")
	}
	// An opponent's Elf, and your Elf on an opponent's turn: nothing.
	advanceToMainOf(t, g, 1)
	b23TapMana(t, g, opp.ID, theirElf)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, theirElf) || counterCount(g, theirElf, "+1/+1") != 0 {
		t.Error("an opponent's Elf is not yours")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(elf) })
	b23TapMana(t, g, me.ID, elf)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, elf) || counterCount(g, elf, "+1/+1") != 1 {
		t.Error("not during your turn")
	}
	// A new turn of yours: it triggers again.
	advanceToMainOf(t, g, 0)
	b23TapMana(t, g, me.ID, elf)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, elf) || counterCount(g, elf, "+1/+1") != 2 {
		t.Error("a new turn: the ability is fresh")
	}
	// Without Dionus the Elves have no such ability.
	b18Kill(t, g, dionus)
	b23TapMana(t, g, me.ID, otherElf)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, otherElf) || counterCount(g, otherElf, "+1/+1") != 1 {
		t.Error("the ability is Dionus's grant, gone with him")
	}
	if spec, _ := Lookup(b23DionusOracle); spec.Completeness != CompletenessFull {
		t.Error("whole")
	}
}
