package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demonlord Belzenlok — Legendary Creature — Elder Demon {4}{B}{B}, 6/6
// (Commander Masters):
//
//	"Flying, trample
//	 When Demonlord Belzenlok enters, exile cards from the top of your
//	 library until you exile a nonland card, then put that card into your
//	 hand. If the card's mana value is 4 or greater, repeat this process.
//	 Demonlord Belzenlok deals 1 damage to you for each card put into your
//	 hand this way."
//
// Each pass exiles from the top until a nonland card (game.ExileUntilHitsForEffect,
// the cards passed over stay in exile) and puts that card into the hand; a card
// of mana value 4 or more starts another pass, and a library that runs out
// ends the process. The damage is dealt once, by Belzenlok, after the last
// pass, one per card that actually reached the hand (a commander's CR 903.9
// offer to stay in the command zone, when answered "yes", is not a card put
// into the hand).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "667ece0e-a474-4e11-8ac4-6c8fc354457c",
		Name:            "Demonlord Belzenlok",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Demonlord Belzenlok — exile cards from the top of your library until you exile a nonland card, then put it into your hand; repeat if its mana value is 4 or greater; it deals 1 damage to you for each card put into your hand this way",
				demonlordBelzenlokDig),
		},
	})
}

func demonlordBelzenlokDig(g *game.Game, item *game.StackItem) error {
	return demonlordBelzenlokPass(NewContext(g, item), 0)
}

// demonlordBelzenlokPass is one pass of the process: exile until a nonland
// card, then put it into the hand through a keyed move, whose tail decides
// whether to go round again once the move has settled (a CR 614 or CR 903.9
// prompt may sit in between). `taken` is how many cards reached the hand so
// far.
func demonlordBelzenlokPass(ctx *Context, taken int) error {
	g := ctx.Game
	_, hits, err := g.ExileUntilHitsForEffect(ctx.Controller(), 1, func(c game.Card) bool { return !c.IsLand() })
	if err != nil {
		return err
	}
	if len(hits) == 0 {
		return demonlordBelzenlokFinish(ctx, taken)
	}
	repeat := 0
	if card, ok := g.LookupCardForEffect(hits[0]); ok && card.ManaValue() >= 4 {
		repeat = 1
	}
	tail := ctx.Tail(demonlordBelzenlokPut)
	tail.Nums = []int{taken, repeat}
	return g.BounceCardsToHandKeyedForEffect(hits, tail)
}

// demonlordBelzenlokPut counts the card that reached the hand and goes round
// again when its mana value was 4 or greater. An on-disk key: never renamed,
// never reused.
//
// It is assigned in init, not at declaration: the body goes round again
// through demonlordBelzenlokPass, which names the tail, and a package-level
// initialiser may not refer to itself.
var demonlordBelzenlokPut game.TailRef

func init() {
	demonlordBelzenlokPut = CardTail("demonlord-belzenlok/put-into-hand", demonlordBelzenlokPutBody)
}

func demonlordBelzenlokPutBody(ctx *Context, o game.TailOutcome) error {
	taken, repeat := 0, 0
	if len(o.Tail.Nums) == 2 {
		taken, repeat = o.Tail.Nums[0], o.Tail.Nums[1]
	}
	taken += len(o.IDs)
	if repeat == 1 {
		return demonlordBelzenlokPass(ctx, taken)
	}
	return demonlordBelzenlokFinish(ctx, taken)
}

// demonlordBelzenlokFinish is the damage: 1 per card put into the hand.
func demonlordBelzenlokFinish(ctx *Context, taken int) error {
	if taken == 0 {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: taken}.Apply(ctx)
}
