package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Eagles Are Coming! — Instant {1}{W} (The Hobbit):
//
//	"Kicker {2}{W}{W}
//	 Choose target creature you own. If this spell was kicked, instead choose
//	 any number of target creatures you own. Return each chosen creature to
//	 your hand. At the beginning of the next upkeep, create a 4/4 white Bird
//	 Soldier creature token with flying for each creature returned to your
//	 hand this way."
//
// The kicked clause is a different clause (#1716, WhenPaid, as Galadriel's
// Dismissal swaps a creature for a player): a kicked cast names any number of
// creatures you OWN (none is allowed, CR 601.2c), an unkicked one names
// exactly one. "You own" is ownership, not control, so a creature of yours an
// opponent has stolen can be chosen.
//
// The creatures go back in ONE simultaneous bounce (BounceCardsToHandKeyedForEffect),
// and the count is the ones the move really settled in a hand: a creature
// that left in response, was bounced by a replacement somewhere else, or was a
// commander taken to the command zone is not "returned to your hand this way"
// and earns no Bird. The delayed trigger is scheduled for the next upkeep of
// any player's turn, controlled by the caster, and carries the count as data.
// Nothing is scheduled for a count of zero.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "169a4a77-12ff-4997-9550-bdcd456aa2f7",
		Name:         "The Eagles Are Coming!",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{WhenPaid(Kicker("{2}{W}{W}"),
			TargetCreature("any number of target creatures you own", YouOwn()).WithCount(0, 0))},
		Targets: TargetCreature("target creature you own", YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ids := legalTargetCards(item, ctx.Game)
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.BounceCardsToHandKeyedForEffect(ids, ctx.ItemTail(eaglesAreComingBirds))
		},
	})
}

// eaglesAreComingBirds is "at the beginning of the next upkeep, create a Bird
// for each creature returned to your hand this way", scheduled once the
// bounce has settled. Keyed (#4253 slice 10a): an on-disk identity.
var eaglesAreComingBirds = CardTail("tail/eagles-are-coming-birds", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:     game.StepUpkeep,
		Label:  "The Eagles Are Coming! — create a 4/4 white Bird Soldier creature token with flying for each creature returned to your hand",
		Body:   eaglesAreComingBirdsBody,
		Params: game.EffectParams{Amount: len(o.IDs)},
	}.Apply(ctx)
})

// The delayed trigger's body: that many 4/4 flying Birds under the
// trigger's controller.
var eaglesAreComingBirdsBody = game.DelayedBody("eagles-are-coming/birds",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if p.Amount <= 0 {
			return nil
		}
		return CreateToken{Controller: item.Controller, Template: TokenCard("4/4 white Bird Soldier with flying"), N: p.Amount}.Apply(NewContext(g, item))
	})
