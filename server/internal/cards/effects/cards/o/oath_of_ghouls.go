package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oath of Ghouls — Enchantment {1}{B} (Exodus, S82 OLD batch 745, #5626):
//
//	"At the beginning of each player's upkeep, that player chooses target
//	 player whose graveyard has fewer creature cards in it than their
//	 graveyard does and is their opponent. The first player may return a
//	 creature card from their graveyard to their hand."
//
// Oath of Druids' shape: the Oath's controller controls the trigger but the
// upkeep player chooses its target (OfThatPlayersChoice), judged against that
// player: an opponent of theirs whose graveyard holds fewer creature cards
// than theirs (FewerCreatureCardsInGraveyardThanChooser). With no such player
// the trigger is removed from the stack (CR 603.3d), and if the target no
// longer qualifies as it resolves the ability does nothing (CR 608.2b). The
// target is only a condition. The upkeep player, not the controller, is asked
// which creature card to take.
//
// No simplification.
func init() {
	const name = "Oath of Ghouls"
	effects.Register(effects.Spec{
		OracleID:     "b9490b52-6c3f-437b-b2c2-9e81c4b43b61",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.AtEachUpkeep(name+" — that player may return a creature card from their graveyard to their hand", effects.ThatPlayerMayReturnACreatureCardToHand(name)),
			effects.WithFewerCreatureCardsInGraveyardThanTheChooser(effects.OfThatPlayersChoice(
				effects.TargetPlayer("target player whose graveyard has fewer creature cards in it than their graveyard does and is their opponent"))),
		)},
	})
}
