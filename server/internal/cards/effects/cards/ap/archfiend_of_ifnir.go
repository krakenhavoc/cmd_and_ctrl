package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archfiend of Ifnir — Creature — Demon {3}{B}{B}, 5/4:
//
//	"Flying
//	 Whenever you cycle or discard another card, put a -1/-1 counter
//	 on each creature your opponents control.
//	 Cycling {2}"
//
// "Cycle or discard" is ONE event, the discard. Cycling's cost is to
// discard the card, so a cycled card is a discarded card (CR 702.29d)
// and the engine emits EventDiscardCard for it; EventCycle follows,
// and watching both would count one cycle twice, which is stronger
// than printed. The trigger watches the discard alone, which catches
// a cycled card and every other discard. It is one trigger per card,
// as printed: a three-card discard puts three counters on each
// creature.
//
// "Another card" rules out Archfiend discarding itself. It cannot be
// on the battlefield and in hand at once, so it only matters for the
// engine's own bookkeeping, and is checked all the same.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "535e2af0-7b08-4026-b1bf-87626407dd42",
		Name:            "Archfiend of Ifnir",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source) && ev.CardID != source.InstanceID
			}, "Archfiend of Ifnir — a -1/-1 counter on each creature your opponents control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.OpponentControls(), effects.Creature())) {
						if err := (effects.AddCounter{Target: c.InstanceID, Kind: game.CounterMinusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
