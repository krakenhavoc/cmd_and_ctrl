package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fractured Identity — Sorcery {3}{W}{U} (EDHREC rank 4507):
//
//	"Exile target nonland permanent. Each player other than its
//	 controller creates a token that's a copy of it."
//
// Five mana that answers any nonland permanent and hands a copy of it
// to everyone else at the table. In a four-player game that is: their
// best thing gone, and three copies of it distributed — including one
// to YOU. It is the quintessential multiplayer political card:
// removal that makes two other players happy about it.
//
// Two halves, in the printed order:
//
//   - The exile is unconditional and answers indestructible,
//     regeneration and death triggers alike.
//   - "EACH PLAYER OTHER THAN ITS CONTROLLER" is read off the
//     permanent BEFORE it is exiled — that is the only place the
//     information still exists — and it is every other player still
//     in the game, the caster included when the caster is not the
//     permanent's controller. Pointing it at your own permanent hands
//     copies to all three opponents and none to you, which is
//     printed and is occasionally the correct play.
//
// The copies are made from the exiled card, so they are copies of the
// PRINTED object (CR 707.2): counters, Auras, Equipment and
// "becomes a copy of" effects do not carry over, and a token copy of
// a legendary permanent is still legendary and still subject to the
// legend rule for whoever got it.
//
// A target that left in response fizzles the whole spell, copies
// included (CR 608.2b). The copies are only made if the exile
// actually landed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1515b0c2-1b55-4cd4-ad81-fb6b1f3e8188",
		Name:         "Fractured Identity",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(item.Targets[0].ID)
			if !ok {
				return nil
			}
			owner := victim.Controller
			copyOf := victim.InstanceID
			return ExileTarget{
				Target: copyOf,
				Then: func(ctx *Context, exiled bool) error {
					if !exiled {
						return nil
					}
					for _, p := range apnapPlayers(ctx.Game) {
						if p == owner {
							continue
						}
						if err := (CreateTokenCopy{Controller: p, Copy: copyOf, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			}.Apply(ctx)
		},
	})
}
