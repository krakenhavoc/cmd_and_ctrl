package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Venture Forth — Sorcery {3}{G} (Commander Legends: Battle for Baldur's Gate,
// S82 GAP #5948):
//
//	"Exile cards from the top of your library until you exile a land card. Put
//	 that card onto the battlefield and the rest on the bottom of your library in
//	 a random order. Exile Venture Forth with three time counters on it.
//	 Suspend 3—{1}{G}"
//
// Solstice Revelations' exile run, stopped at the first land (effects.UntilLand),
// whose continuation is handed the cards that reached exile: the land is put onto
// the battlefield under the caster's control (a land put onto the battlefield is
// not a land play) and the rest go to the bottom of the library in a random order;
// with no land in the library everything goes to the bottom. Then the spell exiles
// itself with three time counters (Inspiring Refrain's step), so it suspends and
// comes back for free when the last counter is removed (Suspend is the engine's
// special action, CR 702.62).
func init() {
	effects.Register(effects.Spec{
		OracleID:       "261f371f-bcda-4761-810f-3216608b31d8",
		Name:           "Venture Forth",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Lands: 1, LandsUntapped: 1},
		SpecialActions: []game.SpecialAction{effects.Suspend(3, "{1}{G}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.MillToZoneKeyedForEffect(item.Controller, 0, game.ZoneExile, effects.UntilLand, ctx.ItemTail(ventureForthExiled))
		},
	})
}

// ventureForthExiled puts the land onto the battlefield, the rest on the bottom
// in a random order, then exiles the spell with three time counters. An on-disk
// key.
var ventureForthExiled = effects.CardTail("then/venture-forth-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	g, you := ctx.Game, ctx.Controller()
	var land uuid.UUID
	var rest []uuid.UUID
	for _, id := range o.IDs {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() && land == uuid.Nil {
			land = id
			continue
		}
		rest = append(rest, id)
	}
	if land != uuid.Nil {
		if err := (effects.ReturnFromExileTogether{Targets: []uuid.UUID{land}, Controller: you}).Apply(ctx); err != nil {
			return err
		}
	}
	if len(rest) > 0 {
		if err := g.PutOnBottomInRandomOrderForEffect(you, game.ZoneExile, rest); err != nil {
			return err
		}
	}
	return effects.Cmdr2S40ExileWithTimeCounters(ctx, 3)
})
