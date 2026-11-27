package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oath of Lieges — Enchantment {1}{W} (Exodus, S82 OLD batch 743, #5626):
//
//	"At the beginning of each player's upkeep, that player chooses target
//	 player who controls more lands than they do and is their opponent. The
//	 first player may search their library for a basic land card, put that
//	 card onto the battlefield, then shuffle."
//
// Oath of Druids' shape: the Oath's controller controls the trigger, but the
// upkeep player chooses its target (OfThatPlayersChoice), judged against the
// upkeep player: an opponent of theirs who controls more lands
// (WithMoreLandsThanTheChooser, #5626). With no such player the trigger is
// removed as it would go on the stack (CR 603.3d), and a target that no longer
// qualifies at resolution makes it do nothing (CR 608.2b). The target is only a
// condition. On resolution the upkeep player is asked, and on a yes searches
// for a basic land card, puts it onto the battlefield and shuffles.
//
// No simplification.
func init() {
	const name = "Oath of Lieges"
	effects.Register(effects.Spec{
		OracleID:     "87038354-6762-4069-be34-7d40ae308a8d",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.AtEachUpkeep(name+" — that player may search for a basic land card and put it onto the battlefield", effects.S82OldB743OathOfLiegesAsk),
			effects.WithMoreLandsThanTheChooser(effects.OfThatPlayersChoice(
				effects.TargetPlayer("target player who controls more lands than they do and is their opponent"))),
		)},
	})
}
