package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disciple of Deceit — Creature — Human Rogue {U}{B}, 1/3 (Journey into Nyx,
// S82 OLD, #5558):
//
//	"Inspired — Whenever this creature becomes untapped, you may discard a
//	 nonland card. If you do, search your library for a card with the same mana
//	 value as that card, reveal it, put it into your hand, then shuffle."
//
// The untap watch is Deepwater Hypnotist's. The discard is a pick over the
// nonland cards in your hand (none, or declining, does nothing); the mana value
// is read from the card in hand before it is discarded, and the search follows
// only when the card really reached the graveyard (Transmute's search).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34ace274-8401-4e3c-89cd-75fe2337bf7d",
		Name:         "Disciple of Deceit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, effects.Self,
				"Disciple of Deceit — you may discard a nonland card; if you do, search your library for a card with the same mana value (inspired)",
				func(g *game.Game, item *game.StackItem) error {
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Hand == nil {
						return nil
					}
					var candidates []uuid.UUID
					for _, c := range p.Hand.Cards {
						if !c.IsLand() {
							candidates = append(candidates, c.InstanceID)
						}
					}
					if len(candidates) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Disciple of Deceit — you may discard a nonland card",
						Cards:    candidates,
						Min:      0,
						Max:      1,
						LiveZone: game.ZoneHand,
						ThenKey:  discipleOfDeceitDiscard,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

// discipleOfDeceitDiscard discards the pick and hands its mana value to the
// search that follows. An on-disk key.
var discipleOfDeceitDiscard = effects.AskedPick("choose-cards/s82-old-b371-disciple-of-deceit-discard",
	func(ctx *effects.Context, r game.CardsChosen) error {
		if len(r.Picked) != 1 {
			return nil
		}
		card, ok := ctx.Game.LookupCardForEffect(r.Picked[0])
		if !ok {
			return nil
		}
		mv, known := card.ParsedManaValue()
		if !known {
			mv = card.ManaValue()
		}
		return ctx.Game.DiscardCardsKeyedForEffect(r.Chooser, r.Picked, r.Source, game.KeyedTail{
			Ref: discipleOfDeceitSearch, Player: r.Chooser, Source: r.Source, N: mv,
		})
	})

// discipleOfDeceitSearch is the "if you do": the search for a card with the
// same mana value, only once the discard landed. An on-disk key.
var discipleOfDeceitSearch = effects.CardTail("then/s82-old-b371-disciple-of-deceit-search", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	mv := o.Tail.N
	return effects.SearchLibrary{
		Player: o.Tail.Player,
		Predicate: func(c game.Card) bool {
			cmv, ok := c.ParsedManaValue()
			return ok && cmv == mv
		},
		Dest:    game.ZoneHand,
		Limit:   1,
		Reveal:  true,
		Shuffle: true,
		Reason:  "Disciple of Deceit — a card with mana value " + effects.NumberWord(mv) + ", to your hand",
		Source:  o.Tail.Source,
	}.Apply(ctx)
})
