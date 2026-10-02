package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_if_dies_g2.go — the helpers ADR 0108 Delivery PR 1's second card
// group (#1886, #1887) shares: damage amounts counted as the spell
// resolves, "loses indestructible until end of turn", Wilt in the Heat's
// "one or more cards left your graveyard this turn", and bargain.
//
// Append-only: add a helper, never change what one means.

// g2CardsInYourGraveyard is a damage (or shrink) amount equal to the
// number of cards in the resolving object's controller's graveyard that
// match — Combustion Technique's Lesson cards, Blitz of the
// Thunder-Raptor's instant and sorcery cards, Necrotic Wound's creature
// cards. Counted as the spell resolves (CR 608.2h), when the spell itself
// is on the stack and so never counts itself. A card in a graveyard has
// no layered characteristics, so `match` reads its printed ones.
func g2CardsInYourGraveyard(match func(game.Card) bool) func(*game.StackItem, *Context) int {
	return func(_ *game.StackItem, ctx *Context) int {
		p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
		if p == nil || p.Graveyard == nil {
			return 0
		}
		count := 0
		for _, c := range p.Graveyard.Cards {
			if match(c) {
				count++
			}
		}
		return count
	}
}

// g2IsLessonCard is "a Lesson card": Lesson is a spell subtype, so the
// printed type line carries it ("Instant — Lesson").
func g2IsLessonCard(c game.Card) bool { return c.HasSubtype("Lesson") }

// g2GreatestPowerAmongYourCreatures is "X is the greatest power among
// creatures you control" (Bouncer's Beatdown), read as the spell
// resolves: layered power, and 0 with no creature.
func g2GreatestPowerAmongYourCreatures() func(*game.StackItem, *Context) int {
	return func(_ *game.StackItem, ctx *Context) int {
		return b42GreatestPowerControlledBy(ctx.Game, ctx.Controller())
	}
}

// g2LosesIndestructibleUntilEOT is "<id> loses indestructible until end
// of turn" (Smite the Deathless, Burn from Within): a layer-6
// ability-removing effect with its own timestamp (CR 613.1f), ending in
// this turn's cleanup (CR 514.2). Lethal damage already marked on the
// creature destroys it at the next state-based check once it has lost
// the keyword (CR 702.12b, 704.5g). A permanent no longer on the
// battlefield gets nothing.
func g2LosesIndestructibleUntilEOT(ctx *Context, id uuid.UUID) error {
	if !onBattlefield(ctx.Game, id) {
		return nil
	}
	return untilEndOfTurn(ctx, id, nil, "Loses indestructible until end of turn",
		game.RemoveKeywordsMod("indestructible"))
}

// g2LosesIndestructibleIfDealtDamage is "If a creature is dealt damage
// this way, it loses indestructible until end of turn" (Burn from
// Within) as a damage continuation: only a creature that was dealt more
// than 0 damage.
func g2LosesIndestructibleIfDealtDamage(item *game.StackItem) RecipientThen {
	return func(g *game.Game, target uuid.UUID, dealt int) error {
		if !dealtDamagePermanent(g, target, dealt, false) {
			return nil
		}
		return g2LosesIndestructibleUntilEOT(NewContext(g, item), target)
	}
}

// g2CardLeftYourGraveyardThisTurn passes when one or more cards left the
// caster's graveyard this turn — Wilt in the Heat's "{2} less". A move
// out of the graveyard is an EventZoneMove, or an EventCast whose
// OldZone is the graveyard (a flashback); "your graveyard" is the one
// the caster owns (b16CardLeftYourGraveyard reads the same two events
// for a trigger). Scanned off g.EventsThisTurn(), which is bounded at
// the real turn boundary. A token is not a card (CR 108.2).
//
// The spell being priced counts itself when it is cast FROM its owner's
// graveyard: CR 601.2a moves it to the stack before CR 601.2f totals the
// cost, so by then it has left the graveyard. The cast's own event is
// emitted after the cost is paid, so that case is read off q.FromZone.
func g2CardLeftYourGraveyardThisTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil {
			return false
		}
		if q.FromZone == game.ZoneGraveyard && q.Card.Owner == q.Controller {
			return true
		}
		for _, ev := range q.Game.EventsThisTurn() {
			if ev.Kind != game.EventZoneMove && ev.Kind != game.EventCast {
				continue
			}
			if ev.OldZone != game.ZoneGraveyard || ev.NewZone == game.ZoneGraveyard || ev.CardID == uuid.Nil {
				continue
			}
			if c, ok := q.Game.LookupCardForEffect(ev.CardID); ok && c.Owner == q.Controller && !IsToken(c) {
				return true
			}
		}
		return false
	}
}

// g2BargainKey is bargain's optional-cost key, read back with
// ctx.OptionalCostTimes. The engine compares keys as strings and
// reads none but kicker, multikicker, buyback and gift itself.
const g2BargainKey = "bargain"

// g2Bargain is the bargain keyword, "Bargain (You may sacrifice an artifact,
// enchantment, or token as you cast this spell.)": an optional
// additional cost (CR 601.2b, 601.2h), the non-mana shape
// KickerSacrifice and BuybackSacrifice already have. A spell whose
// bargain cost was paid has been "bargained"; read it with
// g2WasBargained.
func g2Bargain() game.AdditionalCost {
	return game.AdditionalCost{
		Optional:  true,
		Key:       g2BargainKey,
		Sacrifice: sacrificeSpec("an artifact, enchantment, or token", Or(Artifact(), Enchantment(), IsTokenPredicate())),
		Label:     "Bargain—Sacrifice an artifact, enchantment, or token",
	}
}

// g2WasBargained is "if this spell was bargained".
func g2WasBargained(ctx *Context) bool { return ctx.OptionalCostTimes(g2BargainKey) > 0 }
