package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shocker, Unshakable — Legendary Creature — Human Rogue Villain
// {4}{R}{R}, 5/5 (Marvel's Spider-Man):
//
//	"During your turn, Shocker has first strike.
//	 Vibro-Shock Gauntlets — When Shocker enters, he deals 2 damage to
//	 target creature and 2 damage to that creature's controller."
//
// First strike is the layer 6 grant read on whose turn it is (Triumphant
// Adventurer). The enters trigger has one target; the creature's controller
// is read before the damage, while it is still on the battlefield (Minion
// Missile), so a creature the 2 damage kills still has its controller hit.
// If the target is gone as the trigger resolves, it does nothing at all
// (CR 608.2b). Both damage events come from Shocker by his object id, so
// last-known information is used if he has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "564dac59-97f7-4988-9f30-d82a891e29ef",
		Name:         "Shocker, Unshakable",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "first strike"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Shocker, Unshakable — 2 damage to target creature and 2 damage to that creature's controller",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						victim, ok := g.LookupCardForEffect(id)
						if !ok {
							return nil
						}
						controller := victim.Controller
						if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 2}).Apply(ctx); err != nil {
							return err
						}
						return effects.DealDamage{Source: item.SourceCardID, Target: controller, Amount: 2}.Apply(ctx)
					}),
				effects.TargetCreature("target creature")),
		},
	})
}
