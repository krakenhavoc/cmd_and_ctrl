package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Path of Ancestry — Land (EDHREC rank 14, the highest-ranked card
// this catalog was missing that needed no new machinery):
//
//	"This land enters tapped.
//	 {T}: Add one mana of any color in your commander's color
//	 identity. When that mana is spent to cast a creature spell that
//	 shares a creature type with your commander, scry 1."
//
// The enters-tapped clause is a real CR 614 self-replacement, not an
// OnETB tap — the land is never untapped on the battlefield, the
// distinction the Temple cycle pinned. The mana half is Command
// Tower's identity-narrowed pipe.
//
// SANDBOX SIMPLIFICATION — the scry rider is NOT implemented.
// "When THAT MANA is spent to cast …" is a delayed trigger keyed to
// the provenance of one mana in the pool: the engine's ManaPool
// records colour and nothing about which permanent produced it, so
// there is no seam to hang the trigger on. Wiring it needs
// per-source mana tagging (which also unlocks Cavern of Souls'
// "can't be countered" and Delighted Halfling's spend-restriction),
// and that is a mana-pipeline change, not a card change.
//
// The direction is WEAKER than printed: the land still enters tapped
// and still produces the same mana; it just never scries. A player
// who wants the scry can take it manually — the automation simply
// does not grant it.
func init() {
	Register(Spec{
		OracleID:     "b473e293-59e3-4e04-acf2-622604aeb25f",
		Name:         "Path of Ancestry",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The scry 1 never happens when you spend its mana on a creature that shares a type with your commander."},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color in your commander's color identity",
		}},
	})
}
