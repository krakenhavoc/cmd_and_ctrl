package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonstorm Globe — Artifact {3} (EDHREC rank 2975):
//
//	"Each Dragon you control enters with an additional +1/+1 counter
//	 on it.
//	 {T}: Add one mana of any color."
//
// The Dragon deck's rock. The counter is a CR 614 replacement on the
// entry of any Dragon permanent under the controller's control
// (b28PermanentsYouControlEnterWithACounter, subtype read off the
// entering card), so the counter is on the Dragon before its enters
// trigger or any state check sees it, and a Hardened Scales applies
// on top, as printed. The mana is a five-colour pipe pick, all five
// offered with the commander's identity first, the way every "any
// color" rock is.
//
// Known engine gap, declared: a Dragon TOKEN skips the entry pipeline
// (token creation does not run the CR 614 replacements), so it enters
// without the counter — weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "f6de5bd7-7704-4a0c-a27a-9e565e49f5e9",
		Name:         "Dragonstorm Globe",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Dragon tokens don't get the extra +1/+1 counter — only Dragon cards entering the battlefield do."},
		Replacements: []game.ReplacementEffect{
			b28PermanentsYouControlEnterWithACounter("Dragonstorm Globe: a Dragon enters with an additional +1/+1 counter",
				func(c game.Card) bool { return c.HasSubtype("Dragon") }),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
