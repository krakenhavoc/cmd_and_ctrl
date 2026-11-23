package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shambling Vent — Land:
//
//	"This land enters tapped.
//	 {T}: Add {W} or {B}.
//	 {1}{W}{B}: Until end of turn, this land becomes a 2/3 white and black
//	 Elemental creature with lifelink. It's still a land."
//
// Restless Vents' manland shape: one continuous animation at one timestamp
// (creature type and Elemental subtype in layer 4, colours in layer 5, the
// keyword in layer 6, base 2/3 in layer 7b), pinned to this object until
// end of turn (CR 611.2c). The keyword is part of the animation, not a
// printed keyword of the land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4725abd4-06bc-464a-bc5b-c9e0f71ec079",
		Name:         "Shambling Vent",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|B}",
			Label:    "Add {W} or {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}{B}: Until end of turn, this land becomes a 2/3 white and black Elemental creature with lifelink. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{W}{B}"),
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
						game.SetColorsMod(ctx.TextColors("W", "B")...),
						game.AddKeywordsMod("lifelink"),
					}, game.SetBasePTMods(2, 3)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Shambling Vent — a 2/3 Elemental creature with lifelink until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
