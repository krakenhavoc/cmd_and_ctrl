package legal_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// multi_block_followups_test.go — #1715. A creature given block
// capacity by a resolving effect is offered exactly the attackers it
// has room for, and one that "blocks each attacking creature if able"
// is owed on every attacker: the pass is withheld until it is, the
// required set is one AlwaysLegal move, and every move offered is
// accepted.

// registerOn pins a data record with `mods` to `id` for the turn.
func registerOn(t *testing.T, g *game.Game, id uuid.UUID, mods ...game.Mod) {
	t.Helper()
	ok := false
	g.WithWriteLock(func() {
		ok = g.RegisterScopedEffectForEffect(uuid.New(), g.PinnedObjectsLocked(id), mods, g.UntilEndOfTurnDuration(), "test — #1715")
	})
	if !ok {
		t.Fatal("registered nothing")
	}
}

// followupTable is three attackers against the defender's wall, the
// active player's pass made, the cursor in declare blockers.
func followupTable(t *testing.T, mods ...game.Mod) (g *game.Game, def *game.Player, wall uuid.UUID, atks []uuid.UUID) {
	t.Helper()
	g = newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	def = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(def)
	for _, n := range []string{"A", "B", "C"} {
		atks = append(atks, freshCreature(g, me, "Bear "+n))
	}
	wall = freshCreature(g, def, "Wall")
	advanceTo(t, g, game.StepDeclareAttackers)
	registerOn(t, g, wall, mods...)
	for _, id := range atks {
		if err := g.DeclareAttacker(id, def.ID); err != nil {
			t.Fatal(err)
		}
	}
	// #1501: the defender declares with priority parked; there is no
	// active player's pass to hand it over.
	advanceTo(t, g, game.StepDeclareBlockers)
	return g, def, wall, atks
}

// TestTurnScopedCapacityMovesAgreeWithTheEngine — "+1 this turn": two
// of the three attackers, then nothing, and the third really is refused.
func TestTurnScopedCapacityMovesAgreeWithTheEngine(t *testing.T) {
	g, def, wall, atks := followupTable(t, game.AddBlockCapacityMod(1))
	for i := 0; i < 2; i++ {
		moves := legal.EnumerateFor(g, def.ID)
		dispatchAll(t, g, def.ID, moves)
		if got := blockTargetsOffered(t, moves, wall); len(got) != 3-i {
			t.Fatalf("round %d: wall offered %d attackers, want %d", i, len(got), 3-i)
		}
		if err := g.DeclareBlocker(wall, atks[i]); err != nil {
			t.Fatal(err)
		}
	}
	if got := blockTargetsOffered(t, legal.EnumerateFor(g, def.ID), wall); len(got) != 0 {
		t.Fatalf("a full wall is still offered %v", got)
	}
	var br *game.BlockRefusedError
	if err := g.Clone().DeclareBlocker(wall, atks[2]); !errors.As(err, &br) || br.Reason != game.BlockReasonBlockerCapacity {
		t.Fatalf("a third block: %v, want blocker_capacity", err)
	}
}

// TestBlocksEachAttackerMovesAgreeWithTheEngine — Blaze of Glory's
// record: the finish is withheld (#1501: finish_blocks is the
// declaration's "done" now that priority is parked while the defender
// declares), the required move puts the wall on all three attackers,
// and once it is taken the finish comes back.
func TestBlocksEachAttackerMovesAgreeWithTheEngine(t *testing.T) {
	g, def, wall, atks := followupTable(t, game.BlockAnyNumberMod(),
		game.AddBlockRequirementMod(game.BlockRequirementBlocksEach))

	moves := legal.EnumerateFor(g, def.ID)
	dispatchAll(t, g, def.ID, moves)
	if n := countKind(moves, legal.KindPass) + countKind(moves, legal.KindFinishBlocks); n != 0 {
		t.Fatalf("pass or finish offered while the wall owes blocks: %v", labels(moves))
	}
	var required *legal.Move
	for i, m := range moves {
		if m.Kind == legal.KindBlock && m.AlwaysLegal {
			if required != nil {
				t.Fatalf("two AlwaysLegal block moves: %v", labels(moves))
			}
			required = &moves[i]
		}
	}
	if required == nil || required.Type != legal.TypeDeclareBlockers {
		t.Fatalf("no AlwaysLegal declare_blockers for the required blocks: %v", labels(moves))
	}
	var set struct {
		Blocks []struct{ Blocker, Attacker string }
	}
	if err := json.Unmarshal(required.Params, &set); err != nil {
		t.Fatal(err)
	}
	onWall := map[string]bool{}
	for _, b := range set.Blocks {
		if b.Blocker == wall.String() {
			onWall[b.Attacker] = true
		}
	}
	for _, a := range atks {
		if !onWall[a.String()] {
			t.Fatalf("required blocks %+v leave the wall off %s", set.Blocks, a)
		}
	}
	if err := g.Clone().FinishBlocks(def.ID); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("engine accepted the withheld finish: %v", err)
	}

	decls := make([]game.BlockDeclaration, 0, len(atks))
	for _, a := range atks {
		decls = append(decls, game.BlockDeclaration{Blocker: wall, Attacker: a})
	}
	if err := g.DeclareBlockers(decls); err != nil {
		t.Fatal(err)
	}
	moves = legal.EnumerateFor(g, def.ID)
	dispatchAll(t, g, def.ID, moves)
	if countKind(moves, legal.KindFinishBlocks) != 1 {
		t.Errorf("finish not offered once every attacker is blocked: %v", labels(moves))
	}
}
