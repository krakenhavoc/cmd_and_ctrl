package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bok_b1_helpers.go — the shapes Betrayers of Kamigawa, batch 1 (#5602)
// prints more than once.

// bokB1ShoalPitch is a Shoal's "You may exile a <colour> card with mana
// value X from your hand rather than pay this spell's mana cost." It is
// Shining Shoal's alternative cost (#3175): X is announced (CR 107.3a,
// 601.2b) and the pitched card's mana value must be exactly that X, so on
// the stack X is the exiled card's mana value.
func bokB1ShoalPitch(colourWord, colour string) game.AlternativeCost {
	what := "a " + colourWord + " card with mana value X"
	return Pitch(
		"Exile "+what+" from your hand",
		0,
		CardInYourHand(what+" from your hand", OfColor(colour)).WithManaValueEqualsX(),
		what,
	)
}

// bokB1SamuraiYouControlDied is Indebted Samurai's "a Samurai you control
// dies": a creature that had the Samurai subtype as it last existed on
// the battlefield (CR 603.10a) and left it for a graveyard under the
// source's controller's control. "Samurai" is a word of the source's
// text (CR 612.2).
func bokB1SamuraiYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	if !ok || leftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	samurai := source.TextSubtype("Samurai")
	if was, known := ev.WasSubtype(samurai); known {
		return was
	}
	return dead.HasSubtype(samurai)
}

// bokB1PermanentPutIntoAnOpponentsGraveyard is Patron of the Nezumi's
// "a permanent is put into an opponent's graveyard": a permanent left the
// battlefield for a graveyard and its owner, whose graveyard it went to,
// is an opponent of the source's controller. A token counts: it goes to
// its owner's graveyard before it ceases to exist (CR 111.7). It returns
// that owner.
func bokB1PermanentPutIntoAnOpponentsGraveyard(ev game.Event, source *game.Card, g *game.Game) (uuid.UUID, bool) {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return uuid.Nil, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Owner == uuid.Nil || c.Owner == source.Controller {
		return uuid.Nil, false
	}
	return c.Owner, true
}
