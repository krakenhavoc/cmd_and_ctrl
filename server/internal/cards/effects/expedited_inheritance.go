package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Expedited Inheritance — Enchantment {R}{R} (Murders at Karlov Manor;
// deck request #4724):
//
//	"Whenever a creature is dealt damage, its controller may exile that
//	 many cards from the top of their library. They may play those cards
//	 until the end of their next turn."
//
// Repercussion's trigger: EventDealDamage reaching ANY creature on the
// battlefield, yours or an opponent's (b28CreatureWasDealtDamage). The
// engine opens one damage event per source, so the trigger waits for the
// damage to finish (AtBatchEnd, ThatMuchDamage) and reads the TOTAL the
// creature was dealt at the same time: a creature blocked by two
// creatures triggers once, and Blasphemous Act triggers once per
// creature (CR 120.3, CR 510.2). The amount is the damage actually dealt,
// so prevented damage does not count.
//
// The trigger is the enchantment controller's, but the choice is not:
// "its controller" is the damaged creature's controller as the ability
// resolves, or as it last existed if it has left the battlefield, which a
// creature dealt lethal damage usually has (CR 608.2h). The creature is
// pinned as the object it was when dealt the damage, so one that died and
// came back is not read as the same creature. That player is asked the
// "may" (MayChoice) and, on a yes, exiles that many cards from the top of
// their OWN library and may play them, lands included, until the end of
// THEIR next turn (UntilEndOfYourNextTurnDuration stamped against them).
// A player who has left the game is asked nothing.
//
// No simplification.
func init() {
	const label = "Expedited Inheritance — its controller may exile that many cards and play them until the end of their next turn"
	Register(Spec{
		OracleID:     "11e2f581-7c3c-4f2f-877f-e41540ab967e",
		Name:         "Expedited Inheritance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b28CreatureWasDealtDamage(ev, g)
				return ok
			},
			Key:        label,
			AtBatchEnd: ThatMuchDamage(),
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				c, ok := b28CreatureWasDealtDamage(ev, g)
				if !ok {
					return nil
				}
				item := game.NewTriggeredItem(source, label)
				item.Params.Object = game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}
				item.Params.Player = c.Controller
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: expeditedInheritanceAsk,
		}},
	})
}

// expeditedInheritanceAsk asks the damaged creature's controller whether
// to exile that many cards.
func expeditedInheritanceAsk(g *game.Game, item *game.StackItem) error {
	n := item.Params.Amount
	if n <= 0 {
		return nil
	}
	who := expeditedInheritanceCreaturesController(g, item)
	if p := g.PlayerByIDForEffect(who); p == nil || p.Eliminated {
		return nil
	}
	word := strconv.Itoa(n)
	return MayChoice{
		Player:   who,
		Question: "Expedited Inheritance — exile the top " + word + " cards of your library? You may play them until the end of your next turn.",
		YesLabel: "Exile " + word,
		NoLabel:  "Don't exile",
		Word:     word,
		Then:     expeditedInheritanceExile,
	}.Apply(NewContext(g, item))
}

// expeditedInheritanceCreaturesController is "its controller": the
// creature as it is now while it is still that object on the
// battlefield, and as it last existed once it has left (CR 608.2h). The
// controller it had when the trigger was built is the fallback for an
// object the game no longer remembers.
func expeditedInheritanceCreaturesController(g *game.Game, item *game.StackItem) uuid.UUID {
	if info, ok := g.PermanentForEffect(item.Params.Object); ok && info.Controller != uuid.Nil {
		return info.Controller
	}
	return item.Params.Player
}

// expeditedInheritanceExile is the yes: the chooser exiles that many
// cards from the top of their library and may play them until the end of
// their next turn.
var expeditedInheritanceExile = MayYes("confirm/expedited-inheritance-exile", func(ctx *Context, r game.ConfirmAnswer) error {
	n := wordInt(r)
	who := r.Chooser
	if n <= 0 || who == uuid.Nil {
		return nil
	}
	_, err := ctx.Game.ExileTopWithPermissionForEffect(who, who, n, game.CastPermission{
		Duration: ctx.Game.UntilEndOfYourNextTurnDuration(who),
	})
	return err
})
