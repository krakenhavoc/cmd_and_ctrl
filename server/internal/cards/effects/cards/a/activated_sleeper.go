package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Activated Sleeper — Creature — Phyrexian Shapeshifter {2}{B}, 0/0
// (Dominaria United Commander):
//
//	"Flash
//	 You may have this creature enter as a copy of any creature card in a
//	 graveyard that was put there from the battlefield this turn, except it's
//	 a Phyrexian in addition to its other types."
//
// Body Double's graveyard copy (CR 614.1c, CR 707.2) with the candidates
// narrowed by Samwise the Stouthearted's reading of "put there from the
// battlefield this turn": the card's newest arrival in its graveyard happened
// this turn and came from the battlefield, so a milled, discarded or
// countered card, or one that died last turn, is not offered. The copy takes
// the card's printed values, which the card keeps; "a Phyrexian in addition"
// is an ADD (CR 707.9b). Declining, or a pick that left the graveyard before
// the answer, leaves a 0/0 that dies to CR 704.5f.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e12da2a1-e390-4487-b154-69e439cb7b05",
		Name:            "Activated Sleeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Activated Sleeper", activatedSleeperCandidates,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, src *game.Card) {
					v.AddSubtype(src.TextSubtype("Phyrexian"))
				}),
		},
	})
}

// activatedSleeperCandidates is every creature card in a graveyard that was put
// there from the battlefield this turn.
func activatedSleeperCandidates(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range effects.AnyCreatureCardInAGraveyard(g, controller, self) {
		if effects.PutIntoGraveyardFromBattlefieldThisTurn(g, id) {
			out = append(out, id)
		}
	}
	return out
}
