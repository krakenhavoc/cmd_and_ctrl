package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Throne of the God-Pharaoh — Legendary Artifact {2} (EDHREC rank
// 2628):
//
//	"At the beginning of your end step, each opponent loses life equal
//	 to the number of tapped creatures you control."
//
// The go-wide deck's reach. The controller's own end step; the
// tapped creatures are counted as the trigger RESOLVES, so a
// creature untapped in response is not counted and an attacker that
// died in combat never was. Life loss, not damage — no prevention
// or doubler sees it — and the same amount to every opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea750169-1f6f-40c2-96e9-55719e103a63",
		Name:         "Throne of the God-Pharaoh",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Throne of the God-Pharaoh — each opponent loses life equal to your tapped creatures", func(g *game.Game, item *game.StackItem) error {
				n := effects.B24TappedCreaturesControlled(g, item.Controller)
				if n <= 0 {
					return nil
				}
				return effects.EachOpponentLosesLifeNow(g, item, n)
			}),
		},
	})
}
