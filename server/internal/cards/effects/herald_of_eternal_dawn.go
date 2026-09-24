package effects

// Herald of Eternal Dawn — Creature — Angel {4}{W}{W}{W}, 6/6:
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Flying
//	 You can't lose the game and your opponents can't win the game."
//
// Platinum Angel's static on a white 6/6 with flash, and the card the
// Aang deck (#1306) was waiting on #749 for. The static is the DERIVED
// half of ADR 0057 Decision 4: declared as Spec.GameEndGates and read
// off the battlefield at every loss and every win
// (game.CatalogGameEndGates, keyed by CatalogAbilityKey), never
// written to a player. So:
//
//   - while it is on the battlefield its controller loses to nothing
//     but a concession — 0 or less life, an empty-library draw, ten
//     poison counters, 21 commander damage and "you lose the game" are
//     all stopped (CR 104.3), and every opponent's "you win the game"
//     is prevented (the log says so, once);
//   - the moment it leaves, the next state-based check reads the board
//     again: a controller at 0 or less life loses with no window to
//     respond (the Abyssal Persecutor ruling), while an empty-library
//     draw made under the Herald is forgotten (CR 704.5b reads only
//     draws since the last check);
//   - a Herald that has lost its abilities gates nothing, and one that
//     changes controller gates for its new controller.
//
// "Can't" is not a replacement (CR 614.17): nothing is skipped or
// rewritten, the loss simply doesn't happen. A player behind the Herald
// can pay life only up to what they have — the engine already refuses
// the rest — and can always concede (CR 104.3a).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "4080f7e3-06d3-4d3d-9929-4e826cb66713",
		Name:            "Herald of Eternal Dawn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		GameEndGates:    YouCantLoseOpponentsCantWin(),
	})
}
