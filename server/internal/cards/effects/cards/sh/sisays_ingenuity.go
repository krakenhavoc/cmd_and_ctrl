package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sisay's Ingenuity — Enchantment — Aura {U} (Planeshift, #5615):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature has "{2}{U}: Target creature becomes the color of your
//	 choice until end of turn.""
//
// The draw is the Aura controller's. The ability is Presence of Gond's shape:
// granted to the creature, so its controller activates it and names the
// colour (Illusion's choice, set in layer 5), and it is lost with the Aura.
//
// No simplification.
func init() {
	const grant = "sisays-ingenuity/becomes-a-color"
	const text = "{2}{U}: Target creature becomes the color of your choice until end of turn."
	effects.Register(effects.Spec{
		OracleID:     "3a487f20-6b1d-4be3-b51f-0c8f0a611601",
		Name:         "Sisay's Ingenuity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Cost:    effects.ManaCost("{2}{U}"),
				Targets: effects.TargetCreature("target creature"),
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Effect:  effects.S82OldB636TargetBecomesAColorOfYourChoice("Sisay's Ingenuity", sisaysIngenuityColor),
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sisay's Ingenuity — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

// sisaysIngenuityColor sets the chosen colour on the target until end of turn.
// An on-disk key: never renamed, never reused.
var sisaysIngenuityColor = effects.CardTail("s82-old-b636/sisays-ingenuity-color", func(ctx *effects.Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" {
		return nil
	}
	return effects.Gap2B05TargetBecomesColor(ctx, []string{color}, "Sisay's Ingenuity — "+game.ColorName(color)+" until end of turn")
})
