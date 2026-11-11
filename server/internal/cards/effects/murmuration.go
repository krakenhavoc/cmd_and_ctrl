package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Murmuration — Enchantment {4}{W} (Bloomburrow Commander):
//
//	"Birds you control get +1/+1 and have vigilance.
//	 At the beginning of your end step, for each spell you've cast this turn,
//	 create a 1/2 blue Bird creature token with flying named Storm Crow."
//
// The anthem and the vigilance grant are two statics over Birds you control
// (a Bird token is one, and so is a changeling). The end step trigger counts
// the spells its controller cast this turn off the per-turn tally as it
// resolves, so a spell cast with the trigger on the stack cannot add to it
// and a countered spell still counts (it was cast). The token is the printed
// Storm Crow, a 1/2 blue Bird named Storm Crow, so it takes the bonus too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84e7be36-b8cd-4111-a91b-3628f421d5d5",
		Name:         "Murmuration",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(cmdr2S41BirdsYouControl, 1, 1),
			b16GrantKeywords(cmdr2S41BirdsYouControl, "vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Murmuration — for each spell you've cast this turn, create a 1/2 blue Bird with flying named Storm Crow",
				cmdr2S41StormCrowForEachSpell),
		},
	})
}
