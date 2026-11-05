package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cybernetic Specialist — Artifact Creature — Human Scientist {2}, 2/1
// (Star Trek, #4973):
//
//	"When this creature enters, surveil 1.
//	 {1}, {T}: Put target card from your graveyard on the bottom of your
//	 library."
//
// Coin of Fate's surveil 1 on entry, and Barkform Harvester's graveyard
// tuck behind a {1} and {T} cost (trekB27TuckTargetCardToBottom). "Your
// graveyard" is YouOwn: a card in a graveyard has no controller. The {T}
// is the tap symbol, so a Specialist that entered this turn cannot pay it
// (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9fc49a69-f80c-478f-b237-66f155c0c09a",
		Name:         "Cybernetic Specialist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cybernetic Specialist — surveil 1", Do(Surveil{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Put target card from your graveyard on the bottom of your library",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCardInGraveyard("target card in your graveyard", YouOwn()),
			Effect:  trekB27TuckTargetCardToBottom,
		}},
	})
}
