package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Termination Facilitator — {1}{B} Creature — Human Assassin 1/3 (Jumpstart 2022, #5948):
//
//	"{T}: Put a bounty counter on target creature or planeswalker. Activate
//	 only as a sorcery.
//	 Whenever a creature or planeswalker an opponent controls with a bounty
//	 counter on it is dealt damage, destroy it."
//
// Bounty Board's counter at sorcery speed. The destroy trigger fires per
// damage event that dealt damage (after prevention) to a permanent an
// opponent of the Facilitator's controller controls that has a counter;
// it destroys the permanent only if that same object is still on the
// battlefield (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56edd306-8676-49c4-b3c3-5ff840536c3a",
		Name:         "Termination Facilitator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Put a bounty counter on target creature or planeswalker. Activate only as a sorcery.",
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.AddCounter{Target: t.ID, Kind: "bounty", N: 1}.Apply(ctx)
					}
				}
				return nil
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventDealDamage || ev.Amount <= 0 {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Target)
				if !ok || c.Controller == source.Controller || c.Counters["bounty"] <= 0 {
					return false
				}
				return c.IsCreature() || c.IsPlaneswalker()
			}, "Termination Facilitator — destroy it",
				effects.S82OldB091DestroyTheDamagedPermanent),
		},
	})
}
