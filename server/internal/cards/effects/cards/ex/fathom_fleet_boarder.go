package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fathom Fleet Boarder — Creature — Orc Pirate {2}{B}, 3/3:
//
//	"When this creature enters, you lose 2 life unless you control another
//	 Pirate."
//
// Not an intervening "if": the trigger always goes on the stack, and whether
// another Pirate is controlled is read as it resolves (CR 608.2c). The Boarder
// itself never counts, and a changeling counts as a Pirate. The loss is a life
// change, not damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cffcd887-660c-4646-9333-bcc8549f6609",
		Name:         "Fathom Fleet Boarder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Fathom Fleet Boarder — you lose 2 life unless you control another Pirate",
				effects.Do(effects.If{
					Cond: effects.CondBoard{Name: "you control no other Pirate", Test: func(ctx *effects.Context) bool {
						return !effects.S82OldB219YouControlAnother(ctx, "Pirate")
					}},
					Then: []effects.Applier{effects.GainLife{Amount: -2}},
				})),
		},
	})
}
