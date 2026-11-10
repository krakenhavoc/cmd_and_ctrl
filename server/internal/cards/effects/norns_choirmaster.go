package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Norn's Choirmaster — Creature — Phyrexian Angel {3}{W}{W}, 5/4:
//
//	"Flying, first strike
//	 Whenever a commander you control enters or attacks, proliferate."
//
// One ability watching two events. Choirmaster itself counts when it is
// your commander. The proliferate choice is the player's (CR 701.34).
func init() {
	Register(Spec{
		OracleID:        "bad6a5a9-7010-418d-b042-182926aa999d",
		Name:            "Norn's Choirmaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventAttack}, janiceRandCommanderYouControl,
				"Norn's Choirmaster — proliferate", Do(Proliferate{})),
		},
	})
}
