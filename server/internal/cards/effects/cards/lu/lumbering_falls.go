package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lumbering Falls — Land:
//
//	"This land enters tapped.
//	 {T}: Add {G} or {U}.
//	 {2}{G}{U}: Until end of turn, this land becomes a 3/3 green and blue
//	 Elemental creature with hexproof. It's still a land."
//
// Creeping Tar Pit's manland: one continuous effect pinned to this object
// (CR 611.2c): the creature type and Elemental subtype (layer 4), green and
// blue (layer 5), hexproof (layer 6) and the base 3/3 (layer 7b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2878f4e-3044-4d60-8ec1-b325edfca397",
		Name:         "Lumbering Falls",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|U}",
			Label:    "Add {G} or {U}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}{U}: Until end of turn, this land becomes a 3/3 green and blue Elemental creature with hexproof. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{2}{G}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
						game.SetColorsMod(ctx.TextColors("G", "U")...),
						game.AddKeywordsMod("hexproof"),
					}, game.SetBasePTMods(3, 3)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Lumbering Falls — a 3/3 green and blue Elemental creature with hexproof until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
