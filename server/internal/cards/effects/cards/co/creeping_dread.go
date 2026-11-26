package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Creeping Dread — Enchantment {3}{B} (Shadows over Innistrad, #5543):
//
//	"At the beginning of your upkeep, each player discards a card. Each
//	 opponent who discarded a card that shares a card type with the card you
//	 discarded loses 3 life. (Players reveal the discarded cards
//	 simultaneously.)"
//
// Every player is asked to discard in turn order (Smallpox's fan-out; a
// player with an empty hand is not asked, CR 701.9a), and once the last
// discard has landed the continuation compares them. The types are those of
// the cards as they sit in the graveyards. An opponent loses the life only
// if both they and you discarded a card, and that card shares at least one
// card type with yours; the loss is a life loss, taken in seat order. The
// controller is the run's carry. An on-disk key.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bb16720-18dc-44b7-adb0-c9b7adda9ebe",
		Name:         "Creeping Dread",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Creeping Dread — each player discards a card; each opponent who discarded a card sharing a card type with yours loses 3 life",
				func(g *game.Game, item *game.StackItem) error {
					return g.EachPlayerDiscardsThenKeyedForEffect(uuid.Nil,
						game.DiscardPrompt{Source: item.SourceCardID, N: 1, Question: "Creeping Dread — discard a card"},
						creepingDreadCompare, []uuid.UUID{item.Controller})
				}),
		},
	})
}

var creepingDreadCompare = game.RegisterPromptRunThen("prompt-run/creeping-dread-compare", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	controller := r.Carry[0]
	types := func(ids []uuid.UUID) map[string]bool {
		out := map[string]bool{}
		for _, id := range ids {
			c, ok := g.LookupCardForEffect(id)
			if !ok {
				continue
			}
			_, ts, _ := game.ParseTypeLine(c.TypeLine)
			for _, t := range ts {
				out[t] = true
			}
		}
		return out
	}
	mine := types(r.Discards().By(controller))
	if len(mine) == 0 {
		return nil
	}
	for _, seat := range r.Landed {
		if seat.Seat == controller {
			continue
		}
		for t := range types(seat.Cards) {
			if mine[t] {
				if err := g.ChangePlayerLifeForEffect(r.Source, seat.Seat, -3); err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
})
