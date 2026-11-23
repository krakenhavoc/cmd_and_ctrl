package testsru

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b193_test.go — S82 GAP batch 193 (#5948): twelve Warhammer 40,000
// Commander creatures. Each test plays the real catalog entry.

const (
	gapB193MagusOracle      = "ae9f5c80-bc96-4ab3-bb5b-e8bd470e9eab"
	gapB193SwarmlordOracle  = "ead75ea4-025f-443a-824f-26fef9e424db"
	gapB193FlockOracle      = "6d24fc3e-eaa9-4ac8-b271-7dc83b812acf"
	gapB193MalanthropeOracl = "0f99e921-61e7-4fc8-8115-4162712b4737"
	gapB193RavenerOracle    = "6c4ac289-edca-4dce-8572-817efbdd08b2"
	gapB193TrygonOracle     = "6b2d1d7b-a4b4-46c3-a50b-df7e60f39ee2"
	gapB193HarridanOracle   = "01432542-2ea7-45bc-bc90-151516faee0e"
	gapB193VenomOracle      = "efd70846-2226-43c6-80fe-6acdd15d4246"
	gapB193AcolyteOracle    = "998daf11-c0ed-4801-b013-d8b3136f16d1"
	gapB193ChampionOracle   = "d343c116-b33a-4ed6-ad15-d7270e739b04"
	gapB193BloodcrushOracle = "ae5b9251-2e2d-4e12-b4b4-da051683f970"
	gapB193TerminatorOracle = "e9dfab69-7618-4d90-8bb4-cf23581d1bad"

	gapB193Blaze    = "0596920f-9946-42f4-a03b-24aab67f9f1b"
	gapB193Hellkite = "37a1f340-10d5-407e-bf2b-cbbbb6cab985"
)

func TestGapB193CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB193MagusOracle:      "Magus Lucea Kane",
		gapB193SwarmlordOracle:  "The Swarmlord",
		gapB193FlockOracle:      "Gargoyle Flock",
		gapB193MalanthropeOracl: "Malanthrope",
		gapB193RavenerOracle:    "Ravener",
		gapB193TrygonOracle:     "Trygon Prime",
		gapB193HarridanOracle:   "Tyranid Harridan",
		gapB193VenomOracle:      "Venomthrope",
		gapB193AcolyteOracle:    "Acolyte Hybrid",
		gapB193ChampionOracle:   "Aspiring Champion",
		gapB193BloodcrushOracle: "Bloodcrusher of Khorne",
		gapB193TerminatorOracle: "Chaos Terminator Lord",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// gapB193Settle passes priority until the stack is empty, answering every
// trigger-order prompt and every new-target prompt by keeping `keep`. It
// returns how many target prompts it answered.
func gapB193Settle(t *testing.T, g *game.Game, who uuid.UUID, keep game.TargetRef) int {
	t.Helper()
	answered := 0
	for i := 0; i < 64; i++ {
		if p := effectstest.TriggerOrderPromptFor(g, who); p != nil {
			order := append([]uuid.UUID(nil), p.TriggerOrderIDs...)
			if err := g.ResolveTriggerOrder(p.ID, who, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := effectstest.LatestPickTarget(g, who); p != nil {
			if err := g.ResolvePickTarget(p.ID, who, keep); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			answered++
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return answered
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
	return answered
}

// --- Magus Lucea Kane ---

func TestGapB193MagusPutsACounterOnAnyCreatureAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Magus Lucea Kane", "Legendary Creature — Human Tyranid Wizard", gapB193MagusOracle, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("the opponent's creature has %d counters, want 1 (the target is any creature)", got)
	}
}

func TestGapB193MagusDoesNothingOnAnOpponentsCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B12Push(g, opp.ID, "Magus Lucea Kane", "Legendary Creature — Human Tyranid Wizard", gapB193MagusOracle, 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's Magus put %d counters on a creature at my combat", got)
	}
}

func TestGapB193MagusTapsForTwoColorlessAndCopiesTheNextXSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	magus := effectstest.B12Push(g, me.ID, "Magus Lucea Kane", "Legendary Creature — Human Tyranid Wizard", gapB193MagusOracle, 1, 1)
	effectstest.ActivateManaFor(t, g, me.ID, magus, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("pool %d after the tap, want {C}{C}", got)
	}
	// A spell with no {X} neither fires the trigger nor uses it up.
	effectstest.CastFromHandForTest(t, g, me, "Bear", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers after a non-X spell, want 1", len(g.DelayedTriggers))
	}
	life := opp.Life
	at := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	effectstest.CastXSpell(t, g, "Blaze", "Sorcery", gapB193Blaze, "{X}{R}", 2, []game.TargetRef{at})
	if answered := gapB193Settle(t, g, me.ID, at); answered != 1 {
		t.Errorf("answered %d new-target prompts, want 1 copy", answered)
	}
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want 2 damage from the spell and 2 from the copy", life, opp.Life)
	}
	// Spent: a second X spell is not copied.
	life = opp.Life
	effectstest.CastXSpell(t, g, "Blaze", "Sorcery", gapB193Blaze, "{X}{R}", 1, []game.TargetRef{at})
	if answered := gapB193Settle(t, g, me.ID, at); answered != 0 {
		t.Errorf("a second X spell was copied (%d prompts)", answered)
	}
	if opp.Life != life-1 {
		t.Errorf("opponent life %d -> %d, want 1 damage", life, opp.Life)
	}
}

func TestGapB193MagusCopiesTheNextAbilityWithXInItsCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	magus := effectstest.B12Push(g, me.ID, "Magus Lucea Kane", "Legendary Creature — Human Tyranid Wizard", gapB193MagusOracle, 1, 1)
	kite := effectstest.PushCatalogPermanent(g, me.ID, "Crimson Hellkite", "Creature — Dragon", gapB193Hellkite, false)
	victim := effectstest.B12Push(g, opp.ID, "Big Target", "Creature — Giant", "", 1, 9)
	at := game.TargetRef{Kind: game.TargetCard, ID: victim}
	effectstest.ActivateManaFor(t, g, me.ID, magus, 0, game.ManaAbilityParams{})
	me.ManaPool = game.ManaPool{}
	effectstest.B06AddMana(me, "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, kite, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: []game.TargetRef{at}}); err != nil {
		t.Fatalf("activate the Hellkite for X = 2: %v", err)
	}
	if answered := gapB193Settle(t, g, me.ID, at); answered != 1 {
		t.Errorf("answered %d new-target prompts, want 1 copy", answered)
	}
	if c, ok := effectstest.BattlefieldCard(g, victim); !ok || c.DamageMarked != 4 {
		t.Errorf("the target has %d damage, want 2 + 2 from the copy", c.DamageMarked)
	}
}

// --- The Swarmlord ---

func gapB193CastBig(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, p, tough, x int) uuid.UUID {
	t.Helper()
	id, err := effectstest.SpmB4Cast(t, g, name, typeLine, oracle, cost, p, tough, x)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestGapB193SwarmlordEntersWithTwoCountersPerCommanderCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.CommanderCasts = map[uuid.UUID]int{uuid.New(): 2}
	id := gapB193CastBig(t, g, "The Swarmlord", "Legendary Creature — Tyranid", gapB193SwarmlordOracle, "{3}{G}{U}{R}", 5, 5, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, id); got != 4 {
		t.Errorf("the Swarmlord has %d counters after two commander casts, want 4", got)
	}
}

func TestGapB193SwarmlordEntersBareWithNoCommanderCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := gapB193CastBig(t, g, "The Swarmlord", "Legendary Creature — Tyranid", gapB193SwarmlordOracle, "{3}{G}{U}{R}", 5, 5, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, id); got != 0 {
		t.Errorf("the Swarmlord has %d counters with no commander casts, want 0", got)
	}
}

func TestGapB193SwarmlordDrawsWhenYourCreatureWithACounterDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "The Swarmlord", "Legendary Creature — Tyranid", gapB193SwarmlordOracle, 5, 5)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	marked := effectstest.B12Creature(g, me.ID, "Marked", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, marked, game.CounterPlusOne, 1)
	effectstest.PutCounters(t, g, theirs, game.CounterPlusOne, 1)
	hand := me.Hand.Size()
	effectstest.KillCreature(t, g, me.ID, plain)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("a creature without a counter, or an opponent's, drew a card (hand %d, was %d)", me.Hand.Size(), hand)
	}
	effectstest.KillCreature(t, g, me.ID, marked)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after my creature with a counter died", me.Hand.Size(), hand+1)
	}
}

// --- Gargoyle Flock ---

func TestGapB193GargoyleFlockMakesAGargoyleWhenACreatureEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Gargoyle Flock", "Creature — Tyranid Gargoyle", gapB193FlockOracle, 2, 2)
	effectstest.CastFromHandForTest(t, g, me, "Bear", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardsNamed(g, "Tyranid Gargoyle", me.ID); got != 1 {
		t.Errorf("%d Tyranid Gargoyle tokens, want 1", got)
	}
}

func TestGapB193GargoyleFlockMakesNothingWhenNoCreatureEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Gargoyle Flock", "Creature — Tyranid Gargoyle", gapB193FlockOracle, 2, 2)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardsNamed(g, "Tyranid Gargoyle", me.ID); got != 0 {
		t.Errorf("%d tokens with no creature entering this turn, want 0", got)
	}
}

func TestGapB193GargoyleFlockCountsItselfEnteringAndMakesAFlyingToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gapB193CastBig(t, g, "Gargoyle Flock", "Creature — Tyranid Gargoyle", gapB193FlockOracle, "{2}{G}{U}", 2, 2, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Tyranid Gargoyle" && c.Controller == me.ID {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("no token after the Flock itself entered this turn")
	}
	if !effectstest.HasEffectiveKeyword(t, g, token, "flying") {
		t.Error("the Gargoyle token has no flying")
	}
}

// --- Malanthrope ---

func TestGapB193MalanthropeExilesAGraveyardAndGrowsPerCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c1 := effectstest.PushGraveyardCardTyped(opp, "Dead Bear", "Creature — Bear")
	c2 := effectstest.PushGraveyardCardTyped(opp, "Dead Elf", "Creature — Elf")
	spell := effectstest.PushGraveyardCardTyped(opp, "Dead Spell", "Sorcery")
	mine := effectstest.PushGraveyardCardTyped(me, "My Dead", "Creature — Bear")
	id := gapB193CastBig(t, g, "Malanthrope", "Creature — Tyranid", gapB193MalanthropeOracl, "{1}{G}{U}", 2, 2, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{c1, c2, spell} {
		if !g.Exile.Contains(c) || opp.Graveyard.Contains(c) {
			t.Errorf("a card of the opponent's graveyard was not exiled")
		}
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard was exiled")
	}
	if got := effectstest.PlusOnes(g, id); got != 2 {
		t.Errorf("Malanthrope has %d counters, want 2 (two creature cards)", got)
	}
}

func TestGapB193MalanthropeCanExileItsOwnersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.PushGraveyardCardTyped(me, "My Dead", "Creature — Bear")
	id := gapB193CastBig(t, g, "Malanthrope", "Creature — Tyranid", gapB193MalanthropeOracl, "{1}{G}{U}", 2, 2, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayerTarget(t, g, me.ID, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(mine) {
		t.Error("the targeted player's own card was not exiled")
	}
	if got := effectstest.PlusOnes(g, id); got != 1 {
		t.Errorf("Malanthrope has %d counters, want 1", got)
	}
}

// --- Ravener ---

// gapB193RavenerSettle plays out Ravener's two enters triggers: it orders
// them, answers the creature clause with `creature` and the opponent clause
// with `opp`, and passes until the stack is empty.
func gapB193RavenerSettle(t *testing.T, g *game.Game, me *game.Player, creature, opp uuid.UUID) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if p := effectstest.TriggerOrderPromptFor(g, me.ID); p != nil {
			order := append([]uuid.UUID(nil), p.TriggerOrderIDs...)
			if err := g.ResolveTriggerOrder(p.ID, me.ID, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if len(p.PickTargetCards) > 0 {
				effectstest.PickCard(t, g, me.ID, creature)
			} else {
				effectstest.PickPlayerTarget(t, g, me.ID, opp)
			}
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

func TestGapB193RavenerEntersWithXCountersAndDrawsAtFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	id := gapB193CastBig(t, g, "Ravener", "Creature — Tyranid", gapB193RavenerOracle, "{X}{G}{U}", 0, 0, 5)
	gapB193RavenerSettle(t, g, me, theirs, opp.ID)
	if got := effectstest.PlusOnes(g, id); got != 5 {
		t.Errorf("Ravener has %d counters, want 5", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (X = 5 draws a card)", me.Hand.Size(), hand+1)
	}
}

func TestGapB193RavenerDoesNotDrawBelowFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	id := gapB193CastBig(t, g, "Ravener", "Creature — Tyranid", gapB193RavenerOracle, "{X}{G}{U}", 0, 0, 4)
	gapB193RavenerSettle(t, g, me, theirs, opp.ID)
	if got := effectstest.PlusOnes(g, id); got != 4 {
		t.Errorf("Ravener has %d counters, want 4", got)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d (X = 4 draws nothing)", me.Hand.Size(), hand)
	}
}

func TestGapB193RavenerForcesTheTargetToAttackTheTargetOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	gapB193CastBig(t, g, "Ravener", "Creature — Tyranid", gapB193RavenerOracle, "{X}{G}{U}", 0, 0, 1)
	gapB193RavenerSettle(t, g, me, bear, opp.ID)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != bear {
		t.Fatalf("pass with the Bear home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
}

// --- Trygon Prime ---

func TestGapB193TrygonPrimeGrowsAndMakesAnotherAttackerUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	prime := effectstest.B12Push(g, me.ID, "Trygon Prime", "Creature — Tyranid", gapB193TrygonOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Home", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, prime, bear)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, prime) || effectstest.HasID(pick.PickTargetCards, home) || !effectstest.HasID(pick.PickTargetCards, bear) {
		t.Errorf("offered %v; want only the other attacking creature", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, prime, game.CounterPlusOne); got != 1 {
		t.Errorf("Trygon Prime has %d counters, want 1", got)
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the other attacker has %d counters, want 1", got)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, bear); err == nil {
		t.Error("a creature blocked the unblockable target")
	}
}

func TestGapB193TrygonPrimeAttackingAloneStillGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	prime := effectstest.B12Push(g, me.ID, "Trygon Prime", "Creature — Tyranid", gapB193TrygonOracle, 4, 4)
	effectstest.DeclareAttack(t, g, opp.ID, prime)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, prime, game.CounterPlusOne); got != 1 {
		t.Errorf("Trygon Prime has %d counters attacking alone, want 1", got)
	}
}

// --- Tyranid Harridan ---

func TestGapB193HarridanHasFlyingAndWard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Tyranid Harridan", "Creature — Tyranid", gapB193HarridanOracle, 4, 4)
	effectstest.AssertKeywords(t, g, id, "flying")
	spec, _ := effects.Lookup(gapB193HarridanOracle)
	if len(spec.Triggered) != 2 {
		t.Errorf("%d triggered abilities, want the ward and the Gargoyle trigger", len(spec.Triggered))
	}
}

func TestGapB193HarridanMakesAGargoyleForItselfAndForAnotherTyranid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	harridan := effectstest.B12Push(g, me.ID, "Tyranid Harridan", "Creature — Tyranid", gapB193HarridanOracle, 4, 4)
	other := effectstest.B12Creature(g, me.ID, "Other Tyranid", "Creature — Tyranid", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, harridan, other, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardsNamed(g, "Tyranid Gargoyle", me.ID); got != 2 {
		t.Errorf("%d tokens, want 2 (the Harridan and the other Tyranid, not the Bear)", got)
	}
}

func TestGapB193HarridanIgnoresAnOpponentsTyranid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, opp.ID, "Tyranid Harridan", "Creature — Tyranid", gapB193HarridanOracle, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Tyranid", "Creature — Tyranid", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, mine)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardsNamed(g, "Tyranid Gargoyle", opp.ID); got != 0 {
		t.Errorf("an opponent's Harridan made %d tokens off my Tyranid", got)
	}
}

// --- Venomthrope ---

func TestGapB193VenomthropeHasFlyingDeathtouchAndHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Venomthrope", "Creature — Tyranid", gapB193VenomOracle, 2, 2)
	effectstest.AssertKeywords(t, g, id, "flying", "deathtouch", "hexproof")
}

// --- Acolyte Hybrid ---

func TestGapB193AcolyteHybridDestroysAnArtifactAndItsControllerDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(opp, "A", "B", "C")
	effectstest.AdvanceToMain(t, g)
	acolyte := effectstest.B12Push(g, me.ID, "Acolyte Hybrid", "Creature — Tyranid Human", gapB193AcolyteOracle, 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	hand := opp.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, acolyte)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if opp.Hand.Size() != hand+1 {
		t.Errorf("the artifact's controller's hand %d, want %d", opp.Hand.Size(), hand+1)
	}
}

func TestGapB193AcolyteHybridOwnArtifactMakesYouDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.AdvanceToMain(t, g)
	acolyte := effectstest.B12Push(g, me.ID, "Acolyte Hybrid", "Creature — Tyranid Human", gapB193AcolyteOracle, 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, acolyte)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: the destroyed artifact's controller draws, and that is me", me.Hand.Size(), hand+1)
	}
}

func TestGapB193AcolyteHybridMayChooseNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	acolyte := effectstest.B12Push(g, me.ID, "Acolyte Hybrid", "Creature — Tyranid Human", gapB193AcolyteOracle, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, acolyte)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(acolyte) {
		t.Error("the Acolyte left the battlefield")
	}
}

// --- Aspiring Champion ---

func TestGapB193AspiringChampionSacrificesAndDemonDealsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	champion := effectstest.B12Push(g, me.ID, "Aspiring Champion", "Creature — Astartes Warrior", gapB193ChampionOracle, 3, 3)
	effectstest.SeedLibrary(me, "Filler One", "Filler Two")
	demon := game.Card{InstanceID: uuid.New(), Name: "Big Demon", TypeLine: "Creature — Demon", ManaCost: "{4}{B}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(demon)
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, champion)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(champion) {
		t.Error("the Champion was not sacrificed")
	}
	if !g.Battlefield.Contains(demon.InstanceID) {
		t.Fatal("the Demon did not enter the battlefield")
	}
	// 3 combat damage from the Champion, then the Demon's 5 to each opponent.
	if opp.Life != life-3-5 {
		t.Errorf("opponent life %d -> %d, want 3 combat damage and 5 from the Demon", life, opp.Life)
	}
}

func TestGapB193AspiringChampionNonDemonDealsNoDamageAndTheLibraryIsShuffled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	champion := effectstest.B12Push(g, me.ID, "Aspiring Champion", "Creature — Astartes Warrior", gapB193ChampionOracle, 3, 3)
	effectstest.SeedLibrary(me, "Filler One", "Filler Two")
	bear := game.Card{InstanceID: uuid.New(), Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(bear)
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, champion)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear.InstanceID) {
		t.Fatal("the creature card did not enter the battlefield")
	}
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want only the 3 combat damage", life, opp.Life)
	}
}

func TestGapB193AspiringChampionWithNoCreatureInTheLibraryJustShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	champion := effectstest.B12Push(g, me.ID, "Aspiring Champion", "Creature — Astartes Warrior", gapB193ChampionOracle, 3, 3)
	effectstest.SeedLibrary(me, "Filler One", "Filler Two")
	size := me.Library.Size()
	effectstest.AttackWith(t, g, opp.ID, champion)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(champion) {
		t.Error("the Champion was not sacrificed")
	}
	if me.Library.Size() < size {
		t.Errorf("library %d cards, was %d: a library with no creature card keeps every card", me.Library.Size(), size)
	}
}

// --- Bloodcrusher of Khorne ---

func TestGapB193BloodcrusherGivesTrampleToOthersYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusher := effectstest.B12Push(g, me.ID, "Bloodcrusher of Khorne", "Creature — Demon Knight", gapB193BloodcrushOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AssertKeywords(t, g, crusher, "trample")
	if !effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("my other creature has no trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature has trample")
	}
}

func TestGapB193BloodcrusherOpponentsOneDoesNotHelpMe(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Bloodcrusher of Khorne", "Creature — Demon Knight", gapB193BloodcrushOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("an opponent's Bloodcrusher gave my creature trample")
	}
}

// --- Chaos Terminator Lord ---

func TestGapB193ChaosTerminatorLordGivesAnotherCreatureDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	lord := effectstest.B12Push(g, me.ID, "Chaos Terminator Lord", "Creature — Astartes Warrior", gapB193TerminatorOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	if effectstest.HasID(pick.PickTargetCards, lord) || effectstest.HasID(pick.PickTargetCards, theirs) || !effectstest.HasID(pick.PickTargetCards, bear) {
		t.Errorf("offered %v; want only another creature I control", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("the target has no double strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, lord, "double strike") {
		t.Error("the Lord gave itself double strike")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("double strike outlasted the turn")
	}
}
