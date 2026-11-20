package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skarrgan Pit-Skulk — Creature — Human Warrior {G}, 1/1 (Guildpact,
// #5599):
//
//	"Bloodthirst 1 (If an opponent was dealt damage this turn, this
//	 creature enters with a +1/+1 counter on it.)
//	 Creatures with power less than this creature's power can't block
//	 it."
//
// Battering Wurm's text on a one-drop: the shared Bloodthirst entry
// replacement and Howlgeist's block rule, compared as blockers are
// declared against the Pit-Skulk's power then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2ce99b9-7098-4de1-95bf-609bd448848d",
		Name:         "Skarrgan Pit-Skulk",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Bloodthirst(1, "Skarrgan Pit-Skulk")},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(PowerLessThanSource(), OnSelf(),
				"creatures with power less than Skarrgan Pit-Skulk's can't block it"),
		},
	})
}
