package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Consuming Blob — Creature — Ooze {3}{G}{G}, */*+1:
//
//	"Consuming Blob's power is equal to the number of card types among
//	 cards in your graveyard and its toughness is equal to that number
//	 plus 1.
//	 At the beginning of your end step, create a green Ooze creature
//	 token with "This token's power is equal to the number of card types
//	 among cards in your graveyard and its toughness is equal to that
//	 number plus 1.""
//
// Souls of the Lost's layer 7a characteristic-defining ability counting
// card types (Broodspinner's count). The Ooze carries the same ability on
// its own template (a token that sizes itself, ADR 0083), so it survives
// the Blob leaving; each reads its own controller's graveyard.
func init() {
	Register(Spec{
		OracleID:     "818292cc-6261-47e4-b573-d92f03da6ded",
		Name:         "Consuming Blob",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			consumingBlobCardTypesCDA("Power is the number of card types among cards in your graveyard; toughness is that plus 1."),
		},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Consuming Blob — create an Ooze token that sizes itself",
				Do(CreateToken{Template: consumingBlobOozeToken(), N: 1})),
		},
	})
}
