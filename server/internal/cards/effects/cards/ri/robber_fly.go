package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Robber Fly — Creature — Insect {2}{R}, 1/1 (Mercadian Masques, S82 OLD,
// #5619):
//
//	"Flying
//	 Whenever this creature becomes blocked, defending player discards all
//	 the cards in their hand, then draws that many cards."
//
// "Becomes blocked" is once per blocked attacker (CR 509.3c). The defending
// player is the one the Fly was attacking, read as the trigger resolves; the
// whole hand is discarded first and the same number is drawn after
// (Wheel of Fortune's order), so a player with an empty hand draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0d44d4f3-9713-4d18-a985-9e7ff0df85d3",
		Name:            "Robber Fly",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self,
				"Robber Fly — defending player discards all the cards in their hand, then draws that many cards",
				robberFlyWheel),
		},
	})
}

func robberFlyWheel(g *game.Game, item *game.StackItem) error {
	victim := g.DefendingPlayerForAttackerForEffect(item.SourceCardID)
	if victim == uuid.Nil {
		return nil
	}
	n, err := effects.DiscardWholeHand(g, victim)
	if err != nil || n == 0 {
		return err
	}
	return effects.DrawCards{Player: victim, N: n}.Apply(effects.NewContext(g, item))
}
