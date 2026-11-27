package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gelid Shackles — Snow Enchantment — Aura {W} (Coldsnap, #5597):
//
//	"Enchant creature
//	 Enchanted creature can't block, and its activated abilities can't be
//	 activated.
//	 {S}: Enchanted creature gains defender until end of turn."
//
// Bound in Gold's restriction with the mana half too (the abilities of the creature
// can't be activated, mana abilities included, since the card makes no exception),
// and Bronzehide Lion's ability of the Aura that reaches the creature it enchants.
// {S} is snow mana (Icebind Pillar).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45ca8431-befd-4542-9dd7-503bf1afaac5",
		Name:         "Gelid Shackles",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.RestrictAttached(game.CantBlock | game.CantActivate | game.CantActivateMana)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{S}: Enchanted creature gains defender until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.ManaCost("{S}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.EnchantedCreatureOf(ctx)
				if id == uuid.Nil {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"defender"}, Label: "Gelid Shackles"}.Apply(ctx)
			},
		}},
	})
}
