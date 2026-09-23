package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Morbid Opportunist — Creature — Human Rogue {2}{B}, 1/3
// (EDHREC rank 251):
//
//	"Whenever one or more other creatures die, draw a card. This
//	 ability triggers only once each turn."
//
// Three shared pieces, none of them new: `AnotherCreatureDied`
// (triggers_common.go) is the "any creature, anyone's, other than
// this one" predicate; `OncePerBatch` (#587) collapses a simultaneous
// death batch — a board wipe — into the one trigger the printed text
// describes; and `b11TriggeredThisTurn` (Exemplar of Light's "only
// once each turn" tally) gates the SECOND death-batch of the turn out
// entirely, which OncePerBatch alone does not do since it only
// dedupes within one batch.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "322f44f0-e6da-4ee0-b474-e7d5e9a461c5",
		Name:         "Morbid Opportunist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventLTB, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				if !AnotherCreatureDied(ev, source, ch, g) {
					return false
				}
				return !b11TriggeredThisTurn(g, source.InstanceID, morbidOpportunistDrawLabel)
			}, morbidOpportunistDrawLabel, Do(DrawCards{N: 1}))),
		},
	})
}

const morbidOpportunistDrawLabel = "Morbid Opportunist — draw a card"
