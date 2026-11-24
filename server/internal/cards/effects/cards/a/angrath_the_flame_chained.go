package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angrath, the Flame-Chained — Legendary Planeswalker — Angrath {3}{B}{R},
// loyalty 4:
//
//	"+1: Each opponent discards a card and loses 2 life.
//	 −3: Gain control of target creature until end of turn. Untap it. It gains
//	 haste until end of turn. Sacrifice it at the beginning of the next end
//	 step if it has mana value 3 or less.
//	 −8: Each opponent loses life equal to the number of cards in their
//	 graveyard."
//
// −3 steals, untaps and hastes through the shared Threaten shape, then
// schedules a delayed trigger for the next end step (CR 603.7). The mana
// value is read as that trigger resolves, from the creature as it is then
// (a face-down or changed creature is judged as it is now), and the
// sacrifice happens only if the same object is still on the battlefield.
// −8 reads each opponent's graveyard as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "092ceffb-1ee4-452c-b15e-5b79267d42a6",
		Name:            "Angrath, the Flame-Chained",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Each opponent discards a card and loses 2 life.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := effects.EachOpponentDiscardsOne(g, item); err != nil {
						return err
					}
					return effects.EachOpponentLosesLifeNow(g, item, 2)
				},
			},
			{
				Label:   "−3: Gain control of target creature until end of turn. Untap it. It gains haste until end of turn. Sacrifice it at the beginning of the next end step if it has mana value 3 or less.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := effects.ThreatenAndGrant(ctx, id, "Angrath, the Flame-Chained", "haste"); err != nil {
						return err
					}
					c, ok := effects.BattlefieldCardForEffect(g, id)
					if !ok {
						return nil
					}
					return effects.ScheduleDelayedTrigger{
						At:         game.StepEnd,
						Controller: ctx.Controller(),
						Label:      "Angrath, the Flame-Chained — sacrifice it if it has mana value 3 or less",
						Body:       angrathSacrificeBody,
						Params:     game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
					}.Apply(ctx)
				},
			},
			{
				Label: "−8: Each opponent loses life equal to the number of cards in their graveyard.",
				Cost:  effects.LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, opp := range ctx.Opponents() {
						p := g.PlayerByIDForEffect(opp)
						if p == nil {
							continue
						}
						if n := p.Graveyard.Size(); n > 0 {
							if err := g.ChangePlayerLifeForEffect(ctx.Source(), opp, -n); err != nil {
								return err
							}
						}
					}
					return nil
				},
			},
		},
	})
}

// angrathSacrificeBody is the delayed "sacrifice it if it has mana value 3 or
// less". An on-disk key: never renamed, never reused.
var angrathSacrificeBody = game.DelayedBody("angrath-the-flame-chained/sacrifice-if-mv-3", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	c, ok := g.LookupCardForEffect(p.Object.ID)
	if !ok || !effects.OnBattlefield(g, p.Object.ID) || c.ObjectEpoch != p.Object.Epoch || c.ManaValue() > 3 {
		return nil
	}
	return effects.SacrificePermanent{Target: p.Object.ID}.Apply(effects.NewContext(g, item))
})
