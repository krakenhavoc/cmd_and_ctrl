package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lindblum, Industrial Regency // Mage Siege — Land — Town, with an
// Instant — Adventure half {2}{R} (oracle 4cc014f3…, #3213):
//
//	Lindblum, Industrial Regency
//	  "This land enters tapped.
//	   {T}: Add {R}."
//	Mage Siege
//	  "Create a 0/1 black Wizard creature token with "Whenever you cast
//	   a noncreature spell, this token deals 1 damage to each opponent."
//	   (Then exile this card. You may play the land later from exile.)"
//
// A Town land with an Adventure half; see Zanarkand, Ancient
// Metropolis for the land-play lifecycle (CR 715.3d). The Wizard is the
// shared NoncreatureCastWizardToken. Mage Siege is an instant, so it
// can be cast at any time the engine would let an instant be cast.
//
// No simplification.
const lindblumOracleID = "4cc014f3-05e0-442e-9dee-03eab1aa65a3"

func init() {
	Register(Spec{
		OracleID:     lindblumOracleID,
		Name:         "Lindblum, Industrial Regency",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
	})
	Register(Spec{
		OracleID:     lindblumOracleID + "#1",
		Name:         "Mage Siege",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: item.Controller, Template: NoncreatureCastWizardToken(), N: 1}.Apply(ctx)
		},
	})
}
