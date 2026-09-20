package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain Howler, Sea Scourge — Legendary Creature — Shark Pirate
// {2}{U}{R}, 5/4 (issue #1112):
//
//	"Ward—{2}, Pay 2 life.
//	 Whenever you discard one or more cards, target creature gets
//	 +2/+0 until end of turn for each card discarded this way.
//	 Whenever that creature deals combat damage to a player this
//	 turn, you draw a card."
//
// Three declared gaps, all weaker than printed and none stronger —
// see docs/engine-seams.md, two new rows this card adds:
//
//   - WARD IS A COMPOSITE COST. `effects.Ward` charges exactly one of
//     mana, life or a sacrifice (ward.go's WardCost.validate) — CR
//     702.21a's printed shapes are each single-component, and "{2},
//     pay 2 life" is a composite this card is the first to need.
//     ward.go's own comment already says why a hand-rolled mix is
//     refused rather than half-charged ("{1}, pay 1 life" is not a
//     printed ward). Left off entirely: the creature is a perfectly
//     legal target with no cost offered, which is weaker than
//     printed and never stronger.
//   - "ONE OR MORE … THAT MANY" HAS NO BATCH COUNT. The engine emits
//     one EventDiscardCard per card, and OncePerBatch (#587) only
//     suppresses the later events of a batch — it does not report how
//     many objects were in it, and nothing else in the engine does
//     either (the Ingenious Artillerist audit under #1112 hit the
//     same wall on the same printed idiom). So "for each card
//     discarded this way" cannot scale: OncePerBatch gives one
//     trigger for the whole discard, correctly avoiding a second
//     target prompt and a doubled pump, but the pump it applies is
//     the SINGLE-card amount regardless of how many were actually
//     discarded. Discarding one card is exact; discarding two or more
//     is weaker than printed, never stronger, which is the direction
//     #259 requires.
//   - THE PAYOFF DELAYED TRIGGER HAS NO SHAPE. "Whenever THAT
//     creature deals combat damage to a player THIS TURN" is a
//     delayed triggered ability that has to re-fire for as long as
//     the turn lasts and stay bound to one specific instance —
//     `DelayedTrigger`'s event-conditioned form (#663) fires once and
//     is removed (double strike would need it twice), and nothing
//     else in the engine registers a repeating, instance-scoped
//     watch. Approximating it with a one-shot delayed trigger would
//     silently drop a second connection in the same turn, so it is
//     left out rather than guessed at, per the batch brief. The pump
//     above ships on its own; the draw does not.
func init() {
	Register(Spec{
		OracleID:     "46c99d67-8307-4af0-afd0-377714806614",
		Name:         "Captain Howler, Sea Scourge",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Ward isn't implemented on this creature — it can be targeted for free.",
			"Discarding more than one card at once still only pumps the chosen creature by +2/+0, not +2/+0 for each card discarded.",
			"The bonus card draw for the pumped creature dealing combat damage to a player isn't implemented.",
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Targeting(On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "Captain Howler, Sea Scourge — target creature gets +2/+0 until end of turn",
				captainHowlerPump), TargetCreature("target creature"))),
		},
	})
}

// captainHowlerPump is the resolution body: +2/+0 until end of turn
// on the target chosen when this instance of the trigger went on the
// stack.
//
// Caller holds g.mu.
func captainHowlerPump(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return BoostUntilEOT{
		Target: item.Targets[0].ID,
		Power:  2,
		Label:  "Captain Howler, Sea Scourge — +2/+0 until end of turn",
	}.Apply(NewContext(g, item))
}
