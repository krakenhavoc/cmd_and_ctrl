package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Calibrated Blast — Instant {2}{R} (Modern Horizons 2, #5501):
//
//	"Reveal cards from the top of your library until you reveal a nonland
//	 card. Put the revealed cards on the bottom of your library in a random
//	 order. When you reveal a nonland card this way, Calibrated Blast deals
//	 damage equal to that card's mana value to any target.
//	 Flashback {3}{R}{R}"
//
// The run is revealed (a token is never a hit), every revealed card goes to
// the bottom in a random order, and a reflexive trigger — Breeches' blast —
// deals damage equal to the hit's mana value to a target chosen as it goes on
// the stack (CR 603.12). A library with no nonland card reveals itself and
// triggers nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "d9f5fc77-40df-4872-b633-6bf5a22a8831",
		Name:             "Calibrated Blast",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{R}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			run, hit := effects.RevealUntil(ctx, ctx.Controller(), func(c game.Card) bool { return !c.IsLand() }, "Calibrated Blast")
			mv := 0
			if hit != uuid.Nil {
				if c, ok := ctx.Game.LookupCardForEffect(hit); ok {
					mv = c.ManaValue()
				}
			}
			if err := ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, run); err != nil {
				return err
			}
			if hit == uuid.Nil {
				return nil
			}
			return effects.ReflexiveTrigger{
				Label:  "Calibrated Blast — deals damage equal to the revealed card's mana value to any target",
				Body:   effects.OldB032CalibratedBlastBody,
				Params: game.EffectParams{Amount: mv},
			}.Apply(ctx)
		},
	})
}
