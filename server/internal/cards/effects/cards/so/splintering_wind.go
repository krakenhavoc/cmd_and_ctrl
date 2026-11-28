package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Splintering Wind — Enchantment {2}{G}{G} (Alliances, S82 OLD batch 818):
//
//	"{2}{G}: This enchantment deals 1 damage to target creature. Create a
//	 1/1 green Splinter creature token. It has flying and "Cumulative upkeep
//	 {G}." When it leaves the battlefield, it deals 1 damage to you and each
//	 creature you control."
//
// The token's abilities are printed on the token, so they live on its catalog
// template (PrintedOldB818Splinter), not on this enchantment. The damage is
// dealt first, then the token is made, even if the target left.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "184f68a1-0c13-42f4-aee6-3ca744f6a30b",
		Name:         "Splintering Wind",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: This enchantment deals 1 damage to target creature. Create a 1/1 green Splinter creature token. It has flying and \"Cumulative upkeep {G}.\" When it leaves the battlefield, it deals 1 damage to you and each creature you control.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{2}{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
					if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return effects.CreateToken{Template: effects.TokenFromCatalog(effects.PrintedOldB818Splinter), N: 1}.Apply(ctx)
			},
		}},
	})
}
