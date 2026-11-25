package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aria of Flame — Enchantment {2}{R} (Modern Horizons):
//
//	"When this enchantment enters, each opponent gains 10 life.
//	 Whenever you cast an instant or sorcery spell, put a verse counter on
//	 this enchantment, then it deals damage equal to the number of verse
//	 counters on it to target player or planeswalker."
//
// The enters trigger gives each opponent still in the game 10 life. The cast
// trigger fires on the controller's own instant and sorcery spells, targets as
// it goes on the stack (CR 603.3d), and as it resolves puts the counter first
// and then reads the count (the counter just added included; last-known if
// the enchantment has gone, CR 608.2h) for the damage the enchantment itself
// deals. An enchantment that has left the battlefield by then adds no counter
// and deals no damage for a count of zero.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "406fad02-e1a6-4c24-b46d-976c548e1c78",
		Name:         "Aria of Flame",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Aria of Flame — each opponent gains 10 life", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, opp := range ctx.Opponents() {
					if err := (effects.GainLife{Player: opp, Amount: 10}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.Targeting(effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.InstantOrSorceryCastByYou(ev, source, g)
			}, "Aria of Flame — put a verse counter on it, then it deals damage equal to the number of verse counters on it to target player or planeswalker",
				ariaOfFlameVerse), effects.TargetPlayerOrPlaneswalker()),
		},
	})
}

func ariaOfFlameVerse(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if effects.SourceIsStillThisPermanent(g, item) {
		if err := (effects.AddCounter{Target: item.SourceCardID, Kind: "verse", N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	n := effects.CountersOnThisAsItResolves("verse")(ctx)
	if n <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard || t.Kind == game.TargetPlayer {
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
		}
	}
	return nil
}
