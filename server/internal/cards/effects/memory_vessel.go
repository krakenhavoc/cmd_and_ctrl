package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Memory Vessel — Artifact {3}{R}{R} (#2559):
//
//	"{T}, Exile this artifact: Each player exiles the top seven cards of
//	 their library. Until your next turn, players may play cards they
//	 exiled this way, and they can't play cards from their hand.
//	 Activate only as a sorcery."
//
// Two halves, one window (ADR 0066 amendment 2026-10-10). Each player
// gets a stored permission over the seven cards THEY exiled
// (EachPlayerExilesTopAndMayPlay), and every player gets the stored
// "can't play cards from your hand" record (ModCantPlayFromHand), both
// stamped "until your next turn" against the activator, so both end as
// that player's next turn begins (CR 611.2b, CR 800.4m if they have
// left by then).
//
// What falls out:
//
//   - A player plays only their own exiled cards, at their own normal
//     timing: a creature or a sorcery on their own turn, an instant
//     whenever they could cast one.
//   - A land played from exile is a land play and spends the turn's
//     land drop (CR 305.2); the hand ban leaves a land played from exile
//     alone, and a card in any zone but the hand alone.
//   - Activating abilities of cards in hand (cycling, channel) is not
//     playing them and stays open.
//   - Cards still in exile when the window closes stay there.
//
// Exiling the Vessel is a cost (CR 118.3), so it is gone before the
// ability resolves and is not among anything it exiles.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0179bc62-823e-46b9-b536-342904fedafc",
		Name:         memoryVesselName,
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{T}, Exile this artifact: Each player exiles the top seven cards of their library. Until your next turn, players may play cards they exiled this way, and they can't play cards from their hand. Activate only as a sorcery.",
			Cost:         Plus(TapCost(), ExileThis()),
			SorcerySpeed: true,
			Effect:       memoryVesselResolve,
		}},
	})
}

const memoryVesselName = "Memory Vessel"

// memoryVesselResolve exiles seven off every library with each player's
// permission over their own, then bans every player's hand, all until
// the activator's next turn.
func memoryVesselResolve(g *game.Game, item *game.StackItem) error {
	until := g.UntilYourNextTurnDuration(item.Controller)
	if err := EachPlayerExilesTopAndMayPlay(g, 7, until); err != nil {
		return err
	}
	g.CantPlayFromHandForEffect(item.SourceCardID, uuid.Nil, until,
		memoryVesselName+" — players can't play cards from their hand until its controller's next turn")
	return nil
}
