package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Command Bridge — Land (Edge of Eternities, #4461):
//
//	"This land enters tapped.
//	 When this land enters, sacrifice it unless you tap an untapped
//	 permanent you control.
//	 {T}: Add one mana of any color."
//
// The entry trigger is EntersSacrificeUnlessYouTap's tap payment (game.PayActionTap over
// any permanent) behind the pay-unless prompt the step waits on
// (UpkeepPayUnless reads the step off the cursor, so it is right for an
// entry trigger too), with the shared "sacrifice this permanent" decline.
// With no untapped permanent there is no payment to make, so the
// land is sacrificed. Tapping the permanent is not one of its {T}
// abilities, so an artifact creature that arrived this turn can pay. A
// Bridge that left in response, or left and came back (CR 400.7),
// asks nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87c8e1ed-258a-4a89-bcc6-211405e49692",
		Name:         "Command Bridge",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			EntersSacrificeUnlessYouTap("Command Bridge", "permanent", "permanents", game.PermanentQuery{}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
