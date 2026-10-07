package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Automated Assembly Line — Artifact {1}{W}:
//
//	"Whenever one or more artifact creatures you control deal combat
//	 damage to a player, you get {E} (an energy counter).
//	 Pay {E}{E}{E}: Create a tapped 3/3 colorless Robot artifact
//	 creature token."
//
// The trigger is the "one or more" shape (CR 603.2c): once per player
// the controller's artifact creatures connect with in a combat damage
// step, whatever their number. ADR 0129 §2 (#1995): "Pay {E}{E}{E}" is
// the energy cost component, and the ability may be activated any
// number of times while the energy lasts.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "6e24a5bd-0ce3-4dc1-98c6-60d527bfdcfd",
		Name:         "Automated Assembly Line",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(Artifact(),
				"Automated Assembly Line — you get {E}", Do(GetEnergy{N: 1})), game.Purpose{Energy: 1}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Pay {E}{E}{E}: Create a tapped 3/3 colorless Robot artifact creature token.",
			Cost:    PayEnergy(3),
			Purpose: game.Purpose{Tokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				_, err := g.CreateTokensForEffect(item.Controller, assemblyLineRobotToken(), 1, game.TokenEntryOptions{Tapped: true})
				return err
			},
		}},
	})
}

// assemblyLineRobotToken is the 3/3 colorless Robot artifact creature
// token Automated Assembly Line makes.
func assemblyLineRobotToken() game.Card {
	return game.Card{Name: "Robot", TypeLine: "Token Artifact Creature — Robot", Power: 3, Toughness: 3}
}
