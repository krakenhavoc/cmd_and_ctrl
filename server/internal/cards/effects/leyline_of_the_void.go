package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline of the Void — Enchantment {2}{B}{B} (EDHREC rank 3464):
//
//	"If this card is in your opening hand, you may begin the game with
//	 it on the battlefield.
//	 If a card would be put into an opponent's graveyard from anywhere,
//	 exile it instead."
//
// Rest in Peace's one-sided half: your own graveyard still works, so
// this is the version a reanimator deck can play. The replacement is
// the same shape with one clause added (GraveyardBecomesExile with
// OpponentsOnly), reading the graveyard the move is aimed at — which
// is always the card's owner's.
//
// Cards only, unlike Rest in Peace: a token that would be put into an
// opponent's graveyard is not a card, so Leyline does not touch it.
// That is a distinction without a difference here — a token put into a
// graveyard ceases to exist as a state-based action either way
// (CR 111.8) — and the engine's exit route carries tokens through the
// same window, so the clause is deliberately not written as a filter.
// If a card ever reads "exiled by Leyline", this is the line to
// revisit.
//
// Two declared deviations:
//
//   - the leyline clause itself. There is no "begin the game with it
//     on the battlefield" step in this sandbox; the enchantment is
//     cast for {2}{B}{B} like any other.
//   - an opponent's COMMANDER whose owner takes CR 903.9's offer goes
//     to the command zone rather than to exile, the deviation Rest in
//     Peace, Liesa and Stone of Erech all share.
func init() {
	Register(Spec{
		OracleID:     "f4e32fc1-1b8d-441e-8e76-71f19f98e925",
		Name:         "Leyline of the Void",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You can't begin the game with it on the battlefield from your opening hand — it has to be cast.",
			"An opponent's dying commander whose owner sends it to the command zone goes there instead of being exiled.",
		},
		Replacements: []game.ReplacementEffect{GraveyardBecomesExile{
			OpponentsOnly: true,
			Label:         "Leyline of the Void: exile instead of an opponent's graveyard",
		}.Build()},
	})
}
