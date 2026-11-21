package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #5760: "whenever equipped creature deals combat damage" (no recipient
// named) is ONE trigger per combat damage step. All combat damage of a
// step is dealt simultaneously (CR 510.2), so a trampler that deals
// damage to its blocker and to the player has dealt combat damage once,
// and the ability triggers once for that event (CR 603.2c). The engine
// opens one EventDealDamage per recipient, so each row needs a
// once-per-batch guard. First strike and regular damage are two combat
// damage steps (CR 510.4), so a double striker triggers twice.

// equippedCombatDamageRows are the "whenever equipped creature deals
// combat damage" Equipment in the catalog, with the charge counters one
// trigger puts on the Equipment and the power each counter gives the
// equipped creature (Banshee's Blade's +1/+1 per counter).
var equippedCombatDamageRows = []struct {
	name, oracle    string
	perTrigger      int
	powerPerCounter int
}{
	{"Umezawa's Jitte", effectstest.UmezawasJitteOracle, 2, 0},
	{"Lost Jitte", effectstest.B43LostJitteOracle, 1, 0},
	{"Banshee's Blade", effectstest.MrdBansheesBladeOracle, 1, 1},
}

// onceStepTrampler seats a vanilla creature with trample for seat 0.
func onceStepTrampler(g *game.Game, owner uuid.UUID, power, toughness int, keywords ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Trampler", TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness, Keywords: append([]string{"trample"}, keywords...),
		Owner: owner, Controller: owner,
	})
}

// A 3/3 trampler blocked by a 0/1 assigns 1 to the blocker and 2 to the
// player (CR 702.19b): two recipients, one combat damage event.
func TestEquippedCombatDamageTriggersOnceWhenATramplerHitsItsBlockerAndThePlayer(t *testing.T) {
	for _, row := range equippedCombatDamageRows {
		t.Run(row.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			trampler := onceStepTrampler(g, me.ID, 3, 3)
			eq := effectstest.SeedEquipment(g, me.ID, row.name, row.oracle)
			effectstest.EquipTo(t, g, me.ID, eq, trampler)
			wall := effectstest.B16Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 1)
			life := opp.Life

			effectstest.OnceStepBlock(t, g, opp.ID, trampler, game.StepCombatDamage, wall)
			effectstest.PassPriorityAroundTable(t, g)

			if lost := life - opp.Life; lost != 2 {
				t.Fatalf("the defender lost %d, want 2 trampled over", lost)
			}
			if !opp.Graveyard.Contains(wall) {
				t.Fatal("the blocker was not dealt its lethal 1")
			}
			if n := effectstest.AcCountTriggerAnnouncements(g, row.name+" —"); n != 1 {
				t.Errorf("%d triggers for one combat damage step, want 1", n)
			}
			if n := effectstest.CounterOn(g, eq, game.CounterCharge); n != row.perTrigger {
				t.Errorf("%d charge counters, want %d", n, row.perTrigger)
			}
		})
	}
}

// Two blockers and the player: the CR 510.1c assignment divides a 4/4
// trampler's damage 1 / 1 / 2, three recipients, still one event.
func TestEquippedCombatDamageTriggersOnceWhenATramplerDividesAmongBlockersAndThePlayer(t *testing.T) {
	for _, row := range equippedCombatDamageRows {
		t.Run(row.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			trampler := onceStepTrampler(g, me.ID, 4, 4)
			eq := effectstest.SeedEquipment(g, me.ID, row.name, row.oracle)
			effectstest.EquipTo(t, g, me.ID, eq, trampler)
			w1 := effectstest.B16Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 1)
			w2 := effectstest.B16Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 1)
			life := opp.Life

			effectstest.OnceStepBlock(t, g, opp.ID, trampler, game.StepCombatDamage, w1, w2)
			var prompt *game.PendingChoice
			for _, p := range g.PendingChoices {
				if p != nil && p.Kind == game.PendingChoiceDamageAssignment {
					prompt = p
				}
			}
			if prompt == nil {
				t.Fatal("no damage assignment prompt for a double-blocked attacker")
			}
			entries := []game.DamageAssignmentEntry{{BlockerID: w1, Amount: 1}, {BlockerID: w2, Amount: 1}}
			if err := g.ResolveDamageAssignment(prompt.ID, me.ID, entries, 2); err != nil {
				t.Fatalf("ResolveDamageAssignment: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)

			if lost := life - opp.Life; lost != 2 {
				t.Fatalf("the defender lost %d, want 2 trampled over", lost)
			}
			if !opp.Graveyard.Contains(w1) || !opp.Graveyard.Contains(w2) {
				t.Fatal("a blocker was not dealt its lethal 1")
			}
			if n := effectstest.AcCountTriggerAnnouncements(g, row.name+" —"); n != 1 {
				t.Errorf("%d triggers for one combat damage step, want 1", n)
			}
			if n := effectstest.CounterOn(g, eq, game.CounterCharge); n != row.perTrigger {
				t.Errorf("%d charge counters, want %d", n, row.perTrigger)
			}
		})
	}
}

// Double strike: the first-strike step and the regular step are two
// combat damage steps (CR 510.4), so two triggers, even though the
// first one splits its damage between the blocker and the player. A 5/5
// against a 0/3 puts 3 on the blocker and 2 on the player first, then
// all its power on the player once the blocker is gone (CR 702.19d):
// 5, or 6 with the counter Banshee's Blade banked in the first step.
func TestEquippedCombatDamageTriggersOncePerStepForADoubleStrikingTrampler(t *testing.T) {
	for _, row := range equippedCombatDamageRows {
		t.Run(row.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			trampler := onceStepTrampler(g, me.ID, 5, 5, "double strike")
			eq := effectstest.SeedEquipment(g, me.ID, row.name, row.oracle)
			effectstest.EquipTo(t, g, me.ID, eq, trampler)
			wall := effectstest.B16Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 3)
			life := opp.Life

			effectstest.OnceStepBlock(t, g, opp.ID, trampler, game.StepFirstStrikeDamage, wall)
			effectstest.PassPriorityAroundTable(t, g)
			if lost := life - opp.Life; lost != 2 {
				t.Fatalf("first strike: the defender lost %d, want 2 trampled over", lost)
			}
			if n := effectstest.AcCountTriggerAnnouncements(g, row.name+" —"); n != 1 {
				t.Errorf("first strike: %d triggers, want 1", n)
			}
			if n := effectstest.CounterOn(g, eq, game.CounterCharge); n != row.perTrigger {
				t.Errorf("first strike: %d charge counters, want %d", n, row.perTrigger)
			}

			effectstest.AdvanceTo(t, g, game.StepCombatDamage)
			effectstest.PassPriorityAroundTable(t, g)
			regular := 5 + row.perTrigger*row.powerPerCounter
			if lost := life - opp.Life; lost != 2+regular {
				t.Fatalf("regular damage: the defender lost %d in all, want 2 + %d", lost, regular)
			}
			if n := effectstest.AcCountTriggerAnnouncements(g, row.name+" —"); n != 2 {
				t.Errorf("after both steps: %d triggers, want 2", n)
			}
			if n := effectstest.CounterOn(g, eq, game.CounterCharge); n != 2*row.perTrigger {
				t.Errorf("after both steps: %d charge counters, want %d", n, 2*row.perTrigger)
			}
		})
	}
}

// The Aetherspark: "whenever equipped creature deals combat damage
// during your turn, put that many loyalty counters on The Aetherspark"
// is the same one trigger per step, for the whole of what the creature
// dealt in it: 1 to the blocker plus 2 to the player is one trigger for
// three counters, not two triggers.
func TestAethersparkTriggersOnceForATramplerHittingItsBlockerAndThePlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	trampler := onceStepTrampler(g, me.ID, 3, 3)
	spark := effectstest.PushAetherspark(g, me.ID, 4)
	effectstest.BrAttach(t, g, spark, trampler)
	wall := effectstest.B16Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 1)

	effectstest.OnceStepBlock(t, g, opp.ID, trampler, game.StepCombatDamage, wall)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.AcCountTriggerAnnouncements(g, "The Aetherspark —"); n != 1 {
		t.Errorf("%d triggers for one combat damage step, want 1", n)
	}
	if got := effectstest.LoyaltyCount(g, spark); got != 7 {
		t.Errorf("loyalty %d, want 4 + 3 (1 to the blocker, 2 to the player)", got)
	}
}

// Long River Lurker's "whenever that creature deals combat damage this
// turn" is the same shape. The creature is unblockable that turn, so the
// split comes from a creature that was already blocked when the Lurker's
// trigger resolved; two damage events of one combat damage step stand in
// for it here, and they are one trigger with one "you may".
func TestLongRiverLurkerDelayedTriggerIsOncePerCombatDamageStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	runner := effectstest.B12Creature(g, me.ID, "Runner", "Creature — Bear", 3, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.LrlCastLurker(t, g, me, runner)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: runner, Target: wall, Amount: 1, Combat: true})
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: runner, Target: opp.ID, Amount: 2, Combat: true})
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.AcCountTriggerAnnouncements(g, "Long River Lurker — you may exile"); n != 1 {
		t.Errorf("%d triggers for one combat damage step, want 1", n)
	}
}
