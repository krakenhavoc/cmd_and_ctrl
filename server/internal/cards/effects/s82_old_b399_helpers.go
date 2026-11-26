package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB399PumpEnchanted is "Enchanted creature gets +p/+t until end of turn"
// for an Aura's own activated ability named `name`: the boost goes to the
// creature the Aura is attached to as it resolves, and nothing happens if it is
// no longer attached (CR 608.2b).
func S82OldB399PumpEnchanted(name string, power, toughness int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		host := EnchantedCreatureOf(ctx)
		if host == uuid.Nil {
			return nil
		}
		return BoostUntilEOT{Target: host, Power: power, Toughness: toughness, Label: name + " — " + boostWords(power, toughness) + " until end of turn"}.Apply(ctx)
	}
}

// boostWords words a boost as "+1/+0".
func boostWords(power, toughness int) string {
	sign := func(n int) string {
		if n < 0 {
			return "-" + strconv.Itoa(-n)
		}
		return "+" + strconv.Itoa(n)
	}
	return sign(power) + "/" + sign(toughness)
}

// s82OldB399EnchantedEquipmentHost is the creature the Equipment the resolving
// Aura is attached to is attached to, or uuid.Nil when the Aura is not
// attached, its host is not attached to anything, or what it is attached to
// is not a creature.
func s82OldB399EnchantedEquipmentHost(g *game.Game, aura uuid.UUID) uuid.UUID {
	a, ok := g.LookupCardForEffect(aura)
	if !ok || a.AttachedTo.Kind != game.TargetCard {
		return uuid.Nil
	}
	equipment, ok := g.LookupCardForEffect(a.AttachedTo.ID)
	if !ok || equipment.AttachedTo.Kind != game.TargetCard {
		return uuid.Nil
	}
	host, ok := g.LookupCardForEffect(equipment.AttachedTo.ID)
	if !ok || !host.IsCreature() {
		return uuid.Nil
	}
	return host.InstanceID
}

// S82OldB399EquipmentIsOnACreature is Artificer's Hex's intervening "if
// enchanted Equipment is attached to a creature".
func S82OldB399EquipmentIsOnACreature(ctx *Context) bool {
	return s82OldB399EnchantedEquipmentHost(ctx.Game, ctx.Source()) != uuid.Nil
}

// S82OldB399DestroyEnchantedEquipmentsCreature is "destroy that creature": the
// creature the enchanted Equipment is attached to as the ability resolves.
func S82OldB399DestroyEnchantedEquipmentsCreature(g *game.Game, item *game.StackItem) error {
	host := s82OldB399EnchantedEquipmentHost(g, item.SourceCardID)
	if host == uuid.Nil {
		return nil
	}
	return DestroyTarget{Target: host}.Apply(NewContext(g, item))
}
