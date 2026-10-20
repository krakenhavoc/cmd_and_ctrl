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
// Every ability is written as printed. +1 draws for you first, then for
// the opponent, as one Do so the second draw waits for any prompt the
// first opens (CR 608.2c, Flumph's shape); an opponent who is no longer
// a legal target draws nothing.
//
// −2 asks Teyo's controller "left or right" as it resolves, then gives
// every player "until your next turn, you may attack only your nearest
// opponent that way and planeswalkers they control" (ADR 0206). Left is
// the next player in turn order. A seat whose player has left the game
// is skipped, so the nearest opponent is always someone still playing.
// Battles can't be attacked while it lasts, since only the opponent and
// their planeswalkers are listed. It doesn't stop a creature put onto
// the battlefield attacking (CR 508.4c, and the Teyo ruling). With
// Pramikon or a second Teyo pointing the other way, a player in a game
// of three or more has nobody to attack, which is the ruling.
//
// Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "0dfdf1a1-a3a9-490f-bd9d-dcba088c6785",
		Name:         teyoGeometricTacticianName,
		Completeness: CompletenessFull,
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
		}, {
			Label:  "−2: Choose left or right. Until your next turn, each player may attack only the nearest opponent in the last chosen direction and planeswalkers controlled by that opponent.",
			Cost:   LoyaltyCost(-2),
			Effect: ChooseLeftOrRightThenEachPlayerAttacksOnlyThatWayUntilYourNextTurn(teyoGeometricTacticianName, teyoGeometricTacticianDirection),
		}},
	})
}

const teyoGeometricTacticianName = "Teyo, Geometric Tactician"

// teyoGeometricTacticianDirection is the −2's keyed answer. Its key is
// an on-disk identity: never renamed, never reused.
var teyoGeometricTacticianDirection = AttackDirectionUntilYourNextTurnThen(
	"option-pick/teyo-geometric-tactician-direction", teyoGeometricTacticianName)
