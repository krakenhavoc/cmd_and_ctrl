package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gravegouger — Creature — Nightmare Horror {2}{B}, 3/3:
//
//	"When this creature enters, exile up to two target cards from a
//	 single graveyard.
//	 When this creature leaves the battlefield, return the exiled cards
//	 to their owner's graveyard."
//
// #1807, ADR 0106 §5. The two abilities are linked (CR 607.2a): "the
// exiled cards" are the ones the entry trigger exiled, and only while
// they are still in exile. b27ExiledWith reads that record off the
// event log by the entry trigger's label, the way Angel of Serenity's
// leave trigger does, and a card that has since left exile is not in
// it. The leave trigger reads the record for the incarnation that just
// left (CR 603.10a), and a Gravegouger that comes back is a new object
// with an empty record (CR 400.7).
//
// If Gravegouger leaves before its entry trigger resolves, the leave
// trigger resolves first with nothing to return, and the cards exiled
// afterwards stay exiled — as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2677c0d9-1251-478e-b157-072859c593a9",
		Name:         "Gravegouger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersExileFromASingleGraveyard("Gravegouger", 2, ExileTargetCards),
			On(game.EventLTB, Self, "Gravegouger — return the exiled cards to their owner's graveyard",
				func(g *game.Game, item *game.StackItem) error {
					return g.PutCardsIntoGraveyardThenForEffect(b27ExiledWith(g, item.SourceCardID, gravegougerExileLabel), nil)
				}),
		},
	})
}

// gravegougerExileLabel is the entry trigger's label, which is the key
// the leave trigger reads the linked exile by. It is spelled exactly as
// WhenThisEntersExileFromASingleGraveyard builds it.
const gravegougerExileLabel = "Gravegouger — exile up to two target cards from a single graveyard"
