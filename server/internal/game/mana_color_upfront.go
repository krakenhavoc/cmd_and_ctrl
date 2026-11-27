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
	if TapOthersAnyNumber(ab.TapOthers) {
		// #5447: how many slots there are is decided by how many
		// permanents the activator taps, so no list can be published
		// before the payment is named. Every slot asks its mana_pick.
		return nil
	}
	if TapOthersUntaps(ab.TapOthers) {
		// #6208: WHICH colours the slot offers is decided by which
		// permanent the activator untaps — Benthic Explorers' "one mana
		// of any type that land could produce" is {U} off an Island and
		// {R} off a Mountain. A list published before the payment is
		// named would be one candidate's, and the client would pick a
		// colour the land it chose cannot make. The slot asks its
		// mana_pick instead.
		return nil
	}
	if ab.ChosenPlayerAdds {
		// #6101: whose colour it is waits on the player the activator
		// chooses (Spectral Searchlight), so no list is published: the
		// client asks for the player first, and the colour is then the
		// activator's mana_pick or the chosen player's
		// mana_pick_for_player. A colour named with the activator chosen
		// is still accepted (validateUpfrontManaColors).
		return nil
	}
	if ab.DemandsX() {
		// #6125: how many slots there are is the X the activation
		// announces (Wizard's Rockets), so, as for the tap-any-number
		// form, no list is published and every slot asks its mana_pick
		// unless the colours are named with the X.
		return nil
	}
	slots, err := manaAbilitySlotsLocked(g, playerID, cardID, &ab, g.maxManaPaymentLocked(playerID, cardID, &ab))
	if err == nil {
		// #6210: "mana of a color of your choice instead" offers five.
		slots = g.widenChosenColorSlotsLocked(playerID, cardID, slots, ab.TapCost)
	}
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
	lists, _ := pickingSlotGroups(slots, identity, narrow)
	return lists
}

// pickingSlotGroups is pickingSlotOptions with, beside each list, the
// different-colours group it belongs to (#2558): -1 for an ordinary
// pick, otherwise the index of the slot whose N mana must all differ.
// A "{W|U|B|R|G:2}" slot is TWO lists, one per mana it adds, both in
// one group, so the answer up front is one colour per mana as it is for
// a filter land's two pipes — and the group is what forbids naming one
// colour twice.
func pickingSlotGroups(slots []ProducedManaEntry, identity commanderIdentity, narrow bool) ([][]string, []int) {
	var (
		out    [][]string
		groups []int
	)
	for si, slot := range slots {
		if len(slot.Options) <= 1 {
			continue
		}
		options := manaPickOptions(slot.Options, identity, narrow)
		if len(options) == 0 {
			continue
		}
		if slot.DifferentColors() {
			for k := 0; k < slot.DistinctCount(options); k++ {
				out = append(out, options)
				groups = append(groups, si)
			}
			continue
		}
		out = append(out, options)
		groups = append(groups, -1)
	}
	return out, groups
}

// ManaAbilityAddsDifferentColors reports whether `ab`'s output, read
// now, has an "N mana of different colors" slot (#2558): the bit the
// view publishes beside ColorOptions so a client never offers an answer
// that names one colour twice. Read-only. Caller must hold g.mu.
func ManaAbilityAddsDifferentColors(g *Game, playerID, cardID uuid.UUID, ab ManaAbilityShape) bool {
	if g == nil {
		return false
	}
	slots, err := manaAbilitySlotsLocked(g, playerID, cardID, &ab, g.maxManaPaymentLocked(playerID, cardID, &ab))
	if err != nil {
		return false
	}
	return differentColorSlot(slots) != oneColorSlotNone
}

// validateUpfrontManaColors is the #1443 check: `colors` names one
// colour per picking slot, each one that slot offers. Empty `colors` is
// the ordinary activation and always passes. Called before anything is
// validated or paid, so a refusal leaves the source untapped.
//
// #2558: the colours named for one different-colours slot must also
// differ from each other, so Firemind Vessel's {U}{U} is refused here,
// before the Vessel taps.
//
// Caller must hold g.mu.
//
// #6125: `x` is the X the activation announces. An X ability's slots are
// counted at that X (Wizard's Rockets at X = 3 picks three colours), not
// at the "could produce" stand-in.
func (g *Game) validateUpfrontManaColors(playerID, cardID uuid.UUID, ab ManaAbilityShape, colors []string, x int) error {
	if len(colors) == 0 {
		return nil
	}
	if TapOthersAnyNumber(ab.TapOthers) {
		// No list is published for it (ManaAbilityColorOptions).
		return ErrIllegalManaColor
	}
	payment := g.maxManaPaymentLocked(playerID, cardID, &ab)
	if ab.DemandsX() {
		payment.XValue = x
	}
	slots, err := manaAbilitySlotsLocked(g, playerID, cardID, &ab, payment)
	if err == nil {
		// #6210: "mana of a color of your choice instead" offers five.
		slots = g.widenChosenColorSlotsLocked(playerID, cardID, slots, ab.TapCost)
	}
	if err != nil || !hasMultiOptionSlot(slots) {
		return ErrIllegalManaColor
	}
	identity := commanderIdentityFor(g, g.playerByIDLocked(playerID))
	options, groups := pickingSlotGroups(slots, identity, ab.NarrowToCommanderIdentity)
	if len(options) != len(colors) {
		return ErrIllegalManaColor
	}
	named := map[int]map[string]bool{}
	for i, c := range colors {
		if !containsColor(options[i], c) {
			return ErrIllegalManaColor
		}
		if grp := groups[i]; grp >= 0 {
			if named[grp] == nil {
				named[grp] = map[string]bool{}
			}
			if named[grp][c] {
				return ErrIllegalManaColor
			}
			named[grp][c] = true
		}
	}
	return nil
}
