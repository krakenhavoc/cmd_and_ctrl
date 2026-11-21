package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Too Greedily, Too Deep — Sorcery {5}{B}{R} (Tales of Middle-earth
// Commander):
//
//	"Put target creature card from a graveyard onto the battlefield under your
//	 control. That creature deals damage equal to its power to each other
//	 creature."
//
// Reanimate's shape for any graveyard (the card enters under your control),
// then Nova Flame's single damage event: the creature the card became is the
// source (so its deathtouch or lifelink applies) and its power is read as the
// instruction runs, after it entered with whatever counters or bonuses it
// has. A target that is no longer a legal card in a graveyard when the spell
// resolves is not returned and nothing is dealt (CR 608.2b); the creature
// the spell itself returned is not hit by its own damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af65bbf9-e328-4b8b-8caa-c6dff743867b",
		Name:         "Too Greedily, Too Deep",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		Targets:      effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.ReturnFromGraveyardTogether{
				Targets:    []uuid.UUID{id},
				Controller: ctx.Controller(),
				ThenTail:   s82GapB120TooGreedilyDamage,
			}.Apply(ctx)
		},
	})
}

// s82GapB120TooGreedilyDamage is "that creature deals damage equal to its
// power to each other creature", told the new object the card became.
var s82GapB120TooGreedilyDamage = effects.CardTail("too-greedily-too-deep/damage", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	id := o.IDs[0]
	c, found := ctx.Game.LookupCardForEffect(id)
	if !found || !c.IsCreature() || c.CurrentPower() <= 0 {
		return nil
	}
	power := c.CurrentPower()
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, other := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.NotSelf(id))) {
			if err := (effects.DealDamage{Source: id, Target: other.InstanceID, Amount: power}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
})
