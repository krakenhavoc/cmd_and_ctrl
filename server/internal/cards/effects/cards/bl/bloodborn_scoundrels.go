package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodborn Scoundrels — Creature — Vampire Rogue {5}{B}:
//
//	"Assist (Another player can pay up to {5} of this spell's cost.)
//	 When this creature enters, target opponent loses 2 life and you
//	 gain 2 life."
//
// The ETB is the drain Abundant Maw's cast trigger uses, through the
// same targetOpponentLosesAndYouGain body: the opponent is a Targeting
// pick, and nothing is lost or gained if that opponent has left by the
// time the trigger resolves (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "37c1de8d-30d6-4b08-9368-279cf66d3e85",
		Name:            "Bloodborn Scoundrels",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Bloodborn Scoundrels — target opponent loses 2 life and you gain 2 life",
					func(g *game.Game, item *game.StackItem) error {
						return effects.TargetOpponentLosesAndYouGain(g, item, 2)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
