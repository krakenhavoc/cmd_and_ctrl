package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restless Anchorage — Land:
//
//	"This land enters tapped.
//	 {T}: Add {W} or {U}.
//	 {1}{W}{U}: Until end of turn, this land becomes a 2/3 white and blue
//	 Bird creature with flying. It's still a land.
//	 Whenever this land attacks, create a Map token."
//
// A manland built like Raging Ravine: one continuous effect from a
// resolving ability, pinned to this object, applied in each layer it
// touches (type and subtype in 4, colours in 5, flying in 6, base P/T
// in 7b). The attack trigger is the land's own printed ability, not
// part of the animation, and it can only fire once the land is an
// attacking creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91320daf-f69c-4350-b0fc-4bb37a6904b1",
		Name:         "Restless Anchorage",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U}",
			Label:    "Add {W} or {U}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Restless Anchorage — create a Map token",
				effects.Do(effects.CreateToken{Template: effects.MapToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}{U}: Until end of turn, this land becomes a 2/3 white and blue Bird creature with flying. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{1}{W}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Bird")...),
						game.SetColorsMod(g.TextReaderForEffect(item).TextColors("W", "U")...),
						game.SetBasePowerMod(2),
						game.SetBaseToughnessMod(3),
						game.AddKeywordsMod("flying"),
					},
					Duration: effects.DurationUntilEndOfTurn(effects.NewContext(g, item)),
					Label:    "Restless Anchorage — a 2/3 white and blue Bird creature with flying until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
