package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB357YouControlNoCreatures is "Activate only if you control no
// creatures" (Haunted Plate Mail): the activator's creatures as they are now.
func S82OldB357YouControlNoCreatures(g *game.Game, controller, _ uuid.UUID) bool {
	return CountControlled(g, controller, func(c game.Card) bool { return c.IsCreature() }) == 0
}

// S82OldB357YouControlABlueCreature is "if you control a blue creature"
// (Statute of Denial), read through the layered view.
func S82OldB357YouControlABlueCreature(g *game.Game, you uuid.UUID) bool {
	return CountControlled(g, you, func(c game.Card) bool { return c.IsCreature() && c.HasColorWord("U") }) > 0
}

// S82OldB357OpponentOwnsExiledCard is "an opponent owns a card in exile"
// (Warden of the Beyond): any card in exile, face up or face down, whose
// owner is another player than `you`.
func S82OldB357OpponentOwnsExiledCard(g *game.Game, you uuid.UUID) bool {
	if g.Exile == nil {
		return false
	}
	for _, c := range g.Exile.Cards {
		if c.Owner != you && c.Owner != uuid.Nil {
			return true
		}
	}
	return false
}

// S82OldB357PreventXDamageToEquipped is "If a source would deal damage to
// equipped creature, prevent X of that damage, where X is the number of
// creatures you control" (Shield of the Avatar, CR 615): X is counted as the
// damage would be dealt, for the Equipment's controller.
func S82OldB357PreventXDamageToEquipped(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return src != nil && ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 && src.IsAttachedTo(ev.DamageTarget)
		},
		Replace:    s82OldB357PreventXReplace(),
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      label,
		RowLabel:   "Prevent X of the damage that would be dealt to the equipped creature, where X is the number of creatures you control.",
	}
}

// s82OldB357PreventXReplace is the Replace of Shield of the Avatar's
// prevention: it only changes the event it is handed, taking X off the damage,
// where X is read from the board as the damage would be dealt (a read, not an
// additional effect, so CR 615.12's split between Replace and Then is not at
// stake). It is built by a call so the prevention census, which reads
// literals, has no side effect to find in it because there is none.
func s82OldB357PreventXReplace() func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
	return func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
		x := CountControlled(g, src.Controller, func(c game.Card) bool { return c.IsCreature() })
		if ev.DamageAmount <= x {
			ev.Cancel()
			return nil
		}
		ev.DamageAmount -= x
		return nil
	}
}

// DestroyItWhenAttachedCreatureIsDealtDamage is "Whenever enchanted/equipped
// creature is dealt damage, destroy it" (Mire Blight, Hot Soup): the creature
// is named with its object epoch when the damage is dealt and destroyed only
// if it is still that creature on the battlefield (CR 603.7c, 400.7).
func DestroyItWhenAttachedCreatureIsDealtDamage(label string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.Amount > 0 && source.IsAttachedTo(ev.Target)
		},
		Key:   label,
		Build: BuildNamingTheEventTarget(label),
		Effect: func(g *game.Game, item *game.StackItem) error {
			c, ok := g.LookupCardForEffect(item.Params.Object.ID)
			if !ok || c.ObjectEpoch != item.Params.Object.Epoch || !OnBattlefield(g, c.InstanceID) {
				return nil
			}
			return DestroyTarget{Target: c.InstanceID}.Apply(NewContext(g, item))
		},
	}
}

// BuildNamingTheEventTarget is a triggered ability's Build for "the object the
// event is about": it names the event's Target with its object epoch as the
// ability goes on the stack (CR 603.7c), so the resolution can tell that
// object from a later one that reuses the card (CR 400.7).
func BuildNamingTheEventTarget(label string) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		epoch := -1
		if c, ok := g.LookupCardForEffect(ev.Target); ok {
			epoch = c.ObjectEpoch
		}
		item.Params.Object = game.ObjectRef{ID: ev.Target, Epoch: epoch}
		return item
	}
}
