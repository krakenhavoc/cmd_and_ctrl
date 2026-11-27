package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b721_helpers.go — Urza's Saga enchantments, S82 OLD batch 721
// (#5624): the hidden enchantments that wake when a spell is cast, Presence of
// the Master and Energy Field.

// S82OldB721OpponentCastWhileEnchantment is the trigger condition of the Urza's
// Saga hidden enchantments: "when an opponent casts a <spell>, if this
// permanent is an enchantment" (CR 603.4). A nil spell is any spell. The "if"
// is asked as the ability triggers; ThisEnchantmentBecomesACreature asks it
// again as the ability resolves.
func S82OldB721OpponentCastWhileEnchantment(spell CardPredicate) When {
	cast := AnOpponentCast(spell)
	return func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source.HasCardType("enchantment") && cast(ev, source, lki, g)
	}
}

// S82OldB721OpalTitanBecomes is Opal Titan's effect: it becomes a 4/4 Giant
// creature with protection from each of the colors of the spell that
// triggered it, read as the trigger resolves (the spell is on the stack or
// gone; a colorless spell gives no protection).
func S82OldB721OpalTitanBecomes(g *game.Game, item *game.StackItem) error {
	var mods []game.Mod
	if item.Trigger != nil {
		var colors []string
		if obj := item.Trigger.Object; obj != nil {
			colors = obj.Colors
		} else if spell, ok := g.LookupCardForEffect(item.Trigger.Event.CardID); ok {
			colors = spell.Effective().Colors
		}
		var prot []string
		for _, c := range colors {
			if p := game.ProtectionFromColor(c); p != "" {
				prot = append(prot, p)
			}
		}
		if len(prot) > 0 {
			mods = append(mods, game.AddKeywordsMod(prot...))
		}
	}
	return ThisEnchantmentBecomesACreature("Giant", 4, 4, "Opal Titan — a 4/4 Giant creature", mods...)(g, item)
}

// S82OldB721FromASourceYouDontControl is "by sources you don't control": the
// damage source, as it would deal the damage (its last-known information),
// has a controller other than the permanent's controller (CR 609.7b).
func S82OldB721FromASourceYouDontControl() DamageSourceFilter {
	return func(_ *game.ReplacementEvent, _ *game.Game, lki *game.Characteristic, src *game.Card) bool {
		return lki != nil && lki.Controller != src.Controller
	}
}

// S82OldB721CardPutIntoYourGraveyard is "a card is put into your graveyard
// from anywhere": a card (never a token) its controller owns arriving in the
// graveyard from any other zone. Each route emits one event for the move, as
// ThisWasPutIntoAGraveyardFromAnywhere lists them: a zone move (a death, a
// spell that finished resolving), a mill, a discard and a countered spell.
func S82OldB721CardPutIntoYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.CardID == uuid.Nil {
		return false
	}
	switch ev.Kind {
	case game.EventCounterSpell:
		if ev.Target != ev.CardID {
			return false
		}
	case game.EventZoneMove, game.EventMill, game.EventDiscardCard:
		if ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneGraveyard {
			return false
		}
	default:
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.IsToken() || c.Owner != source.Controller {
		return false
	}
	p := g.PlayerByIDForEffect(source.Controller)
	return p != nil && p.Graveyard.Contains(ev.CardID)
}

// S82OldB721CardPutIntoYourGraveyardEvents are the events it watches.
var S82OldB721CardPutIntoYourGraveyardEvents = []game.EventKind{
	game.EventZoneMove, game.EventMill, game.EventDiscardCard, game.EventCounterSpell,
}
