package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "08af6fdc-4068-4652-b1fc-cda5309c7b71",
		Name:         "Favor of Jukai",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			b43PumpAttachedWhen(cmdr2B19AttachedIsCreature, 3, 3),
			b43GrantToAttachedWhen(cmdr2B19AttachedIsCreature, "reach"),
		},
		Activated: []ActivatedAbility{{
			Label:   "Channel — {1}{G}, Discard this card: Target creature gets +3/+3 and gains reach until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{1}{G}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetCreature("target creature"),
			Effect:  neoB07TargetGetsUntilEOT("Favor of Jukai", 3, 3, "reach"),
		}},
	})
}
