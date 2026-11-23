package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const huntersProwessGrant = "hunters-prowess/draw-that-many"

// Hunter's Prowess — Sorcery {4}{G} (Starter Commander Decks):
//
//	"Until end of turn, target creature gets +3/+3 and gains trample and
//	 "Whenever this creature deals combat damage to a player, draw that many
//	 cards.""
//
// The Pride of Hull Clade's shape: one effect with one timestamp gives the
// pump (layer 7c), trample and the granted trigger (a bundle, ADR 0093) until
// end of turn. The granted trigger draws as many cards as the combat damage
// dealt, its controller drawing, and is gone with the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1a776d2-fb3a-4215-a14e-4b6e9bda95df",
		Name:         "Hunter's Prowess",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose: effects.ForTargets(game.TargetPurpose{
			Slot: 0, UntilEndOfTurn: game.TargetBenefit,
			Pump: &game.Pump{Power: 3, Toughness: 3, Keywords: []string{"trample"}},
		}),
		Grants: []effects.AbilityGrant{{
			Key: huntersProwessGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Hunter's Prowess — draw that many cards",
					func(g *game.Game, item *game.StackItem) error {
						if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
							return nil
						}
						return effects.DrawCards{N: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
					}),
			},
			Text: "Whenever this creature deals combat damage to a player, draw that many cards.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			return effects.GrantAbilitiesFor{
				Target: ts[0].ID,
				Keys:   []string{huntersProwessGrant},
				Also:   []game.Mod{game.ModifyPTMod(3, 3), game.AddKeywordsMod("trample")},
				Label:  "Hunter's Prowess — +3/+3, trample, draws that many cards on combat damage",
			}.Apply(ctx)
		},
	})
}
