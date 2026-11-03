package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b13_helpers.go — helpers for Secrets of Strixhaven Commander, batch
// 13 (#4834). Batch-private so parallel batch PRs do not collide.

// socB13EnchantmentSpell passes on an enchantment spell: "Enchantment
// spells you cast cost {1} less to cast" (Starfield Mystic) is this with
// YourSpell().
func socB13EnchantmentSpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.IsEnchantment() }
}

// socB13EnchantmentYouControlWasPutIntoAGraveyard is "whenever an
// enchantment you control is put into a graveyard from the battlefield"
// (Starfield Mystic): a battlefield-to-graveyard move of a permanent that
// was an enchantment as it last existed (CR 603.10a) under the source's
// controller's control, an Aura included. The source itself counts if it is
// one.
func socB13EnchantmentYouControlWasPutIntoAGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return permanentYouControlWasPutIntoAGraveyard(ev, source, g, "enchantment")
}

// socB13ArtifactOrCreatureCardsThatLeftYourGraveyard lists, for the
// triggering event (a zone move or a cast from the graveyard) of "whenever one or more cards leave your
// graveyard" (Spirit of Resilience), the cards of that same event batch that
// left `you`'s graveyard and are an artifact or creature card as printed:
// "from among those cards". A card that can no longer be found (it left the
// game, or came back as a new object) is not one of them (CR 400.7). The
// types are the printed ones (copiable values), not what a permanent is now.
func socB13ArtifactOrCreatureCardsThatLeftYourGraveyard(g *game.Game, you uuid.UUID, trigger game.Event) []uuid.UUID {
	var out []uuid.UUID
	for _, ev := range g.EventsThisTurn() {
		if ev.Batch != trigger.Batch || (trigger.Batch == 0 && ev.Seq != trigger.Seq) {
			continue
		}
		c, ok := cardThatLeftAGraveyard(ev, g)
		if !ok || c.Owner != you || slices.Contains(out, ev.CardID) {
			continue
		}
		v, ok := g.CopiableValuesForEffect(ev.CardID)
		if !ok || !(v.HasCardType("Artifact") || v.HasCardType("Creature")) {
			continue
		}
		out = append(out, ev.CardID)
	}
	return out
}
