package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drudge Spell — Enchantment {B}{B} (Homelands, S82 OLD b827, #5635):
//
//	"{B}, Exile two creature cards from your graveyard: Create a 1/1 black
//	 Skeleton creature token. It has "{B}: Regenerate this token."
//	 When this enchantment leaves the battlefield, destroy all Skeleton tokens.
//	 They can't be regenerated."
//
// Cryptwailing's graveyard exile as a cost (the activator names the two cards
// as the ability is announced) and Skeletonize's regenerating Skeleton. The
// leave trigger destroys every Skeleton token on the battlefield, whoever
// controls it and whichever card made it, with regeneration ignored (CR
// 701.19c); a Skeleton that is not a token is spared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d63f4eeb-3fa5-4e65-8414-e3fd3c03c128",
		Name:         "Drudge Spell",
		Completeness: effects.CompletenessFull,
		// The leave trigger wipes the Skeleton tokens: a partial creature sweep.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Exile two creature cards from your graveyard: Create a 1/1 black Skeleton creature token. It has \"{B}: Regenerate this token.\"",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.ExileFromGraveyard(2, "two creature cards", effects.MatchCreature)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.OldB516SkeletonToken(), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves("Drudge Spell — destroy all Skeleton tokens",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyAllMatching{
						Match:             effects.And(effects.IsTokenPredicate(), effects.HasSubtype("Skeleton")),
						CantBeRegenerated: true,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
