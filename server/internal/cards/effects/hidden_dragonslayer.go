package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hidden Dragonslayer — Creature — Human Warrior {1}{W}, 2/1 (Murders at
// Karlov Manor Commander):
//
//	"Lifelink
//	 Megamorph {2}{W}
//	 When this creature is turned face up, destroy target creature with
//	 power 4 or greater an opponent controls."
//
// The target is chosen as the trigger goes on the stack and re-judged as
// it resolves (CR 608.2b): a creature that shrank below 4 power in
// response is no longer legal.
//
// No simplification.
func init() {
	up := WhenThisIsTurnedFaceUp("Hidden Dragonslayer — destroy target creature with power 4 or greater an opponent controls",
		destroyFirstLegalTarget)
	up.Targets = TargetPermanent("target creature with power 4 or greater an opponent controls",
		Creature(), PowerGE(4), OpponentControls())
	Register(Spec{
		OracleID:         "862af5c4-cb34-4bfa-a7ff-209e103defd6",
		Name:             "Hidden Dragonslayer",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"lifelink"},
		AlternativeCosts: []game.AlternativeCost{Megamorph("{2}{W}")},
		Triggered:        []game.TriggeredAbility{up},
	})
}
