package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dawnhart Wardens — Creature — Human Warlock {1}{G}{W}, 3/3 (Midnight
// Hunt, #5497):
//
//	"Vigilance
//	 Coven — At the beginning of combat on your turn, if you control
//	 three or more creatures with different powers, creatures you control
//	 get +1/+0 until end of turn."
//
// The affected creatures are fixed as the trigger resolves (CR 611.2c).
//
// No simplification.
func init() {
	const label = "Dawnhart Wardens — Coven: creatures you control get +1/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "0604ed6c-a85f-41da-a163-0c414e57d133",
		Name:            "Dawnhart Wardens",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.MidB05AtBeginningOfCombatIfCoven(label, func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()),
					Power: 1,
					Label: label,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
