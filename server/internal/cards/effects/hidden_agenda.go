package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hidden_agenda.go — the catalog side of conspiracy cards (CR 315) and hidden
// agenda (CR 702.106), #6102, ADR 0227. The engine puts a conspiracy from a
// player's sideboard into their command zone as the game starts, face down
// with a secretly chosen name when it has hidden agenda, and its abilities
// function only while it is face up (CR 315.5). These are the readers its
// "the chosen name" abilities share.

// Spec.HiddenAgenda values.
const (
	// HiddenAgendaOneName is hidden agenda: one card name (CR 702.106a).
	HiddenAgendaOneName = 1
	// HiddenAgendaTwoNames is double agenda: two different card names
	// (CR 702.106f).
	HiddenAgendaTwoNames = 2
)

// HasTheChosenName reports whether c has a name chosen for source's hidden
// agenda (CR 702.106d: the abilities are linked, so only that choice
// counts), compared as every chosen name is (game.CardNameMatches: every
// face, case-insensitive). A face-down object has no name.
func HasTheChosenName(c game.Card, source *game.Card) bool {
	if source == nil || c.FaceDown {
		return false
	}
	for _, n := range game.ChosenNamesOn(source) {
		if game.CardNameMatches(c, n) {
			return true
		}
	}
	return false
}

// CreaturesYouControlWithTheChosenName is the StaticAbility.AppliesTo of
// "creatures you control with the chosen name": a creature (post-layer)
// under the conspiracy's controller with that name.
func CreaturesYouControlWithTheChosenName(target *game.Card, _ *game.Game, source *game.Card) bool {
	return source != nil && target.Controller == source.Controller && target.IsCreature() &&
		HasTheChosenName(*target, source)
}

// SpellsWithTheChosenNameYouCast is the CostPredicate of "spells with the
// chosen name you cast": the caster is the conspiracy's controller and the
// spell, on the face being cast, has the chosen name (game.SpellOrLandHasName).
func SpellsWithTheChosenNameYouCast() CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Controller != q.Source.Controller {
			return false
		}
		for _, n := range game.ChosenNamesOn(&q.Source) {
			if game.SpellOrLandHasName(q.Card, n) {
				return true
			}
		}
		return false
	}
}

// YouCastASpellWithTheChosenName is a cast trigger's When: "whenever you
// cast a <spell> with the chosen name". spell narrows the kind (nil for any
// spell).
func YouCastASpellWithTheChosenName(spell CardPredicate) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventCast || ev.Actor != source.Controller {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || (spell != nil && !spell(g, source.Controller, c)) {
			return false
		}
		for _, n := range game.ChosenNamesOn(source) {
			if game.SpellOrLandHasName(c, n) {
				return true
			}
		}
		return false
	}
}
