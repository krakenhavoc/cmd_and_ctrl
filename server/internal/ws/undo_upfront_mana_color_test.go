package ws

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Shivan Reef
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// undo_upfront_mana_color_test.go — #1443 over the wire and through the
// undo stack. `activate_mana_ability` with a `color` taps a painland,
// deals its damage and puts the named colour in the pool in ONE action,
// so one undo has to take all three back. The undo entry is a
// pre-mutation clone of the whole game, so this is a claim that the
// one-step activation is one entry, not a new undo path.

const upfrontShivanReefOracle = "0fe16212-66c3-4e45-a641-7391e9b2e304"

func upfrontReefRoom(t *testing.T) (*Room, *game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := seedTestGame(t)
	seat := g.Seats[0].ID
	ids, err := g.SpawnCards(seat, seat, game.ZoneBattlefield, game.Card{
		Name: "Shivan Reef", TypeLine: "Land", OracleID: upfrontShivanReefOracle,
	}, 1)
	if err != nil {
		t.Fatalf("spawn Shivan Reef: %v", err)
	}
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), t.TempDir())
	return room, g, seat, ids[0]
}

func upfrontActivate(g *game.Game, seat, land uuid.UUID, extra map[string]any) error {
	params := map[string]any{"card_id": land.String(), "ability_index": 1}
	for k, v := range extra {
		params[k] = v
	}
	raw, _ := json.Marshal(params)
	return actions.Dispatch(g, actions.Action{
		Type: actions.TypeActivateManaAbility, Player: seat, Caller: seat, Params: raw,
	})
}

type upfrontState struct {
	tapped  bool
	life    int
	pool    []string
	pending int
}

func upfrontRead(g *game.Game, seat, land uuid.UUID) upfrontState {
	var s upfrontState
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == land {
				s.tapped = c.Tapped
			}
		}
		p := g.PlayerByIDForEffect(seat)
		s.life = p.Life
		for _, tok := range p.ManaPool {
			s.pool = append(s.pool, tok.Color)
		}
		s.pending = len(g.PendingChoices)
	})
	return s
}

func TestUndoTakesBackAPainlandTappedWithItsColourNamed(t *testing.T) {
	room, g, seat, land := upfrontReefRoom(t)
	before := upfrontRead(g, seat, land)

	if _, _, err := room.Apply(seat, func() error {
		return upfrontActivate(g, seat, land, map[string]any{"color": "R"})
	}); err != nil {
		t.Fatalf("activate_mana_ability {color: R}: %v", err)
	}
	after := upfrontRead(g, seat, land)
	if !after.tapped || after.life != before.life-1 || len(after.pool) != 1 || after.pool[0] != "R" || after.pending != before.pending {
		t.Fatalf("after the one-step activation: %+v, want tapped, 1 damage, pool [R], no new prompt (before %+v)", after, before)
	}

	if _, _, err := room.Undo(uuid.Nil); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	undone := upfrontRead(g, seat, land)
	if undone.tapped || undone.life != before.life || len(undone.pool) != 0 || undone.pending != before.pending {
		t.Errorf("after undo: %+v, want the land untapped, life %d, an empty pool and no prompt", undone, before.life)
	}
}

// The wire's two spellings: `color` for one slot, `colors` for several,
// never both — and an illegal colour is refused with nothing tapped.
func TestActivateManaAbilityColorParamOnTheWire(t *testing.T) {
	_, g, seat, land := upfrontReefRoom(t)
	before := upfrontRead(g, seat, land)

	if err := upfrontActivate(g, seat, land, map[string]any{"color": "R", "colors": []string{"U"}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("color AND colors: err = %v, want ErrInvalidParam", err)
	}
	if err := upfrontActivate(g, seat, land, map[string]any{"color": "G"}); !errors.Is(err, game.ErrIllegalManaColor) {
		t.Errorf("color G on Shivan Reef: err = %v, want ErrIllegalManaColor", err)
	}
	if s := upfrontRead(g, seat, land); s.tapped || s.life != before.life || len(s.pool) != 0 {
		t.Fatalf("a refused activation changed the board: %+v", s)
	}
	if err := upfrontActivate(g, seat, land, map[string]any{"colors": []string{"U"}}); err != nil {
		t.Fatalf("colors [U]: %v", err)
	}
	if s := upfrontRead(g, seat, land); len(s.pool) != 1 || s.pool[0] != "U" {
		t.Errorf("colors [U]: pool = %v, want [U]", s.pool)
	}
}
