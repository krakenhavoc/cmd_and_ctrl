package game

import "github.com/google/uuid"

// mana_color_upfront.go — #1443: the colour of a pipe slot named BEFORE
// the source is tapped.
//
// A mana ability whose output is a choice of colours ("{R|W}",
// "{W|U|B|R|G}", Command Tower's "any color in your commander's color
// identity") used to be answered in two steps: activate, which pays the
// cost and queues a `mana_pick`, then answer the pick. That is the
// right shape for the auto-tapper and the bots, and it stays the
// default. But a player clicking a land has already decided which
// colour they want, and asking again after the land is tapped means the
// question cannot be cancelled — the cost is paid.
//
// So the view publishes, per ability, the colour list each picking slot
// WOULD offer (ManaAbilityColorOptions), and the activation accepts the
// answer up front (ManaAbilityParams.Colors). A named colour must be in
// the list the slot offers, checked before anything is paid; a legal
// one is produced straight into the pool and no `mana_pick` is queued.
//
// One function computes the list for the view, the up-front check and
// the activation's own prompt, and it is the SAME narrowing and
// ordering function the prompt always used (manaPickOptions: CR 903.4f
// narrowing for the four "commander's color identity" cards, #843's
// identity-first order for every other pipe). So the buttons a client
// draws at the card, the colours the server accepts and the colours the
// prompt would have offered are one answer.

// ManaAbilityColorOptions is, for each slot of `ab`'s output that asks
// the activator to pick a colour, the colours that pick would offer
// `playerID` if the ability were activated now — narrowed and ordered
// by manaPickOptions, exactly as the `mana_pick` prompt's
// ColorOptions are. One entry per PICKING slot, in output order: a
// painland's "{R|W}" has one, a filter land's "{W|U}{W|U}" has two, a
// Forest's "{G}" and a Sol Ring's "{C}{C}" have none (nil).
//
// A slot whose printed width is more than one but whose narrowed list
// is EMPTY (Command Tower with no commander, CR 903.4f) adds no mana
// and asks nothing, so it contributes no entry — the same skip the
// activation makes.
//
// A slot counts as picking by its PRINTED width, not its narrowed one,
// for the reason the activation gives: Command Tower under a mono-green
// commander still asks, with one option. The client may answer a
// one-option list without showing it.
//
// The output is read the way every "what would this make" reader reads
// it (manaAbilityProducedLocked), with the largest counter payment the
// cost could take, as ManaAbilityAddsNoMana does. For a derived
// ability (Exotic Orchard) the list is what the board supports now; the
// activation re-reads it after the cost is paid and falls back to the
// prompt if a named colour has gone.
//
// Read-only. Caller must hold g.mu (read or write).
func ManaAbilityColorOptions(g *Game, playerID, cardID uuid.UUID, ab ManaAbilityShape) [][]string {
	if g == nil {
		return nil
	}
	produced := manaAbilityProducedLocked(g, playerID, cardID, &ab, g.maxCounterPaymentLocked(playerID, cardID, ab.RemoveCounters))
	slots, err := ParseProducedMana(produced)
	if err != nil || !hasMultiOptionSlot(slots) {
		return nil
	}
	identity := commanderIdentityFor(g, g.playerByIDLocked(playerID))
	return pickingSlotOptions(slots, identity, ab.NarrowToCommanderIdentity)
}

// pickingSlotOptions is ManaAbilityColorOptions' body with the slots
// and the identity in hand — the one loop the view, the up-front check
// and nothing else reads, so the three agree on which slots pick.
func pickingSlotOptions(slots []ProducedManaEntry, identity commanderIdentity, narrow bool) [][]string {
	var out [][]string
	for _, slot := range slots {
		if len(slot.Options) <= 1 {
			continue
		}
		options := manaPickOptions(slot.Options, identity, narrow)
		if len(options) == 0 {
			continue
		}
		out = append(out, options)
	}
	return out
}

// validateUpfrontManaColors is the #1443 check: `colors` names one
// colour per picking slot, each one that slot offers. Empty `colors` is
// the ordinary activation and always passes. Called before anything is
// validated or paid, so a refusal leaves the source untapped.
//
// Caller must hold g.mu.
func (g *Game) validateUpfrontManaColors(playerID, cardID uuid.UUID, ab ManaAbilityShape, colors []string) error {
	if len(colors) == 0 {
		return nil
	}
	options := ManaAbilityColorOptions(g, playerID, cardID, ab)
	if len(options) != len(colors) {
		return ErrIllegalManaColor
	}
	for i, c := range colors {
		if !containsColor(options[i], c) {
			return ErrIllegalManaColor
		}
	}
	return nil
}
