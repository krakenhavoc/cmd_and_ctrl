package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const spawningPoolRegenGrant = "spawning-pool/skeleton-regenerate"

// Spawning Pool — Land:
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 {1}{B}: This land becomes a 1/1 black Skeleton creature with "{B}:
//	 Regenerate this creature" until end of turn. It's still a land."
//
// The animation is a scoped effect until end of turn: it adds the creature
// type and the Skeleton subtype to the land, sets it black and 1/1, and gives
// it the quoted regeneration ability in the same effect (a bundle granted with
// GrantAbilitiesMod, ADR 0093). It stays a land (CR 205.1b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3bf22cf-0a6f-4fb6-ba82-63ce290308d6",
		Name:         "Spawning Pool",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Grants: []effects.AbilityGrant{{
			Key:       spawningPoolRegenGrant,
			Activated: []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{B}")},
			Text:      "{B}: Regenerate this creature.",
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: This land becomes a 1/1 black Skeleton creature with \"{B}: Regenerate this creature\" until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{1}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtype("Skeleton")),
						game.SetColorsMod(ctx.TextColor("B")),
						game.SetBasePowerMod(1),
						game.SetBaseToughnessMod(1),
						game.GrantAbilitiesMod(spawningPoolRegenGrant),
					},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Spawning Pool — a 1/1 black Skeleton creature with regeneration until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
