package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b119_helpers.go — shared bodies of the OLD batch 119 cards
// (Ikoria: Lair of Behemoths, #5512): Mythos of Snapdax's "choose an
// artifact, a creature, an enchantment and a planeswalker, then sacrifice
// the rest" with a chooser that is not always the permanent's controller,
// and Valiant Rescuer's "for the first time each turn" cycle count.

// S82OldB119KeepOneOfEach is Cataclysmic Gearhulk's "each player chooses an
// artifact, a creature, an enchantment, and a planeswalker from among the
// nonland permanents they control, then sacrifices the rest" for Mythos of
// Snapdax: with youChoose the effect's controller makes every player's
// choice ("If {B}{R} was spent to cast this spell, you choose the
// permanents for each player instead"), otherwise each player chooses their
// own. The sacrifice is one simultaneous event over every seat's rest.
func S82OldB119KeepOneOfEach(g *game.Game, item *game.StackItem, youChoose bool) error {
	var legs []game.ChooseCardsPrompt
	var seats []uuid.UUID
	for _, seat := range ApnapLivingSeats(g) {
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
		chooser := seat
		whose := "your"
		if youChoose {
			chooser = item.Controller
			whose = "that player's"
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    chooser,
			FromPlayer: seat,
			Source:     item.SourceCardID,
			Question:   "Mythos of Snapdax — choose an artifact, a creature, an enchantment, and a planeswalker from among " + whose + " nonland permanents to keep; the rest are sacrificed",
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

// S82OldB119CycledAnotherCardForTheFirstTimeThisTurn is the When of "Whenever
// you cycle another card for the first time each turn": a cycle by the
// source's controller of a card other than the source, with no earlier
// cycle of a card other than the source by that player this turn. It reads
// the turn's event log up to the event being asked about, so the answer is
// the same however often the harvester asks.
func S82OldB119CycledAnotherCardForTheFirstTimeThisTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCycle || source == nil || ev.Actor != source.Controller || ev.CardID == source.InstanceID {
		return false
	}
	for _, prev := range g.EventsThisTurn() {
		if prev.Seq >= ev.Seq {
			break
		}
		if prev.Kind == game.EventCycle && prev.Actor == ev.Actor && prev.CardID != source.InstanceID {
			return false
		}
	}
	return true
}
