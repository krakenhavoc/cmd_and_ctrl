package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// goblin_test_pilot_test.go — #4051: an activated ability's target
// chosen at random. The activator names nothing; the announcement draws
// the target from every legal one (the Pilot included) as part of
// activating, logs it, and an undo replays the draw.

const goblinTestPilotOracle = "4760bb87-56e9-47eb-b5ec-6fa4dd19156e"

// gtpPilotID is fixed so a pinned RNG key gives the same draw every run:
// the keyed stream is seeded from the source's id too.
var gtpPilotID = uuid.MustParse("0b6e7f1a-4051-4c0d-9a7e-000000004051")

type gtpTable struct {
	g     *game.Game
	me    *game.Player
	pilot uuid.UUID
	bear  uuid.UUID
	rock  uuid.UUID
}

func gtpGame(t *testing.T, key byte) gtpTable {
	t.Helper()
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(randomModesKey(key))
	me := g.Seats[0]
	pilot := pushCatalogPermanentWithID(g, gtpPilotID, me.ID, "Goblin Test Pilot", "Creature — Goblin Pilot Wizard", goblinTestPilotOracle, false)
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 4)
	rock := b12Push(g, g.Seats[2].ID, "Rock", "Artifact", "", 0, 0)
	return gtpTable{g: g, me: me, pilot: pilot, bear: bear, rock: rock}
}

// activate announces the Pilot's ability with no target and returns
// the target the announcement drew.
func (tb gtpTable) activate(t *testing.T) game.TargetRef {
	t.Helper()
	if err := tb.g.ActivateCatalogAbility(tb.me.ID, tb.pilot, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if p := latestPickTarget(tb.g, tb.me.ID); p != nil {
		t.Fatalf("a target chosen at random asks nobody: %+v", p)
	}
	item := otItemFrom(tb.g, tb.pilot)
	if item == nil {
		t.Fatal("the ability is not on the stack")
	}
	return soleTarget(t, item)
}

func TestGoblinTestPilotDrawsAnyTargetAtRandom(t *testing.T) {
	seen := map[string]bool{}
	for key := byte(0); key < 96 && len(seen) < 4; key++ {
		tb := gtpGame(t, key)
		g := tb.g
		tgt := tb.activate(t)
		if !findBattlefieldCardForTest(g, tb.pilot).Tapped {
			t.Fatal("the Pilot was not tapped to pay {T}")
		}
		logged := false
		for _, ev := range g.Events {
			if ev.Kind == game.EventOptionChosen && ev.CardID == tb.pilot && strings.HasSuffix(ev.Label, "(at random)") {
				logged = true
			}
		}
		if !logged {
			t.Fatalf("key %d: the draw is not logged", key)
		}
		lives := map[uuid.UUID]int{}
		for _, s := range g.Seats {
			lives[s.ID] = s.Life
		}
		switch {
		case tgt.ID == tb.rock:
			t.Fatalf("key %d: drew the artifact, which is not \"any target\"", key)
		case tgt.Kind == game.TargetPlayer && tgt.ID == tb.me.ID:
			seen["me"] = true
		case tgt.Kind == game.TargetPlayer:
			seen["opponent"] = true
		case tgt.ID == tb.pilot:
			seen["itself"] = true
		case tgt.ID == tb.bear:
			seen["creature"] = true
		default:
			t.Fatalf("key %d: drew %+v", key, tgt)
		}
		passPriorityAroundTable(t, g)
		switch {
		case tgt.Kind == game.TargetPlayer:
			if got := playerByIDForTest(g, tgt.ID).Life; got != lives[tgt.ID]-2 {
				t.Errorf("key %d: the drawn player is at %d life, want %d", key, got, lives[tgt.ID]-2)
			}
		case tgt.ID == tb.bear:
			if d := b12Card(t, g, tb.bear).DamageMarked; d != 2 {
				t.Errorf("key %d: the drawn creature has %d damage, want 2", key, d)
			}
		}
	}
	if len(seen) != 4 {
		t.Errorf("96 draws reached %v; want the activator, an opponent, the Pilot itself and a creature", seen)
	}
}

// No payload may choose the random target, and a refused announcement
// pays nothing.
func TestGoblinTestPilotRefusesANamedTarget(t *testing.T) {
	tb := gtpGame(t, 0)
	err := tb.g.ActivateCatalogAbility(tb.me.ID, tb.pilot, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tb.bear}},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a payload naming the random target: %v, want ErrInvalidParam", err)
	}
	if findBattlefieldCardForTest(tb.g, tb.pilot).Tapped {
		t.Fatal("the refused activation tapped the Pilot")
	}
}

// The draw comes from the keyed stream: an undo, then the same
// activation, draws the same target.
func TestGoblinTestPilotReplaysTheDrawAfterUndo(t *testing.T) {
	for key := byte(0); key < 6; key++ {
		tb := gtpGame(t, key)
		pre := tb.g.Clone()
		first := tb.activate(t)
		tb.g.WithWriteLock(func() { tb.g.RestoreFrom(pre) })
		if again := tb.activate(t); again.ID != first.ID {
			t.Errorf("key %d: undo then redo drew %s then %s", key, first.ID, again.ID)
		}
	}
}
