package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82AfrB16GrantsCascadeWhen is GrantsCascade for a grant that has to read the
// cast event itself rather than only the spell card: Wild-Magic Sorcerer's "the
// first spell you cast from exile each turn has cascade" needs the zone the
// spell was cast FROM (Event.OldZone), which a bare spell card cannot say.
// The trigger is the grantor's own battlefield trigger on the controller's
// cast, with cascade's own effect and its mana-value cap read off the spell.
func S82AfrB16GrantsCascadeWhen(label string, when func(ev game.Event, source *game.Card, g *game.Game) bool) game.TriggeredAbility {
	t := GrantsCascade(label, nil)
	t.AppliesTo = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Actor == source.Controller && when(ev, source, g)
	}
	return t
}

// s82AfrB16CastFromGraveyardByExiling is "You may cast this card from your
// graveyard by exiling <n> <kind> cards from your graveyard in addition to
// paying its other costs" (Demilich): Retrace's model (the printed mana cost
// is still paid, and the card has no exile-on-leaving clause) with an escape
// style exile of n other cards from your graveyard as the added cost. The
// exiled cards are named at announce (CR 601.2h); the card being cast is on the
// stack and so is never among them. The card file still opens the zone:
//
//	CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
//	AlternativeCosts: []game.AlternativeCost{s82AfrB16CastFromGraveyardByExiling("{U}{U}{U}{U}", 4, "instant and/or sorcery", Or(Instant(), Sorcery()))},
func S82AfrB16CastFromGraveyardByExiling(printedCost string, n int, kind string, pred CardPredicate) game.AlternativeCost {
	what := NumberWord(n) + " " + kind + " cards from your graveyard"
	return game.AlternativeCost{
		Key:                "exile_from_graveyard_in_addition",
		Label:              "Exile " + what + " to cast this card from your graveyard",
		ManaCost:           printedCost,
		FromZone:           game.ZoneGraveyard,
		ExileFromGraveyard: CardsInYourGraveyard(n, "Exile "+what, pred),
		PayLabel:           what,
	}
}
