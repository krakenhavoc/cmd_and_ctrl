package game

import (
	"errors"

	"github.com/google/uuid"
)

// aura_host.go — CR 303.4f, the host of an Aura put onto the battlefield
// any way other than by resolving as an Aura spell (#3158, ADR 0154's
// amendment of 2026-10-13).
//
//	303.4f If an Aura is entering the battlefield under a player's
//	control by any means other than by resolving as an Aura spell, and
//	the effect putting it onto the battlefield doesn't specify the
//	object or player the Aura will enchant, that player chooses what it
//	will enchant as the Aura enters the battlefield. The player must
//	choose a legal object or player according to the Aura's enchant
//	ability and any other applicable effects.
//
//	303.4g If an Aura is entering the battlefield and there is no legal
//	object or player for it to enchant, the Aura remains in its current
//	zone, unless that zone is the stack. In that case, the Aura is put
//	into its owner's graveyard instead of entering the battlefield. If
//	the Aura is a token, it isn't created.
//
// # Where the question is asked
//
// Inside the CR 614 window of the entry, as its LAST step: when the
// apply-loop has nothing left to apply (applyReplacementsLocked's
// settled branch), and before anything moves. That is what makes "as
// the Aura enters" observable — the Aura is never on the battlefield
// unattached, so CR 704.5m has nothing to see and no enters trigger of
// its own fires for an Aura that then falls off. The answer rides the
// event (ReplacementEvent.auraHost) to the one finisher, which attaches
// between the zone move and EventETB (announceEntryLocked), exactly
// where a resolving Aura spell is attached (attachResolvedAuraLocked).
// An enters trigger that reads "enchanted creature" — Animate Dead's —
// therefore finds it.
//
// Asking in the window means every entry door that can pause asks, with
// no per-door code: the reanimation, the library search, the exile
// return, the hand / library / graveyard / exile "put" batch and created
// tokens are all entryResumable (#478, #1322, #762), and a paused event
// resumes through the same replacementResume frame every entry question
// uses. A simultaneous entry (Replenish) walks its cards' windows one at
// a time, so each Aura asks its own question, against the board as it
// stood before any of them entered.
//
// # What is a legal host
//
// The Aura's enchant clause is its catalogued TargetSpec (ADR 0036), the
// predicate CR 704.5m holds it to afterwards (attachmentLegalLocked),
// read WITHOUT the targeting gate: choosing a host is not targeting
// (CR 115.1), so hexproof and shroud do not stop it. Protection does —
// CR 702.16c: a permanent or player with protection can't be enchanted
// by an Aura with that quality. ADR 0154's graveyard clauses are part
// of the same predicate (Animate Dead's "creature card in a graveyard").
// Two more exclusions, both CR 303.4f's "as it enters": the Aura itself,
// and every other card of the same simultaneous entry, which is not yet
// on the battlefield to be enchanted and is no longer the card in the
// graveyard (Replenish's 2004 ruling: an Aura it returns can't enchant
// another permanent it returns).
//
// # Which entries ask
//
// An entry asks when every one of these holds:
//
//   - it is a move onto the battlefield from another zone, or a created
//     token;
//   - it is not a resolving spell (stackItem): an Aura spell is attached
//     to its target by attachResolvedAuraLocked (CR 303.4a, 608.3c);
//   - it is not face down (CR 708.2: the object is a 2/2 creature);
//   - it does not enter as a copy (Copy Enchantment's shape has no
//     catalog card yet);
//   - the entering card is an Aura with a catalogued enchant clause. An
//     uncatalogued Aura is a manual object in this sandbox and is left
//     alone, as attachmentLegalLocked leaves it; a Role token has no
//     clause and is created attached by its own effect.
//
// An effect that NAMES the host ("put it onto the battlefield attached
// to Light-Paws", Gift of Immortality, Mantle of the Ancients, a copy of
// an Aura spell) sets auraHostFixed and is not asked; its host is
// checked against the same predicate, and an illegal one leaves the Aura
// where it was (CR 303.4i's "remains in its current zone").
//
// # When nobody can be asked
//
// The sandbox move_card verb is the one entry that cannot pause. It
// keeps its old behaviour: the Aura lands unattached and CR 704.5m takes
// it. Every other entry that cannot put the question to anybody — a
// chooser who has left the game — leaves the Aura where it is, the
// weaker branch.

// PendingChoiceAuraHost is the CR 303.4f prompt: "choose what this Aura
// enchants as it enters". Answered with the pick_target payload — one
// `{kind, id}` ref, a card or a player — because it is the same question
// shape, so the client's board-click flow (including a card in a
// graveyard, through the zone browser) answers it with no new picker.
// The candidates ride PickTargetPlayers / PickTargetCards, min 1 max 1.
//
// It is NOT pick_target: nothing is targeted (hexproof and shroud do not
// apply), and the continuation is a PAUSED ENTRY (replacementResume)
// that resumes through the replacement pipeline's contract, like
// copy_target's.
const PendingChoiceAuraHost PendingChoiceKind = "aura_host"

// enteringAuraLocked reports whether ev is an entry CR 303.4f asks about,
// and returns the entering Aura as it stands in its zone (or staged, for
// a token) with its enchant clause.
//
// Caller must hold g.mu.
func (g *Game) enteringAuraLocked(ev *ReplacementEvent) (Card, *TargetSpec, bool) {
	if ev == nil || ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield ||
		ev.OldZone == ZoneBattlefield || ev.OldZone == ZoneStack || ev.stackItem != nil ||
		ev.FaceDown != FaceDownNone || ev.EntersAsCopyOf != nil || ev.Canceled {
		return Card{}, nil, false
	}
	var card Card
	found := false
	if z := g.findCardZoneLocked(ev.CardID); z != nil {
		if z.Kind == ZoneBattlefield || z.Kind != ev.OldZone {
			return Card{}, nil, false
		}
		for i := range z.Cards {
			if z.Cards[i].InstanceID == ev.CardID {
				card, found = z.Cards[i], true
				break
			}
		}
	} else if tok, ok := g.enteringTokenLocked(ev.CardID); ok {
		card, found = tok, true
	}
	if !found || !card.IsAura() || card.IsRole() || card.IsCreature() {
		return Card{}, nil, false
	}
	spec := TargetSpecFor(catalogKeyOf(&card))
	if spec == nil {
		return Card{}, nil, false
	}
	return card, spec, true
}

// auraHostChooser is the player the Aura is entering under the control
// of: the would-be controller on the event (ADR 0102 keeps it current
// through an entry-controller change), else the card's own.
func auraHostChooser(ev *ReplacementEvent, aura Card) uuid.UUID {
	if ev.Actor != uuid.Nil {
		return ev.Actor
	}
	if aura.Controller != uuid.Nil {
		return aura.Controller
	}
	return aura.Owner
}

// auraHostCandidatesLocked is every object or player the entering Aura
// may enchant as it enters, in the deterministic order the targeting
// enumeration walks (players, then the clause's zones). See the file
// comment for the predicate.
//
// Caller must hold g.mu.
func (g *Game) auraHostCandidatesLocked(ev *ReplacementEvent, aura Card, spec *TargetSpec) LegalTargets {
	chooser := auraHostChooser(ev, aura)
	matches := g.specMatchesLocked(SourceChooser(chooser), spec, false)
	quality := SourceCharacteristics(&aura)
	var out LegalTargets
	for _, id := range matches.Players {
		p := g.playerByIDLocked(id)
		if p == nil || p.Eliminated {
			continue
		}
		if _, refused := g.PlayerProtectedFromLocked(p, quality); refused {
			continue
		}
		out.Players = append(out.Players, id)
	}
	siblings := ev.entryBatchCardIDs()
	for _, id := range matches.Cards {
		if id == aura.InstanceID || siblings[id] {
			continue
		}
		z := g.findCardZoneLocked(id)
		if z == nil {
			// An ability on the stack, or a staged token: neither is an
			// object an Aura can enter attached to.
			continue
		}
		if z.Kind == ZoneBattlefield {
			host := findBattlefieldCard(g, id)
			if host == nil || (HasProtection(host) && ProtectedFrom(host, quality)) {
				continue
			}
		}
		out.Cards = append(out.Cards, id)
	}
	return out
}

// entryBatchCardIDs is the set of cards entering together with this one
// (entry_batch.go), nil for a single entry.
func (ev *ReplacementEvent) entryBatchCardIDs() map[uuid.UUID]bool {
	if ev == nil || ev.entryTail == nil || ev.entryTail.batch == nil {
		return nil
	}
	out := make(map[uuid.UUID]bool, len(ev.entryTail.batch.members))
	for _, m := range ev.entryTail.batch.members {
		out[m.cardID] = true
	}
	return out
}

// auraHostIn reports whether ref is one of the candidates.
func auraHostIn(lt LegalTargets, ref TargetRef) bool {
	switch ref.Kind {
	case TargetPlayer:
		for _, id := range lt.Players {
			if id == ref.ID {
				return true
			}
		}
	case TargetCard:
		for _, id := range lt.Cards {
			if id == ref.ID {
				return true
			}
		}
	}
	return false
}

// auraHostOutcome is what the settled branch of the apply-loop does with
// an entry after CR 303.4f has been considered.
type auraHostOutcome int

const (
	// auraHostProceed: not an Aura entry the rule asks about, or its
	// host is settled. The event proceeds.
	auraHostProceed auraHostOutcome = iota
	// auraHostAsked: a prompt is queued; the apply-loop bails with
	// errReplacementPending and the answer resumes the entry.
	auraHostAsked
	// auraHostStays: no legal host (CR 303.4g), the named host is
	// illegal (CR 303.4i), or nobody can answer. The entry is cancelled,
	// so the Aura stays in its zone and a token is never created.
	auraHostStays
)

// settleEnteringAuraHostLocked is CR 303.4f's step in the apply-loop,
// run once the loop has nothing left to apply. See the file comment.
//
// Caller must hold g.mu.
func (g *Game) settleEnteringAuraHostLocked(ev *ReplacementEvent) auraHostOutcome {
	if ev == nil || ev.auraHostSettled || !ev.entryResumable || ev.mustSettleNow {
		return auraHostProceed
	}
	aura, spec, ok := g.enteringAuraLocked(ev)
	if !ok {
		return auraHostProceed
	}
	cands := g.auraHostCandidatesLocked(ev, aura, spec)
	if ev.auraHostFixed {
		if !auraHostIn(cands, ev.auraHost) {
			return auraHostStays
		}
		ev.auraHostSettled = true
		return auraHostProceed
	}
	n := len(cands.Players) + len(cands.Cards)
	chooser := auraHostChooser(ev, aura)
	if n == 0 || chooser == uuid.Nil || g.chooserGoneLocked(chooser) {
		return auraHostStays
	}
	if n == 1 {
		// One legal host: the choice is forced, and it is still the
		// choice CR 303.4f describes.
		if len(cands.Players) == 1 {
			ev.auraHost = TargetRef{Kind: TargetPlayer, ID: cands.Players[0]}
		} else {
			ev.auraHost = TargetRef{Kind: TargetCard, ID: cands.Cards[0]}
		}
		ev.auraHostSettled = true
		return auraHostProceed
	}
	name := aura.Name
	if name == "" {
		name = "Aura"
	}
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:              PendingChoiceAuraHost,
		Chooser:           chooser,
		Count:             1,
		Source:            aura.InstanceID,
		Reason:            name + " — choose what it enchants as it enters",
		PickTargetPlayers: cands.Players,
		PickTargetCards:   cands.Cards,
		PickTargetMin:     1,
		PickTargetMax:     1,
		replacementResume: &replacementResumeFrame{ev: ev},
	})
	if id == uuid.Nil {
		return auraHostStays
	}
	return auraHostAsked
}

// ResolveAuraHost answers a PendingChoiceAuraHost prompt: `host` is the
// object or player the Aura enters attached to.
//
// The answer is checked against the candidates recomputed now, the
// predicate the prompt was built from. A pick that is not among them is
// refused and the prompt stays open. With no candidate left at all the
// entry takes CR 303.4g's branch and the Aura stays where it is.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveAuraHost(choiceID, chooserID uuid.UUID, host TargetRef) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceAuraHost {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	frame := choice.replacementResume
	if frame == nil || frame.ev == nil {
		return ErrInvalidParam
	}
	ev := frame.ev
	aura, spec, ok := g.enteringAuraLocked(ev)
	var cands LegalTargets
	if ok {
		cands = g.auraHostCandidatesLocked(ev, aura, spec)
	}
	none := len(cands.Players)+len(cands.Cards) == 0
	if !none && !auraHostIn(cands, host) {
		return ErrIllegalTarget
	}
	g.dequeueChoiceLocked(idx)
	if none {
		ev.Cancel()
	} else {
		ev.auraHost = host
		ev.auraHostSettled = true
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	// Through the shared finisher every entry resume uses (#478): a
	// cancelled entry still owes its caller an answer — a search its
	// shuffle, a simultaneous entry the rest of its batch.
	return g.finishSettledReplacementLocked(ev, out)
}

// singleAuraHost is the one card or player in a resolving Aura spell's
// targets, the host attachResolvedAuraLocked would attach it to. False
// for none or several, which attachResolvedAuraLocked leaves alone.
func singleAuraHost(targets []TargetRef) (TargetRef, bool) {
	var host TargetRef
	n := 0
	for _, t := range targets {
		switch t.Kind {
		case TargetCard, TargetPlayer:
			host = t
			n++
		}
	}
	return host, n == 1
}

// attachEnteringAuraLocked attaches a permanent that has just landed to
// the host its entry settled (CR 303.4f's choice, or the host the effect
// named), between the zone move and EventETB. AttachForEffect is the one
// attach verb, so a graveyard host gets its epoch pin (ADR 0154) and a
// host that has gone is CR 701.3b's quiet no-op.
//
// Caller must hold g.mu.
func (g *Game) attachEnteringAuraLocked(entered uuid.UUID, ev *ReplacementEvent) {
	if ev == nil || ev.stackItem != nil || ev.auraHost.Kind == "" {
		return
	}
	if !ev.auraHostSettled && !ev.auraHostFixed {
		return
	}
	_ = g.AttachForEffect(entered, ev.auraHost)
}
