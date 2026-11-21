package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mind into Matter — Sorcery {X}{G}{U} (Secrets of Strixhaven, #4456):
//
//	"Draw X cards. Then you may put a permanent card with mana value X or
//	 less from your hand onto the battlefield tapped."
//
// Two steps, so the put waits for the draw to finish. The put reads the hand
// as it stands after the draw (the cards just drawn are eligible), over the
// permanent cards in it (CR 110.4 drops an instant or a sorcery) whose mana
// value is at most the X the spell was cast with (CR 107.3f; a land is mana
// value 0). It is a "may", so declining is a real answer, and with no
// eligible card there is no question. The permanent arrives tapped, as it
// enters, so nothing that watches for a tap sees one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f9e4be0-814a-4cab-948b-114769035a4a",
		Name:         "Mind into Matter",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{N: ctx.X()}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.PutFromHandOntoBattlefield{
					Match:    effects.And(effects.Permanent(), effects.ManaValueLE(ctx.X())),
					Tapped:   true,
					Optional: true,
					Label:    "Mind into Matter — you may put a permanent card with mana value X or less from your hand onto the battlefield tapped",
				}.Apply(ctx)
			}),
		),
	})
}
