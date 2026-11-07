package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Veteran Ice Climber — Creature — Human Scout {1}{U}, 1/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Vigilance
//	 This creature can't be blocked.
//	 Whenever this creature attacks, up to one target player mills cards
//	 equal to this creature's power."
//
// Vigilance is a printed keyword and "can't be blocked" is RestrictSelf.
// The attack trigger targets zero or one player (Torgaar's "up to one
// target player"); with none chosen it does nothing. The power is read as
// the trigger resolves (sourcePowerNow: the Climber's current power, or
// its last known power if it has left, CR 608.2h), so a pump in response
// mills more. A power of 0 or less mills nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ccaadf8d-22c0-4139-b734-03ac2fcb4494",
		Name:            "Veteran Ice Climber",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			RestrictSelf(game.CantBeBlocked),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Veteran Ice Climber — up to one target player mills cards equal to its power",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					p := targetedPlayerOf(ctx)
					n := sourcePowerNow(ctx)
					if p == uuid.Nil || n <= 0 {
						return nil
					}
					return MillCards{Player: p, N: n}.Apply(ctx)
				}),
				TargetPlayer("up to one target player").WithCount(0, 1)),
		},
	})
}
