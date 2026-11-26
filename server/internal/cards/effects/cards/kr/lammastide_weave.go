package cardskr

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lammastide Weave — Instant {1}{G} (Lorwyn, S82 OLD b546, #5593):
//
//	"Choose a card name, then target player mills a card. If a card with
//	 the chosen name was milled this way, you gain life equal to its mana
//	 value.
//	 Draw a card."
//
// Three keyed steps, so a table waiting on any prompt is a restore point:
// the caster names a card (the target player is carried), the target mills
// one card with the name riding the tail, and when it lands in the graveyard
// the life is gained if the names match (case-insensitively, as card names
// are compared) and then the caster draws, the last sentence. The mana
// value is read from the milled card in the graveyard. A target that is
// illegal as the spell resolves stops all of it, the draw included
// (CR 608.2b). A mill that is replaced away mills nothing, gains nothing,
// and the draw still happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74250f3e-ad88-486f-a10a-90d9a2f0097f",
		Name:         "Lammastide Weave",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			ctx.Game.QueueCardNameChoiceKeyedForEffect(item.Controller, item.SourceCardID, "Lammastide Weave — choose a card name",
				game.KeyedTail{Ref: lammastideWeaveNamed, Player: item.Controller, Source: item.SourceCardID, Carry: []uuid.UUID{item.Targets[0].ID}})
			return nil
		},
	})
}

var lammastideWeaveNamed = game.RegisterTail("lammastide-weave/named", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return lammastideWeaveDraw(g, o.Tail.Player)
	}
	next := o.Tail
	next.Ref = lammastideWeaveMilled
	next.Carry = nil
	next.Word = o.Tail.Word
	return g.MillToZoneKeyedForEffect(o.Tail.Carry[0], 1, game.ZoneGraveyard, game.MillUntil{}, next)
})

var lammastideWeaveMilled = game.RegisterTail("lammastide-weave/milled", func(g *game.Game, o game.TailOutcome) error {
	name := strings.TrimSpace(o.Tail.Word)
	for _, id := range o.IDs {
		c, ok := g.LookupCardForEffect(id)
		if !ok || name == "" || !strings.EqualFold(strings.TrimSpace(c.Name), name) {
			continue
		}
		if mv := c.ManaValue(); mv > 0 {
			if err := g.ChangePlayerLifeForEffect(o.Tail.Source, o.Tail.Player, mv); err != nil {
				return err
			}
		}
	}
	return lammastideWeaveDraw(g, o.Tail.Player)
})

func lammastideWeaveDraw(g *game.Game, player uuid.UUID) error {
	return g.DrawNForEffect(player, 1)
}
