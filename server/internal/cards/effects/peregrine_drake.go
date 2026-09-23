package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Peregrine Drake — "Flying. When this creature enters, untap up
// to five lands."
//
// The blink engine's mana battery: flickering the Drake re-triggers
// the untap.
//
// "Up to five lands" is the controller's choice, made as the trigger
// resolves, among every land on the battlefield — theirs or anyone
// else's, since the clause says neither "target" nor "you control".
// UntapUpToLands asks it as a choose_cards prompt over the tapped
// lands at the table, so zero to five of them untap; the choice is
// untargeted, so a hexproof land is as choosable as any other.
func init() {
	Register(Spec{
		OracleID:        "0bd67481-6bd9-48d6-92bd-8933b5ea1eae",
		Name:            "Peregrine Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Peregrine Drake — untap up to five lands", Do(UntapUpToLands{
				N:        5,
				Question: "Peregrine Drake — untap up to five lands",
			})),
		},
	})
}
