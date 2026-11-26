package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mardu Skullhunter — Creature — Human Warrior {1}{B}, 2/1 (Khans of Tarkir):
//
//	"This creature enters tapped.
//	 Raid — When this creature enters, if you attacked this turn, target
//	 opponent discards a card."
//
// Enters tapped is a self-replacement (CR 614.1c). The raid trigger is an
// intervening "if" (CR 603.4, read as it enters and as it resolves) whose
// target opponent chooses the card they discard (CR 701.9b), as Heartless
// Pillage's discard does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a87e3a9-10e3-4b00-aeb9-d3336b2fa3b3",
		Name:         "Mardu Skullhunter",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OldB223RaidWhenThisEnters("Mardu Skullhunter — raid: target opponent discards a card",
				func(g *game.Game, item *game.StackItem) error {
					opp := effects.TargetedPlayer(effects.NewContext(g, item))
					if opp == uuid.Nil {
						return nil
					}
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: opp, Source: item.SourceCardID, N: 1})
					return nil
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
