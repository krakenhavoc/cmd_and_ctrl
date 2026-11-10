package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Josu Vess, Lich Knight — Legendary Creature — Zombie Knight {2}{B}{B},
// 4/5 (Foundations Commander):
//
//	"Kicker {5}{B}
//	 Menace
//	 When Josu Vess enters, if he was kicked, create eight 2/2 black
//	 Zombie Knight creature tokens with menace."
//
// The kicker is an optional additional cost (CR 702.33a); "if he was kicked"
// is an intervening if (CR 603.4) read off the record the resolution carried
// onto the permanent (CR 400.7d), so a Josu that was reanimated or flickered
// makes none. The Zombie Knight is the printed token of the set: 2/2 black
// with menace.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "974a46f9-aa84-4b34-bee5-c635166e5841",
		Name:            "Josu Vess, Lich Knight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		OptionalCosts:   []game.AdditionalCost{Kicker("{5}{B}")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
				"Josu Vess — kicked, create eight 2/2 black Zombie Knight tokens with menace",
				Do(CreateToken{Template: TokenCard("2/2 black Zombie Knight with menace"), N: 8})),
				game.Purpose{Tokens: 8, CreatureTokens: 8}),
		},
	})
}
