package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Favor of Jukai — Enchantment — Aura {3}{G} (Neon Dynasty, #5493):
//
//	"Enchant artifact or creature
//	 As long as enchanted permanent is a creature, it gets +3/+3 and has
//	 reach.
//	 Channel — {1}{G}, Discard this card: Target creature gets +3/+3 and
//	 gains reach until end of turn."
//
// Careful Cultivation's shape: the enchant clause is the spec the CR 704.5m
// check reruns, so the Aura stays on an artifact that stops being a
// creature, and the pump and reach share one condition read on the host
// (cmdr2B19AttachedIsCreature). Channel works from the hand only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08af6fdc-4068-4652-b1fc-cda5309c7b71",
		Name:         "Favor of Jukai",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(effects.Cmdr2B19AttachedIsCreature, 3, 3),
			effects.B43GrantToAttachedWhen(effects.Cmdr2B19AttachedIsCreature, "reach"),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "Channel — {1}{G}, Discard this card: Target creature gets +3/+3 and gains reach until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(game.TargetPurpose{
				Slot: 0, UntilEndOfTurn: game.TargetBenefit,
				Pump: &game.Pump{Power: 3, Toughness: 3, Keywords: []string{"reach"}},
			})},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Favor of Jukai", 3, 3, "reach"),
		}},
	})
}
