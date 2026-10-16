package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dealt_damage_by.go — "creatures dealt damage by this creature this
// turn" on the catalog side (#1946). The record and its rules are the
// engine's (game/dealt_damage_by.go); these are the shapes cards print.
//
//	ExileIfDealtDamageByThisWouldDie(label)      // Kumano, Master Yamabushi; Frostwielder
//	ExileIfDealtDamageByEnchantedWouldDie(label) // Kumano's Blessing
//	DealtDamageByThisCantBeRegenerated(label)    // Bone Shaman's granted ability
//	dealtDamageByThisTurn(g, source, victim)     // Zurgo Helmsmasher's dies trigger
//
// and, for an effect about one creature's damage from a resolving
// ability, game.CantBeRegeneratedIfDealtDamageByForEffect and
// game.ExileIfDealtDamageByWouldDieThisTurnForEffect (Runesword).

// dealtDamageByThisTurn reports whether `source` dealt damage to
// `victim` this turn, each read as the permanent object it is now or
// most recently was (CR 400.7): either one leaving the battlefield and
// coming back is a new object, and damage prevented to 0 was never
// dealt. A victim that has just died is read as the object that died.
func dealtDamageByThisTurn(g *game.Game, source, victim uuid.UUID) bool {
	return g.PermanentDealtDamageToPermanentThisTurn(source, victim)
}

// ExileIfDealtDamageByThisWouldDie is the static "If a creature dealt
// damage by this creature this turn would die, exile it instead."
// (CR 614.1a, 700.4): Kumano, Master Yamabushi, Frostwielder, Incendiary
// Oracle, Kumano's Pupils. Read live, so a creature damaged before the
// source gained or regained the ability is covered, and nothing is once
// the source has left the battlefield or lost it. Damage the source
// dealt as an earlier object does not count (CR 400.7).
func ExileIfDealtDamageByThisWouldDie(label string) game.ReplacementEffect {
	r := exileIfDealtDamageByWouldDie(label, func(_ *game.Game, src *game.Card) uuid.UUID {
		return src.InstanceID
	})
	r.RowLabel = "If a creature dealt damage by this creature this turn would die, exile it instead."
	return r
}

// ExileIfDealtDamageByEnchantedWouldDie is "If a creature dealt damage by
// enchanted creature this turn would die, exile it instead." (Kumano's
// Blessing): the creature the Aura is attached to as the creature would
// die, so moving the Aura moves whose damage counts.
func ExileIfDealtDamageByEnchantedWouldDie(label string) game.ReplacementEffect {
	r := exileIfDealtDamageByWouldDie(label, func(g *game.Game, src *game.Card) uuid.UUID {
		if host := g.AttachedHostOf(src); host != nil {
			return host.InstanceID
		}
		return uuid.Nil
	})
	r.RowLabel = "If a creature dealt damage by enchanted creature this turn would die, exile it instead."
	return r
}

func exileIfDealtDamageByWouldDie(label string, dealer func(g *game.Game, src *game.Card) uuid.UUID) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventMove || ev.OldZone != game.ZoneBattlefield || ev.NewZone != game.ZoneGraveyard {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || !c.IsCreature() {
				return false
			}
			d := dealer(g, src)
			return d != uuid.Nil && dealtDamageByThisTurn(g, d, ev.CardID)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.NewZone = game.ZoneExile
			ev.NewZoneOwner = uuid.Nil
			return nil
		},
		Controller: replacementSourceController,
		Label:      label,
	}
}

// DealtDamageByThisCantBeRegenerated is the static "Creatures dealt
// damage by this creature this turn can't be regenerated this turn."
// (CR 701.19c), for an AbilityGrant's DealtDamageCantBeRegenerated slot
// (Bone Shaman).
func DealtDamageByThisCantBeRegenerated(label string) []game.DealtDamageCantBeRegenerated {
	return []game.DealtDamageCantBeRegenerated{{Label: label}}
}
