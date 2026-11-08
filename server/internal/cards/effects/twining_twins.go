package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twining Twins // Swift Spiral — Creature — Faerie Wizard {2}{U}{U}, 4/4
// // Instant — Adventure {1}{W} (Wilds of Eldraine, #4471):
//
//	Twining Twins — "Flying, vigilance, ward {1}"
//	Swift Spiral — "Exile target nontoken creature. Return it to the
//	 battlefield under its owner's control at the beginning of the next
//	 end step. (Then exile this card. You may cast the creature later from
//	 exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). Ward is the shared trigger. Swift Spiral is Eerie
// Interlude's exile and delayed return (CR 603.7) on one nontoken
// creature, anyone's: it comes back under its owner's control as a new
// object (CR 400.7). A target that left in response stays where it is
// (CR 608.2b). The Adventure lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const twiningTwinsOracleID = "105aea98-8eb9-4fb2-a0cb-7c7513317c5b"

func init() {
	Register(Spec{
		OracleID:        twiningTwinsOracleID,
		Name:            "Twining Twins",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Twining Twins — ward {1}"),
		},
	})
	Register(Spec{
		OracleID:     twiningTwinsOracleID + "#1",
		Name:         "Swift Spiral",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target nontoken creature", Not(IsTokenPredicate())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileTargetsThenScheduleReturn(ctx, "Swift Spiral — return the exiled creature")
		},
	})
}
