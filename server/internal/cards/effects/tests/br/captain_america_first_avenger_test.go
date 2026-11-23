package testsbr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// captain_america_first_avenger_test.go — #4799: Captain America, First
// Avenger's "Throw ... — {3}, Unattach an Equipment from Captain America:
// He deals damage equal to that Equipment's mana value divided as you
// choose among one, two, or three targets." and "... Catch — At the
// beginning of combat on your turn, attach up to one target Equipment you
// control to Captain America." The engine half of the cost is
// game/unattach_cost_test.go; the enumerator's legal/unattach_cost_test.go;
// the wire's protocol/unattach_cost_view_test.go.

const (
	capFirstAvengerOracle = "7d038055-035d-4698-b6b6-e2859bf2f1d1"
	capFirstAvengerType   = "Legendary Creature — Human Soldier Hero"
)

// capTable is Captain America on the active player's battlefield in their
// main phase, with {3} floating, and an opponent.
func capTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	cap := effectstest.PushCatalogPermanent(g, me.ID, "Captain America, First Avenger", capFirstAvengerType, capFirstAvengerOracle, false)
	effectstest.FillPool(me, 3)
	return g, me, opp, cap
}

func capThrow(g *game.Game, me, cap uuid.UUID, gear []uuid.UUID, targets []game.TargetRef, dist map[uuid.UUID]int) error {
	return g.ActivateCatalogAbility(me, cap, 0, game.ActivateAbilityParams{
		Strict: true, UnattachIDs: gear, Targets: targets, Distribution: dist,
	})
}

func TestCaptainAmericaFirstAvengerIsFull(t *testing.T) {
	spec, ok := effects.Lookup(capFirstAvengerOracle)
	if !ok {
		t.Fatal("Captain America, First Avenger is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v, want Full and none", spec.Completeness, spec.Caveats)
	}
}

// One Equipment attached, mana value 3: the division is 3, the Equipment
// stays on the battlefield unattached, and the {3} is paid.
func TestCaptainAmericaThrowsOneEquipmentForItsManaValue(t *testing.T) {
	g, me, opp, cap := capTable(t)
	shield := effectstest.CapGear(g, me.ID, cap, "Shield", "{1}{W}{W}")
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 30)
	life := opp.Life

	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: wall}}
	if err := capThrow(g, me.ID, cap, []uuid.UUID{shield}, targets, map[uuid.UUID]int{opp.ID: 2, wall: 1}); err != nil {
		t.Fatalf("throw with one Equipment: %v", err)
	}
	if host := effectstest.AttachmentHostOf(t, g, shield); host.Kind != "" {
		t.Errorf("the Shield is still attached to %v after paying", host)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %v, want the {3} spent", me.ManaPool)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d (2 of the 3)", opp.Life, life-2)
	}
	if d := effectstest.E2Card(t, g, wall).DamageMarked; d != 1 {
		t.Errorf("wall damage %d, want 1 (1 of the 3)", d)
	}
	if !g.Battlefield.Contains(shield) {
		t.Error("the Shield left the battlefield; unattaching keeps it there (CR 701.3d)")
	}
}

// Two Equipment attached, one of them an opponent's (CR 301.5d): the
// activator chooses which, and the division is THAT Equipment's mana value.
// A division sized for the other one is refused with nothing paid.
func TestCaptainAmericaDividesTheManaValueOfTheEquipmentHeUnattaches(t *testing.T) {
	g, me, opp, cap := capTable(t)
	knife := effectstest.CapGear(g, me.ID, cap, "Knife", "{1}")
	hammer := effectstest.CapGear(g, opp.ID, cap, "Hammer", "{2}{R}{R}")
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	life := opp.Life

	// The Knife's mana value is 1: two targets cannot divide it.
	err := capThrow(g, me.ID, cap, []uuid.UUID{knife}, effectstest.CardRefs(a, b), map[uuid.UUID]int{a: 1, b: 1})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("two targets for the Knife's 1: %v, want ErrInvalidParam", err)
	}
	// The Hammer's is 4: a division of 2 is short.
	err = capThrow(g, me.ID, cap, []uuid.UUID{hammer}, effectstest.CardRefs(a, b), map[uuid.UUID]int{a: 1, b: 1})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a division of 2 for the Hammer's 4: %v, want ErrInvalidParam", err)
	}
	for _, id := range []uuid.UUID{knife, hammer} {
		if host := effectstest.AttachmentHostOf(t, g, id); host.ID != cap {
			t.Fatalf("a refused activation unattached %s (now on %v)", id, host)
		}
	}
	if me.ManaPool.Total() != 3 {
		t.Fatalf("pool = %v, want the {3} still floating after the refusals", me.ManaPool)
	}

	targets := append(effectstest.CardRefs(a, b), game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if err := capThrow(g, me.ID, cap, []uuid.UUID{hammer}, targets, map[uuid.UUID]int{a: 1, b: 2, opp.ID: 1}); err != nil {
		t.Fatalf("throw the opponent's Hammer for 4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if da, db := effectstest.E2Card(t, g, a).DamageMarked, effectstest.E2Card(t, g, b).DamageMarked; da != 1 || db != 2 {
		t.Errorf("damage %d/%d, want 1/2", da, db)
	}
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
	if host := effectstest.AttachmentHostOf(t, g, hammer); host.Kind != "" {
		t.Errorf("the Hammer is still attached to %v", host)
	}
	if host := effectstest.AttachmentHostOf(t, g, knife); host.ID != cap {
		t.Errorf("the Knife was unattached too (now on %v)", host)
	}
	if c := effectstest.E2Card(t, g, hammer); c.Controller != opp.ID {
		t.Error("unattaching changed the Hammer's controller")
	}
}

// No Equipment attached: the cost cannot be paid (CR 118.3). Naming none
// is refused, and so is naming an Equipment that is attached to nothing or
// to another creature. Nothing is paid.
func TestCaptainAmericaCannotThrowWithNothingAttached(t *testing.T) {
	g, me, opp, cap := capTable(t)
	loose := effectstest.CapGear(g, me.ID, uuid.Nil, "Loose Blade", "{2}")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elsewhere := effectstest.CapGear(g, me.ID, bear, "Bear's Blade", "{2}")
	life := opp.Life
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}

	if err := capThrow(g, me.ID, cap, nil, target, nil); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("no Equipment named: %v, want ErrInvalidParam", err)
	}
	for name, id := range map[string]uuid.UUID{"an unattached Equipment": loose, "another creature's Equipment": elsewhere} {
		if err := capThrow(g, me.ID, cap, []uuid.UUID{id}, target, nil); !errors.Is(err, game.ErrIllegalTarget) {
			t.Errorf("%s: %v, want ErrIllegalTarget", name, err)
		}
	}
	if err := capThrow(g, me.ID, cap, []uuid.UUID{elsewhere, elsewhere}, target, nil); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("two Equipment named for a cost of one: %v, want ErrInvalidParam", err)
	}
	if me.ManaPool.Total() != 3 || opp.Life != life || len(g.Stack.Cards) != 0 {
		t.Errorf("a refused throw paid or did something: pool %v, life %d, stack %d", me.ManaPool, opp.Life, len(g.Stack.Cards))
	}
	if host := effectstest.AttachmentHostOf(t, g, elsewhere); host.ID != bear {
		t.Errorf("the bear's Blade moved to %v", host)
	}
}

// The division is fixed at activation: the throw on the stack survives a
// save and restore with its shares, and resolves on the restored table.
func TestCaptainAmericaThrowSurvivesARestore(t *testing.T) {
	g, me, opp, cap := capTable(t)
	axe := effectstest.CapGear(g, me.ID, cap, "Axe", "{2}")
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 30)
	life := opp.Life
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: wall}, {Kind: game.TargetPlayer, ID: opp.ID}}
	if err := capThrow(g, me.ID, cap, []uuid.UUID{axe}, targets, map[uuid.UUID]int{wall: 1, opp.ID: 1}); err != nil {
		t.Fatalf("throw: %v", err)
	}

	r := effectstest.RestoreThroughJSON(t, g)
	effectstest.PassPriorityAroundTable(t, r)
	if got := effectstest.LifeOf(r, opp.ID); got != life-1 {
		t.Errorf("restored: opponent life %d, want %d", got, life-1)
	}
	if d := effectstest.E2Card(t, r, wall).DamageMarked; d != 1 {
		t.Errorf("restored: wall damage %d, want 1", d)
	}
	if host := effectstest.AttachmentHostOf(t, r, axe); host.Kind != "" {
		t.Errorf("restored: the Axe is attached to %v", host)
	}
}

// Catch: at the beginning of combat on his controller's turn, the chosen
// Equipment is attached to him. The target prompt is a restore point.
func TestCaptainAmericaCatchesAnEquipmentAtBeginningOfCombat(t *testing.T) {
	g, me, opp, cap := capTable(t)
	axe := effectstest.CapGear(g, me.ID, uuid.Nil, "Axe", "{2}")
	theirs := effectstest.CapGear(g, opp.ID, uuid.Nil, "Their Axe", "{2}")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("Catch asks for no target")
	}

	r := effectstest.RestoreThroughJSON(t, g)
	effectstest.B17PickCards(t, r, me.ID, axe)
	effectstest.PassPriorityAroundTable(t, r)
	if host := effectstest.AttachmentHostOf(t, r, axe); host.ID != cap {
		t.Errorf("the Axe is attached to %v, want Captain America", host)
	}
	if host := effectstest.AttachmentHostOf(t, r, theirs); host.Kind != "" {
		t.Errorf("the opponent's Axe moved to %v", host)
	}
}
