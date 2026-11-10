package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// block_not_alone_view_test.go — #5396: a creature that can't block
// alone is stamped cant_block_alone, and while its controller has staged
// no block, every attacker is refused to it as cant_block_alone in
// block_refusals; once another of their creatures blocks, none is.
func TestCantBlockAloneIsStampedAndRefusedInTheView(t *testing.T) {
	const beastOracle = "test-cant-block-alone"
	prev := game.CatalogBlockRules
	t.Cleanup(func() { game.CatalogBlockRules = prev })
	game.CatalogBlockRules = func(key string) []game.BlockRule {
		if key == beastOracle {
			return []game.BlockRule{{NotAlone: true}}
		}
		return nil
	}

	g := threeSeatsInDeclareAttackers(t)
	active := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%3]
	push := func(owner *game.Player, name, oracle string) uuid.UUID {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, OracleID: oracle, TypeLine: "Creature — Beast",
			Power: 4, Toughness: 4, Owner: owner.ID, Controller: owner.ID,
		})
		return id
	}
	raider := push(active, "Raider", "")
	beast := push(def, "Beast", beastOracle)
	bear := push(def, "Bear", "")
	if err := g.DeclareAttacker(raider, def.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("step %s, want declare_blockers", g.Turn.Step)
	}
	cards := func() map[string]CardView {
		out := map[string]CardView{}
		for _, c := range ViewOfGame(g).Battlefield.Cards {
			out[c.InstanceID] = c
		}
		return out
	}

	v := cards()
	if !v[beast.String()].CantBlockAlone || v[bear.String()].CantBlockAlone {
		t.Errorf("cant_block_alone stamps: Beast %v, Bear %v", v[beast.String()].CantBlockAlone, v[bear.String()].CantBlockAlone)
	}
	refs := v[beast.String()].BlockRefusals
	if len(refs) != 1 || refs[0].Attacker != raider.String() || refs[0].Reason != string(game.BlockReasonCantBlockAlone) || refs[0].Message == "" {
		t.Errorf("Beast refusals %+v, want one cant_block_alone with a sentence", refs)
	}
	if len(v[bear.String()].BlockRefusals) != 0 {
		t.Errorf("Bear refusals %+v, want none", v[bear.String()].BlockRefusals)
	}

	if err := g.DeclareBlocker(bear, raider); err != nil {
		t.Fatalf("Bear block: %v", err)
	}
	if refs := cards()[beast.String()].BlockRefusals; len(refs) != 0 {
		t.Errorf("Beast refusals beside the Bear %+v, want none", refs)
	}
}
