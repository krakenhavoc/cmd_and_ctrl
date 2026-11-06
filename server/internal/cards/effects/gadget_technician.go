package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gadget Technician — Creature — Goblin Artificer {2}{U}{R}, 3/2 (Murders
// at Karlov Manor, #4469):
//
//	"When this creature enters or is turned face up, create a 1/1
//	 colorless Thopter artifact creature token with flying.
//	 Disguise {U/R}{U/R} (You may cast this card face down for {3} as a
//	 2/2 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)"
//
// Offender at Large's one ability with two conditions (OnAny). Entering
// face down does not trigger it, because a face-down permanent has no
// abilities (CR 708.3); turning it face up does (CR 708.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d7f777ae-6122-4af9-b962-7574bc8e2691",
		Name:             "Gadget Technician",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Disguise("{U/R}{U/R}")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				OnAny([]game.EventKind{game.EventETB, game.EventTurnedFaceUp}, Self,
					"Gadget Technician — create a 1/1 colorless Thopter artifact creature token with flying",
					Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
				game.Purpose{Tokens: 1}),
		},
	})
}
