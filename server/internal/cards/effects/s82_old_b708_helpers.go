package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b708_helpers.go — the Portal Three Kingdoms damage spells (S82 #5622).

// OldB708DamageEachCreatureAndEachPlayer is "<this> deals X damage to each creature
// that passes `match` and each player": Earthquake's sweep with the creature
// filter as a parameter (Rolling Earthquake, Borrowing the East Wind). X is the
// announced X. The creatures are snapshotted before the first point lands, the
// keyword read post-layer, and then every seated player, the caster included,
// takes X, all as one damage instance. X of 0 deals nothing.
func OldB708DamageEachCreatureAndEachPlayer(match CardPredicate) func(item *game.StackItem, ctx *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		x := ctx.X()
		if x <= 0 {
			return nil
		}
		return ctx.Game.DamageInstanceForEffect(func() error {
			if err := DamageEachMatching(ctx, And(Creature(), match), x); err != nil {
				return err
			}
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Eliminated {
					continue
				}
				if err := (DealDamage{Source: ctx.Source(), Target: p.ID, Amount: x}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	}
}

// OldB708DamageToTargetOpponentOrPlaneswalker is "<this> deals N damage to target
// opponent or planeswalker": the clause's pick, if it is still legal as the spell
// resolves (CR 608.2b). It returns whether the damage was dealt, for the
// "you gain life" rider of Stolen Grain.
func OldB708DamageToTargetOpponentOrPlaneswalker(ctx *Context, n int) (bool, error) {
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return false, nil
	}
	return true, DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
}

// OldB708YouWereAttackedThisStep is "Activate only during the declare attackers
// step and only if you've been attacked this step" (Kongming's Contraptions), the
// activation twin of Deep Wood's cast condition.
func OldB708YouWereAttackedThisStep() ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		return YouWereAttackedThisStep(g, controller, game.Card{})
	}
}

// OldB708DestroyLands destroys the chosen lands one by one.
func OldB708DestroyLands(ctx *Context, ids []uuid.UUID) error {
	for _, id := range ids {
		if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// OldB708FourLandsOf is "<player> destroys four lands they control": it destroys
// them outright when the player has four or fewer (nothing to choose), and
// otherwise asks them to choose four, handing the pick to `key`.
func OldB708FourLandsOf(ctx *Context, player uuid.UUID, key game.ChooseCardsThen, question string, noChoice func(ctx *Context) error) error {
	lands := PermanentsControlledByMatching(ctx.Game, player, Land())
	if len(lands) <= 4 {
		if err := OldB708DestroyLands(ctx, lands); err != nil {
			return err
		}
		return noChoice(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     ctx.Source(),
		Question:   question,
		Cards:      lands,
		Min:        4,
		Max:        4,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    key,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}
