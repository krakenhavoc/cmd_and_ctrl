package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b55_helpers.go — shared pieces for the Wilds of Eldraine leftover
// batches 55–58 (#4471). One seam per section.

// ownsExiledAdventure is "you own a card in exile that has an Adventure":
// any face-up adventurer card in exile whose owner is `you`, whether or
// not it was ever cast as an Adventure (the Memory Theft reading; a
// face-down card shows no characteristics). Caller holds g.mu.
func ownsExiledAdventure(g *game.Game, you uuid.UUID) bool {
	if g.Exile == nil || you == uuid.Nil {
		return false
	}
	for _, c := range g.Exile.Cards {
		if c.Owner == you && !c.FaceDown && cardHasAdventure(c) {
			return true
		}
	}
	return false
}

// cardHasAdventure is "a card that has an Adventure": an adventure-layout
// card whose second face is an Adventure (not a Vengeance or Omen half),
// wherever it is and whichever half is showing (the 2019-10-04 rulings).
func cardHasAdventure(c game.Card) bool {
	if c.Layout != game.LayoutAdventure || len(c.Faces) < 2 {
		return false
	}
	for _, w := range strings.Fields(strings.ReplaceAll(c.Faces[1].TypeLine, "—", " ")) {
		if strings.EqualFold(w, "Adventure") {
			return true
		}
	}
	return false
}

// HasAnAdventure is the card predicate for "a card that has an Adventure".
func HasAnAdventure() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return cardHasAdventure(c) }
}

// AdventureSpellCast is the card predicate for "an Adventure spell": a
// card on the stack whose showing half is the Adventure (its subtype).
func AdventureSpellCast() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.Layout == game.LayoutAdventure && c.HasSubtypeWord("Adventure")
	}
}

// PermanentSpellWithAnAdventure is the cost predicate for "permanent
// spells you cast that have an Adventure" (Beluna Grandsquall): the
// spell is the card's permanent half (an Adventure half is an instant or
// sorcery spell and is not one) and the card has an Adventure.
func PermanentSpellWithAnAdventure() CostPredicate {
	return func(q game.CostQuery) bool {
		return cardHasAdventure(q.Card) && !q.Card.IsInstant() && !q.Card.IsSorcery()
	}
}

// yourLifeIsAtMost is the cost predicate for "if your life total is N or
// less" on a spell's own cost (The End), read for the caster at announce.
func yourLifeIsAtMost(n int) CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil {
			return false
		}
		p := q.Game.PlayerByIDForEffect(q.Controller)
		return p != nil && p.Life <= n
	}
}

// scionDiscountAbility is the activated ability of Rowan, Scion of War and
// Will, Scion of Peace: "{T}: Spells you cast this turn that are <colours>
// cost {X} less to cast, where X is the amount of life you <lost|gained>
// this turn. Activate only as a sorcery." X is read once, as the ability
// resolves, from the per-turn life tally of the activating player (gross: a
// loss ignores gains and a gain ignores losses), and stored as the promise's
// generic reduction (CR 601.2f), which never touches coloured mana. Spells
// with either colour qualify (the filter is "any of"), and the promise lasts
// the turn and applies to every matching spell.
func scionDiscountAbility(label, source, colors string, gained bool) ActivatedAbility {
	return ActivatedAbility{
		Label:        label,
		Cost:         TapCost(),
		SorcerySpeed: true,
		Effect: Do(Instruction(func(ctx *Context) error {
			tally := ctx.Game.TurnTallyFor(ctx.Controller())
			x := tally.LifeLost
			if gained {
				x = tally.LifeGained
			}
			if x <= 0 {
				return nil
			}
			return GrantNextSpellPromise{From: source, Promise: game.NextSpellPromise{
				Filter:             game.PermissionFilter{ColorsAny: colors},
				Reduce:             x,
				EveryMatchingSpell: true,
				Text:               label,
			}}.Apply(ctx)
		})),
	}
}

// YouTappedAnOpponentsCreature is the When for "whenever you tap an untapped
// creature an opponent controls" (Sharae of Numbing Depths, Hylda of the
// Icy Crown): the EventTapCard of a spell or ability YOU control telling a
// creature of an opponent's to tap (game.TapCauseEffect, the tapper as the
// event's Actor, emitted only when the creature was untapped, CR 701.26a). A
// tap you cause by cost, attack or mana ability is not one, and neither is
// a tap an opponent's effect tells them to make of their own creature, as
// the printed rulings say. Watch game.EventTapCard.
func YouTappedAnOpponentsCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventTapCard || ev.TapCause != game.TapCauseEffect || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller != source.Controller
}

// Enchanted is the card predicate for "enchanted creature": a permanent with
// an Aura attached to it, a Role included (CR 303.4). Read off the live
// battlefield at announce and again when the target is re-judged.
func Enchanted() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool { return isEnchantedByAura(g, c.InstanceID) }
}

// exileTheModesTargetPlayersGraveyard is the bullet body "Exile target
// player's graveyard." for a mode whose target clause is a player (Kutzil's
// Flanker, Sentinel of Lost Lore).
func exileTheModesTargetPlayersGraveyard(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return exileGraveyardForEffect(ctx.Game, item, t.ID)
}
