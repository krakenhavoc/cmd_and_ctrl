package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const veiledApparitionUpkeepGrant = "veiled-apparition/upkeep-tax"

// Veiled Apparition — Enchantment {1}{U}:
//
//	"When an opponent casts a spell, if this permanent is an enchantment, it
//	 becomes a 3/3 Illusion creature with flying and "At the beginning of your
//	 upkeep, sacrifice this creature unless you pay {1}{U}.""
//
// The flying and the quoted upkeep ability are given in the same effect as the
// type change (a granted ability bundle, ADR 0093), so they last as long as the
// permanent is this object. An intervening "if" (CR 603.4).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4187fd6d-5660-4ea7-a9dd-402487ec8721",
		Name:         "Veiled Apparition",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: veiledApparitionUpkeepGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Veiled Apparition — sacrifice this creature unless you pay {1}{U}", func(g *game.Game, item *game.StackItem) error {
					return effects.UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     "{1}{U}",
						Question: "Veiled Apparition — pay {1}{U} or sacrifice this creature?",
						Else:     effects.UpkeepSacrificeSource,
					}.Apply(effects.NewContext(g, item))
				}),
			},
			Text: "At the beginning of your upkeep, sacrifice this creature unless you pay {1}{U}.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.OldB717AnOpponentCastIfEnchantment(nil),
				"Veiled Apparition — it becomes a 3/3 Illusion creature with flying and an upkeep tax",
				effects.OldB717BecomesACreature([]string{"Illusion"}, 3, 3, "Veiled Apparition — a 3/3 Illusion creature with flying",
					game.AddKeywordsMod("flying"), game.GrantAbilitiesMod(veiledApparitionUpkeepGrant))),
		},
	})
}
