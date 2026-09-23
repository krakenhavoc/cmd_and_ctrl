package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activation_timing_cards_test.go — the three proof cards for #1208,
// each driven through ActivateCatalogAbility (or ActivateLoyalty) so
// the assertion is about the announce path and not only about the
// predicate.

const (
	leoninShikariOracle      = "857d94f7-113c-45a4-a88a-5d087347d57f"
	bonesplitterTimingOracle = "452e3f5f-ce17-4682-966b-5cc100210aee"
	teferiMasterOfTimeOracle = "e802fb53-7cf5-46bc-8a0b-f99cf5c20f74"
)

// --- Leonin Shikari -------------------------------------------------

// "You may activate equip abilities any time you could cast an
// instant." The whole card.
func TestLeoninShikariEquipsAtInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	sword := pushCatalogPermanent(g, me.ID, "Bonesplitter", "Artifact — Equipment", bonesplitterTimingOracle, false)
	advanceTo(t, g, game.StepEnd)
	fillPool(me, 4)

	equipAt := func() error {
		return g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		})
	}
	if err := equipAt(); err != game.ErrSorcerySpeedRequired {
		t.Fatalf("setup: equip in an end step with no Shikari: got %v, want ErrSorcerySpeedRequired", err)
	}

	pushCatalogPermanent(g, me.ID, "Leonin Shikari", "Creature — Cat Soldier", leoninShikariOracle, false)
	if err := equipAt(); err != nil {
		t.Fatalf("equip under a Leonin Shikari in an end step: %v", err)
	}
}

// The clause is "YOU may activate", so an opponent's Equipment keeps
// its printed window (CR 702.6b).
func TestLeoninShikariDoesNotOpenAnOpponentsEquip(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirBear := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Bear")
	theirSword := pushCatalogPermanent(g, opp.ID, "Bonesplitter", "Artifact — Equipment", bonesplitterTimingOracle, false)
	pushCatalogPermanent(g, me.ID, "Leonin Shikari", "Creature — Cat Soldier", leoninShikariOracle, false)
	advanceTo(t, g, game.StepEnd)
	fillPool(opp, 4)

	err := g.ActivateCatalogAbility(opp.ID, theirSword, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirBear}},
	})
	if err != game.ErrSorcerySpeedRequired {
		t.Errorf("an opponent's Shikari opened their equip: got %v, want ErrSorcerySpeedRequired", err)
	}
}

// --- The Wandering Emperor -----------------------------------------

// "As long as The Wandering Emperor entered this turn, you may
// activate her loyalty abilities any time you could cast an instant."
// The condition is `game.EnteredThisTurn`, so the permanent has to
// ENTER rather than be pushed onto the battlefield.
func TestWanderingEmperorLoyaltyAtInstantSpeedOnlyTheTurnSheEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepEnd)

	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "The Wandering Emperor",
		TypeLine:   "Legendary Planeswalker — Wanderer",
		OracleID:   wanderingEmperorOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield},
		id,
	); err != nil {
		t.Fatalf("flash her in: %v", err)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, 3) })

	// The turn she entered: the −1 (no target) is activatable in an
	// end step, which is the whole point of the card.
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−1 in an end step on the turn she entered: %v", err)
	}
	passPriorityAroundTable(t, g)

	// A later turn: the clause is spent and CR 606.3 is back.
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepEnd)
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{}); err != game.ErrSorcerySpeedRequired {
		t.Errorf("−1 in a later end step: got %v, want ErrSorcerySpeedRequired", err)
	}
}

// `SummonedThisTurn` would have answered "yes" here and
// `EnteredThisTurn` answers "no" — the distinction the constructor's
// comment is about, asserted rather than asserted-in-prose.
func TestSourceEnteredThisTurnIsNotSummoningSickness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushCatalogPermanent(g, me.ID, "Test Walker", "Legendary Planeswalker — Test", "test-entered-this-turn", true)

	var card game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				card = c
			}
		}
	})
	if !card.SummonedThisTurn {
		t.Fatal("setup: the fixture did not set the summoning-sickness marker")
	}
	var entered bool
	g.ReadSnapshot(func() { entered = SourceEnteredThisTurn(g, card) })
	if entered {
		t.Error("a pushed permanent carries SummonedThisTurn but did not ENTER this turn")
	}
}

// --- Teferi, Master of Time ----------------------------------------

// "You may activate loyalty abilities of Teferi on any player's turn
// any time you could cast an instant" — the same clause with no
// condition, so it holds outside every main phase.
func TestTeferiMasterOfTimeActivatesLoyaltyAtInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Master of Time", teferiMasterOfTimeOracle, 3)
	advanceTo(t, g, game.StepEnd)

	// +1: draw then discard, activated outside any main phase. The
	// discard prompt it queues is the card working, so the test stops
	// here rather than stepping past an unanswered question.
	if err := g.ActivateCatalogAbility(me.ID, teferi, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1 in an end step: %v", err)
	}
	if got := loyaltyCount(g, teferi); got != 4 {
		t.Errorf("loyalty after +1: got %d, want 4", got)
	}
}

// "ON ANY PLAYER'S TURN", which is the half CR 606.3 otherwise
// forbids outright — the active-player test in sorcerySpeedOpenLocked.
func TestTeferiMasterOfTimeActivatesOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Master of Time", teferiMasterOfTimeOracle, 5)
	advanceToNextSeatsTurn(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	if active.ID == me.ID {
		t.Fatal("setup: the turn did not move to another seat")
	}
	theirs := pushCreatureToBattlefieldForTest(g, active.ID, "Their Bear")
	advanceTo(t, g, game.StepEnd)

	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err != nil {
		t.Fatalf("−3 on an opponent's end step: %v", err)
	}
	if got := loyaltyCount(g, teferi); got != 2 {
		t.Errorf("loyalty after −3: got %d, want 2", got)
	}
}

// The −3 phases out a creature you don't control, and only one you
// don't control (CR 115.4).
func TestTeferiMasterOfTimeMinusThreePhasesOut(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Master of Time", teferiMasterOfTimeOracle, 3)
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "My Bear")
	theirs := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Bear")
	toMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err != game.ErrIllegalTarget {
		t.Errorf("−3 at your own creature: got %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, teferi, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err != nil {
		t.Fatalf("−3 at an opponent's creature: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirs) {
		t.Error("the −3 did not phase the creature off the battlefield")
	}
}

// --- the equip marker -----------------------------------------------

// Every equip ability in the catalog carries the Equip flag, and
// nothing else does. The guard for the bug EquipOnlyAbility exists to
// prevent: a hand-written equip that Leonin Shikari cannot see.
func TestEveryEquipAbilityIsMarked(t *testing.T) {
	for _, spec := range All() {
		for _, ab := range spec.Activated {
			isEquip := strings.HasPrefix(ab.Label, "Equip ")
			if isEquip && !ab.Equip {
				t.Errorf("%s: %q looks like an equip ability and is not marked — use EquipAbility or EquipOnlyAbility", spec.Name, ab.Label)
			}
			if !isEquip && ab.Equip {
				t.Errorf("%s: %q is marked Equip and does not print the keyword", spec.Name, ab.Label)
			}
		}
	}
}
