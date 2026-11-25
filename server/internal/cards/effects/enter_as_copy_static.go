package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// enter_as_copy_static.go — #6262: a permanent's static "creatures you
// control enter as a copy of …" (CR 614.1c, CR 707.2). The replacement
// is the source's; game.ReplacementEvent.EnterAsCopyOfPermanentForEffect
// stamps the copy, and the entry path lands it as it lands a Clone's.

// CreaturesYouControlEnterAsCopyOfThis is Essence of the Wild's
// "Creatures you control enter as a copy of this creature": every
// creature, token or card, entering under the source's controller.
func CreaturesYouControlEnterAsCopyOfThis(label string) game.ReplacementEffect {
	r := enterAsCopyOf(false, func(_ *game.Game, src *game.Card) uuid.UUID { return src.InstanceID })
	r.Label = label
	return r
}

// NontokenCreaturesYouControlEnterAsCopyOfEnchanted is Infinite
// Reflection's "Nontoken creatures you control enter as a copy of
// enchanted creature". Nothing happens while the Aura is attached to no
// creature.
func NontokenCreaturesYouControlEnterAsCopyOfEnchanted(label string) game.ReplacementEffect {
	r := enterAsCopyOf(true, func(g *game.Game, src *game.Card) uuid.UUID {
		if src.AttachedTo.Kind != game.TargetCard {
			return uuid.Nil
		}
		host, ok := g.LookupCardForEffect(src.AttachedTo.ID)
		if !ok || !host.IsCreature() {
			return uuid.Nil
		}
		return host.InstanceID
	})
	r.Label = label
	return r
}

// enterAsCopyOf is the shared body (the caller sets the Label): a creature (as it would enter,
// CR 614.12) entering under the source's controller, a nontoken one when
// `nontoken`, enters as a copy of the permanent `of` names.
func enterAsCopyOf(nontoken bool, of func(g *game.Game, src *game.Card) uuid.UUID) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || !ev.EnteringAsCreatureForEffect(g) {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || entering.Controller != src.Controller || (nontoken && entering.IsToken()) {
				return false
			}
			return of(g, src) != uuid.Nil
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			ev.EnterAsCopyOfPermanentForEffect(g, of(g, src))
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		// The constructors replace it with the card's own wording.
		Label: "Enters as a copy.",
	}
}
