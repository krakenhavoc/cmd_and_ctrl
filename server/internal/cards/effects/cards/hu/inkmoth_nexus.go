package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inkmoth Nexus — Land:
//
//	"{T}: Add {C}.
//	 {1}: This land becomes a 1/1 Phyrexian Blinkmoth artifact creature with
//	 flying and infect until end of turn. It's still a land."
//
// Mutavault's manland shape: the Artifact and Creature types and the
// Phyrexian Blinkmoth subtypes (layer 4), flying and infect (layer 6) and a
// 1/1 base (layer 7b) until end of turn. It stays colourless.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "675281ff-b81f-4e5e-9f85-9f8cd202b50b",
		Name:         "Inkmoth Nexus",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: This land becomes a 1/1 Phyrexian Blinkmoth artifact creature with flying and infect until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Artifact", "Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Phyrexian", "Blinkmoth")...),
						game.AddKeywordsMod("flying", "infect"),
					}, game.SetBasePTMods(1, 1)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Inkmoth Nexus — a 1/1 Phyrexian Blinkmoth artifact creature with flying and infect until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
