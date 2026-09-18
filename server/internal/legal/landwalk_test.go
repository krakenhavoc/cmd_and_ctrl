package legal_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

const (
	oracleLordOfAtlantis    = "cc7f290f-ca00-4285-9bdb-4b4402444f30"
	oracleTrailblazersBoots = "634d5009-cbf3-44cb-8c15-7057f501a210"
)

// TestLandwalkBlocksAreNeverOffered is #705's agreement test, with the
// two real cards that exercise both halves of landwalk: Lord of
// Atlantis GRANTS islandwalk (a layer-6 keyword on another creature)
// and Trailblazer's Boots grants nonbasic landwalk (a supertype, not a
// land type). The enumerator offers exactly the blocks the engine
// accepts, the withheld ones are really refused, and the offer tracks
// the defending player's lands as they change. Both sides read
// Game.CanBlockLocked, which is what keeps them agreeing (#544).
func TestLandwalkBlocksAreNeverOffered(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(them)

	battlefieldCard(g, me, game.Card{
		Name: "Lord of Atlantis", TypeLine: "Creature — Merfolk", Power: 2, Toughness: 2,
		OracleID: oracleLordOfAtlantis,
	})
	merfolk := battlefieldCard(g, me, game.Card{
		Name: "Merfolk Raider", TypeLine: "Creature — Merfolk Warrior", Power: 1, Toughness: 1,
	})
	booted := freshCreature(g, me, "Booted Bear")
	plain := freshCreature(g, me, "Plain Bear")
	boots := battlefieldCard(g, me, game.Card{
		Name: "Trailblazer's Boots", TypeLine: "Artifact — Equipment", OracleID: oracleTrailblazersBoots,
	})
	attachTo(t, g, boots, booted)

	battlefieldCard(g, them, basic("Island", "Island"))
	wall := freshCreature(g, them, "Wall")

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{merfolk, booted, plain} {
		if err := g.DeclareAttacker(a, them.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)

	offered := func() map[string]bool {
		moves := legal.EnumerateFor(g, them.ID)
		dispatchAll(t, g, them.ID, moves)
		out := map[string]bool{}
		for _, m := range moves {
			if m.Kind == legal.KindBlock && m.Source == wall {
				out[m.Label] = true
			}
		}
		return out
	}
	refusedFor := func(attacker uuid.UUID) game.BlockReason {
		t.Helper()
		err := g.Clone().DeclareBlocker(wall, attacker)
		var br *game.BlockRefusedError
		if !errors.Is(err, game.ErrIllegalBlock) || !errors.As(err, &br) {
			t.Fatalf("the engine accepted a block the enumerator withheld: %v", err)
		}
		return br.Reason
	}

	// A basic Island: the granted islandwalk bites, nonbasic landwalk
	// does not.
	got := offered()
	if got["Block Merfolk Raider with Wall"] {
		t.Error("offered a block against a Merfolk with granted islandwalk while the defender controls an Island")
	}
	if !got["Block Booted Bear with Wall"] || !got["Block Plain Bear with Wall"] || len(got) != 2 {
		t.Errorf("offered %v, want exactly the booted and the plain bear", got)
	}
	if r := refusedFor(merfolk); r != game.BlockReasonLandwalk {
		t.Errorf("islandwalk refusal reason = %q, want landwalk", r)
	}

	// A nonbasic land arrives: now the Boots bite too.
	battlefieldCard(g, them, game.Card{Name: "Tropical Island", TypeLine: "Land — Island Forest"})
	got = offered()
	if !got["Block Plain Bear with Wall"] || len(got) != 1 {
		t.Errorf("with a nonbasic land, offered %v, want only the plain bear", got)
	}
	if r := refusedFor(booted); r != game.BlockReasonLandwalk {
		t.Errorf("nonbasic landwalk refusal reason = %q, want landwalk", r)
	}
}
