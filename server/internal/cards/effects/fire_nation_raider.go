package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire Nation Raider — Creature — Human Soldier {3}{R}, 4/2 (Avatar: The
// Last Airbender, #4459):
//
//	"Raid — When this creature enters, if you attacked this turn, create
//	 a Clue token."
//
// Raid is an intervening "if" (CR 603.4): checked as it triggers and
// again as it resolves. Any attack declared this turn counts, whatever
// has happened to the attacker since (the rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b37e9bc5-6a2c-4f08-b856-3f8d0d16defb",
		Name:         "Fire Nation Raider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, tmcB6YouAttackedThisTurn),
				"Fire Nation Raider — Raid: create a Clue token",
				Do(If{Cond: avatarB11CondYouAttacked{}, Then: []Applier{
					CreateToken{Template: ClueToken(), N: 1},
				}})),
		},
	})
}
