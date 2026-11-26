package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roiling Terrain — Sorcery {2}{R}{R} (Worldwake, S82 OLD b488):
//
//	"Destroy target land, then Roiling Terrain deals damage to that land's
//	 controller equal to the number of land cards in that player's graveyard."
//
// Minion Missile's order: the land's controller is read before the destroy, while it is
// still on the battlefield (CR 608.2h), and the land cards are counted AFTER it, so the land
// just destroyed is in its owner's graveyard when that is the same player. An indestructible
// land is not destroyed and its controller is dealt the count all the same; a target that
// became illegal counters the spell (CR 608.2b) and nothing is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e90f187-42fe-4f32-ad75-302e6709192d",
		Name:         "Roiling Terrain",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, ref := range ctx.LegalTargets() {
				if ref.Kind != game.TargetCard {
					continue
				}
				victim, ok := ctx.Game.LookupCardForEffect(ref.ID)
				if !ok {
					return nil
				}
				if err := (effects.DestroyTarget{Target: ref.ID}).Apply(ctx); err != nil {
					return err
				}
				n := effects.B17LandCardsInGraveyard(ctx.Game, victim.Controller)
				return effects.DealDamage{Source: ctx.Source(), Target: victim.Controller, Amount: n}.Apply(ctx)
			}
			return nil
		},
	})
}
