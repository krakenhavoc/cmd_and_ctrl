package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Compy Swarm — Creature — Dinosaur {1}{B}{G}, 2/2 (Jurassic World
// Collection, #5948):
//
//	"At the beginning of your end step, if a creature died this turn,
//	 create a tapped token that's a copy of this creature."
//
// AtYourEndStepIfAnyCreatureDied is the intervening "if" (CR 603.4, checked
// as the step begins and again on resolution). The copy is made from the
// Swarm's last-known copiable values if it has left (CR 608.2h) and enters
// tapped (TokenEntryOptions.Tapped).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5bd96dd-0d1f-4923-a014-6c8dcb3b2266",
		Name:         "Compy Swarm",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfAnyCreatureDied("Compy Swarm — create a tapped token that's a copy of this creature",
				func(g *game.Game, item *game.StackItem) error {
					tmpl, ok := TokenCopyTemplate(g, item.SourceCardID)
					if !ok {
						return nil
					}
					_, err := g.CreateTokensForEffect(item.Controller, copiedToken(tmpl), 1, game.TokenEntryOptions{Tapped: true})
					return err
				}),
		},
	})
}
