package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dauthi Ghoul — {1}{B} Creature — Dauthi Zombie, 1/1 (Tempest, S82 OLD batch 770,
// #5628):
//
//	"Shadow
//	 Whenever a creature with shadow dies, put a +1/+1 counter on this creature."
//
// Shadow is a printed keyword (CR 702.28). The trigger watches EVERY creature that
// dies, this one's controller's or anyone's, and asks whether it had shadow as it
// last existed on the battlefield (CR 603.10a: Event.WasKeyword), so a creature
// that lost shadow before it died does not count and one that gained it does. A
// Ghoul that dies itself is gone and gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5bf39f9e-4f06-4637-b0fc-5d21b926f6f4",
		Name:            "Dauthi Ghoul",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.S82OldB770ACreatureWithShadowDied,
				"Dauthi Ghoul — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
