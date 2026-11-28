package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b825_helpers.go — Homelands: Auras, batch 825 (S82 older sets,
// #5635): the shared pieces of Serra Bestiary, Funeral March, Ironclaw Curse,
// Orcish Mine and Mammoth Harness.

// EnchantedCreatureTapAbilitiesCantActivate is Serra Bestiary's "Enchanted
// creature's activated abilities with {T} in their costs can't be
// activated" (CR 602.5). The Aura contributes the restriction (so it stops
// the moment the Aura leaves or is moved); it refuses an ability of the
// battlefield object it is attached to whose cost has {T} in it, mana
// abilities included, and nothing else (CR 605.1a: a mana ability is an
// activated ability). Read from ActivationAbility.TapCost, which the engine
// derives from the cost of the shape being activated.
func EnchantedCreatureTapAbilitiesCantActivate(label string) game.ActivationRestriction {
	return game.ActivationRestriction{
		Label: label,
		Forbids: func(q game.ActivationQuery) bool {
			if q.FromZone != game.ZoneBattlefield || !q.Ability.TapCost {
				return false
			}
			return q.Source.IsAttachedTo(q.Card.InstanceID)
		},
	}
}

// WhenEnchantedCreatureLeavesTheBattlefield is "When enchanted creature
// leaves the battlefield" (Funeral March): any departure, not only dying
// (WhenEnchantedCreatureDies). The event is emitted while the Aura is still
// attached, so the creature that left is read off item.Trigger.Event.CardID.
func WhenEnchantedCreatureLeavesTheBattlefield(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return source.IsAttachedTo(ev.CardID)
	}, label, effect)
}

// ItsControllerSacrificesACreature is Funeral March's effect: the creature
// that left the battlefield (item.Trigger.Event.CardID) had a controller as it
// last existed there (CR 608.2h), and that player sacrifices a creature of
// their choice.
func ItsControllerSacrificesACreature(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		info, ok := g.LastKnownPermanentForEffect(item.Trigger.Event.CardID)
		if !ok || info.Controller == uuid.Nil {
			return nil
		}
		g.PlayerSacrificesForEffect(item.SourceCardID, info.Controller,
			SacrificeSpec("a creature", Creature()), label)
		return nil
	}
}

// EnchantedCreatureCantBlockPowerAtLeastItsToughness is Ironclaw Curse's
// "Enchanted creature can't block creatures with power equal to or greater
// than the enchanted creature's toughness": a blocker-side pair rule read
// live, as the block is declared (CR 509.1b), through the layers (so the
// Curse's own -0/-1 counts, CR 208.1).
func EnchantedCreatureCantBlockPowerAtLeastItsToughness(label string) game.BlockRule {
	return game.BlockRule{
		Reason: game.BlockReasonCantBlockAttacker,
		Label:  label,
		Pair: func(_ *game.Game, attacker, blocker, source *game.Card) bool {
			return source != nil && source.IsAttachedTo(blocker.InstanceID) &&
				attacker.PowerForComparison() >= blocker.CurrentToughness()
		},
	}
}

// OreCounter is the kind of Orcish Mine's counters.
const OreCounter = "ore"

// RemoveAnOreCounterFromThis is "remove an ore counter from this Aura"
// (Orcish Mine): nothing happens once there is none to remove. The removal
// emits the counter event the "last ore counter" trigger reads.
func RemoveAnOreCounterFromThis(g *game.Game, item *game.StackItem) error {
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || c.Counters[OreCounter] <= 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: OreCounter, N: -1}.Apply(NewContext(g, item))
}

// TheLastOreCounterWasRemovedFromThis is "When the last ore counter is
// removed from this Aura": a removal that leaves none, off this Aura
// (EventCounterPlaced carries the count after the change).
func TheLastOreCounterWasRemovedFromThis(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventCounterPlaced && ev.Target == source.InstanceID &&
		ev.Label == OreCounter && ev.Amount == 0
}

// DestroyEnchantedLandAndDealDamageToItsController is Orcish Mine's last
// clause: "destroy enchanted land and this Aura deals 2 damage to that
// land's controller". The land's controller is read before the destruction
// (it may take the Aura with it, and the damage is dealt either way).
func DestroyEnchantedLandAndDealDamageToItsController(damage int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		land := AttachedHostFor(g, item.SourceCardID)
		if land == nil {
			return nil
		}
		id, controller := land.InstanceID, land.Controller
		if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		return DealDamage{Source: item.SourceCardID, Target: controller, Amount: damage}.Apply(ctx)
	}
}

// RegenerateTheEnchantedCreatureAsItLastExisted is Carapace's effect: the
// Aura was sacrificed to pay for the ability, so "enchanted creature" is the
// creature it was attached to as it last existed (CR 608.2h). A creature that
// has left the battlefield by resolution gets no shield.
func RegenerateTheEnchantedCreatureAsItLastExisted(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	host := info.AttachedTo.ID
	if z := g.FindCardZoneForEffect(host); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return Regenerate{Target: host}.Apply(ctx)
}

// EnchantedCreatureBlockPairOther is Mammoth Harness's trigger condition and
// its effect target: per blocker-attacker pair (CR 509.3b,
// 509.3d) in which the enchanted creature is one side, the OTHER creature is
// the one the trigger names.
func EnchantedCreatureBlockPairOther(ev game.Event, source *game.Card) (uuid.UUID, bool) {
	if ev.Kind != game.EventBlock || ev.Target == uuid.Nil {
		return uuid.Nil, false
	}
	switch {
	case source.IsAttachedTo(ev.CardID):
		return ev.Target, true
	case source.IsAttachedTo(ev.Target):
		return ev.CardID, true
	}
	return uuid.Nil, false
}
