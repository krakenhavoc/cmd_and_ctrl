package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anvil of Bogardan — Artifact {2}:
//
//	"Players have no maximum hand size.
//	 At the beginning of each player's draw step, that player draws an
//	 additional card, then discards a card."
//
// "Players have no maximum hand size" reaches every player (ADR 0113
// §3, #2074), folded in CR 613.11 timestamp order (the 2009-10-01
// ruling: Null Profusion then Anvil is no maximum, the other order
// two). The trigger is Howling Mine's draw-step trigger, after the
// normal draw (the 2013-04-15 ruling), and then that player chooses a
// card to discard. The discard happens even if the draw was replaced
// away (the 2004-10-04 ruling); two Anvils draw-and-discard twice,
// one at a time.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9deb2a0b-f40c-4d13-8321-41dc41448d24",
		Name:         "Anvil of Bogardan",
		Completeness: CompletenessFull,
		HandSize:     []game.HandSizeStatic{PlayersHaveNoMaxHandSize()},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginDrawStep, AnyPlayer, "Anvil of Bogardan — draw an additional card, then discard a card", func(g *game.Game, item *game.StackItem) error {
				player := triggeringActor(item)
				if player == uuid.Nil {
					return nil
				}
				if err := g.DrawNForEffect(player, 1); err != nil {
					return err
				}
				g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player:   player,
					Source:   item.SourceCardID,
					N:        1,
					Question: "Anvil of Bogardan — discard a card",
				})
				return nil
			}),
		},
	})
}
