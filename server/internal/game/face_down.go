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
	case FaceDownManifested, FaceDownMorphed, FaceDownDisguised, FaceDownCloaked:
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
// answer. The one write path both face-down destinations use — the
// exile route (zone_route.go) and the battlefield entry
// (battlefield_put.go).
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

// revealFaceDownExitLocked is CR 708.9: when a face-down PERMANENT
// moves to another zone, its owner reveals it.
//
// `before` is the card as it was BEFORE the move — MoveCard has
// already cleared the flag by then (CR 400.7), so the question can
// only be asked of the pre-move copy. The reveal itself names the
// card where it landed, which is what "reveals it" means even when
// the destination is a hidden zone.
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
// CR 708.9 is about permanents. The ordinary knowledge rules apply to
// it, which for a public destination already means everyone.
//
// Caller must hold g.mu.
func (g *Game) revealFaceDownExitLocked(before Card) {
	if !before.FaceDownIsPermanent() {
		return
	}
	g.RevealForEffect(RevealSpec{
		Player: before.Owner,
		Source: before.InstanceID,
		Reason: "turned face up on leaving the battlefield",
		Cards:  []uuid.UUID{before.InstanceID},
	})
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
