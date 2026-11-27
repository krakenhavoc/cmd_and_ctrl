package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Infernal Kirin — Legendary Creature — Kirin Spirit {2}{B}{B}, 3/3
// (Saviors of Kamigawa, #5601):
//
//	"Flying
//	 Whenever you cast a Spirit or Arcane spell, target player reveals
//	 their hand and discards all cards with that spell's mana value."
//
// The cast event's Amount is the spell's mana value as cast (CR 202.3).
// The hand is revealed (CR 701.20a), then every card in it with exactly
// that mana value is discarded: nothing is chosen, so no prompt is asked
// (Shatter Assumptions' shape). A hand with none is revealed and nothing
// is discarded (CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4f07a2e6-692b-4d8c-9ecf-4d94b03706a0",
		Name:            "Infernal Kirin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), "Infernal Kirin — target player reveals their hand and discards all cards with that spell's mana value",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok || item.Trigger == nil {
						return nil
					}
					p := g.PlayerByIDForEffect(player)
					if p == nil {
						return nil
					}
					g.RevealHandForEffect(player)
					mv := item.Trigger.Event.Amount
					var ids []uuid.UUID
					for _, c := range p.Hand.Cards {
						if v, ok := g.ManaValueForEffect(c); ok && v == mv {
							ids = append(ids, c.InstanceID)
						}
					}
					if len(ids) == 0 {
						return nil
					}
					return g.DiscardCardsKeyedForEffect(player, ids, ctx.Source(), game.KeyedTail{})
				}),
				effects.TargetPlayer("target player")),
		},
	})
}
