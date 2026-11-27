package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const victualSliverGrant = "victual-sliver/sacrifice-for-life"

// Victual Sliver — Creature — Sliver {G}{W}, 2/2 (Stronghold, S82 OLD b750,
// #5627):
//
//	"All Slivers have "{2}, Sacrifice this permanent: You gain 4 life.""
//
// Clot Sliver's shape: a bundle of one activated ability granted to every Sliver
// on the battlefield, the Victual Sliver itself and every player's included
// (TribalAbilityGrant without YoursOnly). The cost is Compulsory Rest's mana
// plus sacrificing the permanent; the life goes to the ability's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd803b1c-1370-40db-a6c6-abad4b3d1602",
		Name:         "Victual Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: victualSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}, Sacrifice this permanent: You gain 4 life.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Player: item.Controller, Amount: 4}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{2}, Sacrifice this permanent: You gain 4 life.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, victualSliverGrant),
		},
	})
}
