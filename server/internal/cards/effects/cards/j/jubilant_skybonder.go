package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const jubilantSkybonderGrant = "jubilant-skybonder/target-tax"

// Jubilant Skybonder — Creature — Human Wizard {1}{W/U}{W/U}, 2/2 (Starter
// Commander Decks):
//
//	"Flying
//	 Creatures you control with flying have "Spells your opponents cast that
//	 target this creature cost {2} more to cast.""
//
// A granted bundle (ADR 0093) carrying a cost increase, so "this creature" is
// the recipient (the modifier's source, as Firion's granted equip discount
// is). It reads the announced targets (CR 601.2f), applies to opponents'
// spells only, and is gone the moment a creature loses flying or the
// Skybonder leaves. The Skybonder itself has flying, so it grants it to itself.
//
// No simplification.
func init() {
	tax := effects.CostsMore(2,
		"Spells your opponents cast that target this creature cost {2} more to cast.",
		effects.OpponentsSpell(), effects.ThatTargetThisPermanent())
	tax.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:        "aab55af7-8084-47b6-910f-75a75a6313a4",
		Name:            "Jubilant Skybonder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []effects.AbilityGrant{{
			Key:           jubilantSkybonderGrant,
			CostModifiers: []game.CostModifier{tax},
			Text:          "Spells your opponents cast that target this creature cost {2} more to cast.",
		}},
		Static: []game.StaticAbility{{
			Label: "Creatures you control with flying have \"Spells your opponents cast that target this creature cost {2} more to cast.\"",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && game.HasKeyword(target, "flying")
			},
			GrantAbilities: []string{jubilantSkybonderGrant},
		}},
	})
}
