package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b820_helpers.go — Alliances, graveyard recursion, removal, life,
// lands, creatures with keywords, artifacts, enchantments (batch 820).

// OldB820SacrificeSupertypeSubtype is a sacrifice kind with a supertype:
// "a snow Swamp" is ("a snow Swamp", "snow", "Swamp").
func OldB820SacrificeSupertypeSubtype(label, supertype, subtype string) game.SacrificeKind {
	return game.SacrificeKind{Label: label, Subtypes: []string{subtype}, Supertypes: []string{supertype}}
}

// OldB820WasBlockedThisTurn is "it was blocked this turn": some block made
// this turn names the object `id` as the attacker.
func OldB820WasBlockedThisTurn(g *game.Game, id uuid.UUID) bool {
	for _, b := range g.TurnTally.CombatBlocks {
		if b.Attacker.ID == id {
			return true
		}
	}
	return false
}

// OldB820IslandOrBlue and OldB820SwampOrBlack are the permanents Nature's
// Wrath punishes, read as they are now (CR 612.2).
func OldB820IslandOrBlue(c game.Card) bool { return c.HasSubtypeWord("Island") || c.HasColorWord("U") }

func OldB820SwampOrBlack(c game.Card) bool { return c.HasSubtypeWord("Swamp") || c.HasColorWord("B") }

// OldB820PermanentEntersMatching is "whenever a player puts a <kind>
// permanent onto the battlefield, that player sacrifices a <kind> permanent
// of their choice" (Nature's Wrath): the entering permanent's controller
// sacrifices from the permanents of that kind they control, which includes
// the one that entered.
func OldB820PermanentEntersMatching(label, noun string, kind func(game.Card) bool) game.TriggeredAbility {
	return On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.Controller != uuid.Nil && kind(c)
	}, label, func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		c, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
		if !ok {
			return nil
		}
		g.PlayerSacrificesForEffect(item.SourceCardID, c.Controller,
			SacrificeSpec(noun, func(_ *game.Game, _ uuid.UUID, p game.Card) bool { return kind(p) }), "Nature's Wrath — sacrifice "+noun)
		return nil
	})
}
