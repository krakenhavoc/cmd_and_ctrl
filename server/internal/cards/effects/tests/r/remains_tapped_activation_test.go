package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// remains_tapped_activation_test.go is #4061: "for as long as this
// creature remains tapped" is checked from the moment the ability is
// ACTIVATED (CR 611.2b), so a source that untaps while the ability is on
// the stack and is tapped again before it resolves has still broken the
// duration. Preacher is the card that waited on it; Vedalken Shackles, the
// +N/+N Couriers and the tap-and-hold family share the constructors and
// are held to the same reading.

const rtPreacherOracle = "d5ea9176-30c1-4152-9086-322f39245f46"

// preacherTable is Preacher in play and one opponent's creature, with
// the other opponents holding none, so only that opponent can choose.
func preacherTable(t *testing.T) (g *game.Game, me, opp *game.Player, preacher, victim uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	preacher = effectstest.PushCatalogPermanent(g, me.ID, "Preacher", "Creature — Human Cleric", rtPreacherOracle, false)
	victim = effectstest.Pr7Creature(g, opp.ID, "Their Bear", 2, "G")
	return
}

// preacherActivate activates Preacher and has the opponent name their
// creature, leaving the ability on the stack.
func preacherActivate(t *testing.T, g *game.Game, me, opp *game.Player, preacher, victim uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, preacher, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.OtPickFor(t, g, opp.ID, me.ID)
	effectstest.OtAnswer(t, g, opp.ID, effectstest.OtCard(victim))
}

// Preacher takes the creature the opponent names, holds it while Preacher
// stays tapped, and gives it back when Preacher untaps.
func TestPreacherStealsTheOpponentsPickWhileTapped(t *testing.T) {
	g, me, opp, preacher, victim := preacherTable(t)
	other := effectstest.Pr7Creature(g, opp.ID, "Second Bear", 3, "G")
	if err := g.ActivateCatalogAbility(me.ID, preacher, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	p := effectstest.OtPickFor(t, g, opp.ID, me.ID)
	if len(p.PickTargetCards) != 2 {
		t.Fatalf("offered %v, want both of the opponent's creatures and none of mine", p.PickTargetCards)
	}
	effectstest.OtAnswer(t, g, opp.ID, effectstest.OtCard(other))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, other); got != me.ID {
		t.Fatalf("controller %s, want %s: Preacher took what the opponent named", got, me.ID)
	}
	if got := effectstest.ControllerOf(t, g, victim); got != opp.ID {
		t.Errorf("the creature the opponent did not name changed hands")
	}
	effectstest.RtUntap(g, preacher)
	if got := effectstest.ControllerOf(t, g, other); got != opp.ID {
		t.Error("Preacher untapped and the creature stayed stolen (CR 611.2b)")
	}
}

// Preacher's ruling: "If, before the ability resolves, there's any point
// at which Preacher is untapped, the ability has no effect - even if
// Preacher is tapped again by the time the ability resolves."
func TestPreacherUntappedInResponseTakesNothingEvenIfTappedAgain(t *testing.T) {
	g, me, opp, preacher, victim := preacherTable(t)
	preacherActivate(t, g, me, opp, preacher, victim)
	effectstest.RtUntap(g, preacher)
	effectstest.RtRetap(g, preacher)
	if !effectstest.RtTapped(t, g, preacher) {
		t.Fatal("setup: Preacher should be tapped again as the ability resolves")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, victim); got != opp.ID {
		t.Errorf("controller %s, want %s: an untap while the ability waited broke the duration", got, opp.ID)
	}
}

// The control for the test above: tapped the whole time, it works, so the
// failure above is the untap record and not a broken activation.
func TestPreacherTappedTheWholeTimeTakesIt(t *testing.T) {
	g, me, opp, preacher, victim := preacherTable(t)
	preacherActivate(t, g, me, opp, preacher, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Errorf("controller %s, want %s", got, me.ID)
	}
}

// An untap that is replaced (a stun counter) is not an untap: the
// Preacher stays tapped and the ability still works.
func TestPreacherStunReplacedUntapIsNotAnUntap(t *testing.T) {
	g, me, opp, preacher, victim := preacherTable(t)
	preacherActivate(t, g, me, opp, preacher, victim)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == preacher {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterStun: 1}
			}
		}
	})
	effectstest.RtUntap(g, preacher)
	if !effectstest.RtTapped(t, g, preacher) {
		t.Fatal("setup: the stun counter should have replaced the untap")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Errorf("controller %s, want %s: the replaced untap broke the duration", got, me.ID)
	}
}

// Another permanent untapping while Preacher's ability waits changes
// nothing.
func TestPreacherIgnoresOtherPermanentsUntapping(t *testing.T) {
	g, me, opp, preacher, victim := preacherTable(t)
	bystander := effectstest.CtrlPushCreature(g, me.ID, "Bystander")
	effectstest.RtRetap(g, bystander)
	preacherActivate(t, g, me, opp, preacher, victim)
	effectstest.RtUntap(g, bystander)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Errorf("controller %s, want %s: an unrelated untap broke the duration", got, me.ID)
	}
}

// The record survives a restore point: Preacher untapped and tapped
// again while the ability waited, then the table is saved and reloaded.
func TestPreacherUntapRecordSurvivesARestore(t *testing.T) {
	g, me, opp, preacher, victim := preacherTable(t)
	preacherActivate(t, g, me, opp, preacher, victim)
	effectstest.RtUntap(g, preacher)
	effectstest.RtRetap(g, preacher)
	restored := effectstest.RestoreRoundTrip(t, g, true)
	effectstest.PassPriorityAroundTable(t, restored)
	if got := effectstest.ControllerOf(t, restored, victim); got != opp.ID {
		t.Errorf("restored: controller %s, want %s: the untap record was lost", got, opp.ID)
	}
}

// Vedalken Shackles prints the same duration and its rulings say the
// same; it shipped Full with the gap.
func TestVedalkenShacklesUntappedInResponseTakesNothingEvenIfTappedAgain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	shackles := effectstest.PushCatalogPermanent(g, me.ID, "Vedalken Shackles", "Artifact", effectstest.RtShacklesOracle, false)
	bear := effectstest.CtrlPushCreature(g, opp.ID, "Bear")
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, shackles, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(bear)}); err != nil {
		t.Fatal(err)
	}
	effectstest.RtUntap(g, shackles)
	effectstest.RtRetap(g, shackles)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, bear); got != opp.ID {
		t.Errorf("controller %s, want %s", got, opp.ID)
	}
}

// The +N/+N family reads the same way: an Endoskeleton untapped and
// tapped again in response pumps nothing.
func TestEndoskeletonUntappedAndTappedAgainPumpsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	endo := effectstest.PushCatalogPermanent(g, me.ID, "Endoskeleton", "Artifact", effectstest.RtEndoskeletonOracle, false)
	bear := effectstest.CtrlPushCreature(g, me.ID, "Bear")
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, endo, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(bear)}); err != nil {
		t.Fatal(err)
	}
	effectstest.RtUntap(g, endo)
	effectstest.RtRetap(g, endo)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, bear).CurrentToughness(); got != 2 {
		t.Errorf("toughness %d, want 2", got)
	}
}

// Old Man of the Sea reads "this creature remains tapped" the same way:
// untapped and tapped again in response, it steals nothing.
func TestOldManOfTheSeaUntappedAndTappedAgainStealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldMan := effectstest.PushCatalogPermanent(g, me.ID, "Old Man of the Sea", "Creature — Merfolk Wizard",
		"cce84cf1-5574-43b0-9d75-72e6451403a7", false)
	bear := effectstest.Pr7Creature(g, opp.ID, "Weak Bear", 1, "G")
	if err := g.ActivateCatalogAbility(me.ID, oldMan, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(bear)}); err != nil {
		t.Fatal(err)
	}
	effectstest.RtUntap(g, oldMan)
	effectstest.RtRetap(g, oldMan)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, bear); got != opp.ID {
		t.Errorf("controller %s, want %s", got, opp.ID)
	}
}

// The untap hold family: a Mana Leech untapped and tapped again in
// response taps the land and holds nothing.
func TestManaLeechUntappedAndTappedAgainHoldsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leech := effectstest.PushCatalogPermanent(g, me.ID, "Mana Leech", "Creature — Leech", effectstest.RtManaLeechOracle, false)
	land := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	if err := g.ActivateCatalogAbility(me.ID, leech, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(land)}); err != nil {
		t.Fatal(err)
	}
	effectstest.RtUntap(g, leech)
	effectstest.RtRetap(g, leech)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, land)
	if !c.Tapped {
		t.Fatal("the land should still have been tapped")
	}
	if len(c.NextUntapSkips) != 0 {
		t.Errorf("the land is held (%+v) although the Leech untapped while the ability waited", c.NextUntapSkips)
	}
}
