package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necrogenesis — Enchantment {B}{G} (Shards of Alara, #5588):
//
//	"{2}: Exile target creature card from a graveyard. Create a 1/1 green
//	 Saproling creature token."
//
// Cemetery Reaper's activation with no tap and a Saproling: the target
// is a creature card in any graveyard. A target gone in response makes
// the ability fizzle, so no Saproling is made (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad08952a-bcae-41d6-8093-c18535dad7d5",
		Name:         "Necrogenesis",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Exile target creature card from a graveyard. Create a 1/1 green Saproling creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCreatureInAnyGraveyard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 green Saproling"), N: 1}.Apply(ctx)
			},
		}},
	})
}
