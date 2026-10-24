package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elegy Acolyte — Creature — Human Cleric {2}{B}{B}, 4/4 (Edge of
// Eternities):
//
//	"Lifelink
//	 Whenever one or more creatures you control deal combat damage to a
//	 player, you draw a card and lose 1 life.
//	 Void — At the beginning of your end step, if a nonland permanent
//	 left the battlefield this turn or a spell was warped this turn,
//	 create a 2/2 colorless Robot artifact creature token."
//
// The combat trigger is the batch shape (CR 603.2c): once per player
// connected with per damage step, not once per creature. Void is an
// intervening if (void.go). Lifelink is printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f7d52eac-edc9-464b-a163-07c136aa6e4c",
		Name:            "Elegy Acolyte",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Elegy Acolyte — you draw a card and lose 1 life",
				Do(DrawCards{N: 1}, youLoseLife(1))),
			AtYourEndStepIfVoid("Elegy Acolyte — create a 2/2 colorless Robot artifact creature token",
				Do(CreateToken{Template: TokenCard("2/2 colorless Robot artifact"), N: 1})),
		},
	})
}
