package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b720_helpers.go — shared pieces of the S82 OLD batch 720 cards
// (Urza's Saga, #5624): the "hidden" enchantments that become creatures.

// OldB720BecomesACreatureSized is the effect of a hidden enchantment's
// "it becomes a <P>/<T> <subtypes> creature" with a size read as the ability
// resolves (Veiled Sentry's "equal to that spell's mana value"): the CR 603.4
// re-check, then an indefinite type change pinned to this object (CR 611.2c,
// 400.7). It sets the card type (CR 205.1a), so the permanent is no longer an
// enchantment and cannot trigger again; its colour is untouched. `with` is what
// it gains in the same effect.
func OldB720BecomesACreatureSized(subtypes []string, size func(g *game.Game, item *game.StackItem) (power, toughness int), label string, with ...game.Mod) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		self := item.SourceCardID
		if !OnBattlefield(g, self) || SourceIsNewObject(g, item) {
			return nil
		}
		c, ok := g.LookupCardForEffect(self)
		if !ok || !c.HasCardType("enchantment") {
			return nil
		}
		power, toughness := size(g, item)
		return ScopedEffectFor{
			Target: self,
			Mods: append([]game.Mod{
				game.RemoveTypesMod("Enchantment"),
				game.AddTypesMod("Creature"),
				game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes(subtypes...)...),
				game.SetBasePowerMod(power),
				game.SetBaseToughnessMod(toughness),
			}, with...),
			Duration: g.PinnedTo(game.IndefiniteDuration(), self),
			Label:    label,
		}.Apply(NewContext(g, item))
	}
}

// OldB720AnOpponentPlayedALand is the play half of "Whenever an opponent
// plays a land" (CR 305.1): a land that entered because a player PLAYED it,
// not one put onto the battlefield by an effect, whose player is not the
// source's controller. `match` narrows it (nonbasic), nil accepts any land.
func OldB720AnOpponentPlayedALand(match CardPredicate) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if !B20LandPlayed(ev, g) || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && (match == nil || match(g, ev.Actor, c))
	}
}

// OldB720CardPutIntoAGraveyard reports whether the event put a CARD into a
// graveyard from anywhere other than a graveyard (Planar Void). The engine
// emits one event per move: a zone move (a death, a tutor, a spell
// resolving), a discard and a mill each name their destination, and a
// countered spell carries none, so its card is looked for where it is now. A
// token is not a card (CR 108.2).
func OldB720CardPutIntoAGraveyard(ev game.Event, g *game.Game) bool {
	if ev.CardID == uuid.Nil {
		return false
	}
	switch ev.Kind {
	case game.EventZoneMove, game.EventDiscardCard, game.EventMill:
		if ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneGraveyard {
			return false
		}
	case game.EventCounterSpell:
		if ev.Target != ev.CardID {
			return false
		}
		if z := g.FindCardZoneForEffect(ev.CardID); z == nil || z.Kind != game.ZoneGraveyard {
			return false
		}
	default:
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken()
}
