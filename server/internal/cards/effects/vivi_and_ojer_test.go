package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	viviOrnitierOracle = "0452be2a-e97a-4269-a9e3-b616265ecb2e"
	ojerAxonilOracle   = "d3b7b541-6f05-46c1-8031-c848c4bd4635"
)

func TestViviAndOjerAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		viviOrnitierOracle: "Vivi Ornitier",
		ojerAxonilOracle:   "Ojer Axonil, Deepest Might",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- Vivi Ornitier ---------------------------------------------------

// The trigger half: a noncreature spell banks a +1/+1 counter and
// pings every opponent for 1. A creature spell does neither.
func TestViviBanksACounterAndPingsEachOpponentOnANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vivi := b12Push(g, me.ID, "Vivi Ornitier", "Legendary Creature — Wizard", viviOrnitierOracle, 0, 3)

	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}

	castCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, vivi, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d after a noncreature spell, want 1", got)
	}
	for _, p := range g.Seats {
		want := before[p.ID]
		if p.ID != me.ID {
			want--
		}
		if p.Life != want {
			t.Errorf("seat %s life %d, want %d — each OPPONENT takes 1 and you take none", p.Name, p.Life, want)
		}
		before[p.ID] = p.Life
	}

	// "a NONCREATURE spell": a creature spell is not one.
	castCatalogSpell(t, g, "Sandbox Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, vivi, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d after a creature spell, want the same 1", got)
	}
	for _, p := range g.Seats {
		if p.Life != before[p.ID] {
			t.Errorf("seat %s took damage off a creature spell: %d → %d", p.Name, before[p.ID], p.Life)
		}
	}
}

// The mana half: X is Vivi's CURRENT power, each point is its own
// {U}-or-{R} pick, and the ability is gone for the rest of the turn.
func TestViviAddsOneManaPerPointOfPowerOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vivi := b12Push(g, me.ID, "Vivi Ornitier", "Legendary Creature — Wizard", viviOrnitierOracle, 0, 3)
	advanceToMain(t, g)

	// Counters are power, so three of them are three mana.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(vivi, game.CounterPlusOne, 3) })
	activateManaFor(t, g, me.ID, vivi, 0, game.ManaAbilityParams{})

	// "in any combination of {U} and/or {R}" — one pick per slot, so
	// the three can be split. Answer two blue and one red.
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("no mana_pick; each {U|R} slot must ask which colour")
	}
	if len(pick.ColorOptions) != 2 {
		t.Errorf("colour options %v, want exactly {U} and {R}", pick.ColorOptions)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if n := riderAnswerManaPicks(t, g, me.ID, "U"); n != 2 {
		t.Errorf("answered %d further colour picks, want 2 — power 3 is three slots", n)
	}

	if len(me.ManaPool) != 3 {
		t.Fatalf("pool has %d tokens, want 3 — one per point of power", len(me.ManaPool))
	}
	colors := map[string]int{}
	for _, tok := range me.ManaPool {
		colors[tok.Color]++
	}
	if colors["U"] != 2 || colors["R"] != 1 {
		t.Errorf("pool is %v, want two U and one R", colors)
	}

	// "Activate only … once each turn."
	if err := g.ActivateManaAbility(me.ID, vivi, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the second activation this turn was allowed")
	}
	if len(me.ManaPool) != 3 {
		t.Errorf("the refused activation still minted mana: pool has %d tokens", len(me.ManaPool))
	}
}

// "Activate only during your turn" — an opponent's Vivi is inert while
// somebody else has the turn.
func TestViviManaAbilityIsRefusedOnAnotherPlayersTurn(t *testing.T) {
	g := newCatalogGame(t)
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vivi := b12Push(g, them.ID, "Vivi Ornitier", "Legendary Creature — Wizard", viviOrnitierOracle, 0, 3)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(vivi, game.CounterPlusOne, 2) })

	if err := g.ActivateManaAbility(them.ID, vivi, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("Vivi produced mana on someone else's turn")
	}
	if len(them.ManaPool) != 0 {
		t.Errorf("pool has %d tokens after a refused activation, want 0", len(them.ManaPool))
	}
}

// A Vivi at printed power 0 is the turn it lands: X is 0, so the
// activation adds nothing rather than erroring.
func TestViviAtPowerZeroAddsNoMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vivi := b12Push(g, me.ID, "Vivi Ornitier", "Legendary Creature — Wizard", viviOrnitierOracle, 0, 3)
	advanceToMain(t, g)

	activateManaFor(t, g, me.ID, vivi, 0, game.ManaAbilityParams{})
	if len(me.ManaPool) != 0 {
		t.Errorf("pool has %d tokens, want 0 — X is Vivi's power and that is 0", len(me.ManaPool))
	}
}

// --- Ojer Axonil, Deepest Might ---------------------------------------

// Every clause of the damage replacement, in one board: yours-only,
// red-only, opponents-only, and a floor rather than a bonus.
func TestOjerAxonilRaisesYourRedNoncombatDamageToItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Ojer Axonil, Deepest Might", "Legendary Creature — God", ojerAxonilOracle, 4, 4)
	red := b16Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1, "R")
	green := b16Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1, "G")
	theirRed := b16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1, "R")

	before := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(red, opp.ID, 1) })
	if opp.Life != before-4 {
		t.Errorf("your red source's 1 becomes 4: %d → %d", before, opp.Life)
	}

	// "LESS THAN Ojer Axonil's power" — a bigger hit is left alone.
	before = opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(red, opp.ID, 6) })
	if opp.Life != before-6 {
		t.Errorf("6 damage is not lowered to 4: %d → %d", before, opp.Life)
	}

	// "a RED source" — green is not raised.
	before = opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(green, opp.ID, 1) })
	if opp.Life != before-1 {
		t.Errorf("a green source of yours is untouched: %d → %d", before, opp.Life)
	}

	// "YOU CONTROL" — an opponent's red source is not raised.
	before = me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirRed, me.ID, 1) })
	if me.Life != before-1 {
		t.Errorf("an opponent's red source is untouched: %d → %d", before, me.Life)
	}

	// "to an OPPONENT" — damage at your own face is not raised.
	before = me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(red, me.ID, 1) })
	if me.Life != before-1 {
		t.Errorf("damage at your own face is untouched: %d → %d", before, me.Life)
	}
}

// "NONCOMBAT damage": the clause asked directly, because a combat
// damage event only comes out of a real combat.
func TestOjerAxonilLeavesCombatDamageAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ojer := b12Push(g, me.ID, "Ojer Axonil, Deepest Might", "Legendary Creature — God", ojerAxonilOracle, 4, 4)
	red := b16Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1, "R")

	spec, ok := Lookup(ojerAxonilOracle)
	if !ok || len(spec.Replacements) != 1 {
		t.Fatalf("Ojer declares one replacement effect, got %d", len(spec.Replacements))
	}
	src := b12Card(t, g, ojer)

	newEvent := func(combat bool) *game.ReplacementEvent {
		return &game.ReplacementEvent{
			Kind:           game.RepEventDamage,
			DamageSource:   red,
			DamageTarget:   opp.ID,
			DamageAmount:   1,
			IsCombatDamage: combat,
		}
	}
	var noncombat, combat bool
	g.WithWriteLock(func() {
		noncombat = spec.Replacements[0].AppliesTo(newEvent(false), g, &src)
		combat = spec.Replacements[0].AppliesTo(newEvent(true), g, &src)
	})
	if !noncombat {
		t.Error("noncombat damage from your red source is raised")
	}
	if combat {
		t.Error("combat damage is NOT raised — Ojer is not Gratuitous Violence")
	}
}

// The declared simplification, pinned: no dies trigger, and the card
// says so where a player reads it. This test flips when the transform
// verb lands.
func TestOjerAxonilShipsWithoutItsDiesTrigger(t *testing.T) {
	spec, ok := Lookup(ojerAxonilOracle)
	if !ok {
		t.Fatal("Ojer Axonil is registered")
	}
	if len(spec.Triggered) != 0 {
		t.Errorf("the dies trigger is deferred with the transform verb, got %d triggers", len(spec.Triggered))
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Errorf("completeness %v with %d caveats, want caveats with exactly one", spec.Completeness, len(spec.Caveats))
	}
	if !hasAbility(spec.PrintedKeywords, "trample") {
		t.Error("Ojer prints trample")
	}
}
