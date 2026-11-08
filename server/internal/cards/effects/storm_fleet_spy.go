package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Storm Fleet Spy — Creature — Human Pirate {2}{U}, 2/2 (Foundations,
// #4465):
//
//	"Raid — When this creature enters, if you attacked this turn, draw
//	 a card."
//
// Fire Nation Raider's raid enters trigger with a draw for its Clue.
// Raid is an intervening "if" (CR 603.4): checked as it triggers and
// again as it resolves. Any attack declared this turn counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42081a57-b521-45c2-928d-0a0a4641818d",
		Name:         "Storm Fleet Spy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, tmcB6YouAttackedThisTurn),
				"Storm Fleet Spy — Raid: draw a card",
				Do(If{Cond: avatarB11CondYouAttacked{}, Then: []Applier{DrawCards{N: 1}}})),
		},
	})
}
