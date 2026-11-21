package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oust — Sorcery {W}:
//
//	"Put target creature into its owner's library second from the top.
//	 Its controller gains 3 life."
//
// One-mana removal that costs its victim a draw step: the creature is
// the owner's draw after next, not gone. The position is the whole
// card, and it is PutIntoLibrary's Depth (ADR 0088 Decision 4) — the
// tuck route carries it, so a commander whose owner declines the
// command zone still lands second from the top.
//
// "Its controller" is read before the move: the creature's controller
// as it last existed on the battlefield, which is not always its
// owner — an Ousted stolen creature pays the thief. The life is gained
// whether or not the creature reached the library (a commander that
// took the command zone): the second sentence does not depend on the
// first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efc12fda-054b-466a-a863-06cf54878172",
		Name:         "Oust",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			targetID := item.Targets[0].ID
			card, ok := ctx.Game.LookupCardForEffect(targetID)
			if !ok {
				return nil
			}
			// Keyed (#4253 slice 10a): the controller's life gain is the
			// move's tail.
			return effects.PutIntoLibrary{
				Card:     targetID,
				Depth:    2,
				ThenTail: game.KeyedTail{Ref: effects.ThenChangeLife, Player: card.Controller, Source: ctx.Source(), N: 3},
			}.Apply(ctx)
		},
	})
}
