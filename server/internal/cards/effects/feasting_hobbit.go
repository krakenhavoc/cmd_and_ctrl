package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feasting Hobbit — Creature — Halfling Citizen {1}{G}, 2/2 (Tales of
// Middle-earth Commander, #5948):
//
//	"Devour Food 3 (As this creature enters, you may sacrifice any number of
//	 Foods. It enters with three times that many +1/+1 counters on it.)
//	 Creatures with power less than this creature's power can't block it."
//
// Devour over Foods rather than creatures (the shared devourOver, Famished
// Worldsire's land form): the sacrifice is an as-enters choice of any number
// of Foods, and the Hobbit enters with three +1/+1 counters for each. The
// block rule is Battering Wurm's: the comparison is made as blockers are
// declared, against the Hobbit's power then, so the counters raise the bar.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cca33482-0bc5-4fe4-89dc-ba33b2e93831",
		Name:         "Feasting Hobbit",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			devourOver("Feasting Hobbit", 3, "Foods", func(c game.Card) bool { return c.HasSubtype("Food") }),
		},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(PowerLessThanSource(), OnSelf(),
				"creatures with power less than Feasting Hobbit's can't block it"),
		},
	})
}
