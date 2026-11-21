package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Venser, Visionary Traveler — Legendary Planeswalker — Venser {3}{W}{U},
// loyalty 4 (Mystery Booster Commander, #5948):
//
//	"Each nontoken creature you control that wasn't cast from your hand enters
//	 with two additional +1/+1 counters on it.
//	 +1: Exile up to one other target permanent you control. At the beginning of
//	 the next end step, return that card to the battlefield under its owner's
//	 control.
//	 −2: For each opponent, return up to one target nonland permanent that
//	 player controls to its owner's hand.
//	 Venser, Visionary Traveler can be your commander."
//
// The static is a CR 614 entry replacement over a nontoken creature entering
// under Venser's controller's control that was not cast from a hand (a card
// put onto the battlefield, reanimated, flickered or cast from anywhere else):
// the two counters ride the entry (AddCounterAtETB), so Hardened Scales applies
// to them. The +1 is Phelia's exile with a delayed return to the battlefield at
// the next end step under the owner's control (a new object). The −2 is Age of
// Ultron's per-opponent clauses, one "up to one" pick for each seat after the
// chooser, each bound to its opponent, returned to their owners' hands.
// "Can be your commander" is deck-building, read from the oracle text by the deck
// validator.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f4a64308-5f24-47c0-83f9-ee16a3b4f143",
		Name:         "Venser, Visionary Traveler",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{gapB06NotCastFromHandEnterWithCounters()},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Exile up to one other target permanent you control. At the beginning of the next end step, return that card to the battlefield under its owner's control.",
				Cost:    LoyaltyCost(1),
				Targets: Another(TargetPermanent("up to one other target permanent you control", YouControl()).WithCount(0, 1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return exileTargetsThenScheduleReturn(NewContext(g, item),
						"Venser, Visionary Traveler — return the exiled card to the battlefield under its owner's control")
				},
			},
			{
				Label:   "−2: For each opponent, return up to one target nonland permanent that player controls to its owner's hand.",
				Cost:    LoyaltyCost(-2),
				Targets: perOpponentClauses(game.MaxPlayers-1, "nonland permanent", Nonland()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					seen := map[uuid.UUID]bool{}
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard || seen[t.ID] {
							continue
						}
						seen[t.ID] = true
						if err := (BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}

// gapB06NotCastFromHandEnterWithCounters is Venser's replacement: a nontoken
// creature entering under the source controller's control that was not cast from
// a hand enters with two additional +1/+1 counters. A creature spell that was
// cast from the hand (CastProvenance.FromZone) is left alone; a permanent that
// entered with no cast at all has no provenance, so it gets the counters.
func gapB06NotCastFromHandEnterWithCounters() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || !entering.IsCreature() || entering.IsToken() || entering.Controller != src.Controller {
				return false
			}
			return !g.EntryCastFromHandForEffect(ev)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, 2)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Each nontoken creature you control that wasn't cast from your hand enters with two additional +1/+1 counters on it.",
	}
}
