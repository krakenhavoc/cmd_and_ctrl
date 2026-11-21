package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activationsOnStackFrom counts the activated abilities from
// `source` waiting on the stack.
func activationsOnStackFrom(g *game.Game, source uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemActivated && item.SourceCardID == source {
			n++
		}
	}
	return n
}

// --- the keyword action ------------------------------------------------

// Activating once puts the counters on, sets the designation, and
// fires the becomes-monstrous trigger with X on its event.
func TestMonstrosityPlacesCountersSetsTheFlagAndTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.PushMonster(g, me.ID, "Stormbreath Dragon", effectstest.StormbreathOracle, 4, 4)
	effectstest.GiveHandCards(opp, 4)
	hand := len(opp.Hand.Cards)
	life := opp.Life

	effectstest.ActivateMonstrosity(t, g, me.ID, dragon, 0)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.PlusOneCounters(g, dragon); n != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", n)
	}
	if !effectstest.MonstrousNow(g, dragon) {
		t.Error("not monstrous after Monstrosity 3 resolved")
	}
	if got := effectstest.BecameMonstrousEvents(g, dragon); len(got) != 1 || got[0] != 3 {
		t.Errorf("EventBecameMonstrous amounts = %v, want [3]", got)
	}
	if opp.Life != life-hand {
		t.Errorf("opponent life %d → %d, want -%d (the cards in their hand)", life, opp.Life, hand)
	}
	if !effectstest.E2Card(t, g, dragon).Monstrous {
		t.Error("Card.Monstrous not set")
	}
}

// CR 701.37a: a second activation is legal and does nothing — no
// counters, no event, no trigger.
func TestMonstrosityOnAMonstrousCreatureDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.PushMonster(g, me.ID, "Stormbreath Dragon", effectstest.StormbreathOracle, 4, 4)
	effectstest.GiveHandCards(opp, 2)
	hand := len(opp.Hand.Cards)
	life := opp.Life

	for i := 0; i < 2; i++ {
		effectstest.ActivateMonstrosity(t, g, me.ID, dragon, 0)
		effectstest.PassPriorityAroundTable(t, g)
	}

	if n := effectstest.PlusOneCounters(g, dragon); n != 3 {
		t.Errorf("+1/+1 counters = %d after two activations, want 3", n)
	}
	if got := effectstest.BecameMonstrousEvents(g, dragon); len(got) != 1 {
		t.Errorf("EventBecameMonstrous emitted %d times, want 1", len(got))
	}
	if opp.Life != life-hand {
		t.Errorf("opponent life %d → %d, want -%d once (the trigger fired twice?)", life, opp.Life, hand)
	}
}

// CR 701.37a's "if this creature isn't monstrous" is checked as the
// ability RESOLVES: a second activation in response to the first finds
// the creature monstrous by the time it resolves, and does nothing.
func TestMonstrosityInResponseToItselfResolvesToNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.PushMonster(g, me.ID, "Stormbreath Dragon", effectstest.StormbreathOracle, 4, 4)
	effectstest.GiveHandCards(opp, 3)
	hand := len(opp.Hand.Cards)
	life := opp.Life

	effectstest.ActivateMonstrosity(t, g, me.ID, dragon, 0)
	effectstest.ActivateMonstrosity(t, g, me.ID, dragon, 0)
	if n := activationsOnStackFrom(g, dragon); n != 2 {
		t.Fatalf("activations on the stack = %d, want both", n)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.PlusOneCounters(g, dragon); n != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 — the second resolution placed counters on a monstrous creature", n)
	}
	if got := effectstest.BecameMonstrousEvents(g, dragon); len(got) != 1 {
		t.Errorf("EventBecameMonstrous emitted %d times, want 1", len(got))
	}
	if opp.Life != life-hand {
		t.Errorf("opponent life %d → %d, want -%d once", life, opp.Life, hand)
	}
}

// The counters ride the CR 614 pipeline: Doubling Season doubles them.
// X on the trigger stays the announced X (Polukranos's and Hydra
// Broodmaster's rulings), so the tokens are 2/2 and there are two of
// them — times two again from Doubling Season's token half.
func TestMonstrosityDoublingSeasonDoublesCountersNotX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	brood := effectstest.PushMonster(g, me.ID, "Hydra Broodmaster", effectstest.HydraBroodmasterOracle, 7, 7)

	effectstest.ActivateMonstrosity(t, g, me.ID, brood, 2)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.PlusOneCounters(g, brood); n != 4 {
		t.Errorf("+1/+1 counters = %d, want 4 (2 doubled)", n)
	}
	if got := effectstest.BecameMonstrousEvents(g, brood); len(got) != 1 || got[0] != 2 {
		t.Errorf("EventBecameMonstrous amounts = %v, want [2] — the announced X, not the counters that landed", got)
	}
	hydras := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Hydra" && c.Controller == me.ID {
			hydras++
			if c.Power != 2 || c.Toughness != 2 {
				t.Errorf("Hydra token is %d/%d, want 2/2", c.Power, c.Toughness)
			}
		}
	}
	if hydras != 4 {
		t.Errorf("Hydra tokens = %d, want 4 (X=2, doubled by Doubling Season's token half)", hydras)
	}
}

// Two different counter replacements pause the placement on the CR 616
// ordering prompt. The creature is not monstrous — and nothing has
// triggered — until the counters land.
func TestMonstrosityWaitsForAPausedPlacement(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	dragon := effectstest.PushMonster(g, me.ID, "Stormbreath Dragon", effectstest.StormbreathOracle, 4, 4)
	effectstest.GiveHandCards(opp, 1)

	effectstest.ActivateMonstrosity(t, g, me.ID, dragon, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.MonstrousNow(g, dragon) {
		t.Fatal("monstrous while the counters are still owed to the CR 616 prompt")
	}

	effectstest.AnswerOrderBySource(t, g, ds, hs) // 3 → 6 → 7
	if n := effectstest.PlusOneCounters(g, dragon); n != 7 {
		t.Errorf("+1/+1 counters = %d, want 7 ((3*2)+1)", n)
	}
	if !effectstest.MonstrousNow(g, dragon) {
		t.Fatal("not monstrous after the placement settled")
	}
	if got := effectstest.BecameMonstrousEvents(g, dragon); len(got) != 1 || got[0] != 3 {
		t.Errorf("EventBecameMonstrous amounts = %v, want [3]", got)
	}
}

// CR 400.7 + CR 701.37b: a flickered monster is a new object, not
// monstrous, and can become monstrous again.
func TestMonstrousClearsOnFlicker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lion := effectstest.PushMonster(g, me.ID, "Fleecemane Lion", effectstest.FleecemaneLionOracle, 3, 3)
	effectstest.ActivateMonstrosity(t, g, me.ID, lion, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.MonstrousNow(g, lion) {
		t.Fatal("setup: not monstrous")
	}

	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()}
		if err := (effects.Flicker{Target: lion}).Apply(effectstest.CtxFor(g, item)); err != nil {
			t.Fatalf("Flicker: %v", err)
		}
	})
	back := effectstest.FindBattlefieldByName(g, "Fleecemane Lion")
	if back == uuid.Nil {
		t.Fatal("the Lion did not come back")
	}
	if effectstest.MonstrousNow(g, back) || effectstest.E2Card(t, g, back).Monstrous {
		t.Error("the flickered Lion is still monstrous")
	}
	if effectstest.HasString(effectstest.EffectiveAbilities(t, g, back), "hexproof") {
		t.Error("the flickered Lion kept hexproof")
	}

	effectstest.ActivateMonstrosity(t, g, me.ID, back, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.MonstrousNow(g, back) || effectstest.PlusOneCounters(g, back) != 1 {
		t.Error("the new object could not become monstrous")
	}
}

// CR 701.37b: monstrous is not a copiable value. A Clone of a
// monstrous Fleecemane Lion is an ordinary 3/3 with no counters, and
// its gated statics are off.
func TestACopyOfAMonstrousCreatureIsNotMonstrous(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	lion := effectstest.PushMonster(g, active.ID, "Fleecemane Lion", effectstest.FleecemaneLionOracle, 3, 3)
	effectstest.ActivateMonstrosity(t, g, active.ID, lion, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.MonstrousNow(g, lion) {
		t.Fatal("setup: not monstrous")
	}

	cloneID := effectstest.CastCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", effectstest.OracleClone, nil)
	effectstest.ResolveWithCopyChoice(t, g, lion)

	c := effectstest.E2Card(t, g, cloneID)
	if c.Name != "Fleecemane Lion" {
		t.Fatalf("clone name = %q", c.Name)
	}
	if c.Monstrous {
		t.Error("the copy is monstrous")
	}
	if c.Counters[game.CounterPlusOne] != 0 {
		t.Error("the copy took the counters")
	}
	if abs := effectstest.EffectiveAbilities(t, g, cloneID); effectstest.HasString(abs, "hexproof") || effectstest.HasString(abs, "indestructible") {
		t.Errorf("the copy has the monstrous-gated keywords: %v", abs)
	}
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, lion), "hexproof") {
		t.Error("the original lost hexproof")
	}
}

// Undo and the snapshot: the designation is carried, and an undo
// across the resolution takes it back.
func TestMonstrousSurvivesSnapshotAndUndo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lion := effectstest.PushMonster(g, me.ID, "Fleecemane Lion", effectstest.FleecemaneLionOracle, 3, 3)
	pre := g.Clone()

	effectstest.ActivateMonstrosity(t, g, me.ID, lion, 0)
	effectstest.PassPriorityAroundTable(t, g)

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if !effectstest.MonstrousNow(restored, lion) {
		t.Error("the snapshot round-trip dropped the monstrous designation")
	}
	if cl := g.Clone(); !effectstest.MonstrousNow(cl, lion) {
		t.Error("Clone dropped the monstrous designation")
	}

	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	if effectstest.MonstrousNow(g, lion) {
		t.Error("still monstrous after undoing the activation")
	}
	if effectstest.HasString(effectstest.EffectiveAbilities(t, g, lion), "hexproof") {
		t.Error("hexproof survived the undo")
	}
}

// --- the cards ---------------------------------------------------------

// Polukranos: X damage divided among opponents' creatures, and each
// target hits back for its power. A creature of mine is not offered.
func TestPolukranosDividesXAndTakesDamageBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	polu := effectstest.PushMonster(g, me.ID, "Polukranos, World Eater", effectstest.PolukranosOracle, 5, 5)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 2, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 1, 30)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Wall", 1, 30)

	effectstest.ActivateMonstrosity(t, g, me.ID, polu, 3)
	effectstest.PassPriorityAroundTable(t, g)

	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the becomes-monstrous trigger asked for no targets")
	}
	if effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("a creature you control is offered")
	}
	if effectstest.HasID(p.PickTargetCards, polu) {
		t.Error("Polukranos is offered to its own trigger")
	}
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 2}, a, b)
	effectstest.PassPriorityAroundTable(t, g)

	if d1, d2 := effectstest.E2Card(t, g, a).DamageMarked, effectstest.E2Card(t, g, b).DamageMarked; d1 != 1 || d2 != 2 {
		t.Errorf("division landed %d/%d, want 1/2", d1, d2)
	}
	if d := effectstest.E2Card(t, g, polu).DamageMarked; d != 3 {
		t.Errorf("Polukranos took %d, want 3 (powers 2 + 1)", d)
	}
	if n := effectstest.PlusOneCounters(g, polu); n != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", n)
	}
}

// Every target must be assigned at least 1, so X caps the target count.
func TestPolukranosCannotTargetMoreThanX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	polu := effectstest.PushMonster(g, me.ID, "Polukranos, World Eater", effectstest.PolukranosOracle, 5, 5)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)

	effectstest.ActivateMonstrosity(t, g, me.ID, polu, 1)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	refs := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{a: 1, b: 1}); err == nil {
		t.Error("two targets accepted for X = 1")
	}
}

// X = 0: monstrous, but the trigger targets nothing and nothing hits back.
func TestPolukranosAtXZeroTargetsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	polu := effectstest.PushMonster(g, me.ID, "Polukranos, World Eater", effectstest.PolukranosOracle, 5, 5)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 4, 30)

	effectstest.ActivateMonstrosity(t, g, me.ID, polu, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("X = 0 still asked for targets")
	}
	if !effectstest.MonstrousNow(g, polu) {
		t.Error("X = 0 did not make Polukranos monstrous")
	}
	if effectstest.E2Card(t, g, a).DamageMarked != 0 || effectstest.E2Card(t, g, polu).DamageMarked != 0 {
		t.Error("damage was dealt at X = 0")
	}
}

func TestStormbreathDragonKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.PushMonster(g, me.ID, "Stormbreath Dragon", effectstest.StormbreathOracle, 4, 4)
	abs := effectstest.EffectiveAbilities(t, g, dragon)
	for _, kw := range []string{"flying", "haste", "protection from white"} {
		if !effectstest.HasString(abs, kw) {
			t.Errorf("missing %q: %v", kw, abs)
		}
	}
}

// Reach arrives with the designation; before it, the Giant has only
// vigilance.
func TestHundredHandedOneGainsReachWhenMonstrous(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	giant := effectstest.PushMonster(g, me.ID, "Hundred-Handed One", effectstest.HundredHandedOneOracle, 3, 5)
	if abs := effectstest.EffectiveAbilities(t, g, giant); effectstest.HasString(abs, "reach") || !effectstest.HasString(abs, "vigilance") {
		t.Fatalf("before monstrosity: %v, want vigilance and no reach", abs)
	}
	effectstest.ActivateMonstrosity(t, g, me.ID, giant, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, giant), "reach") {
		t.Error("no reach once monstrous")
	}
	if c := effectstest.E2Card(t, g, giant); c.CurrentPower() != 6 || c.CurrentToughness() != 8 {
		p, tg := c.CurrentPower(), c.CurrentToughness()
		t.Errorf("P/T = %d/%d, want 6/8", p, tg)
	}
}

// Each player — you too — sacrifices three lands; a nonland is never
// offered, and a player with two lands loses two.
func TestEmberSwallowerEachPlayerSacrificesThreeLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swallower := effectstest.PushMonster(g, me.ID, "Ember Swallower", effectstest.EmberSwallowerOracle, 4, 5)
	for i := 0; i < 4; i++ {
		effectstest.PushLand(g, me.ID, "My Island")
	}
	effectstest.PushLand(g, opp.ID, "Their Island")
	effectstest.PushLand(g, opp.ID, "Their Island")
	theirBear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	effectstest.ActivateMonstrosity(t, g, me.ID, swallower, 0)
	effectstest.PassPriorityAroundTable(t, g)

	for i := 0; i < 16; i++ {
		c := effectstest.SacrificeChoiceFor(g, me.ID)
		if c == nil {
			c = effectstest.SacrificeChoiceFor(g, opp.ID)
		}
		if c == nil {
			break
		}
		opts := effectstest.SacrificeOptions(c)
		if effectstest.HasID(opts, swallower) || effectstest.HasID(opts, theirBear) {
			t.Fatal("a nonland permanent is offered to a land edict")
		}
		// #3664: a seat owing three lands names all three in one pick.
		if c.ChooseMax != 3 || len(opts) < 3 {
			t.Fatalf("the pick takes %d of %d lands, want 3", c.ChooseMax, len(opts))
		}
		effectstest.AnswerSacrifice(t, g, c.Chooser, opts[:3]...)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountBattlefieldNamed(g, me.ID, "My Island"); n != 1 {
		t.Errorf("my lands = %d, want 1 (4 - 3)", n)
	}
	if n := effectstest.CountBattlefieldNamed(g, opp.ID, "Their Island"); n != 0 {
		t.Errorf("their lands = %d, want 0", n)
	}
	if !g.Battlefield.Contains(theirBear) || !g.Battlefield.Contains(swallower) {
		t.Error("a nonland permanent was sacrificed")
	}
}

// Destroys an opponent's flier; neither your own flier nor their
// ground creature is a legal target.
func TestArborColossusDestroysAnOpposingFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colossus := effectstest.PushMonster(g, me.ID, "Arbor Colossus", effectstest.ArborColossusOracle, 6, 6)
	flier := func(owner uuid.UUID, name string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bird", Keywords: []string{"flying"},
			Power: 1, Toughness: 1, Owner: owner, Controller: owner,
		})
	}
	theirFlier := flier(opp.ID, "Their Bird")
	myFlier := flier(me.ID, "My Bird")
	theirGround := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	effectstest.ActivateMonstrosity(t, g, me.ID, colossus, 0)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(p.PickTargetCards, myFlier) || effectstest.HasID(p.PickTargetCards, theirGround) {
		t.Errorf("illegal targets offered: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, theirFlier)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirFlier) {
		t.Error("the flier survived")
	}
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, colossus), "reach") {
		t.Error("Arbor Colossus has no reach")
	}
}

// With no opposing flier the trigger never asks (CR 603.3d) and the
// Giant is still monstrous.
func TestArborColossusWithNoFlierJustBecomesMonstrous(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colossus := effectstest.PushMonster(g, me.ID, "Arbor Colossus", effectstest.ArborColossusOracle, 6, 6)
	effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.ActivateMonstrosity(t, g, me.ID, colossus, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Error("asked for a target with no flier on the table")
	}
	if !effectstest.MonstrousNow(g, colossus) {
		t.Error("not monstrous")
	}
}

// Hexproof and indestructible only once monstrous.
func TestFleecemaneLionIsProtectedOnlyWhenMonstrous(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lion := effectstest.PushMonster(g, me.ID, "Fleecemane Lion", effectstest.FleecemaneLionOracle, 3, 3)
	if abs := effectstest.EffectiveAbilities(t, g, lion); effectstest.HasString(abs, "hexproof") || effectstest.HasString(abs, "indestructible") {
		t.Fatalf("protected before monstrosity: %v", abs)
	}
	effectstest.ActivateMonstrosity(t, g, me.ID, lion, 0)
	effectstest.PassPriorityAroundTable(t, g)
	abs := effectstest.EffectiveAbilities(t, g, lion)
	if !effectstest.HasString(abs, "hexproof") || !effectstest.HasString(abs, "indestructible") {
		t.Errorf("monstrous Lion abilities = %v, want hexproof and indestructible", abs)
	}
	if p := effectstest.E2Card(t, g, lion).CurrentPower(); p != 4 {
		t.Errorf("power = %d, want 4", p)
	}
}

// Monstrosity X puts X counters on and switches trample on.
func TestDomesticatedHydraMonstrosityX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hydra := effectstest.PushMonster(g, me.ID, "Domesticated Hydra", effectstest.DomesticatedHydraOrcl, 3, 3)
	if effectstest.HasString(effectstest.EffectiveAbilities(t, g, hydra), "trample") {
		t.Fatal("trample before monstrosity")
	}
	effectstest.ActivateMonstrosity(t, g, me.ID, hydra, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, hydra); n != 4 {
		t.Errorf("+1/+1 counters = %d, want 4", n)
	}
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, hydra), "trample") {
		t.Error("no trample once monstrous")
	}
}

// X = 0 places no counters — so no counter event invalidates the layer
// pass — and the gated trample must still switch on: the designation's
// own event is what announces the change.
func TestDomesticatedHydraAtXZeroStillGainsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hydra := effectstest.PushMonster(g, me.ID, "Domesticated Hydra", effectstest.DomesticatedHydraOrcl, 3, 3)
	if effectstest.HasString(effectstest.EffectiveAbilities(t, g, hydra), "trample") {
		t.Fatal("trample before monstrosity")
	}
	effectstest.ActivateMonstrosity(t, g, me.ID, hydra, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOneCounters(g, hydra) != 0 {
		t.Fatal("X = 0 placed counters")
	}
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, hydra), "trample") {
		t.Error("a Hydra made monstrous at X = 0 has no trample")
	}
}

// X X/X Hydras; at X = 0, none — and the Broodmaster is spent.
func TestHydraBroodmasterMakesXHydras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	brood := effectstest.PushMonster(g, me.ID, "Hydra Broodmaster", effectstest.HydraBroodmasterOracle, 7, 7)
	effectstest.ActivateMonstrosity(t, g, me.ID, brood, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Hydra"); n != 3 {
		t.Fatalf("Hydra tokens = %d, want 3", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Hydra" && (c.Power != 3 || c.Toughness != 3) {
			t.Errorf("token is %d/%d, want 3/3", c.Power, c.Toughness)
		}
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	brood2 := effectstest.PushMonster(g2, me2.ID, "Hydra Broodmaster", effectstest.HydraBroodmasterOracle, 7, 7)
	effectstest.ActivateMonstrosity(t, g2, me2.ID, brood2, 0)
	effectstest.PassPriorityAroundTable(t, g2)
	if n := effectstest.CountBattlefieldNamed(g2, me2.ID, "Hydra"); n != 0 {
		t.Errorf("X = 0 made %d tokens", n)
	}
	if !effectstest.MonstrousNow(g2, brood2) {
		t.Error("X = 0 did not make it monstrous")
	}
}
