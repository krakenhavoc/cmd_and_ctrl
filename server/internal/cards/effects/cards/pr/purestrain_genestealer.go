package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Purestrain Genestealer — Creature — Tyranid {2}{G}, 1/1:
//
//	"This creature enters with two +1/+1 counters on it.
//	 Vanguard Species — Whenever this creature attacks, you may remove a
//	 +1/+1 counter from it. If you do, search your library for a basic land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// The counter is removed on the yes answer, as the cost of "if you do"; a
// Genestealer that has lost its counters, or left the battlefield, is not
// offered the choice and does nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d4c7f1a-deca-4bae-90c7-68d623da400e",
		Name:         "Purestrain Genestealer",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters("+1/+1", 2, "Purestrain Genestealer: enters with two +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WheneverThisAttacks("Purestrain Genestealer — you may remove a +1/+1 counter from it; if you do, search for a basic land", purestrainGenestealerAsk),
				game.Purpose{Lands: 1}),
		},
	})
}

func purestrainGenestealerHasCounter(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && effects.OnBattlefield(g, id) && c.Counters[game.CounterPlusOne] > 0
}

func purestrainGenestealerAsk(g *game.Game, item *game.StackItem) error {
	if !purestrainGenestealerHasCounter(g, item.SourceCardID) {
		return nil
	}
	return effects.MayChoice{
		Question: "Purestrain Genestealer — remove a +1/+1 counter to search for a basic land?",
		YesLabel: "Remove a counter",
		NoLabel:  "Do not",
		Then:     purestrainGenestealerAnswer,
	}.Apply(effects.NewContext(g, item))
}

var purestrainGenestealerAnswer = effects.MayYes("confirm/purestrain-genestealer-remove-counter", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	id := ctx.Source()
	if !purestrainGenestealerHasCounter(ctx.Game, id) {
		return nil
	}
	if err := ctx.Game.AddCounterForEffect(id, game.CounterPlusOne, -1); err != nil {
		return err
	}
	return effects.SearchLibrary{
		Player:        ctx.Controller(),
		Predicate:     effects.IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Shuffle:       true,
		TappedOnEntry: true,
		Source:        id,
		Reason:        "Purestrain Genestealer — search for a basic land card",
	}.Apply(ctx)
})
