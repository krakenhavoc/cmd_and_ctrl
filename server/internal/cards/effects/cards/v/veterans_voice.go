package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Veteran's Voice — Enchantment — Aura {R} (Alliances, #5634):
//
//	"Enchant creature you control
//	 Tap enchanted creature: Target creature other than the creature tapped this
//	 way gets +2/+1 until end of turn. Activate only if enchanted creature is
//	 untapped."
//
// The Aura's ability, with the cost "Tap enchanted creature" (a tap-others cost
// on the permanent the Aura is attached to, not the {T} symbol: no summoning
// sickness, CR 302.6) and a target clause that leaves that same creature out
// (OldB816OtherThanEnchantedCreature, "the creature tapped this way"). The cost
// can only be paid with an untapped creature, which is the printed
// restriction.
//
// No simplification.
func init() {
	const name = "Veteran's Voice"
	effects.Register(effects.Spec{
		OracleID:     "68d25ba4-d710-4090-bceb-5cfbd6bb63a1",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap enchanted creature: Target creature other than the creature tapped this way gets +2/+1 until end of turn. Activate only if enchanted creature is untapped.",
			Purpose: veteransVoicePurpose(),
			Cost:    effects.OldB816TapEnchantedCreature(),
			Condition: func(g *game.Game, _, source uuid.UUID) bool {
				return effects.OldB816EnchantedCreatureIsUntapped(g, source)
			},
			Targets: effects.OldB816OtherThanEnchantedCreature(effects.TargetCreature("target creature other than the creature tapped this way")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 1, Label: name + " — +2/+1 until end of turn"}.Apply(ctx)
			},
		}},
	})
}

// veteransVoicePurpose declares the ability's target as a benefit of +2/+1
// until end of turn (#4302), answering with a pump.
func veteransVoicePurpose() game.Purpose {
	p := effects.ForTargets(effects.BenefitUntilEOT(0, 2, 1))
	p.Answers = game.AnswerPump
	return p
}
