package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b089_helpers.go — Zendikar Rising, S82 OLD batch 89 (#5508).

// S82OldB089YouCastAKickedSpell is "whenever you cast a kicked spell"
// (CR 702.33a): the cast event is the source controller's and the spell
// on the stack paid a kicker (multikicker too). The announcement's paid
// optional costs ride the stack item (StackItemPaidForEffect), which is
// still there as the trigger is checked (CR 601.2i).
func S82OldB089YouCastAKickedSpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	return game.KickedTimesPaid(spell, g.StackItemPaidForEffect(ev.CardID).OptionalCosts) > 0
}

// S82OldB089CardHasKicker is "a card with a kicker ability": its catalog
// row declares a kicker cost (CR 702.33).
func S82OldB089CardHasKicker(c game.Card) bool {
	return len(game.KickerIndices(game.OptionalCostsFor(game.CatalogKey(c)))) > 0
}

// S82OldB089PartyCostReduction is "This spell costs {1} less to cast for
// each creature in your party." (CR 700.8), counted as the spell is cast.
func S82OldB089PartyCostReduction() game.CostModifier {
	return CostsLessEach(OldB094PartySizeForCost, "This spell costs {1} less to cast for each creature in your party.")
}

// S82OldB089AnOpponentHasEightInGraveyard is an activation condition:
// an opponent has eight or more cards in their graveyard.
func S82OldB089AnOpponentHasEightInGraveyard(g *game.Game, controller, _ uuid.UUID) bool {
	for _, p := range g.Seats {
		if p != nil && p.ID != controller && !p.Eliminated && p.Graveyard != nil && p.Graveyard.Size() >= 8 {
			return true
		}
	}
	return false
}

// S82OldB089AnotherPartyTypeEnteredThisTurn is "you had another Cleric,
// Rogue, Warrior, or Wizard enter the battlefield under your control this
// turn" (Nimble Trapfinder): the per-turn entry tally counts a permanent
// once for each creature type it had as it entered, so a type the source
// has itself is one entry short of "another".
func S82OldB089AnotherPartyTypeEnteredThisTurn(g *game.Game, source *game.Card) bool {
	for _, t := range FinB9PartyTypes {
		need := 1
		if g.EnteredThisTurn(source.InstanceID) && source.HasSubtype(t) {
			need = 2
		}
		if g.EnteredWithSubtypeThisTurn(source.Controller, t) >= need {
			return true
		}
	}
	return false
}

// S82OldB089OpponentLandWithAnotherLand is Confounding Conundrum's trigger:
// a land an opponent of the source's controller controls entered, and that
// player has had at least two lands enter under their control this turn
// (the tally counts the one that just entered).
func S82OldB089OpponentLandWithAnotherLand(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB {
		return false
	}
	land, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !land.IsLand() || land.Controller == source.Controller {
		return false
	}
	return g.EnteredWithCardTypeThisTurn(land.Controller, "land") >= 2
}

// S82OldB089LandCandidates is a ChoosePermanents offer of exactly one land
// the chooser controls: "return a land you control to its owner's hand"
// (Arid Archway, Confounding Conundrum).
func S82OldB089LandCandidates(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == of && c.IsLand() {
			out = append(out, c.InstanceID)
		}
	}
	return out, 1, 1
}
