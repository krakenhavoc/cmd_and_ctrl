package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Devourer of Memory — Creature — Nightmare {U}{B}, 2/1 (Theros Beyond Death,
// S82 OLD, #5514):
//
//	"Whenever one or more cards are put into your graveyard from your library,
//	 this creature gets +1/+1 until end of turn and can't be blocked this turn.
//	 {1}{U}{B}: Mill a card."
//
// Sidisi, Brood Tyrant's library-to-graveyard watcher (a mill, a surveil, a
// search that puts a card there: every arrival route, each counted once)
// batched so a mill of several cards is one trigger (CR 603.2c). The pump and
// the unblockable are both this creature's, until the end of the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f747b39a-28a0-46a8-a161-182939556d39",
		Name:         "Devourer of Memory",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.OnAny(effects.LibraryToGraveyardKinds, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CardPutIntoYourGraveyardFromYourLibrary(ev, source, g, func(game.Card) bool { return true })
			}, "Devourer of Memory — this creature gets +1/+1 until end of turn and can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1, Label: "Devourer of Memory — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: "Devourer of Memory — can't be blocked"}.Apply(ctx)
				})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}{B}: Mill a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{1}{U}{B}"),
			Effect:  effects.Do(effects.MillCards{N: 1}),
		}},
	})
}
