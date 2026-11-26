package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deus of Calamity — Creature — Spirit Avatar {R/G}{R/G}{R/G}{R/G}{R/G}, 6/6
// (Duel Decks: Heroes vs. Monsters, S82 old sets):
//
//	"Trample
//	 Whenever this creature deals 6 or more damage to an opponent, destroy
//	 target land that player controls."
//
// The trigger fires on one damage event from the Deus of at least six to an
// opponent, combat or not (CR 603.2): trample damage that splits across a
// blocker and the player counts only what reached the player. Mistblade
// Shinobi's "that player controls" is judged against the player this event
// named, and the target is chosen as the trigger goes on the stack (CR
// 603.3d); with no land to target it is removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f263f9d0-2ca9-4d3f-9b41-c67bdc73281a",
		Name:            "Deus of Calamity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventDealDamage, effects.S82OldB398ThisDealtAtLeastToAnOpponent(6),
				"Deus of Calamity — destroy target land that player controls",
				effects.S82OldB398DestroyLegalTarget),
				effects.ControlledByThatPlayer(effects.TargetPermanent("target land that player controls", effects.Land()))),
		},
	})
}
