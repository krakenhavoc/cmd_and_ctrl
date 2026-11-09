package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Public Thoroughfare — Land (Murders at Karlov Manor, #4469):
//
//	"This land enters tapped.
//	 When this land enters, sacrifice it unless you tap an untapped
//	 artifact or land you control.
//	 {T}: Add one mana of any color."
//
// The entry trigger is Koskun Falls' tap payment (game.PayActionTap over
// "artifact or land") behind the pay-unless prompt the step waits on
// (UpkeepPayUnless reads the step off the cursor, so it is right for an
// entry trigger too), with the shared "sacrifice this permanent" decline.
// With no untapped artifact or land there is no payment to make, so the
// land is sacrificed. Tapping the permanent is not one of its {T}
// abilities, so an artifact creature that arrived this turn can pay. A
// Thoroughfare that left in response, or left and came back (CR 400.7),
// asks nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de5b995c-9691-4555-9070-66bcbc29f955",
		Name:         "Public Thoroughfare",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			EntersSacrificeUnlessYouTap("Public Thoroughfare", "artifact or land", "artifacts or lands", QueryTypes("artifact", "land")),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
