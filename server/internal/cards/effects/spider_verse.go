package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spider-Verse — Enchantment {3}{R}{R} (Marvel's Spider-Man promos, #4460):
//
//	"The "legend rule" doesn't apply to Spiders you control.
//	 Whenever you cast a spell from anywhere other than your hand, you may
//	 copy it. If you do, you may choose new targets for the copy. If the
//	 copy is a permanent spell, it gains haste. Do this only once each
//	 turn."
//
// The first line is a legend-rule exemption scoped to the controller's
// Spiders (CR 704.5j; Council of Reeds' creature exemption narrowed to a
// subtype, read through this card's text and off the layered view), read
// live off the battlefield, so two legendary Spiders with one name stay and
// a legendary non-Spider pair still answers to the rule.
//
// The trigger is Nico Minoru's "cast from anywhere other than your hand"
// (the cast event's origin) with Iron Man, Bleeding Edge's optional,
// once-each-turn copy: the "you may" is asked as the trigger resolves (CR
// 603.5), and "only once each turn" limits the DOING, so a declined copy
// does not use the turn up. The copy is made from the spell as it last
// existed if it has left the stack (CR 707.10, 608.2h), its controller may
// choose new targets (CR 707.10c), and a copy of a permanent spell gains
// haste once it exists: an ability-adding effect on the copy (CR 611.2)
// that the token it becomes keeps (CR 400.7a), not a copiable value, so a
// copy of that token does not have it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:             "7ef7d902-1f05-40b3-8997-dec197b0b678",
		Name:                 "Spider-Verse",
		Completeness:         CompletenessFull,
		LegendRuleExemptions: LegendRuleDoesntApplyToYourSubtype("Spider", "Spiders"),
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b25CastFromNotHand(ev, source.Controller) && !b28DoneThisTurn(g, source.InstanceID, spiderVerseDone)
			}, spiderVerseLabel, spiderVerseCopy),
				"Spider-Verse — copy that spell?"),
		},
	})
}

const (
	spiderVerseLabel = "Spider-Verse — copy that spell; a copy of a permanent spell gains haste"
	spiderVerseDone  = "Spider-Verse — copied a spell this turn"
)

func spiderVerseCopy(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || b28DoneThisTurn(g, item.SourceCardID, spiderVerseDone) {
		return nil
	}
	spell := item.Trigger.Event.CardID
	var gains []string
	if c, ok := g.LookupCardForEffect(spell); ok && c.IsPermanent() {
		gains = []string{"haste"}
	}
	b28NoteDoneThisTurn(g, item.SourceCardID, spiderVerseDone)
	err := g.CopyLastKnownSpellGainingKeywordsForEffect(item.SourceCardID, spell, item.Controller, true, gains,
		"Spider-Verse — the copy of a permanent spell gains haste")
	if errors.Is(err, game.ErrCardNotFound) {
		return nil
	}
	return err
}
