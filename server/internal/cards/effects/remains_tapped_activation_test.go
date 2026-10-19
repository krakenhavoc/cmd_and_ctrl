package effects

import (
	"testing"

	"github.com/google/uuid"

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

// rtRetap taps the permanent again, as a second source of tapping would.
func rtRetap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
}

// preacherTable is Preacher in play and one opponent's creature, with
// the other opponents holding none, so only that opponent can choose.
func preacherTable(t *testing.T) (g *game.Game, me, opp *game.Player, preacher, victim uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	preacher = pushCatalogPermanent(g, me.ID, "Preacher", "Creature — Human Cleric", rtPreacherOracle, false)
	victim = pr7Creature(g, opp.ID, "Their Bear", 2, "G")
	return
}

// preacherActivate activates Preacher and has the opponent name their
// creature, leaving the ability on the stack.
func preacherActivate(t *testing.T, g *game.Game, me, opp *game.Player, preacher, victim uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, preacher, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	otPickFor(t, g, opp.ID, me.ID)
	otAnswer(t, g, opp.ID, otCard(victim))
}

// Preacher takes the creature the opponent names, holds it while Preacher
// stays tapped, and gives it back when Preacher untaps.
func TestPreacherStealsTheOpponentsPickWhileTapped(t *testing.T) {
	g, me, opp, preacher, victim := preacherTable(t)
	other := pr7Creature(g, opp.ID, "Second Bear", 3, "G")
	if err := g.ActivateCatalogAbility(me.ID, preacher, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	p := otPickFor(t, g, opp.ID, me.ID)
	if len(p.PickTargetCards) != 2 {
		t.Fatalf("offered %v, want both of the opponent's creatures and none of mine", p.PickTargetCards)
	}
	otAnswer(t, g, opp.ID, otCard(other))
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, other); got != me.ID {
		t.Fatalf("controller %s, want %s: Preacher took what the opponent named", got, me.ID)
	}
	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Errorf("the creature the opponent did not name changed hands")
	}
	rtUntap(g, preacher)
	if got := controllerOf(t, g, other); got != opp.ID {
		t.Error("Preacher untapped and the creature stayed stolen (CR 611.2b)")
	}
}

// Preacher's ruling: "If, before the ability resolves, there's any point
// at which Preacher is untapped, the ability has no effect - even if
// Preacher is tapped again by the time the ability resolves."
func TestPreacherUntappedInResponseTakesNothingEvenIfTappedAgain(t *testing.T) {
	g, me, opp, preacher, victim := preacherTable(t)
	preacherActivate(t, g, me, opp, preacher, victim)
	rtUntap(g, preacher)
	rtRetap(g, preacher)
	if !rtTapped(t, g, preacher) {
		t.Fatal("setup: Preacher should be tapped again as the ability resolves")
	}
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Errorf("controller %s, want %s: an untap while the ability waited broke the duration", got, opp.ID)
	}
}

// The control for the test above: tapped the whole time, it works, so the
// failure above is the untap record and not a broken activation.
func TestPreacherTappedTheWholeTimeTakesIt(t *testing.T) {
	g, me, opp, preacher, victim := preacherTable(t)
	preacherActivate(t, g, me, opp, preacher, victim)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, victim); got != me.ID {
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
	rtUntap(g, preacher)
	if !rtTapped(t, g, preacher) {
		t.Fatal("setup: the stun counter should have replaced the untap")
	}
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Errorf("controller %s, want %s: the replaced untap broke the duration", got, me.ID)
	}
}

// Another permanent untapping while Preacher's ability waits changes
// nothing.
func TestPreacherIgnoresOtherPermanentsUntapping(t *testing.T) {
	g, me, opp, preacher, victim := preacherTable(t)
	bystander := ctrlPushCreature(g, me.ID, "Bystander")
	rtRetap(g, bystander)
	preacherActivate(t, g, me, opp, preacher, victim)
	rtUntap(g, bystander)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Errorf("controller %s, want %s: an unrelated untap broke the duration", got, me.ID)
	}
}

// The record survives a restore point: Preacher untapped and tapped
// again while the ability waited, then the table is saved and reloaded.
func TestPreacherUntapRecordSurvivesARestore(t *testing.T) {
	g, me, opp, preacher, victim := preacherTable(t)
	preacherActivate(t, g, me, opp, preacher, victim)
	rtUntap(g, preacher)
	rtRetap(g, preacher)
	restored := restoreRoundTrip(t, g, true)
	passPriorityAroundTable(t, restored)
	if got := controllerOf(t, restored, victim); got != opp.ID {
		t.Errorf("restored: controller %s, want %s: the untap record was lost", got, opp.ID)
	}
}

// Vedalken Shackles prints the same duration and its rulings say the
// same; it shipped Full with the gap.
func TestVedalkenShacklesUntappedInResponseTakesNothingEvenIfTappedAgain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	shackles := pushCatalogPermanent(g, me.ID, "Vedalken Shackles", "Artifact", rtShacklesOracle, false)
	bear := ctrlPushCreature(g, opp.ID, "Bear")
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, shackles, 0, game.ActivateAbilityParams{Targets: ltCardTarget(bear)}); err != nil {
		t.Fatal(err)
	}
	rtUntap(g, shackles)
	rtRetap(g, shackles)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, bear); got != opp.ID {
		t.Errorf("controller %s, want %s", got, opp.ID)
	}
}

// The +N/+N family reads the same way: an Endoskeleton untapped and
// tapped again in response pumps nothing.
func TestEndoskeletonUntappedAndTappedAgainPumpsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	endo := pushCatalogPermanent(g, me.ID, "Endoskeleton", "Artifact", rtEndoskeletonOracle, false)
	bear := ctrlPushCreature(g, me.ID, "Bear")
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, endo, 0, game.ActivateAbilityParams{Targets: ltCardTarget(bear)}); err != nil {
		t.Fatal(err)
	}
	rtUntap(g, endo)
	rtRetap(g, endo)
	passPriorityAroundTable(t, g)
	if got := layeredCard(t, g, bear).CurrentToughness(); got != 2 {
		t.Errorf("toughness %d, want 2", got)
	}
}

// Old Man of the Sea reads "this creature remains tapped" the same way:
// untapped and tapped again in response, it steals nothing.
func TestOldManOfTheSeaUntappedAndTappedAgainStealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldMan := pushCatalogPermanent(g, me.ID, "Old Man of the Sea", "Creature — Merfolk Wizard",
		"cce84cf1-5574-43b0-9d75-72e6451403a7", false)
	bear := pr7Creature(g, opp.ID, "Weak Bear", 1, "G")
	if err := g.ActivateCatalogAbility(me.ID, oldMan, 0, game.ActivateAbilityParams{Targets: ltCardTarget(bear)}); err != nil {
		t.Fatal(err)
	}
	rtUntap(g, oldMan)
	rtRetap(g, oldMan)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, bear); got != opp.ID {
		t.Errorf("controller %s, want %s", got, opp.ID)
	}
}

// The untap hold family: a Mana Leech untapped and tapped again in
// response taps the land and holds nothing.
func TestManaLeechUntappedAndTappedAgainHoldsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leech := pushCatalogPermanent(g, me.ID, "Mana Leech", "Creature — Leech", rtManaLeechOracle, false)
	land := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	if err := g.ActivateCatalogAbility(me.ID, leech, 0, game.ActivateAbilityParams{Targets: ltCardTarget(land)}); err != nil {
		t.Fatal(err)
	}
	rtUntap(g, leech)
	rtRetap(g, leech)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, land)
	if !c.Tapped {
		t.Fatal("the land should still have been tapped")
	}
	if len(c.NextUntapSkips) != 0 {
		t.Errorf("the land is held (%+v) although the Leech untapped while the ability waited", c.NextUntapSkips)
	}
}
