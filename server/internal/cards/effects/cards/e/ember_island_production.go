package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ember Island Production — Sorcery {3}{U}{U} (Avatar: The Last
// Airbender, #4459):
//
//	"Choose one —
//	 • Create a token that's a copy of target creature you control,
//	   except it's not legendary and it's a 4/4 Hero in addition to its
//	   other types.
//	 • Create a token that's a copy of target creature an opponent
//	   controls, except it's not legendary and it's a 2/2 Coward in
//	   addition to its other types."
//
// Each mode is CreateTokenCopy of its target's copiable values as the
// spell resolves (CR 707.2), with the exceptions applied to them (CR
// 707.9): the legendary supertype struck, base power and toughness set
// to the printed number, and the subtype added beside every type it
// already has. The token is yours either way. Enters abilities of the
// copied creature trigger as the token enters, and an illegal target
// makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b786547-67ed-49fa-9981-35c4c2175f62",
		Name:         "Ember Island Production",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Create a token that's a copy of target creature you control, except it's not legendary and it's a 4/4 Hero in addition to its other types.",
				effects.TargetCreature("target creature you control", effects.YouControl())),
			effects.Mode("Create a token that's a copy of target creature an opponent controls, except it's not legendary and it's a 2/2 Coward in addition to its other types.",
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(1) {
				return emberIslandCopy(ctx, 2, ctx.TextSubtype("Coward"))
			}
			return emberIslandCopy(ctx, 4, ctx.TextSubtype("Hero"))
		},
	})
}

// emberIslandCopy creates the token copy of the chosen mode's target,
// not legendary, with base power and toughness n/n and the subtype
// added.
func emberIslandCopy(ctx *effects.Context, n int, subtype string) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.CreateTokenCopy{
			Controller: ctx.Controller(),
			Copy:       t.ID,
			N:          1,
			Except: func(tok *game.Card) {
				tok.TypeLine = effects.WithSubtypeInAddition(effects.RemoveLegendaryFromTypeLine(tok.TypeLine), subtype)
				tok.Power, tok.Toughness = n, n
				tok.VariableToughness = false
				tok.PrintedPTKnown = true
			},
		}.Apply(ctx)
	}
	return nil
}
