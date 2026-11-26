package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b312_test.go — Battle for Zendikar pump, combat and counter cards
// (#5547). Every test plays the real catalog entry through the engine.

const (
	oldB312Invoker    = "c5dcab94-f4c0-4e5e-ad99-7b3871be6e89"
	oldB312Sharpshoot = "d2f59fa6-8e61-4a70-8044-196d7b26ec9e"
	oldB312Swell      = "e47e0b81-0f0c-4e52-bc58-16af19b0dfe0"
	oldB312Pulse      = "fe3c81b8-9bca-4760-b677-02bfa7e6137b"
	oldB312Captain    = "ffc56a2e-f3bd-493b-9b7b-5f04fad69382"
	oldB312Null       = "0f92daf5-520d-472a-a835-e6680a92fdab"
	oldB312NoyanDar   = "e4190320-5854-479e-be11-a5d5ef3db6a6"
	oldB312Quarantine = "f8b07b1a-f9fb-427d-a5c8-b65a01f7b079"
)

func oldB312Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func TestOldB312EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB312Invoker, oldB312Sharpshoot, oldB312Swell, oldB312Pulse, oldB312Captain, oldB312Null, oldB312NoyanDar, oldB312Quarantine} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestOldB312OranRiefInvokerPumpsForEightMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	inv := effectstest.PushCatalogPermanent(g, me.ID, "Oran-Rief Invoker", "Creature — Human Shaman", oldB312Invoker, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCCCCC")
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only seven mana")
	}
	effectstest.FloatForTest(g, me, "C")
	effectstest.B16Activate(t, g, me.ID, inv, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, inv); p != 6 || tt != 6 {
		t.Errorf("the Invoker is %d/%d, want 6/6 (1/1 base here + 5/+5)", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, inv, "trample") {
		t.Error("the Invoker has no trample")
	}
}

func TestOldB312HagraSharpshooterShrinksAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gun := effectstest.PushCatalogPermanent(g, me.ID, "Hagra Sharpshooter", "Creature — Human Assassin Ally", oldB312Sharpshoot, false)
	mouse := effectstest.B12Creature(g, opp.ID, "Mouse", "Creature — Mouse", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "BCCCC")
	effectstest.B16Activate(t, g, me.ID, gun, 0, game.ActivateAbilityParams{Targets: oldB312Target(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 1 || tt != 1 {
		t.Errorf("the Bear is %d/%d, want 1/1", p, tt)
	}
	effectstest.FloatForTest(g, me, "BCCCC")
	effectstest.B16Activate(t, g, me.ID, gun, 0, game.ActivateAbilityParams{Targets: oldB312Target(mouse)})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, mouse); ok {
		t.Error("the 1/1 survived -1/-1")
	}
}

func TestOldB312SwellOfGrowthPumpsAndMayPutALand(t *testing.T) {
	for _, yes := range []bool{true, false} {
		name := "decline"
		if yes {
			name = "accept"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			land := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
			effectstest.CastCatalogSpell(t, g, "Swell of Growth", "Instant", oldB312Swell, oldB312Target(bear))
			effectstest.PassPriorityAroundTable(t, g)
			if p, tt := effectstest.PtOf(t, g, bear); p != 4 || tt != 4 {
				t.Errorf("the Bear is %d/%d, want 4/4", p, tt)
			}
			if yes {
				effectstest.AnswerChooseCards(t, g, me.ID, land)
			} else {
				effectstest.AnswerChooseCards(t, g, me.ID)
			}
			if got := g.Battlefield.Contains(land); got != yes {
				t.Errorf("the land is on the battlefield = %t, want %t", got, yes)
			}
		})
	}
}

func TestOldB312DampeningPulseShrinksOnlyOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Dampening Pulse", "Enchantment", oldB312Pulse, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if p, _ := effectstest.PtOf(t, g, mine); p != 2 {
		t.Errorf("my creature has power %d, want 2", p)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 1 || tt != 2 {
		t.Errorf("their creature is %d/%d, want 1/2", p, tt)
	}
}

func TestOldB312AngelicCaptainCountsOtherAttackingAllies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cap := effectstest.B12Push(g, me.ID, "Angelic Captain", "Creature — Angel Ally", oldB312Captain, 4, 3)
	a1 := effectstest.B12Creature(g, me.ID, "Ally One", "Creature — Kor Ally", 1, 1)
	effectstest.B12Creature(g, me.ID, "Idle Ally", "Creature — Kor Ally", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, cap, a1, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, cap); p != 5 || tt != 4 {
		t.Errorf("the Captain is %d/%d, want 5/4 (one other attacking Ally)", p, tt)
	}
}

func TestOldB312VoraciousNullSacrificesAnotherCreatureAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	null := effectstest.PushCatalogPermanent(g, me.ID, "Voracious Null", "Creature — Zombie", oldB312Null, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "BC")
	if err := g.ActivateCatalogAbility(me.ID, null, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{null}}); err == nil {
		t.Fatal("the Null sacrificed itself")
	}
	effectstest.B16Activate(t, g, me.ID, null, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the other creature was not sacrificed")
	}
	if n := effectstest.CounterCount(g, null, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2", n)
	}
	// Not outside a main phase.
	more := effectstest.B12Creature(g, me.ID, "More Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatForTest(g, me, "BC")
	if err := g.ActivateCatalogAbility(me.ID, null, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{more}}); err == nil {
		t.Error("activated outside a main phase")
	}
}

func TestOldB312NoyanDarAnimatesALandWithThreeCountersOnAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Noyan Dar, Roil Shaper", "Legendary Creature — Merfolk Ally", oldB312NoyanDar, 4, 4)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Plain Instant", "Instant", "", nil)
	effectstest.B04WaitForPickOrSkip(t, g, me.ID, land)
	if p := effectstest.ConfirmForTriggerPrompt(g, me.ID); p != nil {
		if err := g.ResolveTriggerPrompt(p.ID, me.ID, true); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.ConfirmForTriggerPrompt(g, me.ID); p != nil {
		if err := g.ResolveTriggerPrompt(p.ID, me.ID, true); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	c := effectstest.B12Card(t, g, land)
	if !c.IsLand() || !c.IsCreature() || !c.HasSubtype("Elemental") {
		t.Fatalf("the land is %q, want a land Elemental creature", c.TypeLine)
	}
	if n := effectstest.CounterCount(g, land, game.CounterPlusOne); n != 3 {
		t.Errorf("%d +1/+1 counters, want 3", n)
	}
	if p, tt := effectstest.PtOf(t, g, land); p != 3 || tt != 3 {
		t.Errorf("the land is %d/%d, want 3/3 (0/0 with three counters)", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, land, "haste") {
		t.Error("the land has no haste")
	}
}

func TestOldB312QuarantineFieldExilesOneTargetPerCounterUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Creature(g, opp.ID, "Alpha", "Creature — Bear", 2, 2)
	b := effectstest.B12Permanent(g, opp.ID, "Beta", "Artifact")
	c := effectstest.B12Creature(g, opp.ID, "Gamma", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "WWWWWW")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Quarantine Field", TypeLine: "Enchantment", ManaCost: "{X}{X}{W}{W}",
			OracleID: oldB312Quarantine, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, id, effects.S82OldB312IsolationCounter); n != 2 {
		t.Fatalf("%d isolation counters, want 2", n)
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, gone := range []uuid.UUID{a, b} {
		if _, ok := effectstest.BattlefieldCard(g, gone); ok {
			t.Errorf("%s is still on the battlefield", gone)
		}
	}
	for _, kept := range []uuid.UUID{c, land} {
		if _, ok := effectstest.BattlefieldCard(g, kept); !ok {
			t.Errorf("%s was exiled without being a target", kept)
		}
	}
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	back := 0
	for _, name := range []string{"Alpha", "Beta"} {
		if _, ok := effectstest.B43BattlefieldByName(g, name); ok {
			back++
		}
	}
	if back != 2 {
		t.Errorf("%d of the two exiled permanents came back, want 2", back)
	}
}

func TestOldB312QuarantineFieldTargetsOneClausePerCounter(t *testing.T) {
	spec, ok := effects.Lookup(oldB312Quarantine)
	if !ok || len(spec.Triggered) != 1 || spec.Triggered[0].TargetsFrom == nil {
		t.Fatal("Quarantine Field has no per-counter target clause")
	}
}
