package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phoenix Chick — Creature — Phoenix {R}, 1/1 (Dominaria United):
//
//	"Flying, haste
//	 This creature can't block.
//	 Whenever you attack with three or more creatures, you may pay {R}{R}. If
//	 you do, return this card from your graveyard to the battlefield tapped
//	 and attacking with a +1/+1 counter on it."
//
// Armasaur Guide's count (the EventAttack events of one declaration by you,
// settled as a batch, one trigger when there are at least three) from the
// GRAVEYARD only (Flamewake Phoenix's InGraveyard, CR 113.6). The Chick is
// never among the attackers, so it counts none of the three. "You may pay"
// is MayPay: the return runs only when the mana was really paid, and a card
// that has left the graveyard by then (CR 400.7) stays where it is. It
// enters through the CR 506.3c door tapped and attacking whatever its
// controller picks, with the counter as an entry counter, and was never
// declared as an attacker, so nothing that watches attacks sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "40bb7b6b-a75d-4b59-ba45-a5801dcd59a1",
		Name:            "Phoenix Chick",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			withAtBatchEnd(InGraveyard(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && attackDeclaredByYou(ev, source.Controller)
			}, "Phoenix Chick — you may pay {R}{R}: return it from your graveyard to the battlefield tapped and attacking with a +1/+1 counter on it (you attacked with three or more creatures)",
				phoenixChickMayReturn)),
				ThatManyEventsAtLeast(3)),
		},
	})
}

// phoenixChickMayReturn is the offer, then the put.
func phoenixChickMayReturn(g *game.Game, item *game.StackItem) error {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return MayPay{
		Chooser:    item.Controller,
		Cost:       "{R}{R}",
		Question:   "Phoenix Chick — pay {R}{R}: return it from your graveyard to the battlefield tapped and attacking with a +1/+1 counter?",
		InThisStep: true,
		Then:       phoenixChickPaid,
	}.Apply(NewContext(g, item))
}

// phoenixChickPaid is the paid half, keyed. An on-disk key.
var phoenixChickPaid = PayUnlessBranches("pay-unless/phoenix-chick-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return phoenixChickReturn(ctx, ctx.Source())
	}, nil)
