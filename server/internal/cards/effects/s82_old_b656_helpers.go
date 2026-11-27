package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b656_helpers.go — Invasion, batch 656 (S82, #5616).

// CostsMoreColored is "<matching> spells cost <unit> more to cast" for a
// coloured unit (Sapphire Leech's "Blue spells you cast cost {U} more to
// cast", Ruby Leech's {R}): one point of the increase adds the unit's mana
// to the cost, coloured symbols included (ADR 0048 addendum, open question
// 3). Panics at init on a unit that is not a fixed mana amount.
func CostsMoreColored(unit, label string, when ...CostPredicate) game.CostModifier {
	u, err := game.ParseCost(unit)
	if err != nil || u.XSlots > 0 || u.ManaValue() == 0 {
		panic(fmt.Sprintf("effects.CostsMoreColored: %q is not a fixed mana amount (%v)", unit, err))
	}
	return game.CostModifier{
		Kind:      game.CostIncrease,
		Label:     label,
		AppliesTo: allOf(when),
		Amount:    func(game.CostQuery) int { return 1 },
		Unit:      &u,
	}
}

// BecomesBlockedByAWhiteCreature is Phyrexian Slayer's trigger condition:
// this creature became blocked by a white creature (EventBlock carries the
// blocker in CardID and the attacker in Target; one per blocker, CR 509.3d).
func BecomesBlockedByAWhiteCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !ViciousBattleragerBlocked(ev, source, game.Characteristic{}, g) {
		return false
	}
	blocker, ok := g.LookupCardForEffect(ev.CardID)
	return ok && blocker.HasColorWord("W")
}

// DestroyTheBlockerNoRegen is "destroy that creature. It can't be
// regenerated": the blocker the trigger names, if that object is still on the
// battlefield as the trigger resolves.
func DestroyTheBlockerNoRegen(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Object == nil {
		return nil
	}
	o := item.Trigger.Object
	info, ok := g.PermanentForEffect(game.ObjectRef{ID: o.ID, Epoch: o.Epoch})
	if !ok || info.Left {
		return nil
	}
	return DestroyTarget{Target: o.ID, CantBeRegenerated: true}.Apply(NewContext(g, item))
}
