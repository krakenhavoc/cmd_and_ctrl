package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "57b86d5c-3269-44bc-a838-3c5439d820d9",
		Name:         "Mirrorpool",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{C}, {T}, Sacrifice this land: Copy target instant or sorcery spell you control. You may choose new targets for the copy.",
				Cost:    Plus(ManaCost("{2}{C}"), TapCost(), SacrificeThis()),
				Targets: instantOrSorcerySpell("target instant or sorcery spell you control", YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return CopySpell{StackID: id, Controller: ctx.Controller(), ChooseNewTargets: true}.Apply(ctx)
				},
			},
			{
				Label:   "{4}{C}, {T}, Sacrifice this land: Create a token that's a copy of target creature you control.",
				Cost:    Plus(ManaCost("{4}{C}"), TapCost(), SacrificeThis()),
				Targets: TargetCreature("target creature you control", YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return TokenCopyOfSingleTarget(item, NewContext(g, item))
				},
			},
		},
	})
}
