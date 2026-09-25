package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guiltfeeder — Creature — Horror {3}{B}{B}, 0/4 (EDHREC rank 6848):
//
//	"Fear
//	 Whenever this creature attacks and isn't blocked, defending player
//	 loses 1 life for each card in their graveyard."
//
// A proof card for #1279 — see Swamp Mosquito for the trigger. The
// graveyard is counted as the ability RESOLVES (CR 608.2h), so a card
// milled in response counts and one exiled in response does not. Life
// loss, not damage: nothing prevents it and lifelink has nothing to
// do with it. Fear is the engine's CR 702.36 block restriction.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "0285bd20-f49e-48f7-8c8c-960f9fbe4d34",
		Name:            "Guiltfeeder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Triggered: []game.TriggeredAbility{
			WhenAttacksAndIsNotBlockedEffect("Guiltfeeder — defending player loses 1 life for each card in their graveyard", guiltfeederDrain),
		},
	})
}

func guiltfeederDrain(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
	if !defendingPlayerStillIn(g, defender) {
		return nil
	}
	p := g.PlayerByIDForEffect(defender)
	if p.Graveyard == nil || len(p.Graveyard.Cards) == 0 {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, defender, -len(p.Graveyard.Cards))
}
