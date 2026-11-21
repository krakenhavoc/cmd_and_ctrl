package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nexos — Creature — Human Tyranid Advisor {1}{G}, 2/2:
//
//	"Strategic Coordinator — Basic lands you control have '{T}: Add {C}{C}.
//	 Spend this mana only on costs that contain {X}.'"
//
// A layer-6 grant of a mana ability stamped with game.ManaRestrictXCost, the
// restriction Elementalist's Palette uses (CR 106.6): the mana pays a spell
// whose mana cost has {X}, or an activated ability whose cost has {X}.
const nexosGrant = "nexos/colorless-for-x"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b9a2d17-205a-49cf-98c3-fd833015da18",
		Name:         "Nexos",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: nexosGrant,
			Mana: []effects.ManaAbility{{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Produced:     "{C}{C}",
				Restrictions: []string{game.ManaRestrictXCost},
				Label:        "{T}: Add {C}{C}. Spend this mana only on costs that contain {X}",
			}},
			Text: "{T}: Add {C}{C}. Spend this mana only on costs that contain {X}.",
		}},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.LandsYouControl(target, g, source) && target.HasSupertype("basic")
			}, nexosGrant), "Basic lands you control have \"{T}: Add {C}{C}. Spend this mana only on costs that contain {X}.\""),
		},
	})
}
