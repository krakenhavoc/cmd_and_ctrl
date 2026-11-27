package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b700_helpers.go — Mercadian Masques, "Counterspells and bounce,
// Artifacts" (S82 OLD batch 700, #5619). Append-only.

// S82OldB700ItemTargetsACreature is "spell or ability that targets a creature"
// (Diplomatic Escort): one of the item's target slots is a creature on the
// battlefield. A target that has left the battlefield is not a creature any
// more, so an item whose only creature target is gone is not a legal pick.
func S82OldB700ItemTargetsACreature(g *game.Game, _ uuid.UUID, item *game.StackItem) bool {
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if c, ok := g.LookupCardForEffect(t.ID); ok && c.IsCreature() {
			return true
		}
	}
	return false
}

// S82OldB700CounterTheTargetItem is "Counter target spell" / "Counter target
// spell or ability" as an ability's effect (Diplomatic Escort, Teferi's Care):
// the target stack item, spell or ability, countered if it is still legal as
// this resolves (CR 608.2b).
func S82OldB700CounterTheTargetItem(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return CounterTarget{StackID: t.ID}.Apply(ctx)
		}
	}
	return nil
}

// S82OldB700EachUpkeepThatPlayerMills is "At the beginning of each player's
// upkeep, that player mills a card" (Worry Beads): the player whose upkeep it
// is, read off the event.
func S82OldB700EachUpkeepThatPlayerMills(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Actor == uuid.Nil {
		return nil
	}
	return MillCards{Player: item.Trigger.Event.Actor, N: 1}.Apply(NewContext(g, item))
}

// S82OldB700DamageToAPlayerExilesThatMany is "If damage would be dealt to a
// player, that player exiles that many cards from the top of their library
// instead" (Crumbling Sanctuary): a CR 614.1a replacement, not a prevention
// (damage that can't be prevented is still replaced). The damage to any
// player, combat or not, is cancelled and that many cards come off the top of
// their library into exile. A player with fewer cards exiles what they have.
// The affected player orders it against other replacements (CR 616.1).
func S82OldB700DamageToAPlayerExilesThatMany() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return src != nil && ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 &&
				g.PlayerByIDForEffect(ev.DamageTarget) != nil
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
			n, player := ev.DamageAmount, ev.DamageTarget
			ev.Cancel()
			return g.ExileTopKeyedForEffect(player, player, n, game.KeyedTail{})
		},
		Controller: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
			return ev.DamageTarget
		},
		Prevention: false,
		Label:      "Crumbling Sanctuary — that player exiles that many cards from the top of their library instead",
	}
}
