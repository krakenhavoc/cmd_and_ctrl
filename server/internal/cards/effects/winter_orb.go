package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Winter Orb — Artifact for {2}:
//
//	"As long as this artifact is untapped, players can't untap more
//	 than one land during their untap steps."
//
// #826 / ADR 0070: the first card on the untap-step CAP. It held out
// for this ADR by owner decision (ADR 0058 question 3) rather than
// shipping with the engine picking a land, because under a Winter Orb
// lock which land untaps is the decision the card is about.
//
// The cap is a ceiling, not a restriction: with two or more tapped
// lands the untap step stops and asks the active player which one
// untaps. With one tapped land the cap does not bind and there is no
// prompt — CR 502.3's "normally, all of a player's permanents untap"
// still does the work, and the one land is untapped without a click.
//
// "As long as this artifact is untapped" is read at the instant the
// step asks, so a Winter Orb tapped on a previous turn (Icy
// Manipulator, a Fatestitcher) caps nothing.
//
// Everything that is not a land untaps as normal and is never offered:
// the cap counts lands, and the prompt only ever contains permanents a
// cap counts or a clause lets its controller hold back.
func init() {
	Register(Spec{
		OracleID:     "1dcbd583-3388-4b34-a7cd-131648aa6abd",
		Name:         "Winter Orb",
		Completeness: CompletenessFull,
		UntapCaps: []game.UntapCap{
			cantUntapMoreThanWhileSourceUntapped(
				"Winter Orb — no more than one land", 1, Land()),
		},
	})
}
