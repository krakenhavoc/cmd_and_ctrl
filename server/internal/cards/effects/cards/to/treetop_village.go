package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treetop Village — Land (Avatar: The Last Airbender Eternal, #5075):
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 {1}{G}: This land becomes a 3/3 green Ape creature with trample
//	 until end of turn. It's still a land."
//
// Creeping Tar Pit's manland shape: one continuous effect pinned to this
// object, the creature type and Ape (layer 4), green (layer 5), trample
// (layer 6) and the base 3/3 (layer 7b), so a +1/+1 counter or a pump
// still applies on top (the rulings). Summoning sickness is the
// engine's (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b53f216d-1592-4eee-b204-502a805fbc8c",
		Name:         "Treetop Village",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: This land becomes a 3/3 green Ape creature with trample until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{1}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Ape")...),
						game.SetColorsMod(ctx.TextColors("G")...),
						game.AddKeywordsMod("trample"),
					}, game.SetBasePTMods(3, 3)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Treetop Village — a 3/3 green Ape creature with trample until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
