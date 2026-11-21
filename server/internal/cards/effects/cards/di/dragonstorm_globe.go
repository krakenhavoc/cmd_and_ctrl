package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// A Dragon TOKEN gets the counter too, since #762: a created token
// runs the same battlefield-entry pipeline every other permanent
// runs, so the counter is added to its entry event and placed through
// the CR 614 counter pipeline like any other.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6de5bd7-7704-4a0c-a27a-9e565e49f5e9",
		Name:         "Dragonstorm Globe",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B28PermanentsYouControlEnterWithACounter("Dragonstorm Globe: a Dragon enters with an additional +1/+1 counter",
				// "Dragon" is a word of the source's text, read through the
				// text the replacement lends its match (CR 612.2, #3984).
				func(c game.Card) bool { return c.HasSubtypeWord("Dragon") }),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
