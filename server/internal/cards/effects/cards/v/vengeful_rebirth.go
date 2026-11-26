package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vengeful Rebirth — Sorcery {4}{R}{G} (Commander 2011, #5574):
//
//	"Return target card from your graveyard to your hand. If you return a
//	 nonland card to your hand this way, Vengeful Rebirth deals damage equal
//	 to that card's mana value to any target.
//	 Exile Vengeful Rebirth."
//
// Two target clauses, each judged on its own at resolution (CR 608.2b): with
// the card gone nothing is returned and no damage is dealt; with the damage
// target gone the card is still returned. The damage is dealt only if the
// card really reached the hand and is not a land, and its mana value is read
// from the card in the hand. The spell then exiles itself instead of going to
// the graveyard, even when the rest of it did nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "522297b1-493a-4c62-a82a-3017c44183e8",
		Name:         "Vengeful Rebirth",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn()),
			effects.TargetAny(),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			card, ok := ctx.ClauseTarget(0)
			if ok && card.Kind == game.TargetCard {
				if err := (effects.ReturnFromGraveyard{Target: card.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
				if z := g.FindCardZoneForEffect(card.ID); z != nil && z.Kind == game.ZoneHand {
					if c, found := g.LookupCardForEffect(card.ID); found && !c.IsLand() {
						if dmg, ok := ctx.ClauseTarget(1); ok {
							if mv, known := g.ManaValueForEffect(c); known && mv > 0 {
								if err := (effects.DealDamage{Source: item.SourceCardID, Target: dmg.ID, Amount: mv}).Apply(ctx); err != nil {
									return err
								}
							}
						}
					}
				}
			}
			if effects.SourceIsNewObject(g, item) {
				return nil
			}
			return g.ExileCardForEffect(ctx.Source())
		},
	})
}
