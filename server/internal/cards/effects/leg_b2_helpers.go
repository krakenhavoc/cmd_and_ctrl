package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// leg_b2_helpers.go — the shape Legends batch 2 (#5639) prints more than
// once. Its own file, as #231 asks of a batch, so a concurrent batch
// never collides here.
//
// Append-only: add a builder, never change what one means.

// legSelfPump is "<cost>: <subject> gets +1/+<toughness> until end of
// turn." (Carrion Ants and Killer Bees +1/+1; Pavel Maliki and Wall of
// Opposition +1/+0). No tap in the cost, so it can be activated any number of
// times, summoning sick or not (CR 302.6 restricts only the tap and
// untap symbols); each activation is its own boost on the source, and the
// boosts end in the cleanup step (CR 514.2). `subject` is the printed
// name of the creature in the label: "This creature" or the card's own
// name. `name` labels the boost.
func legSelfPump(cost string, toughness int, subject, name string) ActivatedAbility {
	const power = 1
	return ActivatedAbility{
		Label:   cost + ": " + subject + " gets " + signedPT(power, toughness) + " until end of turn.",
		Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: power, Toughness: toughness}},
		Cost:    ManaCost(cost),
		Effect:  thisGetsUntilEndOfTurn(power, toughness, name+" — "+signedPT(power, toughness)+" until end of turn"),
	}
}
