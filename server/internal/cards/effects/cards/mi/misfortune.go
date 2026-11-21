package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Misfortune — Sorcery {1}{B}{R}{G}:
//
//	"An opponent chooses one —
//	 • You put a +1/+1 counter on each creature you control and gain 4
//	   life.
//	 • You put a -1/-1 counter on each creature that player controls and
//	   Misfortune deals 4 damage to that player."
//
// The opponent chooses the bullet as the spell is cast (CR 700.2e,
// OpponentChoosesOne), and the caster picks which opponent when there is
// more than one. "That player" is that opponent (ctx.ModeChooser). Each
// creature set is read as the spell resolves, before the first counter
// goes on (CR 608.2c); Misfortune, the spell, is the damage's source.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e282eb5b-f9e6-4fb3-9edf-44e23359357d",
		Name:         "Misfortune",
		Completeness: effects.CompletenessFull,
		Modes: effects.OpponentChoosesOne(
			effects.ModeWithPurpose(effects.Mode("You put a +1/+1 counter on each creature you control and gain 4 life."), game.Purpose{LifeGain: 4}),
			// "That player" is an opponent: a one-sided -1/-1 (ADR 0126's
			// OpponentsOnly, as "target player controls" is).
			effects.ModeWithPurpose(effects.Mode("You put a -1/-1 counter on each creature that player controls and Misfortune deals 4 damage to that player."),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}}),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				if err := misfortuneCounters(ctx, ctx.Controller(), game.CounterPlusOne); err != nil {
					return err
				}
				return effects.GainLife{Player: ctx.Controller(), Amount: 4}.Apply(ctx)
			}
			if !ctx.HasMode(1) {
				return nil
			}
			that := ctx.ModeChooser()
			if ctx.Game.PlayerByIDForEffect(that) == nil {
				return nil
			}
			if err := misfortuneCounters(ctx, that, game.CounterMinusOne); err != nil {
				return err
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: that, Amount: 4}.Apply(ctx)
		},
	})
}

// misfortuneCounters puts one counter of `kind` on each creature `player`
// controls, the set read before the first one goes on.
func misfortuneCounters(ctx *effects.Context, player uuid.UUID, kind string) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller == player {
			ids = append(ids, c.InstanceID)
		}
	}
	group := ctx.AsGroupMember()
	for _, id := range ids {
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (effects.AddCounter{Target: id, Kind: kind, N: 1}).Apply(group); err != nil {
			return err
		}
	}
	return nil
}
