package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Applied Geometry — Sorcery {2}{G}{U} (Secrets of Strixhaven, #4456):
//
//	"Create a token that's a copy of target non-Aura permanent you control,
//	 except it's a 0/0 Fractal creature in addition to its other types.
//	 Put six +1/+1 counters on it."
//
// The copy is the copiable values of the target as it is on resolution
// (CR 707.2) with the exception applied to them (CR 707.9): base power
// and toughness 0/0 (a printed 0, not a `*` stand-in, as eternalize's
// exception sets), and the creature card type and the Fractal subtype
// added to its own. The six counters are then PUT on the token after it
// exists, so a copy that enters through a replacement or waits on a
// question still gets them. If the target is no longer legal the spell
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88fc179c-055b-47c2-b830-1fd2b1377147",
		Name:         "Applied Geometry",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target non-Aura permanent you control", effects.Not(effects.Subtype("Aura")), effects.YouControl()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(appliedGeometryCopy),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.SocB19PutCountersOnNewTokens(ctx, appliedGeometryKey, 6)
			}),
		),
	})
}

const appliedGeometryKey = "applied-geometry/copy"

// appliedGeometryCopy creates the token copy of the target.
func appliedGeometryCopy(ctx *effects.Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		fractal := ctx.TextSubtype("Fractal")
		ctx.RecordStep(appliedGeometryKey, nil, int(effects.B25LastEventSeq(ctx.Game)))
		return effects.CreateTokenCopy{
			Controller: ctx.Controller(),
			Copy:       t.ID,
			N:          1,
			Except: func(tok *game.Card) {
				tok.Power, tok.Toughness = 0, 0
				tok.VariableToughness = false
				tok.PrintedPTKnown = true
				tok.TypeLine = effects.SocB19WithCreatureAndSubtypeInAddition(tok.TypeLine, fractal)
			},
		}.Apply(ctx)
	}
	return nil
}
