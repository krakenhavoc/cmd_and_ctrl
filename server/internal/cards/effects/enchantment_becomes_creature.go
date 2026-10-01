package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// enchantment_becomes_creature.go — the Mirage/Visions "hidden"
// enchantments (ADR 0107 PR 1, #1858): "When <state>, if this permanent
// is an enchantment, it becomes a <P>/<T> <Subtype> creature." Opal
// Avenger, Lurking Jackals, Veiled Crocodile.
//
// Each is a CR 603.8 state trigger with an intervening "if" (CR 603.4):
// the "is an enchantment" half is part of the trigger condition, so it is
// asked as the state is, and it is asked again as the ability resolves.
// "Becomes a … creature" with no "in addition to its other types" sets
// the card type (CR 205.1a), so the animated permanent is no longer an
// enchantment — and the intervening "if" is what stops the state, still
// true, from triggering it again. The effect has no duration, so it lasts
// as long as the permanent is this object (CR 611.2c, 400.7), as Lurking
// Evil's does.
//
// Append-only, per the shared-vocabulary rule.

// whenStateIfThisIsAnEnchantment is the trigger: `state` and the
// intervening "if this permanent is an enchantment", then the effect.
func whenStateIfThisIsAnEnchantment(label string, state StateCondition, effect Effect) game.TriggeredAbility {
	return WhenState(label, func(g *game.Game, source *game.Card, controller uuid.UUID) bool {
		return source.HasCardType("enchantment") && state(g, source, controller)
	}, effect)
}

// thisEnchantmentBecomesACreature is the effect: the CR 603.4 re-check,
// then the indefinite type change pinned to this object. The colour is
// untouched.
func thisEnchantmentBecomesACreature(subtype string, power, toughness int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		self := item.SourceCardID
		if !onBattlefield(g, self) || sourceIsNewObject(g, item) {
			return nil
		}
		c, ok := g.LookupCardForEffect(self)
		if !ok || !c.HasCardType("enchantment") {
			return nil
		}
		return ScopedEffectFor{
			Target: self,
			Mods: []game.Mod{
				game.RemoveTypesMod("Enchantment"),
				game.AddTypesMod("Creature"),
				game.AddSubtypesMod(subtype),
				game.SetBasePowerMod(power),
				game.SetBaseToughnessMod(toughness),
			},
			Duration: g.PinnedTo(game.IndefiniteDuration(), self),
			Label:    label,
		}.Apply(NewContext(g, item))
	}
}

// livePlayers are the players still in the game.
func livePlayers(g *game.Game) []*game.Player {
	out := make([]*game.Player, 0, len(g.Seats))
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			out = append(out, p)
		}
	}
	return out
}
