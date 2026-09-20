package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_swing.go — the shared vocabulary for clauses whose size is
// "the amount of life you gained this turn" or "the amount of life
// you lost this turn" (#1117, the Betor deck).
//
// Both numbers live on Game.TurnTally, counted off EventChangeLife as
// it happens (#586), so nothing here walks the event log. Read them
// through g.TurnTallyFor(player); a scan bounded by the upkeep event
// misses the untap step, which is the bug the tally replaced.
//
// Append to this file; do not change what is already in it.

// manaValueWithinLifeSwing is the shape behind both predicates below:
// "mana value less than or equal to <some running per-turn total>".
//
// The bound is recomputed on every snapshot, so a picker opened
// before a drain and answered after it offers the wider set — which
// is what the printed card says, since the comparison is made as the
// target is chosen and again as the ability resolves (CR 608.2b).
//
// A card whose cost the engine cannot read never passes, matching
// ManaValueLE rather than treating it as zero.
func manaValueWithinLifeSwing(bound func(game.PlayerTurnTally) int) CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		mv, ok := g.ManaValueForEffect(c)
		return ok && mv <= bound(g.TurnTallyFor(caster))
	}
}

// ManaValueAtMostLifeLostThisTurn — "mana value less than or equal to
// the amount of life you lost this turn" (Betor, Ancestor's Voice).
// "You" is the ability's controller, which is the player the
// predicate is handed.
func ManaValueAtMostLifeLostThisTurn() CardPredicate {
	return manaValueWithinLifeSwing(func(t game.PlayerTurnTally) int { return t.LifeLost })
}

// ManaValueAtMostLifeGainedThisTurn — "mana value X or less, where X
// is the amount of life you gained this turn" (Rodolf Duskbringer).
func ManaValueAtMostLifeGainedThisTurn() CardPredicate {
	return manaValueWithinLifeSwing(func(t game.PlayerTurnTally) int { return t.LifeGained })
}

// drainTargetedOpponent is the body of "target opponent loses that
// much life": the amount is captured by value when the trigger is
// built, so a life total that moves again before it resolves does not
// change the drain.
//
// It is life LOSS, not damage, so prevention and damage doublers
// never see it — the CR 614 life window still runs.
func drainTargetedOpponent(amount int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if amount <= 0 {
			return nil
		}
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetPlayer {
				continue
			}
			if err := g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -amount); err != nil {
				return err
			}
		}
		return nil
	}
}
