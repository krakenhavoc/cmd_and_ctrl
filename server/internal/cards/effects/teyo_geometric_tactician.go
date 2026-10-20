package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teyo, Geometric Tactician — Legendary Planeswalker — Teyo {2}{W},
// loyalty 3:
//
//	"When Teyo enters, create a 0/4 white Wall creature token with
//	 defender and flying.
//	 +1: You and target opponent each draw a card.
//	 −2: Choose left or right. Until your next turn, each player may
//	     attack only the nearest opponent in the last chosen direction
//	     and planeswalkers controlled by that opponent."
//
// The enters trigger and +1 are written as printed. +1 draws for you
// first, then for the opponent, as one Do so the second draw waits for
// any prompt the first opens (CR 608.2c, Flumph's
// shape); an opponent who is no longer a legal target draws nothing.
//
// −2 is left out. The engine has no seat direction: "the nearest
// opponent in the chosen direction" and an attack restriction keyed to
// it are one seam with Pramikon, Sky Rampart's static (#4197).
// Offering the −2 as a loyalty cost that restricts nothing would be the
// lie ADR 0032 forbids, so the row is absent rather than empty.
//
// Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "0dfdf1a1-a3a9-490f-bd9d-dcba088c6785",
		Name:         "Teyo, Geometric Tactician",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Its −2 (choose left or right to limit attacks) isn't available yet — use it by hand."},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Teyo, Geometric Tactician — create a 0/4 white Wall creature token with defender and flying",
				Do(CreateToken{Template: TokenCard("0/4 white Wall with defender and flying"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "+1: You and target opponent each draw a card.",
			Purpose: game.Purpose{Draws: 1, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1})},
			Cost:    LoyaltyCost(1),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect:  youAndTargetOpponentEachDraw(),
		}},
	})
}
