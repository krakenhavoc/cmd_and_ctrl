package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// damage_split_banding_test.go — ADR 0224 (#5723). A division banding
// hands to the defending player (CR 702.22j) is offered to THAT player,
// first with the split that loses them the least, and is never marked
// covers_lethal, so the browser does not send it unasked.

// bandingBlock: a 4/4 trampler blocked by a 1/1 with banding and two
// more creatures of the given sizes; the cursor in combat damage.
func bandingBlock(t *testing.T, others ...[2]int) (*game.Game, uuid.UUID, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(def)
	atk := battlefieldCard(g, me, game.Card{
		Name: "Beast", TypeLine: "Creature — Beast", Power: 4, Toughness: 4, Keywords: []string{"trample"},
	})
	ids := []uuid.UUID{battlefieldCard(g, def, game.Card{
		Name: "Benalish Hero", TypeLine: "Creature — Human Soldier", Power: 1, Toughness: 1,
		Keywords: []string{game.KeywordBanding},
	})}
	for _, pt := range others {
		ids = append(ids, battlefieldCard(g, def, game.Card{
			Name: "Blocker", TypeLine: "Creature — Soldier", Power: pt[0], Toughness: pt[1],
		}))
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, def.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, id := range ids {
		if err := g.DeclareBlocker(id, atk); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepCombatDamage)
	return g, def.ID, atk, ids
}

func TestBandingSplitLosesTheDefenderOneCreature(t *testing.T) {
	// 4 damage, a 1/1 bander and two 2/2s: no split saves all three, so
	// all 4 go on the creature worth least, the Hero.
	g, def, _, ids := bandingBlock(t, [2]int{2, 2}, [2]int{2, 2})
	s := damageSplit(t, g)
	got := shares(s)
	if got[ids[0]] != 4 || got[ids[1]] != 0 || got[ids[2]] != 0 || s.TrampleTo != 0 {
		t.Fatalf("split %v, trample %d; want all 4 on the Hero", got, s.TrampleTo)
	}
	if s.CoversLethal {
		t.Error("a banding split is marked covers_lethal")
	}
	// The defending player is offered it first, and the engine takes it.
	m, offered, trample := assignmentMove(t, g, def)
	if offered[ids[0]] != 4 || trample != 0 {
		t.Fatalf("the defender is offered %v (trample %d) first", offered, trample)
	}
	dispatchAll(t, g, def, []legal.Move{m})
}

func TestBandingSplitSparesEveryCreatureWhenItCan(t *testing.T) {
	// 4 damage, a 1/1 bander and a 5/5: the 5/5 takes all 4 and lives.
	g, _, _, ids := bandingBlock(t, [2]int{5, 5})
	got := shares(damageSplit(t, g))
	if got[ids[0]] != 0 || got[ids[1]] != 4 {
		t.Fatalf("split %v; want all 4 on the 5/5, the Hero spared", got)
	}
}
