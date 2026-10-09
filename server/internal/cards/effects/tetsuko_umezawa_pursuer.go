package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tetsuko Umezawa, Pursuer — Legendary Creature — Human Mercenary
// {3}{R}, 2/4:
//
//	"Double strike
//	 Prowess
//	 Whenever a creature an opponent controls with power or toughness 1
//	 or less blocks, Tetsuko Umezawa deals 1 damage to that creature's
//	 controller."
//
// EventBlock is per (blocker, attacker) pair, so a creature that blocks
// two attackers triggers it twice; CR 509.3a would give one trigger for
// "blocks" with no object, and this card's wording ("a creature ...
// blocks") is that case. The trigger therefore reads only the first
// pair of a blocker's declaration (selfBlocksOnce's Amount rule).
// Power and toughness are the blocker's effective ones as it blocks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "708d4567-a5a3-437f-bae4-ae10f1537aa9",
		Name:            "Tetsuko Umezawa, Pursuer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "prowess"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, rfCreatureEWeakOpposingBlocker,
				"Tetsuko Umezawa, Pursuer — 1 damage to the blocker's controller", rfCreatureETetsukoPing),
		},
	})
}
