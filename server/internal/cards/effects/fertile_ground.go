package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fertile Ground — Enchantment — Aura {1}{G} (#302):
//
//	"Enchant land
//	 Whenever enchanted land is tapped for mana, its controller adds an
//	 additional one mana of any color."
//
// The colour choice lives inside the TRIGGER rather than on the land's
// own ability, which is the shape #763's ADR had to answer. It reuses
// the two modes the mana pipeline already had for a multi-option slot,
// and adds no third:
//
//   - tapped BY HAND, it queues the same mana_pick a Birds of Paradise
//     activation queues, all five colours with the commander's
//     identity first;
//   - tapped by the AUTO-TAPPER, it makes the colour the plan booked
//     for it, or picks greedily against what the cast still owes when
//     the plan booked none, because the auto-tapper's contract is "no
//     further player decisions" and a prompt appearing halfway through
//     a cast would break it.
//
// The pick the trigger queues is NOT marked as a tap (ManaTapped), so
// answering it adds mana and stops: a triggered mana ability does not
// re-trigger, and nothing was tapped for its mana.
//
// The auto-tap planner counts the extra mana as a five-colour pick and
// books it like any other slot (ADR 0074, amendment of 2026-10-30), so
// the pick cannot take a pip that another source in the plan was booked
// for.
//
// No simplification of the card itself.
func init() {
	Register(Spec{
		OracleID:     "cf14d4e5-5965-45ad-97f7-26facf2884b5",
		Name:         "Fertile Ground",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		ManaTriggers: []game.ManaTrigger{
			WheneverAttachedTapsForMana(
				"Fertile Ground — add an additional one mana of any color",
				"{W|U|B|R|G}"),
		},
	})
}
