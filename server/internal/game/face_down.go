package game

import "github.com/google/uuid"

// face_down.go — face-down objects (CR 406.3a, CR 708), ADR 0069.
//
// Everything a face-down object needs is a pure function of ONE field,
// Card.FaceDownKind:
//
//	who may look at it        faceDownViewersLocked   (decision 2)
//	what its characteristics  faceDownCharacteristic  (decision 3)
//	whether it has any text   CatalogKey's guard      (decision 4)
//
// The kind rather than the zone, because a Card knows nothing about
// where it is and the engine has no Card.Zone. That is what lets the
// projection sit on a Card method with no *Game in reach, and it is
// what makes "is this a CR 708.2 object" one comparison at every read
// point instead of a zone lookup.

// FaceDownKind is WHY an object is face down. The zero value is "face
// up"; a non-empty kind and Card.FaceDown are set and cleared
// together by SetFaceDown / ClearFaceDown.
type FaceDownKind string

const (
	// FaceDownNone is face up. Named so a switch reads as one.
	FaceDownNone FaceDownKind = ""

	// FaceDownExiled is CR 406.3's plain face-down exile: NO player
	// may examine it, the player who exiled it included. Necropotence
	// is the only card in the catalog that makes one.
	FaceDownExiled FaceDownKind = "exiled"

	// FaceDownForetold is CR 702.143b's foretell exile. Its OWNER may
	// look at it (CR 702.143d, permitted by CR 406.3a); nobody else
	// may. Not made by anything yet — #658.
	FaceDownForetold FaceDownKind = "foretold"

	// FaceDownManifested is CR 701.40's manifest: the top card of a
	// library put onto the battlefield face down as a 2/2.
	FaceDownManifested FaceDownKind = "manifested"

	// FaceDownMorphed is CR 702.37's morph — a spell cast face down
	// and the permanent it resolves into. Not made by anything yet —
	// #95.
	FaceDownMorphed FaceDownKind = "morphed"

	// FaceDownDisguised is CR 702.168's disguise: morph plus ward
	// {2}. Not made by anything yet — #95.
	FaceDownDisguised FaceDownKind = "disguised"

	// FaceDownCloaked is CR 701.58's cloak: manifest plus ward {2}.
	// Not made by anything yet — #95.
	FaceDownCloaked FaceDownKind = "cloaked"

	// FaceDownTurned is CR 708.2a: a permanent that was ALREADY on
	// the battlefield, face up, and was turned face down by a spell
	// or ability that listed no characteristics for it — Ixidron,
	// Backslide, Cyber Conversion. It is a 2/2 with no text like the
	// other four, and it is the only one of the six that is not a
	// keyword's doing.
	//
	// A SEVENTH KIND RATHER THAN A REUSE, and the two candidates fail
	// for rules reasons rather than tidiness ones (ADR 0082's
	// 2026-09-23 amendment, #1209):
	//
	//   - `morphed` would price the wrong thing. TurnFaceUpOffer's
	//     morph arm demands that the CARD's declared face-down cast
	//     match the state's kind, which is right for a cast (you cast
	//     it face down USING that keyword) and wrong for a turn:
	//     CR 702.37e and CR 702.168d both key the turn-face-up
	//     permission on the card HAVING the ability, whatever put the
	//     permanent face down. A Backslid disguise creature would be
	//     refused its own way back up.
	//   - `disguised` and `cloaked` would hand it WARD {2}. That body
	//     is the one those two keywords list (CR 702.168a,
	//     CR 701.58a); CR 708.2a's is "no text", full stop, and
	//     HasWard is false here for exactly that reason.
	//
	// It also has to be distinguishable at a glance from the four
	// keyword states, because CR 708.7 ("the ability or rules that
	// allow a permanent to be face down MAY ALSO allow the
	// permanent's controller to turn it face up") is a question about
	// WHICH rules did it — which is what FaceDownKind is for.
	FaceDownTurned FaceDownKind = "turned"
)

// IsPermanentState reports whether this kind is one of the CR 708.2
// object states — a face-down permanent, or a morph or disguise on
// the stack — rather than one of the two face-down EXILE states.
//
// This is the partition the whole file turns on. A CR 708.2 object
// has the 2/2 body and no text; a face-down card in exile has no
// characteristics at all (CR 406.3a) and keeps its catalog entry,
// because the cast out of exile needs it.
func (k FaceDownKind) IsPermanentState() bool {
	switch k {
	case FaceDownManifested, FaceDownMorphed, FaceDownDisguised, FaceDownCloaked, FaceDownTurned:
		return true
	}
	return false
}

// HasWard reports whether this kind's CR 708.2 body has ward {2} —
// disguise (CR 702.168a) and cloak (CR 701.58a).
//
// Nothing consumes this yet. The ward itself is NOT a keyword in
// Characteristic.Abilities: canonicalKeywords is a closed set with no
// "ward" token, deliberately, because a bare token has nowhere to put
// the cost (keywords.go, cards/effects/attachments_batch3.go). It is a
// CR 702.21a triggered ability, built by effects.Ward, and #95 hangs
// it off this predicate when it ships disguise and cloak. See ADR 0069
// decision 3.
func (k FaceDownKind) HasWard() bool {
	return k == FaceDownDisguised || k == FaceDownCloaked
}

// CatalogFaceDownWard builds the ward {2} a DISGUISED (CR 702.168a)
// or CLOAKED (CR 701.58a) object has — the one ability a CR 708.2
// object does have.
//
// A hook rather than a value, for the boundary reason ADR 0069
// decision 3 gave and ADR 0082 decision 8 restates: ward is a
// CR 702.21a TRIGGERED ABILITY, not a token in `canonicalKeywords`
// (that table is closed, and a bare token has nowhere to put the
// cost), it is built by `effects.Ward`, and `effects` is a package
// `game` cannot import. So it arrives the way every other catalog
// fact does — a function variable the effects package sets at init.
//
// Nil in a game-package test, which is why every reader goes through
// faceDownWardLocked rather than calling this.
var CatalogFaceDownWard func() []TriggeredAbility

// faceDownWardLocked is the ward a face-down object has right now: the
// disguise and cloak rows of the kind table, and nothing for the other
// four kinds or for any face-up card in the game.
//
// It is the FIRST thing TriggersForCard answers, before the catalog
// read, because the catalog read is precisely what CR 708.2a silences:
// this ability belongs to the face-down STATE and not to the card
// underneath, so a disguised Sheoldred has ward {2} and nothing else.
func faceDownWardLocked(c Card) []TriggeredAbility {
	if !c.FaceDown || !c.FaceDownKind.HasWard() || CatalogFaceDownWard == nil {
		return nil
	}
	return CatalogFaceDownWard()
}

// FaceDownCast is the CR 708 half of morph (CR 702.37b), megamorph
// (CR 702.109a) and disguise (CR 702.168a): the keyword's permission
// to cast the card FACE DOWN, and the price of turning the permanent
// it becomes back face up.
//
// It hangs off AlternativeCost, because casting face down IS an
// alternative cost — a price paid instead of the mana cost, chosen at
// CR 601.2b, that changes what the spell is. The {3} lives on the
// AlternativeCost's own ManaCost, because that half is the KEYWORD's
// and is the same on every card that prints it.
//
// FaceUpCost is here rather than in a declaration of its own for the
// one reason ADR 0082 decision 4 is arranged around: a face-down
// permanent has no catalog entry, so the price of turning it up
// cannot be read off the object. It has to be read off the CARD, and
// the card's "I may be cast face down" declaration is the only place
// in the catalog that already knows which face-down state this card
// produces. One declaration cannot be half-written.
type FaceDownCast struct {
	// Kind is the CR 708.2 state the cast produces —
	// FaceDownMorphed for morph and megamorph, FaceDownDisguised for
	// disguise. Must be a permanent state; Register refuses anything
	// else at boot.
	Kind FaceDownKind

	// FaceUpCost is the MORPH COST (CR 702.37b) or DISGUISE COST
	// (CR 702.168b) — what the controller pays to turn the permanent
	// face up, in Scryfall brace notation. Empty is a free
	// turn-face-up, which Zoetic Cavern's "Morph {0}"-shaped cards
	// really are.
	FaceUpCost string

	// FaceUpCounter is megamorph's "turn it face up, then put a
	// +1/+1 counter on it" (CR 702.109b). False for morph and
	// disguise.
	FaceUpCounter bool
}

// FaceDownCastFor returns the card's own CR 708.4 face-down cast
// offer, or nil when the card prints none.
//
// It reads the catalog under the key given, so a caller holding a
// FACE-DOWN card has to decide for itself which key it means: the
// OBJECT's (CatalogKey, which is "" — a face-down permanent has no
// text) or the CARD's (faceUpCatalogKey, whose two permitted callers
// its own doc comment names).
func FaceDownCastFor(catalogKey string) *AlternativeCost {
	for _, ac := range AlternativeCostsFor(catalogKey) {
		if ac.FaceDown != nil {
			out := ac
			return &out
		}
	}
	return nil
}

// faceUpCatalogKey is the catalog key of the CARD UNDERNEATH a
// face-down object — CatalogKey with the CR 708.2a suppression lifted.
//
// EXACTLY ONE RULE may ask, and it is CR 708.6: "any time you have
// priority, you may turn this permanent face up by paying its morph
// cost". That cost is printed on the CARD, and the player reading it
// is the one CR 708.5 allows to look at it. Every other reader in the
// engine wants the OBJECT and must keep using CatalogKey, which
// answers "" — a face-down permanent has no text, no abilities and no
// catalog entry (ADR 0069 decision 4).
//
// So it is unexported and it has exactly TWO callers, both named
// here and both asking about a price the CARD prints:
//
//  1. TurnFaceUpOffer — CR 708.6's morph cost, above.
//  2. castOfferKey — the CR 601.2b re-derivation of the
//     alternative cost a face-down cast is ALREADY paying, below.
//
// A third caller is a rules bug. It caches nothing and materialises
// nothing: the read is a pure function of the card, taken at the
// moment the price is quoted, and the instant ClearFaceDown runs
// CatalogKey answers on its own again.
func faceUpCatalogKey(c Card) string {
	up := c
	up.ClearFaceDown()
	return CatalogKey(up)
}

// castOfferKey is the catalog key an ALTERNATIVE-COST CLAIM is
// resolved against — CatalogKey for every card in the game, and the
// card underneath for a face-down spell.
//
// A face-down spell (CR 708.4) is the one object whose claimed cost
// cannot be re-derived from its own catalog entry: the stamp that
// made it a CR 708.2 object silenced the very entry the claim came
// out of. But the claim itself is legitimate and was validated
// against the CARD before the stamp — CR 601.2b chooses the cost
// while the card is still a card in a zone — and morph's {3} is a
// price the CARD prints, not text the OBJECT has.
//
// So this is a re-read of a decision already made, not a new look at
// a face-down permanent's text. It cannot widen what a cast may
// claim: CastSpell resolved the claim before stamping anything, and a
// key the card does not offer was refused there.
//
// Locked in name only — it takes no game state — but it is the
// alternative-cost path's half of the pair and reads the way its
// caller does.
func castOfferKey(c Card) string {
	if c.FaceDownIsPermanent() {
		return faceUpCatalogKey(c)
	}
	return CatalogKey(c)
}

// TurnFaceUpOffer is CR 708.6's answer for ONE face-down permanent:
// may this be turned face up, and at what price. It is the whole of
// ADR 0082 decision 5, and the ONE place the per-kind rule is written
// down — the engine, the legal-move enumerator and the wire
// projection all reach it through SpecialActionOffered.
//
//	morphed      the card declares a morph cast    its morph cost    CR 702.37b
//	disguised    the card declares a disguise cast its disguise cost CR 702.168b
//	manifested   the card is a CREATURE CARD       its mana cost     CR 701.34d
//	cloaked      the card is a CREATURE CARD       its mana cost     CR 701.58b
//	turned       the card declares EITHER cast     that cast's cost  CR 702.37e / CR 702.168d
//
// nil means "no", and the three ways to get one all matter: a
// manifested Mountain (CR 701.34d — it stays face down forever), an
// Ixidron'd Sheoldred (CR 708.7 — a permanent turned face down by an
// effect that did not give it a way back up, and whose own card
// prints no way either, can never be turned face up), and every
// face-up permanent in the game.
//
// "Creature card" is PrintedIsCreature deliberately: that accessor is
// the copiable-value surface (CR 707.2) and answers "what does this
// CARD say", which is exactly what CR 701.34d asks. The face-down
// projection would answer "yes, a 2/2 creature" for a manifested
// Island.
//
// CR 701.34e — a manifested card that ALSO has morph may be turned up
// for either cost — is out of scope (ADR 0082 decision 5): this is one
// offer per permanent, and no card in the catalog reaches the case.
func TurnFaceUpOffer(c Card) *SpecialAction {
	if !c.FaceDownIsPermanent() {
		return nil
	}
	switch c.FaceDownKind {
	case FaceDownMorphed, FaceDownDisguised:
		// CR 708.6: "its morph cost" is printed on the CARD, which is
		// what the controller is allowed to look at (CR 708.5).
		alt := FaceDownCastFor(faceUpCatalogKey(c))
		if alt == nil || alt.FaceDown == nil || alt.FaceDown.Kind != c.FaceDownKind {
			// The permanent is in a state its card does not print a
			// way out of. CR 708.7: it can never be turned face up.
			//
			// The kind EQUALITY is the whole of this guard, and it
			// belongs here and nowhere else: a morphed or disguised
			// permanent is one a CAST made, using that keyword, so
			// the card's declaration and the state cannot disagree
			// unless something is wrong. The turned arm below
			// deliberately does not ask (#1209).
			return nil
		}
		return &SpecialAction{
			Kind:          SpecialActionTurnFaceUp,
			Cost:          alt.FaceDown.FaceUpCost,
			FaceUpCounter: alt.FaceDown.FaceUpCounter,
			Label:         turnFaceUpLabel(alt.FaceDown.FaceUpCost),
		}
	case FaceDownTurned:
		// CR 708.2a put it here, and CR 708.7 says the rules that did
		// so "may also allow the permanent's controller to turn it
		// face up" — Ixidron's and Cyber Conversion's do not, and
		// that is the nil below. But morph and disguise are not the
		// effect's to give or withhold: CR 702.37e is "any time you
		// have priority, you may turn a face-down permanent you
		// control WITH A MORPH ABILITY face up", and CR 702.168d says
		// the same for disguise. Both key on the CARD having the
		// ability rather than on how the permanent came to be face
		// down — which is exactly why Backslide is printed "target
		// creature with a morph ability", and why Master of the
		// Veil's loop is a loop.
		//
		// So EITHER declared cast opens the door and the kind
		// equality the arm above makes is deliberately absent here. A
		// megamorph card re-hidden this way still gets CR 702.109b's
		// +1/+1 counter when its megamorph cost is paid to turn it
		// up, because the counter rides the cost that was paid.
		alt := FaceDownCastFor(faceUpCatalogKey(c))
		if alt == nil || alt.FaceDown == nil {
			return nil
		}
		return &SpecialAction{
			Kind:          SpecialActionTurnFaceUp,
			Cost:          alt.FaceDown.FaceUpCost,
			FaceUpCounter: alt.FaceDown.FaceUpCounter,
			Label:         turnFaceUpLabel(alt.FaceDown.FaceUpCost),
		}
	case FaceDownManifested, FaceDownCloaked:
		// CR 701.34d / CR 701.58b: a manifested or cloaked CREATURE
		// CARD may be turned face up for its mana cost. Anything else
		// stays face down for as long as it is on the battlefield.
		if !c.PrintedIsCreature() {
			return nil
		}
		return &SpecialAction{
			Kind:  SpecialActionTurnFaceUp,
			Cost:  c.ManaCost,
			Label: turnFaceUpLabel(c.ManaCost),
		}
	}
	return nil
}

// turnFaceUpLabel is the menu row and the log line. The COST is in it
// and the card's name is not: the row is shown to the controller, who
// can already see the card, and the log line is read by the whole
// table, who may not (CR 708.5).
func turnFaceUpLabel(cost string) string {
	if cost == "" {
		return "Turn face up"
	}
	return "Turn face up " + cost
}

// turnFaceUpLocked is the turn_face_up special action's performer,
// run by PerformSpecialAction once the morph cost is paid
// (CR 708.6, CR 116.2g).
//
// CR 708.8: turning a permanent face up DOES NOT make it a new
// object. So this touches nothing that identifies one — InstanceID,
// ObjectEpoch, counters, damage, attachments, the combat
// declarations, EnteredBattlefieldAt and SummonedThisTurn all ride
// through untouched. A morph that is attacking stays attacking, and
// one that has been out since last turn can attack the moment it is
// turned up.
//
// The ORDER is the rule. The state is cleared BEFORE the event is
// emitted, because the trigger harvester reads a source's abilities
// through CatalogKey: emitting first would harvest against an object
// with no text and drop the very "when this is turned face up"
// trigger CR 708.8 exists for.
//
// Caller must hold g.mu (write).
func (g *Game) turnFaceUpLocked(p *Player, cardID uuid.UUID, sa SpecialAction) error {
	idx := -1
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrCardNotFound
	}
	c := &g.Battlefield.Cards[idx]
	// Re-checked here rather than trusted from the dispatcher: the
	// offer was priced against this card, and between then and now
	// nothing may have changed it — but "nothing may have" is not a
	// guarantee the performer is entitled to make about itself.
	if !c.FaceDownIsPermanent() || c.Controller != p.ID {
		return ErrSpecialActionNotOffered
	}
	c.ClearFaceDown()
	// The card's OWN cached resolution is nilled at the mutation site
	// rather than left to the listener, for the window between the
	// two: EmitEvent dispatches synchronously, and a listener earlier
	// in the slice than layerVersionBump would otherwise read the 2/2
	// off the card it has just been told is face up. The same
	// treatment, for the same reason, that
	// TransformPermanentForEffect gives a transform (ADR 0079).
	c.effective = nil
	// The battlefield is a public zone and the object is public again
	// (CR 708.2 stops applying the moment the permanent is face up),
	// so every seat becomes a knower — the mirror of the face-down
	// landing, which REPLACED the marking rather than adding to it.
	g.markCardKnownInZoneLocked(g.Battlefield, cardID)
	if sa.FaceUpCounter {
		// CR 702.109b: megamorph turns it face up AND puts a +1/+1
		// counter on it — before the event, so a "when this is turned
		// face up" trigger already sees it.
		if err := g.applyCounterByLocked(cardID, "+1/+1", 1, p.ID, cardID); err != nil {
			return err
		}
	}
	g.EmitEvent(Event{
		Kind:   EventTurnedFaceUp,
		Actor:  p.ID,
		Source: cardID,
		CardID: cardID,
	})
	return nil
}

// CanTurnFaceDown reports whether an instruction to turn this
// permanent face down can do anything to it.
//
// Two refusals, and both are "nothing happens" rather than an error
// — the rules say so in those words:
//
//   - CR 708.2b: "A face-down permanent can't be turned face down. If
//     a spell or ability attempts to turn a face-down permanent face
//     down, nothing happens and that effect doesn't change any of its
//     characteristics or their copiable values." Ixidron resolving
//     over a board that already holds a morph is the case, and
//     Ixidron's own ruling says it in as many words: "turning a
//     face-down creature face-down typically has no effect".
//   - CR 712.16: "Melded permanents and other double-faced permanents
//     can't be turned face down." The predicate is transform.go's,
//     because the layout allowlist that decides what "double-faced"
//     means has to have one definition — an `adventure` card also has
//     two Faces and is not a double-faced card.
//
// A TOKEN is deliberately NOT refused. Nothing in CR 708 or CR 111
// stops one being turned face down: it becomes a nameless 2/2 and
// goes on sitting on the battlefield, and Card.IsToken reads the
// PRINTED type line rather than the CR 708.2 projection, so CR 704.5d
// still removes it the instant it leaves. Ixidron says "nontoken" in
// its own text, which is that card's restriction and not the rule's;
// Cyber Conversion happily hits one.
func CanTurnFaceDown(c Card) bool {
	if c.FaceDown {
		return false
	}
	return !isDoubleFacedPermanent(c)
}

// TurnFaceDownForEffect turns the named face-up battlefield
// permanents face down (CR 708.2a) — Ixidron's whole board, Cyber
// Conversion's one target, Backslide's re-hidden morph. `source` is
// the object doing it; the variadic tail is the batch, and one ID is
// the ordinary case.
//
// Returns the IDs that actually turned over, in the order given: a
// card that is not on the battlefield is skipped (CR 608.2b's per-slot
// existence re-check for free), and so is one CanTurnFaceDown refuses.
// A caller that has to count what it did — Ixidron sizing itself —
// counts this.
//
// NOT A NEW OBJECT, which is the whole of the contract. CR 708.8
// gives the rule for the other direction and CR 613.7f gives it for
// both ("a permanent receives a new timestamp each time it turns face
// up or face down" is a statement about one permanent, not two), so
// this touches nothing that identifies one: InstanceID, ObjectEpoch,
// counters, marked damage, attachments, tap state, the combat
// declarations, EnteredBattlefieldAt and SummonedThisTurn all ride
// through untouched. An Ixidron'd attacker goes on attacking as a
// 2/2. The Equipment stays equipped.
//
// The AURA stays attached too, and then may not survive the next
// state-based check — CR 704.5m, and the engine needs no line here
// for it: attachmentSBALocked re-asks the enchant restriction every
// pass, and a "enchant creature with flying" Aura on something that
// is now a vanilla 2/2 is put into its owner's graveyard by the same
// sweep that handles a creature losing flying any other way.
//
// WHO MAY LOOK is reset to the CR 708.5 answer — the controller and
// nobody else — through applyFaceDownLandingLocked, the same one
// writer the exile route and the face-down battlefield entry use. It
// is a genuine narrowing: every player SAW the creature a moment ago,
// and CR 708.6's differentiation rule assumes at a paper table they
// remember. ADR 0082's 2026-09-23 amendment decision A3 records why
// the engine forgets instead — KnownBy has always modelled "may look
// at", CR 708.5 is explicit that you may not look at another player's
// face-down permanent, and the log line the change produces is
// redacted by the very same knower predicate, so the identity does
// not leak back out through the history.
//
// THE ORDER IS MUTATE, THEN ANNOUNCE, and it is worth being explicit
// because the mirror of it looks tempting. ADR 0082 decision 7 clears
// the face-down state BEFORE emitting EventTurnedFaceUp so that the
// trigger harvester, which reads a source's abilities through
// CatalogKey, finds the card's text again. Reading that as "the
// readable state comes first" would put the emit before the mutation
// here. It is not the rule. The rule is that an event is emitted
// after the change it reports — layerVersionBump must not be told
// about a change that has not happened, and a listener earlier in the
// slice must not read face-up characteristics off a card it has just
// been told is face down. What the order costs is the permanent's OWN
// "when this is turned face down" trigger, which CR 708.2a has
// already silenced by then; no printed card has one, and the
// asymmetry belongs to the rules rather than to this function.
//
// The whole batch is turned over before the FIRST event goes out, for
// phaseOutLocked's reason: Ixidron turns every other creature face
// down at once, and a trigger that fired halfway through would read a
// board that never existed.
//
// Caller must hold g.mu (write). Added in S46 (ADR 0082's 2026-09-23
// amendment, #1209).
func (g *Game) TurnFaceDownForEffect(source uuid.UUID, ids ...uuid.UUID) []uuid.UUID {
	if g.Battlefield == nil || len(ids) == 0 {
		return nil
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	actors := make(map[uuid.UUID]uuid.UUID, len(ids))
	turned := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil || seen[id] {
			continue
		}
		seen[id] = true
		idx := findCardOnBattlefield(g, id)
		if idx < 0 {
			continue
		}
		if !CanTurnFaceDown(g.Battlefield.Cards[idx]) {
			continue
		}
		// Read before the write: the CR 708.2 body has no controller
		// of its own to report, and layer 2 rides through unchanged
		// anyway (faceDownCharacteristic keeps baseController).
		actors[id] = g.Battlefield.Cards[idx].Controller
		g.applyFaceDownLandingLocked(g.Battlefield, id, FaceDownTurned)
		// The card's own cached resolution is nilled at the mutation
		// site rather than left to the listener, for the window
		// between the two: EmitEvent dispatches synchronously, and a
		// listener earlier in the slice than layerVersionBump would
		// otherwise read the real card off the permanent it has just
		// been told is a 2/2. The same treatment, for the same
		// reason, that turnFaceUpLocked gives the other direction.
		g.Battlefield.Cards[idx].effective = nil
		turned = append(turned, id)
	}
	for _, id := range turned {
		g.EmitEvent(Event{
			Kind:   EventTurnedFaceDown,
			Actor:  actors[id],
			Source: source,
			CardID: id,
		})
	}
	if len(turned) == 0 {
		return nil
	}
	return turned
}

// SetFaceDown puts the card into a face-down state, or takes it out of
// one when kind is FaceDownNone. The ONLY writer of the pair, so
// "FaceDown true with no kind" cannot be built.
//
// It does not touch KnownBy: who may look is the caller's business,
// because the answer needs the seat list (faceDownViewersLocked).
func (c *Card) SetFaceDown(kind FaceDownKind) {
	if kind == FaceDownNone {
		c.ClearFaceDown()
		return
	}
	c.FaceDown = true
	c.FaceDownKind = kind
}

// ClearFaceDown turns the object face up. Called unconditionally by
// MoveCard on every zone change (CR 400.7).
func (c *Card) ClearFaceDown() {
	c.FaceDown = false
	c.FaceDownKind = FaceDownNone
}

// FaceDownIsPermanent reports whether this card is a CR 708.2 object
// right now — face down AND in one of the four permanent states.
//
// The one predicate behind the 2/2 projection (printedCharacteristic
// and the four accessors that bypass it) and behind catalog
// suppression (CatalogKey). If it is true, this object is a 2/2
// colourless creature with no name and no text, whatever the card
// underneath says.
func (c Card) FaceDownIsPermanent() bool {
	return c.FaceDown && c.FaceDownKind.IsPermanentState()
}

// faceDownCharacteristic is the CR 708.2 body: a 2/2 creature with no
// name, no mana cost, no subtypes and no colour.
//
// It is returned from printedCharacteristic, which is layer 0 — the
// baseline the whole CR 613 pass is applied to — so an anthem still
// pumps it to 3/3, "creatures you control" still counts it, targeting
// still sees a legal creature and the wire still ships 2/2, with no
// further plumbing anywhere. See ADR 0069 decision 3 for why this is
// not a layer-1 copy override.
//
// The layer-2 control baseline rides through unchanged: turning a
// permanent face down does not change who controls it.
func faceDownCharacteristic(c Card) Characteristic {
	ch, _ := FaceDownBody(c.FaceDownKind)
	ch.Controller = c.baseController()
	return ch
}

// FaceDownBody is the CR 708.2 body as a pure function of the kind,
// and the ONE definition of the 2/2. ok is false — and the
// Characteristic is the zero value — for a kind that is not a CR 708.2
// object state, so a caller's guard and this one cannot drift.
//
// Exported for the protocol layer, which has to re-stamp the body onto
// a redacted CardView and has a kind string rather than a Card. A
// face-down permanent's 2/2 is PUBLIC: every player can see it and it
// identifies nothing. The #646 redaction strips name, type line,
// colours and P/T because on an ordinary card those name it; on a
// face-down permanent they ARE this projection, so the protocol layer
// puts them back rather than widening the redaction allowlist for
// every card in the game.
func FaceDownBody(kind FaceDownKind) (Characteristic, bool) {
	if !kind.IsPermanentState() {
		return Characteristic{}, false
	}
	return Characteristic{
		Power:     2,
		Toughness: 2,
		Types:     []string{"Creature"},
	}, true
}

// faceDownViewersLocked is WHO may look at a face-down object —
// ADR 0069 decision 2, and the only definition of it:
//
//	exiled                              nobody         CR 406.3
//	foretold                            the owner      CR 702.143d
//	manifested/morphed/disguised/cloaked the controller CR 708.5
//
// The result is written into KnownBy (replacing whatever was there,
// not added to it: a scryed library card carried into a face-down
// exile would otherwise leave exactly one seat able to read a card
// the rules say nobody can). KnownBy stays the single source of truth
// the wire projection reads; this function decides what it is set to.
//
// A nil or empty result means "nobody", which is a legitimate answer
// and the Necropotence one. Returns nothing for a card that is not
// face down.
//
// Caller must hold g.mu.
func (g *Game) faceDownViewersLocked(c Card) []uuid.UUID {
	if !c.FaceDown {
		return nil
	}
	var viewer uuid.UUID
	switch c.FaceDownKind {
	case FaceDownExiled:
		return nil
	case FaceDownForetold:
		viewer = c.Owner
	default:
		// CR 708.5: the controller of a face-down permanent may look
		// at it. Falls back to the owner for a card that has no
		// controller stamped — a fixture, or a permanent whose
		// controller has left the game.
		viewer = c.Controller
		if viewer == uuid.Nil || g.playerByIDLocked(viewer) == nil {
			viewer = c.Owner
		}
	}
	if viewer == uuid.Nil || g.playerByIDLocked(viewer) == nil {
		return nil
	}
	return []uuid.UUID{viewer}
}

// applyFaceDownLandingLocked puts the card at cardID in zone into the
// face-down state `kind` and sets its knowers to the decision-2
// answer. The one write path every face-down state uses — the exile
// route (zone_route.go), the battlefield entry (battlefield_put.go)
// and, since #1209, the CR 708.2a turn-face-down above, which is the
// only one of the three whose card was already sitting in its zone
// face up.
//
// The caller must NOT also call markCardKnownInZoneLocked: exile and
// the battlefield are both public zones, so the ordinary path would
// mark every seat, and skipping it is the one line that separates
// "public zone" from "public object".
//
// Caller must hold g.mu.
func (g *Game) applyFaceDownLandingLocked(zone *Zone, cardID uuid.UUID, kind FaceDownKind) {
	if zone == nil || kind == FaceDownNone {
		return
	}
	for i := range zone.Cards {
		if zone.Cards[i].InstanceID != cardID {
			continue
		}
		zone.Cards[i].SetFaceDown(kind)
		viewers := g.faceDownViewersLocked(zone.Cards[i])
		zone.Cards[i].ClearKnown()
		zone.Cards[i].AddKnowersAll(viewers)
		return
	}
}

// revealFaceDownExitLocked is CR 708.9: a face-down object leaving the
// zone it is a CR 708.2 object in is revealed by its owner.
//
// TWO exits reach it, and both are the rule rather than one being an
// accident of the other:
//
//   - a face-down PERMANENT leaving the BATTLEFIELD — a manifest
//     destroyed, a morph bounced or tucked into a library;
//   - a face-down SPELL leaving the STACK — a countered morph, one
//     that fizzled, one Hinder shuffles away (#1194). The table finds
//     out what the {3} was buying, which is the answer the rules give
//     and the one a player would insist on at a paper table.
//
// `before` is the card as it was BEFORE the move — MoveCard has
// already cleared the flag by then (CR 400.7), so the question can
// only be asked of the pre-move copy. The reveal itself names the
// card where it landed, which is what "reveals it" means even when
// the destination is a hidden zone.
//
// `from` is the zone it left, and it is carried in ONLY to say so in
// the log line: the rule is the same on both exits, and a reveal that
// told the table a countered spell "left the battlefield" would be
// describing a game that did not happen.
//
// It runs BEFORE the destination's own knowledge rule, and that order
// is the rule: the reveal is what every player SAW, and the
// destination then decides what they still KNOW. A morph tucked into
// a library is revealed to the table and then lost in it (CR 401.2).
//
// It goes through the S22 reveal frame rather than anything new:
// RevealForEffect marks every seated player a knower and emits one
// EventRevealCards, and it moves nothing — a reveal is not a zone
// change, which is exactly the shape CR 708.9 wants.
//
// A face-down EXILE that leaves exile is not revealed by this rule:
// its kind is not a CR 708.2 object state, so FaceDownIsPermanent is
// false for it. The ordinary knowledge rules apply, which for a public
// destination already means everyone.
//
// Caller must hold g.mu.
func (g *Game) revealFaceDownExitLocked(before Card, from ZoneKind) {
	if !before.FaceDownIsPermanent() {
		return
	}
	g.RevealForEffect(RevealSpec{
		Player: before.Owner,
		Source: before.InstanceID,
		Reason: faceDownExitReason(from),
		Cards:  []uuid.UUID{before.InstanceID},
	})
}

// faceDownExitReason is the one-line log reason for a CR 708.9 reveal,
// in the words of the zone the object actually left.
func faceDownExitReason(from ZoneKind) string {
	switch from {
	case ZoneStack:
		return "turned face up as it left the stack"
	case ZoneBattlefield:
		return "turned face up on leaving the battlefield"
	}
	return "turned face up on leaving the " + string(from)
}

// revealFaceDownOwnedByLocked is CR 702.143f's half that can be built
// today: when a player leaves the game, all cards they own that are
// face down in exile are revealed.
//
// It is called from leaveGameObjectsLocked BEFORE the leaver's objects
// are swept out of the game — after the sweep there is nothing left to
// reveal. Nothing is foretold until #658, so today this reveals a
// Necropotence exile, which is the same instruction read literally
// (CR 406.3 stops players examining it while the game is running; it
// stops nothing once the owner is gone) and is the behaviour the hook
// exists to already have when #658 arrives.
//
// The game-end half of CR 702.143f has no game-end sweep to hang on
// yet and is left to #658.
//
// Caller must hold g.mu.
func (g *Game) revealFaceDownOwnedByLocked(playerID uuid.UUID) {
	if g.Exile == nil || playerID == uuid.Nil {
		return
	}
	var ids []uuid.UUID
	for i := range g.Exile.Cards {
		c := &g.Exile.Cards[i]
		if c.FaceDown && c.Owner == playerID {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return
	}
	g.RevealForEffect(RevealSpec{
		Player: playerID,
		Reason: "face-down cards revealed as their owner leaves",
		Cards:  ids,
	})
}
