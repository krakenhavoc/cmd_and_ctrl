package effects

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keep_one_of_each_type.go — Cataclysmic Gearhulk's "chooses an artifact, a
// creature, an enchantment, and a planeswalker ... then sacrifices the
// rest", shared with Ajani, Nacatl Avenger's −4 (opponents only).

// gearhulkTypes are the four types a player keeps one permanent of.
var gearhulkTypes = []string{"artifact", "creature", "enchantment", "planeswalker"}

func gearhulkTypesOf(c game.Card) []string {
	var out []string
	for _, t := range gearhulkTypes {
		if c.HasCardType(t) {
			out = append(out, t)
		}
	}
	return out
}

func gearhulkMask(types []string) int {
	mask := 0
	for i, name := range gearhulkTypes {
		if slices.Contains(types, name) {
			mask |= 1 << i
		}
	}
	return mask
}

// gearhulkLegs are the per-seat set rules, one for each non-empty set of the
// four types a board can have. The keys name the types and are on-disk
// identities: never renamed, never reused.
var gearhulkLegs = func() (legs [1 << 4]game.ChooseCardsThen) {
	for mask := 1; mask < 1<<len(gearhulkTypes); mask++ {
		var present []string
		for i, name := range gearhulkTypes {
			if mask&(1<<i) != 0 {
				present = append(present, name)
			}
		}
		legs[mask] = game.RegisterOrderedPickLegRule(
			"choose-cards/cataclysmic-gearhulk-keep-"+strings.Join(present, "-"),
			LilianaKeepsOneOfEachType(present))
	}
	return legs
}()

// gearhulkBoard is a seat's nonland permanents that the effect may
// sacrifice, with the cards read through the layered view.
func gearhulkBoard(g *game.Game, seat uuid.UUID) []game.Card {
	var out []game.Card
	for _, id := range g.EffectSacrificeCandidatesForEffect(seat, nil) {
		if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
			out = append(out, c)
		}
	}
	return out
}

// gearhulkChoosable is the part of a board that has one of the four types.
func gearhulkChoosable(board []game.Card) []game.Card {
	var out []game.Card
	for _, c := range board {
		if len(gearhulkTypesOf(c)) > 0 {
			out = append(out, c)
		}
	}
	return out
}

// gearhulkForced reports that a board needs no choice: everything on it is
// choosable and the whole of it is the one keep-set (nothing is sacrificed).
func gearhulkForced(board []game.Card) bool {
	choosable := gearhulkChoosable(board)
	if len(choosable) != len(board) {
		return false
	}
	present := gearhulkPresent(choosable)
	return LilianaKeepsOneOfEachType(present)(choosable)
}

func gearhulkPresent(cards []game.Card) []string {
	var present []string
	for _, t := range gearhulkTypes {
		for _, c := range cards {
			if c.HasCardType(t) {
				present = append(present, t)
				break
			}
		}
	}
	return present
}

// GearhulkKeepOneOfEach is the Gearhulk's resolution for every player, or,
// with opponentsOnly, for each opponent of the effect's controller (Ajani,
// Nacatl Avenger's −4).
func GearhulkKeepOneOfEach(g *game.Game, item *game.StackItem, opponentsOnly bool) error {
	var legs []game.ChooseCardsPrompt
	var seats []uuid.UUID
	for _, seat := range ApnapLivingSeats(g) {
		if opponentsOnly && seat == item.Controller {
			continue
		}
		board := gearhulkBoard(g, seat)
		if len(board) == 0 {
			continue
		}
		seats = append(seats, seat)
		choosable := gearhulkChoosable(board)
		if gearhulkForced(board) || len(choosable) == 0 {
			continue
		}
		present := gearhulkPresent(choosable)
		ids := make([]uuid.UUID, len(choosable))
		for i, c := range choosable {
			ids[i] = c.InstanceID
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    seat,
			FromPlayer: seat,
			Source:     item.SourceCardID,
			Question:   "Cataclysmic Gearhulk — choose an artifact, a creature, an enchantment, and a planeswalker from among your nonland permanents to keep; the rest are sacrificed",
			Cards:      ids,
			Min:        1,
			Max:        len(present),
			Zone:       game.ZoneBattlefield,
			ThenKey:    gearhulkLegs[gearhulkMask(present)],
		})
	}
	if len(seats) == 0 {
		return nil
	}
	if len(legs) == 0 {
		return gearhulkSacrificeRest(g, item.SourceCardID, seats, nil)
	}
	return g.ChoosePermanentsInOrderKeyedForEffect(legs, gearhulkKeepRest, item.SourceCardID, game.RunData{Carry: seats})
}

// gearhulkKeepRest is "then sacrifices the rest", keyed (#4290). Carry is
// every seat with a nonland permanent, in turn order.
var gearhulkKeepRest = game.RegisterPromptRunThen("prompt-run/cataclysmic-gearhulk-keep", func(g *game.Game, r game.RunDone) error {
	return gearhulkSacrificeRest(g, r.Source, r.Carry, r.Picks().Cards())
})

// gearhulkSacrificeRest sacrifices, at once, every nonland permanent the
// seats control that is neither among `picked` nor on a board that needed
// no choice (where everything is kept).
func gearhulkSacrificeRest(g *game.Game, source uuid.UUID, seats, picked []uuid.UUID) error {
	kept := map[uuid.UUID]bool{}
	for _, id := range picked {
		kept[id] = true
	}
	var rest []uuid.UUID
	for _, seat := range seats {
		board := gearhulkBoard(g, seat)
		forced := gearhulkForced(board)
		for _, c := range board {
			if kept[c.InstanceID] || (forced && len(gearhulkTypesOf(c)) > 0) {
				continue
			}
			rest = append(rest, c.InstanceID)
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return g.SacrificeAllKeyedForEffect(source, rest, game.KeyedTail{})
}
