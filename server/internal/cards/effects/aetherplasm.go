package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aetherplasm — Creature — Illusion {2}{U}{U}, 1/1 (Ravnica Remastered,
// #5948):
//
//	"Whenever this creature blocks a creature, you may return this creature to
//	 its owner's hand. If you do, you may put a creature card from your hand
//	 onto the battlefield blocking that creature."
//
// One trigger per attacker blocked (CR 509.3b, b28SelfBlocked), and the "you
// may" is asked as it resolves (CR 603.5). The return is BounceToHand's, whose
// tail is told whether the card reached a hand: "if you do" is that, and a
// Aetherplasm that has left or come back as a new object (CR 400.7) returns
// nothing and puts nothing. The creature card is then chosen from the hand as
// it is NOW, the Aetherplasm itself among the candidates, and put onto the
// battlefield (not cast) blocking the attacker (CR 506.3d): it was never
// declared as a blocker (CR 509.4), so nothing that watches blocks sees it,
// and a creature put in blocks only if the attacker is still attacking
// (PutOnBattlefieldBlockingForEffect).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0cd99210-8d34-4112-8e41-7e8c5d2a95b9",
		Name:         "Aetherplasm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b28SelfBlocked(ev, source)
			}, "Aetherplasm — return it to its owner's hand; if you do, you may put a creature card from your hand onto the battlefield blocking that creature",
				gapB75AetherplasmBounce),
				"Aetherplasm — return it to its owner's hand?"),
		},
	})
}

// gapB75AetherplasmBounce returns the Aetherplasm to its owner's hand and
// hands the attacker it blocked to the tail that follows.
func gapB75AetherplasmBounce(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	tail := ctx.ItemTail(gapB75AetherplasmReturned)
	tail.Carry = []uuid.UUID{item.Trigger.Event.Target}
	return BounceToHand{Target: item.SourceCardID, ThenTail: tail}.Apply(ctx)
}

// gapB75AetherplasmReturned is "If you do, you may put a creature card from
// your hand onto the battlefield blocking that creature": told the card when it
// reached a hand and nothing when it did not. An on-disk key.
var gapB75AetherplasmReturned = CardTail("gap-b75/aetherplasm-returned", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	candidates := handCardsMatching(ctx.Game, ctx.Controller(), Creature())
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Destination: game.ChooseOntoBattlefield,
		Chooser:     ctx.Controller(),
		Source:      ctx.Source(),
		Question:    "Aetherplasm — you may put a creature card from your hand onto the battlefield blocking that creature",
		Cards:       candidates,
		Min:         0,
		Max:         1,
		ThenKey:     gapB75AetherplasmPut,
		LiveZone:    game.ZoneHand,
		Carry:       []uuid.UUID{o.Tail.Carry[0]},
		Asked:       game.NewAskedItem(ctx.Item),
	})
	return nil
})

// gapB75AetherplasmPut puts the chosen card onto the battlefield and, once it
// is there, blocking the attacker the Aetherplasm blocked. Keys
// "choose-cards/aetherplasm-put" and "gap-b75/aetherplasm-blocks".
var gapB75AetherplasmPut = game.RegisterChooseCardsThen("choose-cards/aetherplasm-put", func(g *game.Game, r game.CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 || len(r.Carry) == 0 || r.Asked == nil {
		return nil
	}
	card := r.Picked[0]
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneHand {
		return nil
	}
	tail := NewContext(g, r.Asked).ItemTail(gapB75AetherplasmBlocks)
	tail.Carry = []uuid.UUID{r.Carry[0]}
	return g.PutFromHandOntoBattlefieldKeyedForEffect(card, game.HandEntryOptions{Controller: r.Asked.Controller}, tail)
})

// gapB75AetherplasmBlocks makes the creature that entered a blocker of the
// attacker, if that is still attacking. An on-disk key.
var gapB75AetherplasmBlocks = EnteredTail("gap-b75/aetherplasm-blocks", func(ctx *Context, entered uuid.UUID, o game.TailOutcome) error {
	if entered == uuid.Nil || len(o.Tail.Carry) == 0 {
		return nil
	}
	ctx.Game.PutOnBattlefieldBlockingForEffect(entered, o.Tail.Carry[0])
	return nil
})
