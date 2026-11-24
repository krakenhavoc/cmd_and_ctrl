package cardsex

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Excise — Instant {X}{W} (Prophecy, #5617):
//
//	"Exile target attacking creature unless its controller pays {X}."
//
// The creature's controller is asked for X as the spell resolves, and the
// table waits for the answer (#5793, CR 118.12). X is the announced X. With
// X = 0 the payment is {0}, which is always made, so nothing is asked and
// nothing is exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28e1a613-30ad-4013-be0b-b0365f627201",
		Name:         "Excise",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			legal := ctx.LegalTargets()
			if len(legal) == 0 {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(legal[0].ID)
			if !ok {
				return nil
			}
			cost := "{" + strconv.Itoa(x) + "}"
			return effects.PayUnless{
				Chooser:  c.Controller,
				Cost:     cost,
				Question: "Excise — pay " + cost + ", or " + c.Name + " is exiled?",
				Else:     exciseExile,
			}.Apply(ctx)
		},
	})
}

// exciseExile is the decline. Keyed, an on-disk identity.
var exciseExile = effects.PayUnlessElse("pay-unless/excise-exile",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.Targets() {
			if t.Kind == game.TargetCard {
				return effects.ExileTarget{Target: t.ID}.Apply(ctx)
			}
		}
		return nil
	})
