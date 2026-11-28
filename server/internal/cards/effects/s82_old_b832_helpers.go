package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// PreventNextOneToYouB832 is "Prevent the next 1 damage that would be dealt to
// you this turn" (Shield of the Ages; S82 OLD batch 832, #5636).
func PreventNextOneToYouB832(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return PreventNextDamage{Target: ctx.Controller(), Amount: 1, Label: "Shield of the Ages: prevent the next 1 damage"}.Apply(ctx)
}

// PreventAllCombatDamageB832 is "Prevent all combat damage that would be dealt
// this turn."
func PreventAllCombatDamageB832(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return PreventAllCombatDamageThisTurn{Label: name + " — prevent all combat damage this turn"}.Apply(NewContext(g, item))
	}
}

// DuringDeclareBlockersB832 is "Activate only during the declare blockers
// step".
func DuringDeclareBlockersB832() ActivationCondition {
	return func(g *game.Game, _, _ uuid.UUID) bool { return g.Turn.Step == game.StepDeclareBlockers }
}

// JarkeldSwitchB832 is General Jarkeld's ability: of the two target blocked
// attacking creatures, if each could be blocked by all the creatures the other
// is blocked by, every creature blocking exactly one of them stops blocking it
// and blocks the other (CR 509.1h, #6193).
func JarkeldSwitchB832(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) != 2 {
		return nil
	}
	g.SwitchBlockedAttackersForEffect(ids[0], ids[1])
	return nil
}

// SappersUnblockableB832 is "Target creature you control can't be blocked this
// turn. Destroy it [and this creature] at end of combat." withSelf adds the
// Sappers to the destruction.
func SappersUnblockableB832(withSelf bool) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return nil
		}
		if err := (RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Goblin Sappers — can't be blocked"}).Apply(ctx); err != nil {
			return err
		}
		ids := []uuid.UUID{id}
		refs := []game.ObjectRef{{ID: id, Epoch: c.ObjectEpoch}}
		if withSelf && item.SourceCardID != id {
			if self, ok := g.LookupCardForEffect(item.SourceCardID); ok && OnBattlefield(g, self.InstanceID) {
				ids = append(ids, self.InstanceID)
				refs = append(refs, game.ObjectRef{ID: self.InstanceID, Epoch: self.ObjectEpoch})
			}
		}
		for i := range ids {
			if err := (ScheduleDelayedTrigger{
				At:     game.StepEndCombat,
				Label:  "Goblin Sappers — destroy it at end of combat",
				Cards:  []uuid.UUID{ids[i]},
				Body:   DestroyTheObjectBody,
				Params: game.EffectParams{Object: refs[i]},
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// AmuletOfQuozB832 is the Amulet's ability with its ante text inert (#5724,
// CR 407): the opponent "may ante the top card of their library" and, there
// being no ante, doesn't, so the coin is flipped; if you win the flip that
// player loses the game, if you lose it you lose the game.
func AmuletOfQuozB832(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opp = t.ID
		}
	}
	if opp == uuid.Nil {
		return nil
	}
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Amulet of Quoz — call the coin flip",
		ThenKey:  amuletOfQuozFlipped,
		Asked:    game.NewAskedItem(item),
		Carry:    []uuid.UUID{opp},
	})
	return nil
}

var amuletOfQuozFlipped = CoinFlipThen("coin-flip/old-b832-amulet-of-quoz", func(ctx *Context, r game.CoinFlipped) error {
	if len(r.Carry) == 0 {
		return nil
	}
	if WonFirstFlip(r) {
		return LoseTheGame{Player: r.Carry[0]}.Apply(ctx)
	}
	return LoseTheGame{}.Apply(ctx)
})
