package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tura Kennerüd, Skyknight — Legendary Creature — Human Knight
// {2}{W}{U}{U}, 3/3 (Dominaria United, #5487):
//
//	"Flying
//	 Whenever you cast an instant or sorcery spell, create a 1/1 white
//	 Soldier creature token."
//
// Young Pyromancer's cast trigger with a Soldier in place of the
// Elemental: a spell you cast, not a copy of one (CR 707.12).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4cb3a303-9ded-46d5-90e7-1cc1c65a75da",
		Name:            "Tura Kennerüd, Skyknight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Tura Kennerüd, Skyknight — create a 1/1 white Soldier",
				Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1})),
		},
	})
}
