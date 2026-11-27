package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eternal Dominion — Sorcery {7}{U}{U}{U} (Saviors of Kamigawa, #5601):
//
//	"Search target opponent's library for an artifact, creature,
//	 enchantment, or land card. Put that card onto the battlefield under
//	 your control. Then that player shuffles.
//	 Epic (For the rest of the game, you can't cast spells. At the
//	 beginning of each of your upkeeps, copy this spell except for its epic
//	 ability. You may choose a new target for the copy.)"
//
// Bribery's search over a target opponent's library (the caster chooses and
// controls what is found; the searched library is shuffled either way), with
// a quality, so it may find nothing (CR 701.23b). Epic is effects.Epic
// (CR 702.50a); each copy may be aimed at a different opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ede2a435-bf9a-4f87-b579-259cfdebc13f",
		Name:         "Eternal Dominion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.Epic(func(_ *game.StackItem, ctx *effects.Context) error {
			victim, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return effects.SearchLibrary{
				Player:       ctx.Controller(),
				LibraryOwner: victim,
				Predicate: func(c game.Card) bool {
					return c.IsArtifact() || c.IsCreature() || c.IsEnchantment() || c.IsLand()
				},
				Dest:    game.ZoneBattlefield,
				Limit:   1,
				Reveal:  true,
				Shuffle: true,
				Reason:  "Eternal Dominion — search target opponent's library for an artifact, creature, enchantment, or land card",
				Source:  ctx.Source(),
			}.Apply(ctx)
		}),
	})
}
