package aiseat_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// block_tax_test.go — the bot half of ADR 0221 (#5375): the CR 509.1d
// block tax is paid as the declaration completes, so the enumerator
// offers only blocks the seat could still pay for, and finish_blocks
// carries the price and auto_tap. Not behind AISEAT_GAME_TESTS, for
// attack_tax_test.go's reason.

const botArchangelOfTithesOracle = "ff5caed4-0276-476d-8fae-edae2536df7f"

// tithesBlockTable: seat 0 attacks seat 1 with a real Archangel of
// Tithes and a 4/4 Ogre; seat 1 has two 2/2 Bears and `lands` untapped
// Forests, and is declaring blockers.
func tithesBlockTable(t *testing.T, lands int) (g *game.Game, ogre uuid.UUID, bears [2]uuid.UUID) {
	t.Helper()
	g = newSettledTable(t, 41)
	me, them := g.Seats[0], g.Seats[1]
	push := func(owner *game.Player, name, typeLine, oracle string, power, tough int) uuid.UUID {
		c := game.Card{
			InstanceID: uuid.New(),
			Name:       name, TypeLine: typeLine, OracleID: oracle,
			Power: power, Toughness: tough,
			Owner: owner.ID, Controller: owner.ID,
		}
		c.AddKnowersAll(seatIDs(g))
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	angel := push(me, "Archangel of Tithes", "Creature — Angel", botArchangelOfTithesOracle, 3, 5)
	ogre = push(me, "Ogre", "Creature — Ogre", "", 4, 4)
	bears[0] = push(them, "Bear One", "Creature — Bear", "", 2, 2)
	bears[1] = push(them, "Bear Two", "Creature — Bear", "", 2, 2)
	for i := 0; i < lands; i++ {
		push(them, "Forest", "Basic Land — Forest", "", 0, 0)
	}
	advanceToStep(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: angel, Target: them.ID},
		{Attacker: ogre, Target: them.ID},
	}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	advanceToStep(t, g, game.StepDeclareBlockers)
	if g.BlockDeclarationStatusOf(them.ID) != game.BlockDeclarationPending {
		t.Fatal("the defender is not declaring")
	}
	return g, ogre, bears
}

func movesOfKind(moves []legal.Move, k legal.Kind) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Kind == k {
			out = append(out, m)
		}
	}
	return out
}

func dispatchBlockTaxMove(t *testing.T, g *game.Game, caller uuid.UUID, m legal.Move) error {
	t.Helper()
	return actions.Dispatch(g, actions.Action{
		Type:   actions.Type(m.Type),
		Player: m.Player,
		Caller: caller,
		Params: m.Params,
	})
}

func TestTheBotIsOfferedNoBlockItCannotPayFor(t *testing.T) {
	g, _, _ := tithesBlockTable(t, 0)
	def := g.Seats[1].ID
	moves := legal.EnumerateFor(g, def)
	if b := movesOfKind(moves, legal.KindBlock); len(b) != 0 {
		t.Fatalf("offered %d blocks with no mana under Archangel of Tithes: %v", len(b), attackLabels(b))
	}
	fin := movesOfKind(moves, legal.KindFinishBlocks)
	if len(fin) != 1 || !fin[0].AlwaysLegal || fin[0].Cost != nil {
		t.Fatalf("want one free AlwaysLegal No blocks, got %+v", fin)
	}
	if err := dispatchBlockTaxMove(t, g, def, fin[0]); err != nil {
		t.Fatalf("No blocks was refused: %v", err)
	}
}

func TestTheBotsBlockMovesStayWithinWhatItCanPay(t *testing.T) {
	g, _, _ := tithesBlockTable(t, 1)
	def := g.Seats[1].ID
	blocks := movesOfKind(legal.EnumerateFor(g, def), legal.KindBlock)
	if len(blocks) == 0 {
		t.Fatal("one Forest covers {1} but no block was offered")
	}
	for _, m := range blocks {
		if m.Cost == nil || m.Cost.Mana != "{1}" {
			t.Errorf("block %q costs %#v, want {1}", m.Label, m.Cost)
		}
	}
	if err := dispatchBlockTaxMove(t, g, def, blocks[0]); err != nil {
		t.Fatalf("the engine refused an enumerated block: %v", err)
	}
	// One Forest pays for one blocker: a second Bear is not offered.
	moves := legal.EnumerateFor(g, def)
	for _, m := range movesOfKind(moves, legal.KindBlock) {
		if m.Cost != nil && m.Cost.Mana != "" {
			t.Errorf("offered %q, a second taxed blocker the seat cannot pay for", m.Label)
		}
	}
	fin := movesOfKind(moves, legal.KindFinishBlocks)
	if len(fin) != 1 || fin[0].Cost == nil || fin[0].Cost.Mana != "{1}" {
		t.Fatalf("finish_blocks should carry the {1}: %+v", fin)
	}
	var p struct {
		AutoTap bool `json:"auto_tap"`
	}
	if err := json.Unmarshal(fin[0].Params, &p); err != nil || !p.AutoTap {
		t.Fatalf("finish_blocks params %s, want auto_tap", fin[0].Params)
	}
	if err := dispatchBlockTaxMove(t, g, def, fin[0]); err != nil {
		t.Fatalf("the engine refused the enumerated finish: %v", err)
	}
	if g.BlockDeclarationStatusOf(def) != game.BlockDeclarationDeclared {
		t.Error("the finish did not complete the declaration")
	}
	tapped := 0
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if c := &g.Battlefield.Cards[i]; c.Name == "Forest" && c.Tapped {
				tapped++
			}
		}
	})
	if tapped != 1 {
		t.Errorf("%d Forests tapped for the block tax, want 1", tapped)
	}
}

// TestTheBotDeclaresBlocksUnderABlockTaxWithoutWedging drives the
// policy: whatever it picks, the engine accepts, and the declaration
// ends.
func TestTheBotDeclaresBlocksUnderABlockTaxWithoutWedging(t *testing.T) {
	for _, lands := range []int{0, 1, 2} {
		g, _, _ := tithesBlockTable(t, lands)
		def := g.Seats[1].ID
		for i := 0; g.BlockDeclarationStatusOf(def) == game.BlockDeclarationPending; i++ {
			if i > 8 {
				t.Fatalf("lands=%d: still declaring after %d decisions", lands, i)
			}
			moves := legal.EnumerateFor(g, def)
			d, err := heuristic.New().Decide(context.Background(), aiseat.Input{
				View:  protocol.ViewOfGameFor(g, def.String()),
				Seat:  def,
				Moves: moves,
			})
			if err != nil {
				t.Fatalf("policy: %v", err)
			}
			idx := d.Index
			if idx < 0 {
				idx = aiseat.SafeIndex(moves)
			}
			if idx < 0 || idx >= len(moves) {
				t.Fatalf("lands=%d: no move to take (%d)", lands, idx)
			}
			if err := dispatchBlockTaxMove(t, g, def, moves[idx]); err != nil {
				t.Fatalf("lands=%d: the engine refused %q: %v", lands, moves[idx].Label, err)
			}
		}
	}
}
