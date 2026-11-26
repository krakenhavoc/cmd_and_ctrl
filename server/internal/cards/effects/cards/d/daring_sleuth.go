package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Daring Sleuth // Bearer of Overwhelming Truths — Creature — Human Rogue
// {1}{U}, 2/1 // Creature — Human Wizard 3/2 (Shadows over Innistrad,
// #5543):
//
//	Front: "When you sacrifice a Clue, transform this creature."
//	Back:  "Prowess
//	        Whenever this creature deals combat damage to a player,
//	        investigate."
//
// "You sacrifice a Clue" is Tireless Tracker's reading: the sacrifice event
// is emitted before the Clue leaves, so its subtype is read on the
// battlefield. The transform is the in-place verb (CR 701.27), ignored once
// the creature has turned over since the trigger went on the stack
// (CR 701.27f), so two Clues sacrificed together turn it over once. The
// back's trigger is the usual combat-damage-to-a-player one.
//
// No simplification.
func init() {
	const oracle = "69ca69c7-403b-4b5d-be08-d9dce8ce410e"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Daring Sleuth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, effects.LonisYouSacrificedAClue,
				"Daring Sleuth — transform this creature", effects.Do(effects.TransformThis{})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Bearer of Overwhelming Truths",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Bearer of Overwhelming Truths — investigate", effects.Do(effects.Investigate{N: 1})),
		},
	})
}
