package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stirring Wildwood — Land:
//
//	"This land enters tapped.
//	 {T}: Add {G} or {W}.
//	 {1}{G}{W}: Until end of turn, this land becomes a 3/4 green and white
//	 Elemental creature with reach. It's still a land."
//
// Restless Vents' manland shape: one continuous animation at one timestamp
// (creature type and Elemental subtype in layer 4, colours in layer 5, the
// keyword in layer 6, base 3/4 in layer 7b), pinned to this object until
// end of turn (CR 611.2c). The keyword is part of the animation, not a
// printed keyword of the land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e900f211-dc8b-40f7-a217-1b1462b15c26",
		Name:         "Stirring Wildwood",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|W}",
			Label:    "Add {G} or {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}{W}: Until end of turn, this land becomes a 3/4 green and white Elemental creature with reach. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{G}{W}"),
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
						game.SetColorsMod(ctx.TextColors("G", "W")...),
						game.AddKeywordsMod("reach"),
					}, game.SetBasePTMods(3, 4)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Stirring Wildwood — a 3/4 Elemental creature with reach until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
