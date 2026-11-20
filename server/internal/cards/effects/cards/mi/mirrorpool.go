package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirrorpool — Land:
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {2}{C}, {T}, Sacrifice this land: Copy target instant or sorcery spell
//	 you control. You may choose new targets for the copy.
//	 {4}{C}, {T}, Sacrifice this land: Create a token that's a copy of
//	 target creature you control."
//
// Choreographed Sparks' spell copy (CopySpell with the CR 707.10c retarget
// prompt) and Cackling Counterpart's token copy, each behind a
// tap-and-sacrifice cost with a {C} (colourless-only) mana symbol.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57b86d5c-3269-44bc-a838-3c5439d820d9",
		Name:         "Mirrorpool",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{C}, {T}, Sacrifice this land: Copy target instant or sorcery spell you control. You may choose new targets for the copy.",
				Cost:    effects.Plus(effects.ManaCost("{2}{C}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetInstantOrSorcerySpell("target instant or sorcery spell you control", effects.YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.CopySpell{StackID: id, Controller: ctx.Controller(), ChooseNewTargets: true}.Apply(ctx)
				},
			},
			{
				Label:   "{4}{C}, {T}, Sacrifice this land: Create a token that's a copy of target creature you control.",
				Cost:    effects.Plus(effects.ManaCost("{4}{C}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.TokenCopyOfSingleTarget(item, effects.NewContext(g, item))
				},
			},
		},
	})
}
