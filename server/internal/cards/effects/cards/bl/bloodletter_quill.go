package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodletter Quill — Artifact {3} (Ravnica: City of Guilds, S82 OLD b582):
//
//	"{2}, {T}, Put a blood counter on this artifact: Draw a card, then you lose
//	 1 life for each blood counter on this artifact.
//	 {U}{B}: Remove a blood counter from this artifact."
//
// Netherborn Altar's counter-as-cost (AddCounterToThis), so the first
// activation already counts it. The life waits for the draw through the draw
// continuation (CR 608.2c) and counts the counters as the Quill is then. The
// second ability is an effect, not a cost, so it does nothing with no counter
// to remove.
//
// No simplification.
func init() {
	const blood = "blood"
	effects.Register(effects.Spec{
		OracleID:     "79cf9c52-56e0-4237-a595-ec07824e5eed",
		Name:         "Bloodletter Quill",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}, Put a blood counter on this artifact: Draw a card, then you lose 1 life for each blood counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.AddCounterToThis(blood, 1)),
				Effect:  effects.S82OldB582BloodletterQuillDraw,
			},
			{
				Label:   "{U}{B}: Remove a blood counter from this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{U}{B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) || effects.B39CountersOn(g, item.SourceCardID, blood) == 0 {
						return nil
					}
					return g.AddCounterForEffect(item.SourceCardID, blood, -1)
				},
			},
		},
	})
}
