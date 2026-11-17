package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honden of Night's Reach — Legendary Enchantment — Shrine {3}{B}
// (Champions of Kamigawa, #5603):
//
//	"At the beginning of your upkeep, target opponent discards a card
//	 for each Shrine you control."
//
// The opponent is chosen as the trigger goes on the stack and the
// Shrines are counted as it resolves (CR 608.2h). The opponent picks the
// cards (Mind Rot's discard prompt).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3776d11b-18fb-4451-a8f8-bda4c7c23776",
		Name:         "Honden of Night's Reach",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(chkB1Honden("Honden of Night's Reach — target opponent discards a card for each Shrine you control",
				func(g *game.Game, item *game.StackItem, shrines int) error {
					for _, t := range NewContext(g, item).LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: t.ID, Source: item.SourceCardID, N: shrines})
						return nil
					}
					return nil
				}), TargetPlayer("target opponent", Opponent())),
		},
	})
}
