package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oath of Gideon — Legendary Enchantment {2}{W}:
//
//	"When Oath of Gideon enters, create two 1/1 white Kor Ally creature
//	 tokens.
//	 Each planeswalker you control enters with an additional loyalty
//	 counter on it."
//
// The enters trigger is an ordinary stack trigger making two plain
// tokens.
//
// "Enters with an additional loyalty counter" is a CR 614.1c
// replacement on how the walker enters, and it lands on one of two
// events. A walker entering the ordinary way gets its starting loyalty
// from the CR 306.5b stamp (ReplacementEvent.StartingLoyalty, the event
// Compleated also reads), and the Oath adds one to that placement. It
// can't ride the entry's own "enters with" counters there, because the
// stamp skips a walker that already has loyalty counters and the walker
// would arrive with one. A walker that some other effect puts onto the
// battlefield WITH loyalty counters skips the stamp, so for that entry
// the Oath adds its counter to the entry's counters instead. The two
// cannot both apply to one entry. Either way the counter is the
// walker's controller's ("you control" is the controller it enters
// under, the move's actor), a counter replacement (Doubling Season)
// sees it as part of the same placement, and an Oath that has lost its
// abilities adds nothing.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "ea168acc-22a1-4ba5-affd-fb813b93535a",
		Name:         "Oath of Gideon",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Oath of Gideon — create two 1/1 white Kor Ally creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 white Kor Ally"), N: 2})),
		},
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventCounterPlaced, game.EventZoneMove},
			AppliesTo: oathOfGideonApplies,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				if ev.Kind == game.RepEventMove {
					ev.AddCounterAtETB(game.CounterLoyalty, 1)
					return nil
				}
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label:    "Oath of Gideon — an additional loyalty counter",
			RowLabel: "Each planeswalker you control enters with an additional loyalty counter on it.",
		}},
	})
}

// oathOfGideonApplies picks the event the walker's entering loyalty
// rides: the starting-loyalty stamp for an ordinary entry, or the
// entry's own "enters with" counters when another effect already gives
// it loyalty counters (and the stamp is then skipped).
func oathOfGideonApplies(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if src == nil {
		return false
	}
	switch ev.Kind {
	case game.RepEventCounter:
		if !ev.StartingLoyalty || ev.CounterName != game.CounterLoyalty || ev.CounterDelta <= 0 {
			return false
		}
		walker, ok := g.LookupCardForEffect(ev.CounterTarget)
		return ok && walker.IsPlaneswalker() && walker.Controller == src.Controller
	case game.RepEventMove:
		if ev.NewZone != game.ZoneBattlefield || ev.EntersWithCounters[game.CounterLoyalty] <= 0 || ev.Actor != src.Controller {
			return false
		}
		entering, ok := g.LookupCardForEffect(ev.CardID)
		return ok && entering.IsPlaneswalker()
	}
	return false
}
