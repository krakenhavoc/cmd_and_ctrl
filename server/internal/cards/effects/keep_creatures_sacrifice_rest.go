package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keep_creatures_sacrifice_rest.go — "chooses up to two creatures they
// control, then sacrifices the rest" (Archfiend of Depravity) and "each
// player chooses any number of creatures they control with total power
// 4 or less, then sacrifices all other creatures they control" (Slaughter
// the Strong). CR 608.2c, 701.21.
//
// The shape is a keep-set, not a sacrifice-one: each seat is asked once
// for the creatures it KEEPS, over its own board, with the set rule on
// the picks (a ceiling on how many, or on their total power), and only
// when every seat has answered does everything else it controls go in
// one simultaneous sacrifice, so a Blood Artist caught in it sees every
// death. The old shape asked "sacrifice one" over and over until the
// board fit, which made the player name the cards they lose in a fixed
// order and could not express a power ceiling at all.
//
// Nothing is targeted, so hexproof and protection are irrelevant, and a
// seat with no creatures is not asked.

// keepCreaturesThenSacrificeTheRest asks each seat in `seats` (the
// caller passes them in APNAP order, CR 101.4) for the creatures it
// keeps — at most `most` of them, or any number when `most` is 0 — and
// then sacrifices every other creature those seats control. `valid`,
// when set, is the rule on the picked set (total power 4 or less).
//
// A seat that can keep everything it controls — `most` or fewer
// creatures, and the set rule holds for all of them — is not asked:
// keeping everything is the only answer worth giving.
func keepCreaturesThenSacrificeTheRest(g *game.Game, item *game.StackItem, seats []uuid.UUID, question string, most int, valid func(picked []game.Card) bool) error {
	var legs []game.ChooseCardsPrompt
	asked := map[uuid.UUID]bool{}
	for _, seat := range seats {
		p := g.PlayerByIDForEffect(seat)
		if p == nil || p.Eliminated {
			continue
		}
		creatures := permanentsControlledByMatching(g, seat, Creature())
		if len(creatures) == 0 || (keepsAll(g, creatures, most, valid)) {
			continue
		}
		asked[seat] = true
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    seat,
			FromPlayer: seat,
			Source:     item.SourceCardID,
			Question:   question,
			Cards:      creatures,
			Min:        0,
			Max:        most,
			Zone:       game.ZoneBattlefield,
			Validate:   valid,
		})
	}
	if len(legs) == 0 {
		return nil
	}
	source := item.SourceCardID
	_, err := g.ChooseCardsRunThenForEffect(legs, func(g *game.Game, picks game.PromptedPicks) error {
		kept := map[uuid.UUID]bool{}
		for _, id := range picks.Cards() {
			kept[id] = true
		}
		var rest []uuid.UUID
		for seat := range asked {
			for _, id := range permanentsControlledByMatching(g, seat, Creature()) {
				if !kept[id] {
					rest = append(rest, id)
				}
			}
		}
		if len(rest) == 0 {
			return nil
		}
		return g.SacrificeAllThenForEffect(source, rest, nil)
	})
	return err
}

// keepsAll reports whether a seat's whole creature list is a legal keep
// set, so nothing would be sacrificed whatever it chose.
func keepsAll(g *game.Game, creatures []uuid.UUID, most int, valid func(picked []game.Card) bool) bool {
	if most > 0 && len(creatures) > most {
		return false
	}
	if valid == nil {
		return true
	}
	cards := make([]game.Card, 0, len(creatures))
	for _, id := range creatures {
		if c, ok := g.LookupCardForEffect(id); ok {
			cards = append(cards, c)
		}
	}
	return valid(cards)
}

// totalPowerAtMost is a set rule: the picked creatures' powers sum to
// `limit` or less. Power is read after layers, so an anthem counts.
func totalPowerAtMost(limit int) func(picked []game.Card) bool {
	return func(picked []game.Card) bool {
		total := 0
		for _, c := range picked {
			total += c.CurrentPower()
		}
		return total <= limit
	}
}
