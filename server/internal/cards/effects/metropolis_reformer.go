package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Metropolis Reformer — Creature — Angel Cleric {2}{W}, 2/3:
//
//	"Flying, vigilance
//	 You have hexproof.
//	 Whenever this creature is dealt damage, you gain that much life."
//
// All damage dealt at once (one combat damage step, one instruction)
// is one trigger for the total (CR 603.2c, ThatMuchDamage).
func init() {
	Register(Spec{
		OracleID:        "baec04a1-3bc9-4618-afab-e42013e36648",
		Name:            "Metropolis Reformer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		PlayerKeywords:  []string{KeywordHexproof},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			},
			Key:        "Metropolis Reformer — you gain that much life",
			AtBatchEnd: ThatMuchDamage(),
			Effect:     cmdrB1GainThatMuchDamage,
		}},
	})
}
