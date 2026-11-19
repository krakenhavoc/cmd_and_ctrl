package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaalia of the Vast — Legendary Creature — Human Cleric {1}{R}{W}{B}, 2/2
// (Modern Horizons 3):
//
//	"Flying
//	 Whenever Kaalia attacks an opponent, you may put an Angel, Demon, or Dragon
//	 creature card from your hand onto the battlefield tapped and attacking that
//	 opponent."
//
// The creature enters attacking the opponent Kaalia attacked, not a defender of
// the controller's choice (CR 508.4). It triggers nothing as an attacker.
func init() {
	const label = "Kaalia of the Vast — you may put an Angel, Demon, or Dragon creature card from your hand onto the battlefield tapped and attacking that opponent"
	Register(Spec{
		OracleID:        "cb8d80c9-ed58-4f2d-aa8c-c383370c7f1a",
		Name:            "Kaalia of the Vast",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, gapB50AttacksAnOpponent, label, func(g *game.Game, item *game.StackItem) error {
				cands := handCardsMatching(g, item.Controller,
					And(Creature(), Or(Subtype("Angel"), Subtype("Demon"), Subtype("Dragon"))))
				if len(cands) == 0 {
					return nil
				}
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Destination: game.ChooseOntoBattlefieldTapped,
					Chooser:     item.Controller,
					Source:      item.SourceCardID,
					Question:    label,
					Cards:       cands,
					Min:         0,
					Max:         1,
					ThenKey:     gapB50KaaliaPick,
					LiveZone:    game.ZoneHand,
					Carry:       []uuid.UUID{triggerAttackTarget(item)},
					Asked:       game.NewAskedItem(item),
				})
				return nil
			}),
		},
	})
}
