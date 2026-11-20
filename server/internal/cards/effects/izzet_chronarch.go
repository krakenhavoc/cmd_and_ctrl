package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Izzet Chronarch — Creature — Human Wizard {3}{U}{R}, 2/2 (Guildpact,
// #5599):
//
//	"When this creature enters, return target instant or sorcery card
//	 from your graveyard to your hand."
//
// Archaeomancer's trigger: the graveyard target is chosen as it goes on
// the stack and re-checked as it resolves (returnTargetCardToHand).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1da438f3-db1c-4713-a60c-e078f31d809c",
		Name:         "Izzet Chronarch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetCardInGraveyard("target instant or sorcery card from your graveyard", Or(Instant(), Sorcery()), YouOwn()),
			Key:       "Izzet Chronarch — return an instant or sorcery card to hand",
			Effect:    returnTargetCardToHand,
		}},
	})
}
