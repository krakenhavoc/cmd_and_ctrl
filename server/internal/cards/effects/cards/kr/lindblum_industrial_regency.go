package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
const LindblumOracleID = "4cc014f3-05e0-442e-9dee-03eab1aa65a3"

func init() {
	effects.Register(effects.Spec{
		OracleID:     LindblumOracleID,
		Name:         "Lindblum, Industrial Regency",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     LindblumOracleID + "#1",
		Name:         "Mage Siege",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: item.Controller, Template: effects.NoncreatureCastWizardToken(), N: 1}.Apply(ctx)
		},
	})
}
