package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b090_helpers.go — Zendikar Rising kicker cards, S82 OLD batch 90 (#5508).

// S82OldB090YouCastAKickedSpell is "whenever you cast a kicked spell"
// (CR 702.33a): the cast event is the source controller's and the spell
// on the stack paid a kicker (multikicker too). The announcement's paid
// optional costs ride the stack item (StackItemPaidForEffect), which is
// still there as the trigger is checked (CR 601.2i).
func S82OldB090YouCastAKickedSpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	return game.KickedTimesPaid(spell, g.StackItemPaidForEffect(ev.CardID).OptionalCosts) > 0
}

// S82OldB090CardHasKicker is "a card with a kicker ability": its catalog
// row declares a kicker cost (CR 702.33).
func S82OldB090CardHasKicker(c game.Card) bool {
	return len(game.KickerIndices(game.OptionalCostsFor(game.CatalogKey(c)))) > 0
}

// S82OldB090HighestLifeTotal is the highest life total among the players
// still in the game (Scourge of the Skyclaves).
func S82OldB090HighestLifeTotal(g *game.Game) int {
	best, seen := 0, false
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if !seen || p.Life > best {
			best, seen = p.Life, true
		}
	}
	return best
}

// S82OldB090EachPlayerLosesHalfTheirLife is "each player loses half their
// life, rounded up": every player's loss is worked out from the totals as
// the effect begins, then applied (CR 120.3).
func S82OldB090EachPlayerLosesHalfTheirLife(g *game.Game, item *game.StackItem) error {
	type loss struct {
		player uuid.UUID
		n      int
	}
	var losses []loss
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.Life <= 0 {
			continue
		}
		losses = append(losses, loss{p.ID, (p.Life + 1) / 2})
	}
	for _, l := range losses {
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, l.player, -l.n); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB090AlsoAPartyCreature is "<this card> is also a Cleric, Rogue,
// Warrior, and Wizard." (CR 700.8; Burakos, Tajuru Paragon): a layer 4 type
// effect on the source itself that adds the four party types it lacks.
func S82OldB090AlsoAPartyCreature(label string) game.StaticAbility {
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, t := range FinB9PartyTypes {
				if !slices.Contains(c.Subtypes, t) {
					c.Subtypes = append(c.Subtypes, t)
				}
			}
		},
	}
}
