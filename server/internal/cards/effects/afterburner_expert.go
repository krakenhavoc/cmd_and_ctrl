package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Afterburner Expert — Creature — Goblin Artificer {2}{G}, 4/2:
//
//	"Exhaust — {2}{G}{G}: Put two +1/+1 counters on this creature.
//	 (Activate each exhaust ability only once.)
//	 Whenever you activate an exhaust ability, return this card from
//	 your graveyard to the battlefield."
//
// The same watch as Rangers' Refueler (#1184) in the one zone that
// makes it mean something. "Return this card FROM YOUR GRAVEYARD"
// says where the ability functions (CR 113.6): a battlefield copy of
// it would have nothing to return, so `InGraveyard` REPLACES the
// default zone rather than adding to it — #925's rule, and the reason
// the wrapper exists.
//
// Two consequences fall out of that and are worth stating, because
// both look like bugs until you read the card again:
//
//   - the Goblin's OWN exhaust ability cannot wake it up. It is
//     activated from the battlefield, and from the battlefield this
//     trigger does not exist; by the time the card is in the
//     graveyard the ability is unactivatable. It comes back off
//     SOMEBODY ELSE'S exhaust ability — yours, on another permanent.
//   - "you" is the card's OWNER (CR 108.4), not its controller: a
//     card in a graveyard has no controller. The harvest stamps the
//     owner onto the source it hands the predicate, so `ByYou` inside
//     the trigger reads as printed and the returned creature arrives
//     under its owner's control.
//
// The return is MANDATORY — no "you may" — which is the difference
// from Bloodghast beside it, and it is a real difference: a player
// who would rather leave the Goblin in the yard for a later
// reanimation has no say.
//
// A new object arrives (CR 400.7), so the returning creature's
// exhaust ability is unspent whatever the old one did. That is the
// record's key doing the work rather than anything here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4d19e3a-c2f8-4503-bcfa-1b46bde972e5",
		Name:         "Afterburner Expert",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {2}{G}{G}: Put two +1/+1 counters on this creature.",
			Exhaust: true,
			Cost:    ManaCost("{2}{G}{G}"),
			Effect:  plusOneCountersOnThis(2),
		}},
		Triggered: []game.TriggeredAbility{
			InGraveyard(WheneverYouActivateAnExhaustAbility(
				"Afterburner Expert — return it from your graveyard to the battlefield",
				returnThisCardFromYourGraveyard)),
		},
	})
}
