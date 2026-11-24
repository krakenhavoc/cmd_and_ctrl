package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Omnath, Locus of Creation — Legendary Creature — Elemental {R}{G}{W}{U},
// 4/4:
//
//	"When Omnath enters, draw a card.
//	 Landfall — Whenever a land you control enters, you gain 4 life if this is
//	 the first time this ability has resolved this turn. If it's the second
//	 time, add {R}{G}{W}{U}. If it's the third time, Omnath deals 4 damage to
//	 each opponent and each planeswalker you don't control."
//
// The count is Elrond's: the per-turn tally of this Omnath's resolutions of
// this ability (Game.ResolvedThisTurn), which includes the one resolving, per
// object (CR 400.7), so an Omnath that left and came back starts again. Each
// resolution does exactly one of the three; a fourth or later does nothing. The
// third deals its 4 damage as one instruction (CR 120) to every opponent and
// every planeswalker its controller doesn't control, Omnath being the source
// whether or not it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd133d30-51ff-4114-a7d7-029345f0f0d7",
		Name:         "Omnath, Locus of Creation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Omnath, Locus of Creation — draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.Landfall(omnathCreationLabel, omnathCreationLandfall),
		},
	})
}

const omnathCreationLabel = "Omnath, Locus of Creation — gain 4 life the first time this resolves each turn; add {R}{G}{W}{U} the second; 4 damage to each opponent and each planeswalker you don't control the third (landfall)"

func omnathCreationLandfall(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	switch g.ResolvedThisTurn(item.SourceCardID, omnathCreationLabel) {
	case 1:
		return effects.GainLife{Amount: 4}.Apply(ctx)
	case 2:
		return effects.AddMana{Player: ctx.Controller(), Produced: "{R}{G}{W}{U}"}.Apply(ctx)
	case 3:
		var targets []game.Card
		for _, c := range g.Battlefield.Cards {
			if c.IsPlaneswalker() && c.Controller != ctx.Controller() {
				targets = append(targets, c)
			}
		}
		return g.DamageInstanceForEffect(func() error {
			for _, opp := range ctx.Opponents() {
				if err := (effects.DealDamage{Source: item.SourceCardID, Target: opp, Amount: 4}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, c := range targets {
				if err := (effects.DealDamage{Source: item.SourceCardID, Target: c.InstanceID, Amount: 4}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return nil
}
