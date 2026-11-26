package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// PreventAllDamageDealtByAttached is "Prevent all damage that would be
// dealt by enchanted creature" (Defang): a CR 615.1a prevention whose
// source is the permanent the Aura is attached to, whatever it would
// hit. The attachment is read as the damage would be dealt.
func PreventAllDamageDealtByAttached(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true,
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return src != nil && ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 &&
				src.IsAttachedTo(ev.DamageSource)
		},
		Replace:    preventThatDamage,
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      label,
		RowLabel:   rowLabelFromHeader(label),
	}
}

// AddAttachedColors is "enchanted creature is <colour> in addition to its
// other colors": a layer 5 addition (CR 613.1e) that keeps the colours it
// already has. The colours are words of the Aura's text (CR 612.2).
func AddAttachedColors(colors ...string) game.StaticAbility {
	return game.StaticAbility{
		Label:     "Attached permanent gains a color.",
		Layer:     game.Layer5Color,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			for _, col := range source.TextColors(colors...) {
				if !slices.Contains(c.Colors, col) {
					c.Colors = append(c.Colors, col)
				}
			}
		},
	}
}

// AddAttachedSubtypes is "enchanted creature is a Zombie in addition to
// its other types": a layer 4 addition (CR 613.1d). The subtypes are words
// of the Aura's text (CR 612.2).
func AddAttachedSubtypes(subtypes ...string) game.StaticAbility {
	return game.StaticAbility{
		Label:     attachedTypesLabel(nil, subtypes),
		Layer:     game.Layer4Type,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			for _, st := range source.TextSubtypes(subtypes...) {
				if !slices.Contains(c.Subtypes, st) {
					c.Subtypes = append(c.Subtypes, st)
				}
			}
		},
	}
}

// AttachedTargetedByASpell is "when enchanted creature becomes the target
// of a spell": the Aura's own trigger, fired when the creature it is on is
// targeted by a spell (not an ability).
func AttachedTargetedByASpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventBecomesTarget && ev.CardID != uuid.Nil &&
		source.IsAttachedTo(ev.CardID) && B40TargetedByASpell(ev, g)
}

// SacrificeThisSource sacrifices the permanent whose trigger this is.
func SacrificeThisSource(g *game.Game, item *game.StackItem) error {
	return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
}

// HostControllerControlsNoOtherCreatures is "as long as its controller
// controls no other creatures" for an attached host (Predator's Gambit):
// the host's controller controls no creature but the host.
func HostControllerControlsNoOtherCreatures(host *game.Card, g *game.Game, _ *game.Card) bool {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != host.InstanceID && c.Controller == host.Controller && c.IsCreature() {
			return false
		}
	}
	return true
}
