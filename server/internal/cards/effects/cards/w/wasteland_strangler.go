package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wasteland Strangler — Creature — Eldrazi Processor {2}{B}, 3/2 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 When this creature enters, you may put a card an opponent owns from exile
//	 into that player's graveyard. If you do, target creature gets -3/-3 until
//	 end of turn."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) and
// re-checked as the -3/-3 is applied (CR 608.2b).
//
// No simplification.
func init() {
	const name = "Wasteland Strangler"
	effects.Register(effects.Spec{
		OracleID:        "197f3b17-ce21-430e-8e41-f5d02d5e3637",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.OldB305WhenThisEntersMayProcess(name, "wasteland-strangler",
				"target creature gets -3/-3 until end of turn",
				effects.TargetCreature("target creature"),
				func(ctx *effects.Context) error {
					for _, t := range ctx.LegalTargets() {
						return effects.BoostUntilEOT{Target: t.ID, Power: -3, Toughness: -3,
							Label: "Wasteland Strangler — -3/-3 until end of turn"}.Apply(ctx)
					}
					return nil
				}),
		},
	})
}
