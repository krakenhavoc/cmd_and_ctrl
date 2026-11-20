package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suffocating Fumes — Instant {2}{B} (Ikoria: Lair of Behemoths, #5512):
//
//	"Creatures your opponents control get -1/-1 until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// A one-sided -1/-1 over the opponents' creatures as it resolves; a
// creature it takes to 0 toughness dies to state-based actions. Cycling
// from hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "80c7b1b4-4e37-4aef-9432-d629d71b7045",
		Name:     "Suffocating Fumes",
		// ADR 0126 §6: the -1/-1 on your opponents' creatures removes the x/1s among them.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(BoostUntilEOT{
			Match: And(Creature(), OpponentControls()), Power: -1, Toughness: -1,
			Label: "Suffocating Fumes — creatures your opponents control get -1/-1",
		}),
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
