package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const prideOfHullCladeGrant = "the-pride-of-hull-clade/draw-cards-equal-to-toughness"

// The Pride of Hull Clade — Legendary Creature — Crocodile Elk Turtle
// {10}{G} (Murders at Karlov Manor, #4469):
//
//	"This spell costs {X} less to cast, where X is the total toughness of
//	 creatures you control.
//	 Defender
//	 {2}{U}{U}: Until end of turn, target creature you control gets +1/+0,
//	 gains 'Whenever this creature deals combat damage to a player, draw
//	 cards equal to its toughness,' and can attack as though it didn't have
//	 defender."
//
// The discount is the spell's own cost modifier (SelfCostModifiers): the
// total toughness is read as the cost is determined (the layered view, so an
// anthem counts), and CR 118.7 stops it at {G} (a reduction never touches the
// coloured symbol); the mana value stays 11. The activated ability is one
// effect with one timestamp: +1/+0, the granted trigger and the defender
// waiver (GrantAbilitiesFor with Also). The granted draw counts the
// creature's toughness as it resolves, or as it last existed if it has left
// the battlefield (CR 608.2h, the ruling).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bbd7b756-bbee-451e-9446-06019791cb36",
		Name:            "The Pride of Hull Clade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int { return effects.B30TotalToughnessControlled(q.Game, q.Controller) },
				"This spell costs {X} less to cast, where X is the total toughness of creatures you control."),
		},
		Grants: []effects.AbilityGrant{{
			Key: prideOfHullCladeGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("The Pride of Hull Clade — draw cards equal to its toughness",
					func(g *game.Game, item *game.StackItem) error {
						n := effects.ToughnessNowOrLastKnown(g, item.SourceCardID)
						if n <= 0 {
							return nil
						}
						return effects.DrawCards{N: n}.Apply(effects.NewContext(g, item))
					}),
			},
			Text: "Whenever this creature deals combat damage to a player, draw cards equal to its toughness.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}{U}: Until end of turn, target creature you control gets +1/+0, gains \"Whenever this creature deals combat damage to a player, draw cards equal to its toughness,\" and can attack as though it didn't have defender.",
			Cost:    effects.ManaCost("{2}{U}{U}"),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantAbilitiesFor{
					Target: id,
					Keys:   []string{prideOfHullCladeGrant},
					Also:   []game.Mod{game.ModifyPTMod(1, 0), game.WaiveDefenderMod()},
					Label:  "The Pride of Hull Clade — +1/+0, draws cards equal to its toughness on combat damage, can attack as though it didn't have defender",
				}.Apply(ctx)
			},
		}},
	})
}
