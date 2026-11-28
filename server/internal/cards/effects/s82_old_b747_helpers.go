package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b747_helpers.go — shared pieces of S82 OLD batch 747 (#5627).

// DamageToPlayerOrWalkerAndTheirCreatures is "<this> deals N damage to target
// player or planeswalker and each creature that player or that planeswalker's
// controller controls" (Flame Wave): the player is the target, or the
// controller of the targeted planeswalker, read before any damage lands; the
// creatures are snapshotted first and every point is one simultaneous damage
// instance (CR 608.2). With its one target illegal the spell does nothing
// (CR 608.2b).
func DamageToPlayerOrWalkerAndTheirCreatures(ctx *Context, n int) error {
	targets := ctx.LegalTargets()
	if n <= 0 || len(targets) == 0 {
		return nil
	}
	t := targets[0]
	player := t.ID
	if t.Kind == game.TargetCard {
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		player = c.Controller
	}
	var creatures []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, Creature()) {
		if c.Controller == player {
			creatures = append(creatures, c.InstanceID)
		}
	}
	return ctx.Game.DamageInstanceForEffect(func() error {
		if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}).Apply(ctx); err != nil {
			return err
		}
		for _, id := range creatures {
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: n}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// DamageToTargetEqualToCreaturesYouControl is "<this> deals damage to target
// player or planeswalker equal to the number of creatures you control" (Mob
// Justice): the creatures are counted as the spell resolves (CR 608.2b), so
// one that died in response is not counted.
func DamageToTargetEqualToCreaturesYouControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller == item.Controller {
			n++
		}
	}
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
	}
	return nil
}

// DestroyTheDamagedCreatureAtEndOfCombat is the trigger "whenever this
// creature deals damage to a creature, destroy that creature at end of combat"
// (Lowland Basilisk): the damaged creature is recorded as the object it is
// (CR 400.7) as the trigger goes on the stack, and the destruction is a
// delayed trigger at the beginning of the end of combat step (CR 511.2) that
// destroys that object if it is still on the battlefield.
func DestroyTheDamagedCreatureAtEndOfCombat(label string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventDealDamage},
		AppliesTo: ThisDealtDamageToACreature,
		Key:       label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Object = DamagedCreatureRef(g, ev.Target)
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			ref := item.Params.Object
			if ref.ID == uuid.Nil {
				return nil
			}
			return ScheduleDelayedTrigger{
				At:     game.StepEndCombat,
				Label:  label,
				Cards:  []uuid.UUID{ref.ID},
				Body:   DestroyTheObjectBody,
				Params: game.EffectParams{Object: ref},
			}.Apply(NewContext(g, item))
		},
	}
}

// ThisIsDealtDamageReflected is the trigger "whenever this creature is dealt
// damage, it deals that much damage to target opponent or planeswalker" (Mogg
// Maniac), and with combatOnly the "dealt combat damage" of Wall of Souls. It
// is Brash Taunter's shape waiting for the damage to finish, so the TOTAL
// dealt together is "that much" (CR 120.3, CR 510.2), aimed at an opponent or
// planeswalker chosen as the trigger goes on the stack.
func ThisIsDealtDamageReflected(key string, combatOnly bool) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return (!combatOnly || ev.Combat) && B35SelfWasDealtDamage(ev, source)
		},
		Targets:    TargetOpponentOrPlaneswalker(),
		Key:        key,
		AtBatchEnd: ThatMuchDamage(),
		Effect:     DamagedCreatureReflectsChosenAmount,
	}
}
