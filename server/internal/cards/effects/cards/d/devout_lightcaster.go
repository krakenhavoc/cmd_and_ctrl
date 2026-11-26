package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devout Lightcaster — Creature — Kor Cleric {W}{W}{W}, 2/2 (Zendikar, S82 OLD
// b495):
//
//	"Protection from black
//	 When this creature enters, exile target black permanent."
//
// Protection from black is a printed keyword. The enters trigger is mandatory and
// targets any black permanent, whoever controls it (ExileChosenTarget, re-checked
// as it resolves); with none the trigger is not put on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a6f14387-4dfd-4d6b-a6b1-2ae658a70de6",
		Name:            "Devout Lightcaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Devout Lightcaster — exile target black permanent", effects.Do(effects.ExileChosenTarget{})),
				effects.TargetPermanent("target black permanent", effects.OfColor("B"))),
		},
	})
}
