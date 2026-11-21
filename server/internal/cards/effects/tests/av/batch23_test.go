package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b23Cast seeds a fully-specified card — colours and mana cost
// included, for a colour or mana-value check — into the active
// seat's hand and casts it from a main phase.
func b23Cast(t *testing.T, g *game.Game, name, typeLine, manaCost, oracle string, colors []string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.HandCardFull(active, name, typeLine, manaCost, oracle, colors)
	effectstest.AdvanceToMain(t, g)
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

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Hordeling
// Outburst was already on main from the S21 token work and is pinned
// here so the table matches the issue's 21.
func TestBatch23CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B23HordelingOutburstAlreadyOID: "Hordeling Outburst",
		effectstest.B23BorealDruidOracle:           "Boreal Druid",
		effectstest.B23DiscipleOfTheVaultOracle:    "Disciple of the Vault",
		effectstest.B23PrimordialHydraOracle:       "Primordial Hydra",
		effectstest.B23IvyLaneDenizenOracle:        "Ivy Lane Denizen",
		effectstest.B23EncroachingDragonstormOID:   "Encroaching Dragonstorm",
		effectstest.B23PyrohemiaOracle:             "Pyrohemia",
		effectstest.B23InsightEngineOracle:         "Insight Engine",
		effectstest.B23RegisaurAlphaOracle:         "Regisaur Alpha",
		effectstest.B23DionusOracle:                "Dionus, Elvish Archdruid",
		effectstest.B23ForcedFruitionOracle:        "Forced Fruition",
		effectstest.B23FangsOfKaloniaOracle:        "Fangs of Kalonia",
		effectstest.B23NivMizzetFiremindOracle:     "Niv-Mizzet, the Firemind",
		effectstest.B23ProteanHulkOracle:           "Protean Hulk",
		effectstest.B23RipApartOracle:              "Rip Apart",
		effectstest.B23CollectorOupheOracle:        "Collector Ouphe",
		effectstest.B23RunicArmasaurOracle:         "Runic Armasaur",
	}
	if len(want) != 17 {
		t.Fatalf("the batch registers 16 cards plus Hordeling Outburst, the table lists %d", len(want))
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
	}
	// Ilharg, the Raze-Boar (the card put from hand at resolution) came off
	// the declared skips with GAP batch 74 (#5948): its attack trigger is
	// Kinscaer Sentry's hand_attacker_put.go seam.
}

// --- the mana dork and the spells ----------------------------------

func TestB23BorealDruidTapsForColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	druid := effectstest.B21Push(g, me.ID, "Boreal Druid", "Snow Creature — Elf Druid", effectstest.B23BorealDruidOracle, 1, 1, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.B23TapMana(t, g, me.ID, druid)
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want C", got)
	}
	if !effectstest.B16Tapped(t, g, druid) {
		t.Error("the Druid taps")
	}
	sick := effectstest.CastCatalogSpell(t, g, "Boreal Druid", "Snow Creature — Elf Druid", effectstest.B23BorealDruidOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a creature's tap mana ability waits out summoning sickness")
	}
	if spec, _ := effects.Lookup(effectstest.B23BorealDruidOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("a mana dork is whole")
	}
}

func TestB23RipApartBurnsOrDestroys(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	big := effectstest.B16Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 5, 5, "G")
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jace", TypeLine: "Legendary Planeswalker — Jace", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	shrine := effectstest.B12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	effectstest.CastModal(t, g, "Rip Apart", "Sorcery", effectstest.B23RipApartOracle, []int{0}, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("3 damage kills the 2/2")
	}
	effectstest.CastModal(t, g, "Rip Apart", "Sorcery", effectstest.B23RipApartOracle, []int{0}, effectstest.B16TargetCard(walker))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, walker, game.CounterLoyalty) != 1 {
		t.Errorf("3 damage to a planeswalker removes 3 loyalty: %d, want 1", effectstest.CounterCount(g, walker, game.CounterLoyalty))
	}
	effectstest.CastModal(t, g, "Rip Apart", "Sorcery", effectstest.B23RipApartOracle, []int{1}, effectstest.B16TargetCard(shrine))
	effectstest.PassPriorityAroundTable(t, g)
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
		id := effectstest.HandCardFull(me, "Rip Apart", "Sorcery", "{R}{W}", effectstest.B23RipApartOracle, nil)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{tc.mode}, Targets: effectstest.B16TargetCard(tc.target)}); err == nil {
			t.Errorf("%s is refused", tc.what)
		}
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 5/5 was never touched")
	}
}

func TestB23FangsOfKaloniaGrowsThenDoubles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bare := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	laden := effectstest.B16Creature(g, me.ID, "Hydra", "Creature — Hydra", 0, 0, "G")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(laden, "+1/+1", 3) })
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.CastCatalogSpell(t, g, "Fangs of Kalonia", "Sorcery", effectstest.B23FangsOfKaloniaOracle, effectstest.B16TargetCard(bare))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bare, "+1/+1"); n != 2 {
		t.Errorf("a bare creature: one counter, doubled to 2, got %d", n)
	}
	effectstest.CastCatalogSpell(t, g, "Fangs of Kalonia", "Sorcery", effectstest.B23FangsOfKaloniaOracle, effectstest.B16TargetCard(laden))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, laden, "+1/+1"); n != 8 {
		t.Errorf("three counters: one more makes 4, doubled to 8, got %d", n)
	}
	// Target creature YOU control.
	id := effectstest.HandCardFull(me, "Fangs of Kalonia", "Sorcery", "{1}{G}", effectstest.B23FangsOfKaloniaOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(theirs)}); err == nil {
		t.Error("an opponent's creature is not a legal target")
	}
	// Overloaded: each creature you control, no target, opponents'
	// creatures untouched.
	effectstest.CastWithAltCost(t, g, "Fangs of Kalonia", "Sorcery", effectstest.B23FangsOfKaloniaOracle, "overload")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bare, "+1/+1"); n != 6 {
		t.Errorf("overload on the 2-counter Bear: 3 doubled to 6, got %d", n)
	}
	if n := effectstest.CounterCount(g, laden, "+1/+1"); n != 18 {
		t.Errorf("overload on the 8-counter Hydra: 9 doubled to 18, got %d", n)
	}
	if n := effectstest.CounterCount(g, theirs, "+1/+1"); n != 0 {
		t.Errorf("an opponent's creature gets nothing: %d", n)
	}
	if spec, _ := effects.Lookup(effectstest.B23FangsOfKaloniaOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("whole")
	}
}

// TestB23FangsOfKaloniaDoublesAfterAPausedPlacement is #1290's real-
// card exit criterion for the batch shape: with Doubling Season AND
// Hardened Scales both on the board, the first +1/+1 counter's
// placement pauses on a CR 616 ordering prompt, and "had a +1/+1
// counter put on it this way" — the gate on whether the second,
// doubling placement happens at all — has to be read AFTER that
// placement lands. Before the fix, the growth check ran on the next
// line while the placement was still owed to the prompt, always saw
// no growth, and silently skipped the doubling.
func TestB23FangsOfKaloniaDoublesAfterAPausedPlacement(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	bare := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")

	effectstest.CastCatalogSpell(t, g, "Fangs of Kalonia", "Sorcery", effectstest.B23FangsOfKaloniaOracle, effectstest.B16TargetCard(bare))
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CounterCount(g, bare, "+1/+1"); n != 0 {
		t.Fatalf("counters landed with the first placement still owed to a CR 616 prompt: %d", n)
	}

	// Doubling Season first each time: the first placement's window
	// (1 → 2 → 3) and, once that settles and growth is detected, the
	// doubling placement's own window (3 → 6 → 7) — the doubling adds
	// however many the creature has, and that add is itself subject
	// to the same two replacements.
	for i := 0; len(g.PendingChoices) > 0; i++ {
		if i >= 5 {
			t.Fatalf("still answering CR 616 prompts after %d rounds — counters = %d", i, effectstest.CounterCount(g, bare, "+1/+1"))
		}
		effectstest.AnswerOrderBySource(t, g, ds, hs)
	}

	if n := effectstest.CounterCount(g, bare, "+1/+1"); n != 10 {
		t.Errorf("+1/+1 counters = %d, want 10 (3 from the first placement's window, then 7 more from doubling those 3 through the same two replacements)", n)
	}
}

// --- the creatures: statics and ETB --------------------------------

func TestB23RegisaurAlphaMakesAHastyDinosaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dino := effectstest.B16Creature(g, me.ID, "My Raptor", "Creature — Dinosaur", 2, 2, "R")
	bear := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Raptor", "Creature — Dinosaur", 2, 2, "R")
	alpha := effectstest.CastCatalogSpell(t, g, "Regisaur Alpha", "Creature — Dinosaur", effectstest.B23RegisaurAlphaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	tokens := effectstest.BattlefieldIDsNamed(g, "Dinosaur")
	if len(tokens) != 1 {
		t.Fatalf("one Dinosaur token, got %d", len(tokens))
	}
	token := tokens[0]
	c := effectstest.CardByID(g, token)
	if effectstest.ControllerOf(t, g, token) != me.ID || effectstest.EffectivePower(t, g, token) != 3 || effectstest.EffectiveToughness(t, g, token) != 3 || len(c.Colors) != 1 || c.Colors[0] != "G" {
		t.Error("a 3/3 green Dinosaur under your control")
	}
	if !effectstest.HasEffectiveKeyword(t, g, token, "trample") {
		t.Error("with trample")
	}
	if !effectstest.HasEffectiveKeyword(t, g, token, "haste") || !effectstest.HasEffectiveKeyword(t, g, dino, "haste") {
		t.Error("other Dinosaurs you control have haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, alpha, "haste") || effectstest.HasEffectiveKeyword(t, g, bear, "haste") || effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("not the Alpha itself, not a Bear, not an opponent's Dinosaur")
	}
}

func TestB23IvyLaneDenizenGrowsATargetPerGreenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	denizen := effectstest.B21Push(g, me.ID, "Ivy Lane Denizen", "Creature — Elf Warrior", effectstest.B23IvyLaneDenizenOracle, 2, 3, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	elf := b23Cast(t, g, "Elf", "Creature — Elf", "{G}", "", []string{"G"}, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, elf) || !effectstest.HasID(p.PickTargetCards, denizen) || !effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("target CREATURE — the entering one, the Denizen, an opponent's")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, theirs, "+1/+1") != 1 {
		t.Error("the chosen creature gets the counter")
	}
	// A green token counts; a colourless artifact creature and a
	// red creature do not; an opponent's green creature is not yours.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Dinosaur with trample"), 1) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, elf, "+1/+1") != 1 {
		t.Error("a green token entering triggers")
	}
	b23Cast(t, g, "Myr", "Artifact Creature — Myr", "{2}", "", nil, nil)
	effectstest.PassPriorityAroundTable(t, g)
	b23Cast(t, g, "Goblin", "Creature — Goblin", "{R}", "", []string{"R"}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a colourless or red creature entering triggers nothing")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	b23Cast(t, g, "Their Elf", "Creature — Elf", "{G}", "", []string{"G"}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's green creature is not one you control")
	}
}

func TestB23EncroachingDragonstormRampsAndReturnsOnADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Mountain", "Basic Land — Mountain"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Command Tower", "Land"),
	)
	storm := effectstest.CastCatalogSpell(t, g, "Encroaching Dragonstorm", "Enchantment", effectstest.B23EncroachingDragonstormOID, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("three basics for up to two: the chooser asks")
	}
	if c.SearchMax != 2 || effectstest.SearchOptionNamed(g, c, "Command Tower") != uuid.Nil {
		t.Error("up to two BASIC land cards")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest", "Mountain")
	for _, name := range []string{"Forest", "Mountain"} {
		id := effectstest.FindBattlefieldByName(g, name)
		if id == uuid.Nil {
			t.Fatalf("%s enters", name)
		}
		effectstest.Top100AssertEnteredTapped(t, g, id, name)
	}
	if !g.Battlefield.Contains(storm) {
		t.Fatal("the enchantment stays")
	}
	// A non-Dragon: nothing. An opponent's Dragon: nothing. Your
	// Dragon: back to hand.
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(storm) {
		t.Error("a Bear is not a Dragon")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.CastCatalogSpell(t, g, "Their Dragon", "Creature — Dragon", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(storm) {
		t.Error("an opponent's Dragon is not one you control")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.CastCatalogSpell(t, g, "Dragon", "Creature — Dragon", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(storm) || !me.Hand.Contains(storm) {
		t.Error("a Dragon you control entering returns the enchantment to your hand")
	}
	_ = opp
}

func TestB23ProteanHulkFetchesCreaturesWithinSixManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hulk := effectstest.B21Push(g, me.ID, "Protean Hulk", "Creature — Beast", effectstest.B23ProteanHulkOracle, 6, 6, "G")
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Elf", TypeLine: "Creature — Elf", ManaCost: "{G}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Hydra", TypeLine: "Creature — Hydra", ManaCost: "{X}{G}{G}"},
		game.Card{Name: "Titan", TypeLine: "Creature — Giant", ManaCost: "{4}{G}{G}"},
		game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"},
	)
	elf, bear, hydra, titan, bolt := ids[0], ids[1], ids[2], ids[3], ids[4]
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(hulk) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("a bounce is not a death")
	}
	hulk = effectstest.B21Push(g, me.ID, "Protean Hulk", "Creature — Beast", effectstest.B23ProteanHulkOracle, 6, 6, "G")
	effectstest.B18Kill(t, g, hulk)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("dying: the chooser asks")
	}
	if c.SearchMax < 4 {
		t.Errorf("any number: max %d", c.SearchMax)
	}
	for _, id := range []uuid.UUID{elf, bear, hydra, titan} {
		if !effectstest.HasID(c.SearchCards, id) {
			t.Errorf("%s (a creature card) is offered", effectstest.CardByID(g, id).Name)
		}
	}
	if effectstest.HasID(c.SearchCards, bolt) {
		t.Error("an instant is not")
	}
	// Fail to find is legal: nothing taken, the library shuffled.
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	if me.Library.Size() != 5 {
		t.Error("nothing taken")
	}
	// Again. Total mana value 6 or less: Titan (6) plus Bear (2) is
	// refused; Elf (1) + Bear (2) + Hydra (2, X as zero) is taken.
	again := effectstest.B21Push(g, me.ID, "Protean Hulk", "Creature — Beast", effectstest.B23ProteanHulkOracle, 6, 6, "G")
	effectstest.B18Kill(t, g, again)
	c = effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the second Hulk asks too")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{titan, bear}); err == nil {
		t.Error("a set over six mana value is refused")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, elf, bear, hydra)
	for _, name := range []string{"Elf", "Bear", "Hydra"} {
		if id := effectstest.FindBattlefieldByName(g, name); id == uuid.Nil || effectstest.ControllerOf(t, g, id) != me.ID {
			t.Errorf("%s enters under your control", name)
		}
	}
	if !me.Library.Contains(titan) || !me.Library.Contains(bolt) {
		t.Error("the rest stays in the library")
	}
	if spec, _ := effects.Lookup(effectstest.B23ProteanHulkOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("whole")
	}
}

// --- the creatures: X, upkeep and counters -------------------------

func TestB23PrimordialHydraEntersWithXAndDoublesEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	hydra := effectstest.CastXSpell(t, g, "Primordial Hydra", "Creature — Hydra", effectstest.B23PrimordialHydraOracle, "{X}{G}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hydra) {
		t.Fatal("X=3: the 0/0 survives on its counters")
	}
	if n := effectstest.CounterCount(g, hydra, "+1/+1"); n != 3 || effectstest.CardByID(g, hydra).CurrentPower() != 3 {
		t.Errorf("enters with X counters: %d, power %d", n, effectstest.CardByID(g, hydra).CurrentPower())
	}
	if effectstest.HasEffectiveKeyword(t, g, hydra, "trample") {
		t.Error("no trample under ten counters")
	}
	// An opponent's upkeep does nothing; yours doubles.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, hydra, "+1/+1"); n != 3 {
		t.Errorf("an opponent's upkeep: still 3, got %d", n)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if effectstest.TriggerOnStack(g, hydra) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("your upkeep: the doubling is a trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, hydra, "+1/+1"); n != 6 || effectstest.CardByID(g, hydra).CurrentPower() != 6 {
		t.Errorf("doubled: 6 counters, power 6 — got %d, power %d", n, effectstest.CardByID(g, hydra).CurrentPower())
	}
	// Ten or more: trample, the moment the tenth lands.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(hydra, "+1/+1", 4) })
	if !effectstest.HasEffectiveKeyword(t, g, hydra, "trample") {
		t.Error("ten counters: trample")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(hydra, "+1/+1", -1) })
	if effectstest.HasEffectiveKeyword(t, g, hydra, "trample") {
		t.Error("nine counters: no trample")
	}
	// #1002: the X counters ride the CR 614 entry pipeline now, so the
	// gap they used to declare is gone and the card is complete.
	if spec, _ := effects.Lookup(effectstest.B23PrimordialHydraOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Primordial Hydra is complete: %v %v", spec.Completeness, spec.Caveats)
	}
}

func TestB23InsightEngineChargesAndDrawsPerCharge(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	engine := effectstest.B21Push(g, me.ID, "Insight Engine", "Artifact", effectstest.B23InsightEngineOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, engine, 0, game.ActivateAbilityParams{})
	if !effectstest.B16Tapped(t, g, engine) {
		t.Error("the Engine taps")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, engine, "charge") != 1 || me.Hand.Size() != hand+1 {
		t.Errorf("first activation: 1 charge, 1 card — %d charges, drew %d", effectstest.CounterCount(g, engine, "charge"), me.Hand.Size()-hand)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(engine) })
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, engine, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, engine, "charge") != 2 || me.Hand.Size() != hand+3 {
		t.Errorf("second activation: 2 charges, 2 more cards — %d charges, drew %d", effectstest.CounterCount(g, engine, "charge"), me.Hand.Size()-hand)
	}
	// Tapped: cannot activate again.
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a tapped Engine cannot activate")
	}
}

// --- the creatures: cast, draw, tap and dies triggers --------------

func TestB23ForcedFruitionFeedsOpponentsSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Forced Fruition", "Enchantment", effectstest.B23ForcedFruitionOracle, 0, 0, "U")
	before := b23Hands(g)
	effectstest.CastCatalogSpell(t, g, "My Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before[0] {
		t.Error("your own spell draws nothing")
	}
	effectstest.B13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", effectstest.B16TargetPlayer(me.ID))
	if effectstest.TriggerOnStack(g, uuid.Nil) == nil && len(g.PendingTriggers) == 0 && len(g.StackMeta) == 0 {
		t.Fatal("an opponent's spell: the draw is a trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != before[1]+7 {
		t.Errorf("that player draws seven: hand %d, want %d", opp.Hand.Size(), before[1]+7)
	}
	if g.Seats[2].Hand.Size() != before[2] {
		t.Error("only the caster draws")
	}
}

func TestB23NivMizzetPingsPerDrawAndTapsToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	niv := effectstest.B21Push(g, me.ID, "Niv-Mizzet, the Firemind", "Legendary Creature — Dragon Wizard", effectstest.B23NivMizzetFiremindOracle, 4, 4, "U", "R")
	if !effectstest.HasEffectiveKeyword(t, g, niv, "flying") {
		t.Error("flying")
	}
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 1, "G")
	b23Draw(g, me.ID, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetPlayers, opp.ID) || !effectstest.HasID(p.PickTargetPlayers, me.ID) || !effectstest.HasID(p.PickTargetCards, bear) || !effectstest.HasID(p.PickTargetCards, niv) {
		t.Error("any target: a player (you included), a creature")
	}
	effectstest.PickPlayer(t, g, me.ID, other.ID)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if other.Life != 39 {
		t.Errorf("two draws: two triggers; the first pings the chosen player: life %d, want 39", other.Life)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("… the second kills the 1/1")
	}
	// An opponent's draw is not yours.
	b23Draw(g, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an opponent drawing triggers nothing")
	}
	// {T}: draw a card — and the draw pings.
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, niv, 0, game.ActivateAbilityParams{})
	if !effectstest.B16Tapped(t, g, niv) {
		t.Error("Niv taps")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Error("the activation draws a card")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Errorf("the drawn card pings: life %d, want 39", opp.Life)
	}
}

func TestB23DiscipleOfTheVaultDrainsPerArtifactDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B21Push(g, me.ID, "Disciple of the Vault", "Creature — Human Cleric", effectstest.B23DiscipleOfTheVaultOracle, 1, 1, "B")
	theirRock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	myMyr := effectstest.B16Creature(g, me.ID, "My Myr", "Artifact Creature — Myr", 1, 1)
	shrine := effectstest.B12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	exiled := effectstest.B12Permanent(g, opp.ID, "Their Other Rock", "Artifact")
	// An opponent's artifact destroyed: you may, target opponent.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirRock) })
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetPlayers, me.ID) || !effectstest.HasID(p.PickTargetPlayers, opp.ID) || !effectstest.HasID(p.PickTargetPlayers, other.ID) {
		t.Error("target OPPONENT")
	}
	effectstest.PickPlayer(t, g, me.ID, other.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if other.Life != 39 || opp.Life != 40 {
		t.Errorf("the chosen opponent loses 1: %d / %d", other.Life, opp.Life)
	}
	// Your own artifact creature dying, and a Treasure you sacrifice
	// for mana: each asks. Declining does nothing.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(myMyr) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if other.Life != 39 {
		t.Error("declined: no loss")
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	treasure := effectstest.FindBattlefieldByName(g, "Treasure")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, treasure, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack the Treasure: %v", err)
	}
	// #730: the colour pick gates the table. Answer it so the death
	// trigger the sacrifice also queued can be passed around.
	effectstest.RiderAnswerManaPicks(t, g, me.ID, "B")
	if !effectstest.B19HasTriggerPromptFor(g, me.ID) && effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("a Treasure sacrificed for mana is an artifact put into a graveyard")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Errorf("the Treasure drains: life %d, want 39", opp.Life)
	}
	// An enchantment dying, an artifact exiled: nothing.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(shrine) })
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(exiled) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Error("an enchantment is not an artifact; exile is not a graveyard")
	}
}

func TestB23PyrohemiaPingsEverythingAndLeavesWithTheCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pyro := effectstest.B21Push(g, me.ID, "Pyrohemia", "Enchantment", effectstest.B23PyrohemiaOracle, 0, 0, "R")
	mine := effectstest.B16Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1, "R")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1, "R")
	big := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	before := effectstest.B17Life(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R")
	effectstest.B16Activate(t, g, me.ID, pyro, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != before[i]-1 {
			t.Errorf("seat %d life %d, want %d — each player, you included", i, p.Life, before[i]-1)
		}
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("each creature takes 1: the 1/1s die, yours included")
	}
	if !g.Battlefield.Contains(big) || effectstest.DamageMarkedOn(g, big) != 1 {
		t.Error("the 2/2 takes 1 and lives")
	}
	// The end step with a creature out: the enchantment stays.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(pyro) {
		t.Fatal("a creature on the battlefield keeps Pyrohemia")
	}
	// No creatures at the end step, but one arrives in response: the
	// intervening-if is re-checked and it stays.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.B18Kill(t, g, big)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if len(g.PendingTriggers) == 0 && effectstest.TriggerOnStack(g, pyro) == nil {
		t.Fatal("no creatures at the end step: the sacrifice triggers (any player's end step)")
	}
	effectstest.B16Creature(g, opp.ID, "Late Bear", "Creature — Bear", 2, 2, "G")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(pyro) {
		t.Fatal("a creature arriving in response keeps Pyrohemia")
	}
	// No creatures, nothing in response: sacrificed.
	effectstest.AdvanceToMainOf(t, g, 2)
	effectstest.B18Kill(t, g, effectstest.FindBattlefieldByName(g, "Late Bear"))
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pyro) || !me.Graveyard.Contains(pyro) {
		t.Error("with no creatures at the end step, Pyrohemia is sacrificed")
	}
}

func TestB23DionusUntapsAndGrowsEachElfOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dionus := effectstest.B21Push(g, me.ID, "Dionus, Elvish Archdruid", "Legendary Creature — Elf Druid", effectstest.B23DionusOracle, 3, 3, "G")
	elf := effectstest.B21Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, 1, 1, "G")
	otherElf := effectstest.B21Push(g, me.ID, "Fyndhorn Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, 1, 1, "G")
	bird := effectstest.B21Push(g, me.ID, "Birds", "Creature — Bird", effectstest.LlanowarElvesOracle, 0, 1, "G")
	theirElf := effectstest.B21Push(g, opp.ID, "Their Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, 1, 1, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.B23TapMana(t, g, me.ID, elf)
	// #4928: a trigger a mana activation queues reaches the stack at once,
	// and a granted trigger's source is the Elf that has it.
	if len(g.PendingTriggers) == 0 && effectstest.TriggerOnStack(g, dionus) == nil && effectstest.TriggerOnStack(g, elf) == nil {
		t.Fatal("an Elf you control tapped on your turn: the granted ability triggers")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, elf) || effectstest.CounterCount(g, elf, "+1/+1") != 1 {
		t.Errorf("untapped with a counter: tapped %v, counters %d", effectstest.B16Tapped(t, g, elf), effectstest.CounterCount(g, elf, "+1/+1"))
	}
	// Tap it again this turn: once each turn, per Elf — the other Elf
	// still has its own trigger.
	effectstest.B23TapMana(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, elf) || effectstest.CounterCount(g, elf, "+1/+1") != 1 {
		t.Error("the ability triggers only once each turn")
	}
	effectstest.B23TapMana(t, g, me.ID, otherElf)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, otherElf) || effectstest.CounterCount(g, otherElf, "+1/+1") != 1 {
		t.Error("each Elf carries its own once-per-turn ability")
	}
	// A non-Elf tapping: nothing.
	effectstest.B23TapMana(t, g, me.ID, bird)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, bird) || effectstest.CounterCount(g, bird, "+1/+1") != 0 {
		t.Error("a Bird is not an Elf")
	}
	// Dionus himself is an Elf: attacking taps him, and he untaps
	// and grows.
	effectstest.DeclareAttack(t, g, opp.ID, dionus)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, dionus) || effectstest.CounterCount(g, dionus, "+1/+1") != 1 {
		t.Error("an attacking Elf becomes tapped: it untaps and grows")
	}
	// An opponent's Elf, and your Elf on an opponent's turn: nothing.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.B23TapMana(t, g, opp.ID, theirElf)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, theirElf) || effectstest.CounterCount(g, theirElf, "+1/+1") != 0 {
		t.Error("an opponent's Elf is not yours")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(elf) })
	effectstest.B23TapMana(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, elf) || effectstest.CounterCount(g, elf, "+1/+1") != 1 {
		t.Error("not during your turn")
	}
	// A new turn of yours: it triggers again.
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.B23TapMana(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, elf) || effectstest.CounterCount(g, elf, "+1/+1") != 2 {
		t.Error("a new turn: the ability is fresh")
	}
	// Without Dionus the Elves have no such ability.
	effectstest.B18Kill(t, g, dionus)
	effectstest.B23TapMana(t, g, me.ID, otherElf)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, otherElf) || effectstest.CounterCount(g, otherElf, "+1/+1") != 1 {
		t.Error("the ability is Dionus's grant, gone with him")
	}
	if spec, _ := effects.Lookup(effectstest.B23DionusOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("whole")
	}
}
