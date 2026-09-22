package game

import (
	"github.com/google/uuid"
)

// special_action.go — CR 116.2 special actions, ADR 0062 Decision 4.
//
// A special action is a game action a player takes WITHOUT using the
// stack and without passing priority. CR 116.2 lists seven of them;
// three are built here:
//
//	foretell      CR 116.2h, 702.143a  pay {2}, exile the card face down
//	suspend       CR 116.2f, 702.62a   pay the suspend cost, exile it with
//	                                   N time counters
//	turn_face_up  CR 116.2g, 708.6     pay a face-down permanent's morph
//	                                   cost and turn it face up
//
// ONE VERB, not one per keyword. The rules group these precisely
// because their contract is identical — no stack, no announce, no
// response window, no targets — and four handlers would be four
// copies of `requirePriorityHolder`, four param blocks and four
// enumerator cases for a difference that is three fields wide (where
// the card is, when the window is open, and what it costs). That
// difference is three per-kind tables — specialActionZone, the TIMING
// TABLE below, and the `perform` switch at the bottom — and nothing
// else in the engine forks. ADR 0082 decision 6 added the zone table
// when turn_face_up arrived, which is the first kind whose card is
// not in a hand at all; the claim held.
//
// # The timing table
//
//	kind          rule                 window
//	foretell      CR 702.143a, 116.2h  any time you have priority during
//	                                   YOUR turn; LEGAL under split second
//	                                   (CR 702.61b)
//	suspend       CR 702.62a, 116.2f   any time you could begin to CAST the
//	                                   card — sorcery timing for a sorcery,
//	                                   instant timing for an instant — and
//	                                   therefore NOT legal under split
//	                                   second
//	turn_face_up  CR 702.37c, 116.2g   any time you have priority, on any
//	                                   turn, with anything on the stack;
//	                                   LEGAL under split second, for the
//	                                   same reason foretell is
//
// The split-second asymmetry is the one thing in this file that is
// easy to get wrong and invisible when you do. CR 702.61b stops
// players casting spells and activating abilities that are not mana
// abilities; it says nothing about special actions, so foretelling
// under a Trickbind is legal. Suspend is barred anyway — not by
// CR 702.61b but by its own wording, which imports every restriction
// on beginning to cast the card (CR 702.62c). ADR 0062 spells this
// out and tells the enumerator NOT to copy the `SplitSecondActive`
// early return that opens `legal/cast.go` and `legal/abilities.go`.

// SpecialActionKind is which CR 116.2 special action is being taken.
// The zero value is not a kind.
type SpecialActionKind string

const (
	// SpecialActionForetell is CR 116.2h / CR 702.143a — pay {2} and
	// exile a card from your hand face down.
	SpecialActionForetell SpecialActionKind = "foretell"

	// SpecialActionSuspend is CR 116.2f / CR 702.62a — pay the
	// suspend cost and exile a card from your hand with N time
	// counters on it.
	SpecialActionSuspend SpecialActionKind = "suspend"

	// SpecialActionTurnFaceUp is CR 116.2g / CR 708.6 — pay a
	// face-down permanent's morph cost and turn it face up.
	//
	// The first kind that is not a hand keyword: its card is on the
	// BATTLEFIELD, and the actor is its CONTROLLER rather than its
	// owner (CR 708.5). That is one more per-kind table
	// (specialActionZone) and nothing else — see ADR 0082 decision 6.
	SpecialActionTurnFaceUp SpecialActionKind = "turn_face_up"
)

// SpecialAction is one CR 116.2 special action a card offers, as the
// catalog declares it. Pure data — no closures — so it rides
// CardDef, the wire and the enumerator without any of them learning
// what the kinds mean.
//
// Build one with effects.Foretell / effects.Suspend rather than by
// hand: the fixed halves of each keyword (foretell's {2}, the label
// shape the client renders) belong to the keyword and not to the
// card file.
type SpecialAction struct {
	// Kind is which special action this is.
	Kind SpecialActionKind

	// Cost is the mana cost of TAKING the action, in Scryfall brace
	// notation — always "{2}" for foretell (CR 702.143a), the printed
	// suspend cost for suspend. Empty is a free action, which
	// Lotus Bloom's "Suspend 3—{0}" really is.
	Cost string

	// CastCost is what the card costs to cast LATER, when the special
	// action opens a cast. Foretell's foretell cost (CR 702.143a);
	// empty for suspend, whose cast is free (CR 702.62b) and is
	// priced by the grant rather than by an offer the caster claims.
	CastCost string

	// Counters is how many time counters suspend exiles the card with
	// — the N of "Suspend N—cost" (CR 702.62a). Zero for foretell.
	Counters int

	// FaceUpCounter is megamorph's "turn it face up, then put a +1/+1
	// counter on it" (CR 702.109b) — the one thing turning a
	// permanent face up does beyond turning it face up. False for
	// plain morph, for disguise, and for every other kind.
	//
	// It rides the OFFER rather than being re-derived by the
	// performer, because TurnFaceUpOffer is the one place the
	// per-kind rule is written down and the performer must not open a
	// second door onto the card under a face-down object (ADR 0082
	// decision 4).
	FaceUpCounter bool

	// Label is the menu row and the log line: "Foretell {2}",
	// "Suspend 1—{R}", "Turn face up {1}{U}".
	Label string
}

// CatalogSpecialActions is the catalog hook the effects package wires
// at init, mirroring CatalogAlternativeCosts and friends. Nil, or a
// nil return, means the card offers no special action — which is
// every card but the foretell and suspend families.
var CatalogSpecialActions func(oracleID string) []SpecialAction

// SpecialActionsFor returns the special actions a card offers.
func SpecialActionsFor(oracleID string) []SpecialAction {
	if CatalogSpecialActions == nil || oracleID == "" {
		return nil
	}
	return CatalogSpecialActions(oracleID)
}

// SpecialActionsOfferedByCard is every CR 116.2 special action this
// card offers right now — the list form of SpecialActionOffered, and
// what the wire projection and the legal-move enumerator walk.
//
// It is the ONE place the two sources are joined: the DECLARED
// keywords out of the catalog (foretell, suspend), and the DERIVED
// turn-face-up offer, which is nowhere in the catalog and cannot be
// (ADR 0082 decision 5). A caller that walked SpecialActionsFor alone
// would silently never offer a morph its way back up.
//
// Order is derived-first, which is also cheapest-to-decide-first: the
// derived offer exists only for a face-down permanent, and a
// face-down permanent's catalog entry is empty.
func SpecialActionsOfferedByCard(c Card) []SpecialAction {
	var out []SpecialAction
	if up := TurnFaceUpOffer(c); up != nil {
		out = append(out, *up)
	}
	return append(out, SpecialActionsFor(CatalogKey(c))...)
}

// SpecialActionOffered returns this card's offer of `kind`, or nil
// when the card does not offer it. THE accessor: the engine, the
// legal-move enumerator and the wire projection all ask through it,
// so no two of them can disagree about what a card offers.
//
// Two sources, and which one answers is the kind's business:
//
//   - DECLARED — foretell and suspend are printed keywords, read out
//     of the catalog through CatalogKey, so a face-down object offers
//     neither (CR 708.2a) for free.
//   - DERIVED — turn_face_up is nowhere in the catalog and cannot be
//     (a manifested Mountain has no declaration and must not be
//     turnable; a manifested Grizzly Bears has none and must be). It
//     is a function of the face-down kind and the card underneath —
//     TurnFaceUpOffer, ADR 0082 decision 5.
func SpecialActionOffered(c Card, kind SpecialActionKind) *SpecialAction {
	if kind == SpecialActionTurnFaceUp {
		return TurnFaceUpOffer(c)
	}
	for _, sa := range SpecialActionsFor(CatalogKey(c)) {
		if sa.Kind == kind {
			out := sa
			return &out
		}
	}
	return nil
}

// specialActionZone is WHERE the card a kind acts on lives:
//
//	foretell, suspend  the actor's HAND  CR 702.143a, CR 702.62a
//	turn_face_up       the BATTLEFIELD   CR 708.6
//
// The third per-kind table, beside the timing table and the performer
// switch, and the only thing a kind that is not a hand keyword
// forks — ADR 0062 Decision 4's "one verb" survives intact.
//
// The empty zone is a kind the engine does not carry out.
func specialActionZone(kind SpecialActionKind) ZoneKind {
	switch kind {
	case SpecialActionForetell, SpecialActionSuspend:
		return ZoneHand
	case SpecialActionTurnFaceUp:
		return ZoneBattlefield
	}
	return ""
}

// SpecialActionParams is the announce-time payload. A special action
// has no targets, no modes and no choices — the only thing the caller
// can say is how the mana is found, which is the same pair every
// other payment on the wire carries.
type SpecialActionParams struct {
	// Strict refuses the action when the pool cannot cover the cost,
	// rather than waving it through with an EventCostWarning.
	Strict bool

	// AutoTap asks the server to plan and execute a tap-and-fill
	// before the cost check. Implies Strict.
	AutoTap bool
}

// SpecialActionTimingOKLocked is the per-kind timing table, and the
// ONE place either window is written down. Read by the engine when
// the action is taken and by the legal-move enumerator when it
// decides whether to offer it, so the two cannot disagree.
//
// `card` is the card the action would be taken on, in the player's
// hand. It matters only for suspend, whose window is the card's own
// casting window.
//
// Caller must hold g.mu.
func (g *Game) SpecialActionTimingOKLocked(playerID uuid.UUID, card Card, kind SpecialActionKind) bool {
	switch kind {
	case SpecialActionForetell:
		// CR 702.143a: "during your turn". Priority is the caller's
		// business (the dispatcher's requirePriorityHolder), and the
		// stack need not be empty — foretelling in response to a
		// spell, on your own turn, is legal and is the whole reason
		// the keyword is a special action rather than a sorcery.
		//
		// Deliberately NOT gated on SplitSecondActive: CR 702.61b
		// stops casts and activations, and a special action is
		// neither.
		return g.activeSeatIDLocked() == playerID

	case SpecialActionSuspend:
		// CR 702.62c: "any time you could begin to cast this card".
		// That imports the card's own timing — sorcery speed for a
		// sorcery, instant speed for an instant or a card with flash
		// — and every restriction on beginning a cast, of which
		// split second is the one the engine models.
		if g.SplitSecondActive {
			return false
		}
		if card.IsInstant() || HasKeyword(&card, "flash") {
			return true
		}
		return g.sorcerySpeedOpenLocked(playerID)

	case SpecialActionTurnFaceUp:
		// CR 702.37c / CR 708.6: "any time you have priority". No
		// turn restriction, no sorcery-speed gate, no empty-stack
		// requirement — turning a Willbender face up in response to
		// the spell it will redirect is the whole point of the card.
		//
		// LEGAL UNDER SPLIT SECOND, and this is the second row of
		// this table whose entire content is that asymmetry:
		// CR 702.61b stops players CASTING spells and ACTIVATING
		// abilities that are not mana abilities, and a special action
		// is neither.
		//
		// Priority is the dispatcher's requirePriorityHolder, and
		// whether the permanent is face down and the actor's to turn
		// is SpecialActionOffered's answer and PerformSpecialAction's
		// re-check. Nothing left for the window itself to say.
		return true
	}
	return false
}

// PerformSpecialAction is THE entry point for every CR 116.2 special
// action: one function, one timing table, one payment, no stack.
//
// It refuses before paying anything when the card does not offer the
// action (ErrSpecialActionNotOffered) or when the window is shut
// (ErrSpecialActionTiming), and the payment is the last thing that
// can fail — so a refused action leaves the hand, the pool and the
// board exactly as they were.
//
// The action does not use the stack and grants no window: CR 116.2
// says a special action "doesn't use the stack and can't be
// responded to". The caller keeps priority, which is what the
// dispatcher's requirePriorityHolder already established.
func (g *Game) PerformSpecialAction(playerID, cardID uuid.UUID, kind SpecialActionKind, params SpecialActionParams) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	// WHERE the card is, is the kind's business (specialActionZone).
	// Foretell and suspend act on a card in its owner's HAND
	// (CR 702.143a, CR 702.62a), and a hand in this engine holds only
	// its owner's cards, so finding it there is the whole of the "is
	// this yours" check (CR 108.4). Turning face up acts on a
	// BATTLEFIELD permanent, which holds everybody's cards, so that
	// arm checks CONTROL instead — CR 708.6 names the controller, and
	// a stolen morph is turned up by the thief.
	card, err := g.specialActionCardLocked(p, cardID, kind)
	if err != nil {
		return err
	}
	// The layers decide whether a card is an instant for suspend's
	// window, and whether it has flash. Fast-path no-op when nothing
	// changed.
	g.RecomputeLayersIfStaleLocked()
	sa := SpecialActionOffered(card, kind)
	if sa == nil {
		return ErrSpecialActionNotOffered
	}
	// The performer is looked up BEFORE the cost is paid, so a kind
	// the engine declares but cannot yet carry out refuses for free
	// rather than charging for nothing. That is not hypothetical
	// bookkeeping: the verb, foretell and suspend are three commits,
	// and this is the line that makes each of them true on its own.
	perform := specialActionPerformer(kind)
	if perform == nil {
		return ErrSpecialActionNotOffered
	}
	if !g.SpecialActionTimingOKLocked(playerID, card, kind) {
		return ErrSpecialActionTiming
	}
	// CR 116.2: taking a special action means paying its cost. The
	// mana goes through the same helper an activated ability's mana
	// component uses, so auto-tap, permissive mode and the
	// EventManaSpent breadcrumb behave identically.
	//
	// The spend context is the ZERO one on purpose: a special action
	// is neither a cast nor an activation, so mana restricted to
	// either ("spend this mana only to cast creature spells") cannot
	// pay for it. Conservative in the direction #259 requires.
	if sa.Cost != "" {
		// #1184: the helper takes a parsed cost now. A special action
		// is not an activation (CR 116.2), so nothing prices it
		// through the CR 601.2f pass — it parses its printed string
		// and pays that, exactly as before.
		saCost, perr := ParseCost(sa.Cost)
		if perr != nil {
			return ErrInvalidParam
		}
		if _, err := g.payAbilityManaCostLocked(p, cardID, card.Name, saCost, ActivateAbilityParams{
			Strict:  params.Strict,
			AutoTap: params.AutoTap,
		}, ManaSpendContext{}, nil); err != nil {
			return err
		}
	}
	if err := perform(g, p, cardID, *sa); err != nil {
		return err
	}
	g.EmitEvent(Event{
		Kind:   EventSpecialAction,
		Actor:  playerID,
		CardID: cardID,
		Label:  sa.Label,
	})
	g.runStateChecksLocked()
	return nil
}

// specialActionPerformer is what a kind DOES, once its cost is paid.
// One function per kind, each in the file that owns the keyword
// (foretell.go, suspend.go), so the verb knows nothing about either
// mechanic and neither mechanic has an entry point of its own.
//
// A kind with no performer is a kind the engine does not carry out
// yet: PerformSpecialAction refuses it before charging for it, and
// effects.Register refuses a card that declares it at boot.
func specialActionPerformer(kind SpecialActionKind) func(*Game, *Player, uuid.UUID, SpecialAction) error {
	switch kind {
	case SpecialActionForetell:
		return (*Game).foretellLocked
	case SpecialActionSuspend:
		return (*Game).suspendLocked
	case SpecialActionTurnFaceUp:
		return (*Game).turnFaceUpLocked
	}
	return nil
}

// specialActionCardLocked finds the card a special action would be
// taken on, in the zone its kind names, and refuses it when the actor
// has no claim to it.
//
// Caller must hold g.mu.
func (g *Game) specialActionCardLocked(p *Player, cardID uuid.UUID, kind SpecialActionKind) (Card, error) {
	switch specialActionZone(kind) {
	case ZoneHand:
		if p.Hand == nil {
			return Card{}, ErrCardNotFound
		}
		for _, c := range p.Hand.Cards {
			if c.InstanceID == cardID {
				return c, nil
			}
		}
	case ZoneBattlefield:
		if g.Battlefield == nil {
			return Card{}, ErrCardNotFound
		}
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != cardID {
				continue
			}
			// CR 708.6: the CONTROLLER turns it face up. Anyone else
			// is told the card is not there rather than that it is
			// not theirs — a face-down permanent's identity is not
			// theirs to probe.
			if c.Controller != p.ID {
				return Card{}, ErrCardNotFound
			}
			return c, nil
		}
	}
	return Card{}, ErrCardNotFound
}

// SpecialActionKindBuilt reports whether the engine can carry out
// this kind. Read by effects.Register, which panics at boot on a card
// declaring a special action nothing can perform — the same treatment
// an ability with a cost its zone cannot pay gets.
func SpecialActionKindBuilt(kind SpecialActionKind) bool {
	return specialActionPerformer(kind) != nil
}
