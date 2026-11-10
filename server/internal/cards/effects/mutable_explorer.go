package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mutable Explorer — Creature — Shapeshifter {2}{G}, 1/1 (Lorwyn
// Eclipsed, #4458):
//
//	"Changeling (This card is every creature type.)
//	 When this creature enters, create a tapped Mutavault token."
//
// The token is the card Mutavault by name (CR 111.11, the card's
// ruling): MutavaultToken carries Mutavault's oracle ID, so its mana
// ability and its animation are the catalog card's. It enters tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "34b93a66-de1e-4133-a479-e0d90ba94a3a",
		Name:            "Mutable Explorer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mutable Explorer — create a tapped Mutavault token",
				Do(CreateTokenAdvanced{Spec: Token(MutavaultToken()).EntersTapped(), N: 1})),
		},
	})
}
