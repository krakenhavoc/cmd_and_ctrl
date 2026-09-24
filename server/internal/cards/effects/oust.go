package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oust — Sorcery {W}:
//
//	"Put target creature into its owner's library second from the top.
//	 Its controller gains 3 life."
//
// One-mana removal that costs its victim a draw step: the creature is
// the owner's draw after next, not gone. The position is the whole
// card, and it is PutIntoLibrary's Depth (ADR 0088 Decision 4) — the
// tuck route carries it, so a commander whose owner declines the
// command zone still lands second from the top.
//
// "Its controller" is read before the move: the creature's controller
// as it last existed on the battlefield, which is not always its
// owner — an Ousted stolen creature pays the thief. The life is gained
// whether or not the creature reached the library (a commander that
// took the command zone): the second sentence does not depend on the
// first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "efc12fda-054b-466a-a863-06cf54878172",
		Name:         "Oust",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			targetID := item.Targets[0].ID
			card, ok := ctx.Game.LookupCardForEffect(targetID)
			if !ok {
				return nil
			}
			controller := card.Controller
			source := ctx.Source()
			return PutIntoLibrary{
				Card:  targetID,
				Depth: 2,
				Then: func(g *game.Game, _ bool) error {
					return g.ChangePlayerLifeForEffect(source, controller, 3)
				},
			}.Apply(ctx)
		},
	})
}
