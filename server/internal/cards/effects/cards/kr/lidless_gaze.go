package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lidless Gaze — Sorcery {2}{B}{R} (Tales of Middle-earth Commander):
//
//	"Exile the top card of each player's library. Until the end of your next
//	 turn, you may play those cards, and mana of any type can be spent to
//	 cast those spells.
//	 Flashback {2}{B}{R} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Bloodsoaked Insight's impulse draw (ExileTopWithPermission, the window
// ending as your next turn ends and any type of mana spendable), made once
// per seat in turn order from the caster's own library round the table
// (every player's library, yours included). Flashback is the shared
// alternative cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "287aae84-b098-456e-b9cb-8eac51c36360",
		Name:             "Lidless Gaze",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerValue},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{2}{B}{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			window := ctx.Game.UntilEndOfYourNextTurnDuration(item.Controller)
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Eliminated {
					continue
				}
				if err := (effects.ExileTopWithPermission{
					From:    p.ID,
					GrantTo: item.Controller,
					N:       1,
					AnyType: true,
					Window:  &window,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
