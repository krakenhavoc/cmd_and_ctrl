package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abyssal Nightstalker — Creature — Nightstalker {3}{B}, 2/2 (EDHREC
// rank 27645):
//
//	"Whenever this creature attacks and isn't blocked, defending player
//	 discards a card."
//
// A proof card for #1279 — see Swamp Mosquito for the trigger. The
// defending player chooses the card (CR 701.8a: a discard nobody says
// is random is the discarding player's choice), so the payoff is the
// discard PROMPT, not DiscardCards, which is for "at random" only.
// Nothing hangs off the answer, so the fire-and-forget queue is right.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "10733767-1c97-4e2e-b02b-54bf346f6583",
		Name:         "Abyssal Nightstalker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenAttacksAndIsNotBlocked("Abyssal Nightstalker — defending player discards a card", defendingPlayerDiscardsACard),
		},
	})
}

func defendingPlayerDiscardsACard(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
	if !defendingPlayerStillIn(g, defender) {
		return nil
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player: defender,
		Source: item.SourceCardID,
		N:      1,
	})
	return nil
}
