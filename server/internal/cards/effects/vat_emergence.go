package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vat Emergence — Sorcery {4}{B} (Phyrexia: All Will Be One, #5479):
//
//	"Put target creature card from a graveyard onto the battlefield under
//	 your control. Proliferate."
//
// Reanimate's put without the life loss. The proliferate is the entry's keyed
// tail (thenProliferateAgain, one pass), because a returning Clone or a
// creature with an entry question pauses the entry and the proliferate waits
// for it. With its only target gone the spell does not resolve at all
// (CR 608.2b), so there is no proliferate either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03e59169-d8f8-40da-b5f7-645d16365e7b",
		Name:         "Vat Emergence",
		Completeness: CompletenessFull,
		Targets:      targetCreatureInAnyGraveyard(),
		Purpose:      ForTargets(Reanimates(0)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return ctx.Game.ReturnFromGraveyardToBattlefieldKeyedForEffect(
				item.Targets[0].ID, ctx.Controller(), false, ctx.TailN(thenProliferateAgain, 1))
		},
	})
}
