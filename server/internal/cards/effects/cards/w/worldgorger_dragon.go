package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Worldgorger Dragon — Creature — Nightmare Dragon {3}{R}{R}{R}, 7/7
// (Dominaria Remastered, #5948):
//
//	"Flying, trample
//	 When this creature enters, exile all other permanents you control.
//	 When this creature leaves the battlefield, return the exiled cards to
//	 the battlefield under their owners' control."
//
// Two triggers, linked (CR 607.2a): the enters trigger exiles every other
// permanent its controller controls in one move and the leaves trigger
// brings back what that ability exiled and is still in exile, all as one
// entry (ReturnFromExileTogether), under their owners' control. Tokens
// cease to exist in exile and do not return. If the Dragon leaves before
// its enters trigger resolves, the leaves trigger finds nothing and the
// permanents are exiled for good, as the rules have it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a628186d-b7d9-40a5-9ae2-fbc9d2a14c7c",
		Name:            "Worldgorger Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(worldgorgerDragonExileLabel, effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				self := ctx.Source()
				return effects.ExileAllMatching{Match: func(_ *game.Game, caster uuid.UUID, c game.Card) bool {
					return c.Controller == caster && c.InstanceID != self
				}}.Apply(ctx)
			}))),
			effects.On(game.EventLTB, effects.Self, "Worldgorger Dragon — return the exiled cards to the battlefield under their owners' control",
				effects.B41ReturnCardsExiledWithToTheBattlefield(worldgorgerDragonExileLabel)),
		},
	})
}

const worldgorgerDragonExileLabel = "Worldgorger Dragon — exile all other permanents you control"
