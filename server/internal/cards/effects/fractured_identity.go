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
// The players create their copies at the same time (CR 101.4, CR 608.2f),
// so every copy enters in one event and each sees the others enter
// (CR 603.6a, #4867): Fractured Identity on a Soul Warden gains each new
// controller a life for every other copy. Each player's creation is still
// their own, and so are the replacement effects applying to its token
// (CR 701.7b): a Doubling Season doubles only its controller's copy
// (game.CreateTokensForPlayersTogetherForEffect).
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
				Target:   copyOf,
				ThenTail: withCarry(ctx.ItemTail(fracturedIdentityExiled), copyOf, owner),
			}.Apply(ctx)
		},
	})
}

// fracturedIdentityExiled is the clause after the exile (#4253 slice 9).
var fracturedIdentityExiled = CardTail("fractured-identity/after-exile", func(ctx *Context, o game.TailOutcome) error {
	copyOf, owner := o.Tail.Carry[0], o.Tail.Carry[1]
	exiled := len(o.IDs) == 1

	if !exiled {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(ctx.Game, copyOf)
	if !ok {
		return nil
	}
	var specs []game.TokenCreation
	for _, p := range apnapPlayers(ctx.Game) {
		if p == owner {
			continue
		}
		specs = append(specs, game.TokenCreation{
			Controller: p, Source: ctx.Source(),
			Groups: []game.TokenGroup{{Template: copiedToken(tmpl), Count: 1}},
		})
	}
	return ctx.Game.CreateTokensForPlayersTogetherForEffect(specs)
})
