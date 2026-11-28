package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forbidden Lore — Enchantment — Aura {2}{G} (Ice Age, S82 OLD #5636):
//
//	"Enchant land
//	 Enchanted land has "{T}: Target creature gets +2/+1 until end of turn.""
//
// Pursuit of Flight's grant, with a tap cost and a target.
//
// No simplification.
const forbiddenLoreGrant = "forbidden-lore/pump"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2853bad-a751-45ee-8269-24b0cdbc5c6b",
		Name:         "Forbidden Lore",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: forbiddenLoreGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Target creature gets +2/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostFirstTargetB830("Forbidden Lore", 2, 1),
			}},
			Text: "{T}: Target creature gets +2/+1 until end of turn.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(forbiddenLoreGrant)},
	})
}
