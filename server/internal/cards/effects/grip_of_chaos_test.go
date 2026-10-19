package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// grip_of_chaos_test.go — #4050: "Whenever a spell or ability is put
// onto the stack, if it has a single target, reselect its target at
// random." One test per way an object reaches the stack (cast, spell
// copy, activation, placed trigger, ability copy), plus the intervening
// if, the loop guard and the replay.

const (
	gripOfChaosOracle     = "ca2e0c8f-7ece-465f-9103-ba60712eacf7"
	gripBombardmentOracle = "edad60c6-80de-4033-af1b-a703ac332983" // Goblin Bombardment
	gripLegalCreatures    = 2
)

// gripItems is every Grip of Chaos trigger currently on the stack.
func gripItems(g *game.Game) []*game.StackItem {
	var out []*game.StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemTriggered && strings.HasPrefix(it.Label, "Grip of Chaos") {
				out = append(out, it)
			}
		}
	})
	return out
}

// gripTargets is the targets of the stack item `id`.
func gripTargets(g *game.Game, id uuid.UUID) []game.TargetRef {
	var out []game.TargetRef
	g.ReadSnapshot(func() {
		if it := g.StackMeta[id]; it != nil {
			out = append([]game.TargetRef(nil), it.Targets...)
		}
	})
	return out
}

// gripSettle passes priority until no Grip trigger is left on the
// stack, and no further: the object Grip watched is still there to
// look at.
func gripSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 24 && len(gripItems(g)) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				return
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if n := len(gripItems(g)); n > 0 {
		t.Fatalf("%d Grip of Chaos triggers never resolved", n)
	}
}

// Doom Blade aimed at one of two creatures: Grip's trigger sits above it,
// resolves first and the Blade then goes at a creature drawn at random
// from both (the one it had included). Over many keys both are reached,
// nobody is ever asked, and the creature the Blade finally kills is the
// one it ended up targeting.
func TestGripOfChaosReselectsASingleTargetSpell(t *testing.T) {
	seen := map[int]bool{}
	for key := byte(0); key < 48 && len(seen) < gripLegalCreatures; key++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		me, opp := g.Seats[0], g.Seats[1]
		pushCatalogPermanent(g, me.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
		a := pushTypedCard(g, opp.ID, "Bear A", "Creature — Bear", "{1}{G}")
		b := pushTypedCard(g, opp.ID, "Bear B", "Creature — Bear", "{1}{G}")

		spell := castCatalogSpell(t, g, "Doom Blade", "Instant", doomBladeOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: a}})
		if n := len(gripItems(g)); n != 1 {
			t.Fatalf("key %d: %d Grip triggers on the stack after the cast, want 1", key, n)
		}
		gripSettle(t, g)
		if p := latestPickTarget(g, me.ID); p != nil {
			t.Fatalf("key %d: a target chosen at random asks nobody: %+v", key, p)
		}
		got := gripTargets(g, spell)
		if len(got) != 1 || (got[0].ID != a && got[0].ID != b) {
			t.Fatalf("key %d: the Blade now targets %+v, want one of the two bears", key, got)
		}
		final := got[0].ID
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(final) {
			t.Errorf("key %d: Doom Blade did not destroy the creature it ended up targeting", key)
		}
		other := a
		if final == a {
			other = b
		}
		if !g.Battlefield.Contains(other) {
			t.Errorf("key %d: the creature it did not target died", key)
		}
		if final == a {
			seen[0] = true
		} else {
			seen[1] = true
		}
	}
	if len(seen) != gripLegalCreatures {
		t.Errorf("48 draws reached %d of the 2 creatures", len(seen))
	}
}

// The draw is the keyed stream: an undo of the trigger's resolution and a
// second resolution points the spell where the first did.
func TestGripOfChaosReplaysTheDrawAfterUndo(t *testing.T) {
	seen := map[uuid.UUID]bool{}
	for key := byte(0); key < 12; key++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		me, opp := g.Seats[0], g.Seats[1]
		pushCatalogPermanent(g, me.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
		var ids [3]uuid.UUID
		for i := range ids {
			ids[i] = pushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
		}
		spell := castCatalogSpell(t, g, "Doom Blade", "Instant", doomBladeOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: ids[0]}})
		pre := g.Clone()
		resolve := func() uuid.UUID {
			gripSettle(t, g)
			got := gripTargets(g, spell)
			if len(got) != 1 {
				t.Fatalf("key %d: targets %+v", key, got)
			}
			return got[0].ID
		}
		first := resolve()
		g.WithWriteLock(func() { g.RestoreFrom(pre) })
		second := resolve()
		if first != second {
			t.Errorf("key %d: undo then redo drew %s then %s", key, first, second)
		}
		seen[first] = true
	}
	if len(seen) < 2 {
		t.Errorf("12 keys always drew the same bear: %v", seen)
	}
}

// With no other legal target the target stays, and the table is told.
func TestGripOfChaosKeepsTheOnlyLegalTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
	only := pushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	spell := castCatalogSpell(t, g, "Doom Blade", "Instant", doomBladeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: only}})
	gripSettle(t, g)
	if got := gripTargets(g, spell); len(got) != 1 || got[0].ID != only {
		t.Fatalf("the only creature is the only target: %+v", got)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(only) {
		t.Error("the Blade should have resolved against the bear")
	}
}

// "Single target" is CR 115.9a: Arc Trail has two, so Grip never
// triggers for it.
func TestGripOfChaosLeavesATwoTargetSpellAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
	big := pushCostedPermanentForTest(g, opp.ID, "Big", "Creature — Beast", "{2}{G}")
	pushCostedPermanentForTest(g, opp.ID, "Other", "Creature — Beast", "{2}{G}")
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: big, Slot: 1}}
	spell := castCatalogSpell(t, g, "Arc Trail", "Sorcery", arcTrailOracle, targets)
	if n := len(gripItems(g)); n != 0 {
		t.Fatalf("%d Grip triggers for a two-target spell, want 0", n)
	}
	got := gripTargets(g, spell)
	if len(got) != 2 || got[0].ID != opp.ID || got[1].ID != big {
		t.Errorf("Arc Trail's targets changed: %+v", got)
	}
}

// A spell with no target is not touched either.
func TestGripOfChaosIgnoresAnUntargetedSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
	castCatalogSpell(t, g, "Divination Stand-in", "Sorcery", "", nil)
	if n := len(gripItems(g)); n != 0 {
		t.Fatalf("%d Grip triggers for an untargeted spell, want 0", n)
	}
}

// An activated ability is "put onto the stack" too: Goblin Bombardment
// pointed at a player is redirected to anything Bombardment's "any
// target" admits.
func TestGripOfChaosReselectsAnActivatedAbility(t *testing.T) {
	seen := map[uuid.UUID]bool{}
	var bear uuid.UUID
	for key := byte(0); key < 64 && len(seen) < 3; key++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		me, opp := g.Seats[0], g.Seats[1]
		pushCatalogPermanent(g, me.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
		bomb := pushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", gripBombardmentOracle, false)
		fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
		bear = pushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")

		if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
			SacrificeIDs: []uuid.UUID{fodder},
			Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		items := gripItems(g)
		if len(items) != 1 {
			t.Fatalf("key %d: %d Grip triggers after the activation, want 1", key, len(items))
		}
		var ability uuid.UUID
		g.ReadSnapshot(func() {
			for id, it := range g.StackMeta {
				if it != nil && it.Kind == game.StackItemActivated {
					ability = id
				}
			}
		})
		if ability == uuid.Nil {
			t.Fatal("the activation is not on the stack")
		}
		gripSettle(t, g)
		got := gripTargets(g, ability)
		if len(got) != 1 {
			t.Fatalf("key %d: ability targets %+v, want one", key, got)
		}
		seen[got[0].ID] = true
	}
	if len(seen) < 3 {
		t.Errorf("64 draws reached %d distinct targets, want the 4 players and the bear to be reachable (saw %v, bear %v)", len(seen), seen, bear)
	}
}

// A triggered ability is put onto the stack by the CR 603.3b drain, and
// Grip's trigger lands above it: Acidic Slime picks a land, and Grip
// redraws among everything the Slime could have targeted, itself
// included in the legal set (an enchantment).
func TestGripOfChaosReselectsAPlacedTrigger(t *testing.T) {
	seen := map[uuid.UUID]bool{}
	for key := byte(0); key < 48 && len(seen) < 3; key++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		me, opp := g.Seats[0], g.Seats[1]
		grip := pushCatalogPermanent(g, me.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
		mountain := pushTypedCard(g, opp.ID, "Mountain", "Basic Land — Mountain", "")
		island := pushTypedCard(g, opp.ID, "Island", "Basic Land — Island", "")
		slime := castAndResolveCreature(t, g, "Acidic Slime", "Creature — Ooze", acidicSlimeOracle)
		pickCard(t, g, me.ID, mountain)
		trig := triggerOnStack(g, slime)
		if trig == nil {
			t.Fatalf("key %d: the Slime's trigger is not on the stack", key)
		}
		if n := len(gripItems(g)); n != 1 {
			t.Fatalf("key %d: %d Grip triggers above the Slime's, want 1", key, n)
		}
		gripSettle(t, g)
		if p := latestPickTarget(g, me.ID); p != nil {
			t.Fatalf("key %d: asked %+v", key, p)
		}
		got := gripTargets(g, trig.ID)
		if len(got) != 1 {
			t.Fatalf("key %d: the Slime's trigger targets %+v", key, got)
		}
		if id := got[0].ID; id != mountain && id != island && id != grip {
			t.Fatalf("key %d: the Slime's trigger now targets %v, not a land or an enchantment", key, id)
		}
		seen[got[0].ID] = true
	}
	if len(seen) < 3 {
		t.Errorf("48 draws reached %d of the 3 legal permanents", len(seen))
	}
}

// Grip's own trigger targets nothing, so it never triggers itself: the
// stack settles even with two Grips watching each other's triggers.
func TestTwoGripsOfChaosDoNotLoop(t *testing.T) {
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(randomModesKey(3))
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
	pushCatalogPermanent(g, opp.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
	a := pushTypedCard(g, opp.ID, "Bear A", "Creature — Bear", "{1}{G}")
	b := pushTypedCard(g, opp.ID, "Bear B", "Creature — Bear", "{1}{G}")
	spell := castCatalogSpell(t, g, "Doom Blade", "Instant", doomBladeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}})
	if n := len(gripItems(g)); n != 2 {
		t.Fatalf("%d Grip triggers for the one spell, want one from each Grip", n)
	}
	gripSettle(t, g)
	if got := gripTargets(g, spell); len(got) != 1 || (got[0].ID != a && got[0].ID != b) {
		t.Fatalf("after two redraws the Blade targets %+v", got)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) && g.Battlefield.Contains(b) {
		t.Error("the Blade killed nothing")
	}
}

// Intervening if (CR 603.4): the spell is countered while Grip's trigger
// waits, so the trigger finds nothing to reselect and does nothing.
func TestGripOfChaosDoesNothingForASpellThatLeftTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
	a := pushTypedCard(g, opp.ID, "Bear A", "Creature — Bear", "{1}{G}")
	pushTypedCard(g, opp.ID, "Bear B", "Creature — Bear", "{1}{G}")
	spell := castCatalogSpell(t, g, "Doom Blade", "Instant", doomBladeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}})
	// The Blade leaves the stack with Grip's trigger still on it.
	g.WithWriteLock(func() {
		delete(g.StackMeta, spell)
		g.Stack.Remove(spell)
	})
	gripSettle(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) {
		t.Error("nothing was left to resolve against the bear")
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventEffectError {
			t.Errorf("Grip's trigger errored for a spell that left: %+v", ev)
		}
	}
}

// The same trigger re-checks "a single target" as it resolves: the spell
// gained a second target in the meantime and is left alone.
func TestGripOfChaosRechecksTheSingleTargetAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
	a := pushTypedCard(g, opp.ID, "Bear A", "Creature — Bear", "{1}{G}")
	b := pushTypedCard(g, opp.ID, "Bear B", "Creature — Bear", "{1}{G}")
	spell := castCatalogSpell(t, g, "Doom Blade", "Instant", doomBladeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}})
	g.WithWriteLock(func() {
		it := g.StackMeta[spell]
		it.Targets = append(it.Targets, game.TargetRef{Kind: game.TargetCard, ID: b, Slot: 1})
	})
	gripSettle(t, g)
	got := gripTargets(g, spell)
	if len(got) != 2 || got[0].ID != a || got[1].ID != b {
		t.Errorf("a spell with two targets was reselected: %+v", got)
	}
}

// A spell copy is put onto the stack too (CR 707.10, EventSpellCopied).
func TestGripOfChaosReselectsASpellCopy(t *testing.T) {
	seen := map[uuid.UUID]bool{}
	for key := byte(0); key < 48 && len(seen) < 2; key++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		me, opp := g.Seats[0], g.Seats[1]
		a := pushTypedCard(g, opp.ID, "Bear A", "Creature — Bear", "{1}{G}")
		b := pushTypedCard(g, opp.ID, "Bear B", "Creature — Bear", "{1}{G}")
		// The spell goes on the stack before Grip is there to see it, so
		// the only trigger is the copy's.
		spell := castCatalogSpell(t, g, "Doom Blade", "Instant", doomBladeOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: a}})
		pushCatalogPermanent(g, me.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
		var copyID uuid.UUID
		g.WithWriteLock(func() {
			id, err := g.CopySpellKeepingTargetsForEffect(spell, me.ID)
			if err != nil {
				t.Fatalf("copy: %v", err)
			}
			copyID = id
		})
		// A copy made inside a resolution is followed by the priority
		// boundary that puts its triggers on the stack; made from outside
		// one, the test runs that boundary itself.
		g.RunStateChecksForTest()
		if n := len(gripItems(g)); n != 1 {
			t.Fatalf("key %d: %d Grip triggers for the copy, want 1", key, n)
		}
		gripSettle(t, g)
		got := gripTargets(g, copyID)
		if len(got) != 1 || (got[0].ID != a && got[0].ID != b) {
			t.Fatalf("key %d: the copy targets %+v", key, got)
		}
		if orig := gripTargets(g, spell); len(orig) != 1 || orig[0].ID != a {
			t.Fatalf("key %d: the original was reselected: %+v", key, orig)
		}
		seen[got[0].ID] = true
	}
	if len(seen) < 2 {
		t.Errorf("48 draws reached %d of the 2 creatures", len(seen))
	}
}

// A copy of an ability is put onto the stack too, and it is neither
// activated nor triggered (CR 707.10a), so it needs EventAbilityCopied.
// The event exists only while something watches it.
func TestGripOfChaosReselectsAnAbilityCopyAndTheEventIsGated(t *testing.T) {
	setup := func(key byte, withGrip bool) (g *game.Game, ability, copyOf uuid.UUID, bear, other game.TargetRef) {
		g = newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		me, opp := g.Seats[0], g.Seats[1]
		bomb := pushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", gripBombardmentOracle, false)
		fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
		bear = game.TargetRef{Kind: game.TargetCard, ID: pushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")}
		other = game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
		if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
			SacrificeIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{other},
		}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		g.ReadSnapshot(func() {
			for id, it := range g.StackMeta {
				if it != nil && it.Kind == game.StackItemActivated {
					ability = id
				}
			}
		})
		if withGrip {
			pushCatalogPermanent(g, me.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
		}
		g.WithWriteLock(func() {
			if err := g.CopyAbilityForEffect(ability, me.ID, false); err != nil {
				t.Fatalf("copy ability: %v", err)
			}
		})
		g.ReadSnapshot(func() {
			for id, it := range g.StackMeta {
				if it != nil && it.IsCopy {
					copyOf = id
				}
			}
		})
		return
	}

	g, _, _, _, _ := setup(1, false)
	if n := acCountEvents(g, game.EventAbilityCopied); n != 0 {
		t.Errorf("%d ability_copied events with nothing watching, want 0", n)
	}

	seen := map[uuid.UUID]bool{}
	for key := byte(0); key < 64 && len(seen) < 2; key++ {
		g, ability, cp, _, _ := setup(key, true)
		if cp == uuid.Nil {
			t.Fatal("no copy on the stack")
		}
		if n := acCountEvents(g, game.EventAbilityCopied); n != 1 {
			t.Fatalf("key %d: %d ability_copied events, want 1", key, n)
		}
		g.RunStateChecksForTest()
		if n := len(gripItems(g)); n != 1 {
			t.Fatalf("key %d: %d Grip triggers for the copy, want 1", key, n)
		}
		gripSettle(t, g)
		if orig := gripTargets(g, ability); len(orig) != 1 || orig[0].Kind != game.TargetPlayer {
			t.Fatalf("key %d: the original ability was reselected: %+v", key, orig)
		}
		got := gripTargets(g, cp)
		if len(got) != 1 {
			t.Fatalf("key %d: the copy targets %+v", key, got)
		}
		seen[got[0].ID] = true
	}
	if len(seen) < 2 {
		t.Errorf("64 draws moved the copy to %d distinct targets, want at least 2", len(seen))
	}
}

// Grip watches the five ways an object reaches the stack, and its own
// trigger declares no target (the loop guard).
func TestGripOfChaosWatchesEveryWayOntoTheStackAndTargetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Grip of Chaos", "Enchantment", gripOfChaosOracle, false)
	trig := game.TriggersForCard(game.Card{OracleID: gripOfChaosOracle})
	if len(trig) != 1 {
		t.Fatalf("Grip of Chaos declares %d triggers, want 1", len(trig))
	}
	want := map[game.EventKind]bool{
		game.EventCast: true, game.EventSpellCopied: true, game.EventActivateAbility: true,
		game.EventTriggerPlaced: true, game.EventAbilityCopied: true,
	}
	if len(trig[0].Watches) != len(want) {
		t.Fatalf("Grip watches %v", trig[0].Watches)
	}
	for _, k := range trig[0].Watches {
		if !want[k] {
			t.Errorf("Grip watches %q, which is not a way onto the stack", k)
		}
	}
	if trig[0].Targets != nil || trig[0].TargetsFrom != nil {
		t.Error("Grip's own trigger must target nothing, or it would trigger itself")
	}
}
