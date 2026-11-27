package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b717_helpers.go — shared pieces of the S82 OLD batch 717 cards
// (Urza's Legacy, #5623).

// OldB717BecomesACreature is the effect of the Urza's Legacy "hidden"
// enchantments that become a creature with more than one subtype: the
// CR 603.4 re-check, then an indefinite type change pinned to this object
// (ThisEnchantmentBecomesACreature, for a creature type list). The colour is
// untouched; `with` is what it gains in the same effect.
func OldB717BecomesACreature(subtypes []string, power, toughness int, label string, with ...game.Mod) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		self := item.SourceCardID
		if !OnBattlefield(g, self) || SourceIsNewObject(g, item) {
			return nil
		}
		c, ok := g.LookupCardForEffect(self)
		if !ok || !c.HasCardType("enchantment") {
			return nil
		}
		return ScopedEffectFor{
			Target: self,
			Mods: append([]game.Mod{
				game.RemoveTypesMod("Enchantment"),
				game.AddTypesMod("Creature"),
				game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes(subtypes...)...),
				game.SetBasePowerMod(power),
				game.SetBaseToughnessMod(toughness),
			}, with...),
			Duration: g.PinnedTo(game.IndefiniteDuration(), self),
			Label:    label,
		}.Apply(NewContext(g, item))
	}
}

// OldB717AnOpponentCastIfEnchantment is the cast half of "When an opponent
// casts a <spell>, if this permanent is an enchantment, <effect>".
func OldB717AnOpponentCastIfEnchantment(spell CardPredicate) When {
	cast := AnOpponentCast(spell)
	return func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return source.HasCardType("enchantment") && cast(ev, source, ch, g)
	}
}

// OldB717UpkeepWipe is "At the beginning of your upkeep, if there are N or
// more <permanents> on the battlefield, sacrifice this enchantment and destroy
// all <permanents>." (Planar Collapse, Impending Disaster). The count is
// checked as the trigger fires and again as it resolves (CR 603.4). Once it
// resolves the destruction happens whether or not the enchantment could be
// sacrificed: the two are one instruction each, not "if you do".
func OldB717UpkeepWipe(label string, atLeast int, match CardPredicate, cantBeRegenerated bool) game.TriggeredAbility {
	count := func(g *game.Game, _ uuid.UUID) bool {
		n := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if match(g, uuid.Nil, c) {
				n++
			}
		}
		return n >= atLeast
	}
	return IfVentureCond(count, AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if err := SacrificeThisIfStillHere(ctx); err != nil {
			return err
		}
		return DestroyAllMatching{Match: match, CantBeRegenerated: cantBeRegenerated}.Apply(ctx)
	}))
}

// OldB717AuraOnACreature is "Aura attached to a creature" (Devout Harpist):
// an Aura on the battlefield whose host is a creature.
func OldB717AuraOnACreature(g *game.Game, _ uuid.UUID, c game.Card) bool {
	if !c.IsAura() || c.AttachedTo.Kind != game.TargetCard {
		return false
	}
	host, ok := g.LookupCardForEffect(c.AttachedTo.ID)
	return ok && host.IsCreature()
}
