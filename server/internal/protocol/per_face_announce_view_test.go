package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// per_face_announce_view_test.go — #992, ADR 0034.
//
// `CardView` carries the announce surface of the face that is UP, and
// for a card in hand that is always face 0. A modal DFC and an
// adventure card are two castable objects sharing one instance
// (CR 712.12a, CR 715.3), the halves have different catalog entries,
// and most printed Adventure halves TARGET — so a client asked to cast
// face 1 had `cardAsFace` clear the front's answers and nothing to put
// back. It reached cast_spell with no target picker ever opening.
//
// What is pinned here: every face a cast may choose carries its own
// block, computed by the same castStampsFor walk as the card's; the
// per-viewer / public split of #1055 applies to a face exactly as it
// does to a card; and a grant that NAMES faces narrows which faces
// carry one at all, so the view can never offer a picker row
// faceForCastLocked would refuse with ErrInvalidFace.

// adventureOracle is the fixture's card: a 1/1 creature whose
// Adventure half deals damage to any target, which is Bonecrusher
// Giant // Stomp in miniature.
const adventureOracle = "view-per-face-adventure"

// withFaceCatalog stubs the target clause, target mode and MODAL
// clause of the ADVENTURE half only — key "<oracle>#1" — so a stamp
// that reads the card's bare oracle ID for face 1 shows up as an empty
// answer rather than as the front's.
//
// The modal half is #1172's: `modes` is a public field that carries a
// legal set per targeted bullet, so a face's block has the nested
// split to pin as well as the top-level one.
func withFaceCatalog(t *testing.T) {
	t.Helper()
	prevSpec, prevMode, prevModes := game.CatalogTargetSpec, game.CatalogTargetMode, game.CatalogModeSpec
	game.CatalogTargetSpec = func(id string) *game.TargetSpec {
		if id != adventureOracle+"#1" {
			return nil
		}
		return adventureClause()
	}
	game.CatalogTargetMode = func(id string) string {
		if id == adventureOracle+"#1" {
			return "any"
		}
		return ""
	}
	game.CatalogModeSpec = func(id string) *game.ModeSpec {
		if id != adventureOracle+"#1" {
			return nil
		}
		return &game.ModeSpec{
			Prompt: "Choose one —",
			Min:    1, Max: 1,
			Options: []game.ModeOption{
				{Label: "Fixture Stomp deals 2 damage to target creature.", Targets: adventureClause()},
				{Label: "Draw a card."},
			},
		}
	}
	t.Cleanup(func() {
		game.CatalogTargetSpec, game.CatalogTargetMode, game.CatalogModeSpec = prevSpec, prevMode, prevModes
	})
}

// adventureClause is the Adventure half's target clause: one creature
// on the battlefield. A fresh value per call, because the mode option
// and the face's own clause must not share one.
func adventureClause() *game.TargetSpec {
	return &game.TargetSpec{
		Mode:  "any",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1,
	}
}

// adventureCard is the fixture card, front face up, known to the
// seats named.
func adventureCard(owner uuid.UUID, knowers ...uuid.UUID) game.Card {
	c := game.NewCard("Fixture Giant", owner)
	c.OracleID = adventureOracle
	c.ScryfallID = "fixture-scryfall"
	c.Layout = game.LayoutAdventure
	c.Faces = []game.Face{
		{Name: "Fixture Giant", TypeLine: "Creature — Giant", ManaCost: "{2}{R}", Power: 4, Toughness: 3},
		{Name: "Fixture Stomp", TypeLine: "Instant — Adventure", ManaCost: "{1}{R}"},
	}
	c.SetFace(0)
	c.KnownBy = map[uuid.UUID]bool{}
	for _, k := range knowers {
		c.KnownBy[k] = true
	}
	return c
}

// faceOf finds one card in a projected zone and returns the face
// block the picker would read for face `i`.
func faceOf(t *testing.T, z ZoneView, id uuid.UUID, i int) CardFaceView {
	t.Helper()
	c := cardInSeatZone(t, z, id)
	if i < 0 || i >= len(c.Faces) {
		t.Fatalf("card %s ships %d faces, want at least %d", id, len(c.Faces), i+1)
	}
	return c.Faces[i]
}

// The canonical case: an adventure card in its owner's hand ships the
// Adventure half's target clause on `faces[1]`, and the creature
// half's silence on `faces[0]`.
func TestAdventureHalfShipsItsOwnTargetClause(t *testing.T) {
	withFaceCatalog(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	bear := game.NewCard("Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(bear)

	c := adventureCard(me.ID, me.ID)
	id := c.InstanceID
	me.Hand.PushTop(c)

	hand := ViewOfGameFor(g, me.ID.String()).Seats[0].Hand
	card := cardInSeatZone(t, hand, id)

	// The card itself is still the ACTIVE face's answer, unchanged:
	// face 0, the creature, which targets nothing.
	if card.TargetMode != "" || card.LegalTargets != nil {
		t.Errorf("the card carries the Adventure half's clause on its top-level block: mode %q targets %+v",
			card.TargetMode, card.LegalTargets)
	}

	front := faceOf(t, hand, id, 0)
	if front.TargetMode != "" || front.LegalTargets != nil {
		t.Errorf("faces[0] (the creature) carries a target clause: mode %q targets %+v",
			front.TargetMode, front.LegalTargets)
	}

	back := faceOf(t, hand, id, 1)
	if back.TargetMode != "any" {
		t.Errorf("faces[1].target_mode = %q, want \"any\" — this is the field the picker opens on", back.TargetMode)
	}
	if back.LegalTargets == nil {
		t.Fatalf("faces[1] ships no legal_targets; the client has nothing to open a picker over")
	}
	if len(back.LegalTargets.Cards) != 1 || back.LegalTargets.Cards[0] != bear.InstanceID.String() {
		t.Errorf("faces[1].legal_targets.cards = %v, want just the bear", back.LegalTargets.Cards)
	}
	// Priced off the FACE, not off the card: the two halves have
	// different mana costs and a stamp that read the card's would
	// size every cost-shaped answer against the wrong one.
	if back.ManaCost != "{1}{R}" {
		t.Errorf("faces[1].mana_cost = %q, want the Adventure's {1}{R}", back.ManaCost)
	}
}

// A single-faced card carries no `faces` at all, so it carries no
// per-face blocks either — which is what bounds the cost of this to
// the number of multi-face cards on the table rather than to the
// ~33,000 ordinary oracle IDs.
func TestSingleFacedCardShipsNoPerFaceBlock(t *testing.T) {
	withFaceCatalog(t)
	g := buildActiveGame(t)
	me := g.Seats[0]

	c := game.NewCard("Ordinary Instant", me.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{R}"
	c.OracleID = adventureOracle
	c.KnownBy = map[uuid.UUID]bool{me.ID: true}
	id := c.InstanceID
	me.Hand.PushTop(c)

	card := cardInSeatZone(t, ViewOfGameFor(g, me.ID.String()).Seats[0].Hand, id)
	if len(card.Faces) != 0 {
		t.Errorf("a single-faced card ships %d faces", len(card.Faces))
	}
}

// The per-viewer half, and it is the #1055 sentence one level down: a
// face's legal target set is narrowed by hexproof, shroud, protection
// and "target opponent" exactly as a card's is, so it is the asking
// seat's answer and nobody else's — including on a revealed hand card,
// which is the surface #1166 is about.
func TestPerFaceLegalTargetsAreTheViewersOwn(t *testing.T) {
	withFaceCatalog(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	bear := game.NewCard("Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(bear)

	// Revealed: the opponent is a knower, so keepKnownInHandZone keeps
	// the card and the assertion below is answered by the STAMP rather
	// than by the redaction.
	c := adventureCard(me.ID, me.ID, opp.ID)
	id := c.InstanceID
	me.Hand.PushTop(c)

	mine := faceOf(t, ViewOfGameFor(g, me.ID.String()).Seats[0].Hand, id, 1)
	if mine.LegalTargets == nil {
		t.Fatalf("the hand's owner lost their own per-face legal target set")
	}
	// #1172: and the NESTED one, on the same face. The owner keeps it
	// because the promotion assigns the whole block.
	if mine.Modes == nil {
		t.Fatalf("the hand's owner lost the Adventure half's `modes`; the nested assertions are vacuous")
	}
	if lt := mine.Modes.Options[0].LegalTargets; lt == nil || len(lt.Cards) != 1 {
		t.Errorf("the hand's owner lost their own per-face NESTED legal target set: %+v", lt)
	}

	theirs := faceOf(t, ViewOfGameFor(g, opp.ID.String()).Seats[0].Hand, id, 1)
	if theirs.LegalTargets != nil {
		t.Errorf("a knower of a revealed hand card got its OWNER's per-face legal target set: %+v", theirs.LegalTargets)
	}
	// #1169 + #1172: in a HAND the parent goes too, so there is no
	// surviving `modes` for a nested set to ride in on. Asserted on
	// the parent AND on the nested field, because the two are
	// different failures — a hand that kept `modes` would be the
	// #1169 regression, and a `modes` that kept its legal sets on a
	// public pile is the #1172 one.
	if theirs.Modes != nil {
		t.Errorf("a knower of a revealed hand card got its owner's per-face `modes`: %+v", theirs.Modes)
	}
	// `target_mode` is printed text about the card and stays, the way
	// it does on the card's own block — it is what a player reading a
	// revealed card in paper already knows.
	if theirs.TargetMode != "any" {
		t.Errorf("faces[1].target_mode = %q on a revealed card, want the public \"any\"", theirs.TargetMode)
	}

	// The spectator, who has no seat to cast from, gets a stronger
	// answer than an unset block: hideZoneContents drops every hand
	// from their frame, faces and all.
	if got := FilterViewFor(ViewOfGame(g), "").Seats[0].Hand.Cards; len(got) != 0 {
		t.Errorf("a spectator received %d hand cards", len(got))
	}
}

// The per-viewer promotion must not bleed between frames. FilterViewFor
// runs once per viewer over ONE GameView, and `faces` is a slice header
// every copy shares — so a promotion written in place would reach every
// frame built after it.
func TestPerFacePromotionDoesNotLeakBetweenFrames(t *testing.T) {
	withFaceCatalog(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	bear := game.NewCard("Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(bear)

	c := adventureCard(me.ID, me.ID, opp.ID)
	id := c.InstanceID
	me.Hand.PushTop(c)

	// Build ONE view and filter it twice, owner first. The owner's
	// promotion happens on the first pass; the opponent's pass must
	// not see it.
	v := ViewOfGame(g)
	own := faceOf(t, FilterViewFor(v, me.ID.String()).Seats[0].Hand, id, 1)
	if own.LegalTargets == nil {
		t.Fatalf("owner's frame: per-face legal targets missing")
	}
	theirs := faceOf(t, FilterViewFor(v, opp.ID.String()).Seats[0].Hand, id, 1)
	if theirs.LegalTargets != nil {
		t.Errorf("the owner's promotion reached a later frame: %+v", theirs.LegalTargets)
	}
	// And the owner's own frame is stable across repeated filtering,
	// which is the property `knowers` and `castOffers` are both
	// cleared on the way out for.
	again := faceOf(t, FilterViewFor(v, me.ID.String()).Seats[0].Hand, id, 1)
	if again.LegalTargets == nil {
		t.Errorf("owner's second frame lost the per-face legal targets; the stamps were consumed rather than copied")
	}
}

// CR 715.4's grant names the CREATURE half, so the only face a cast
// out of exile may choose is face 0 — and the view must publish a
// block for that face and no other, or the client's picker offers a
// row the announce path refuses.
func TestAGrantThatNamesAFaceNarrowsThePerFaceBlocks(t *testing.T) {
	withFaceCatalog(t)
	g := buildActiveGame(t)
	me := g.Seats[0]

	c := adventureCard(me.ID, me.ID, g.Seats[1].ID)
	id := c.InstanceID
	g.Exile.PushTop(c)
	g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
		Player:   me.ID,
		Duration: game.WhileInZoneDuration(),
		CastOnly: true,
		Faces:    []int{0},
		Label:    "Adventure — cast Fixture Giant from exile",
	})

	card := cardInSeatZone(t, ViewOfGameFor(g, me.ID.String()).Exile, id)
	if len(card.Faces) != 2 {
		t.Fatalf("faces = %d, want both printed halves on the wire", len(card.Faces))
	}
	// Face 1 carries nothing: the grant does not open it, so there is
	// no announce surface to describe.
	if card.Faces[1].TargetMode != "" || card.Faces[1].LegalTargets != nil {
		t.Errorf("faces[1] carries an announce block under a grant that names face 0 only: %+v", card.Faces[1])
	}
}

// An impulse grant over the same card names NO face, so both halves
// are open (the 2026-09-18 ADR 0034 amendment: "a Ragavan that
// impulse-exiles an adventure card opens BOTH halves") — and both
// carry a block, for the holder alone.
func TestAFacelessGrantOpensEveryFaceForTheHolderAlone(t *testing.T) {
	withFaceCatalog(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	bear := game.NewCard("Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(bear)

	c := adventureCard(me.ID, me.ID, opp.ID)
	id := c.InstanceID
	g.Exile.PushTop(c)
	g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
		Player:   me.ID,
		Duration: game.WhileInZoneDuration(),
	})

	holder := faceOf(t, ViewOfGameFor(g, me.ID.String()).Exile, id, 1)
	if holder.TargetMode != "any" || holder.LegalTargets == nil {
		t.Errorf("the grant holder cannot see the Adventure half's clause: mode %q targets %+v",
			holder.TargetMode, holder.LegalTargets)
	}
	// #1172: including the nested set inside the public `modes`
	// field, which the promotion hands over with the rest of the
	// block.
	if holder.Modes == nil || holder.Modes.Options[0].LegalTargets == nil {
		t.Errorf("the grant holder cannot see the Adventure half's nested legal targets: %+v", holder.Modes)
	}

	// Exile has no owner, so NOTHING about a holder's answer is
	// public there (#978) — a bystander gets the printed face list
	// and no announce block on any face.
	bystander := faceOf(t, ViewOfGameFor(g, opp.ID.String()).Exile, id, 1)
	if bystander.TargetMode != "" || bystander.LegalTargets != nil {
		t.Errorf("a bystander got the holder's per-face announce block: %+v", bystander)
	}
	// #1172: and no nested one either — asserted separately from the
	// block above, because a nested set is what survived a strip of
	// the fields around it for the whole of #1169.
	if bystander.Modes != nil {
		t.Errorf("a bystander got the holder's per-face `modes`: %+v", bystander.Modes)
	}
}

// A non-knower gets no faces at all, so there is nothing for a
// per-face block to hang off — the same allowlist the rest of the
// cost surface is redacted by.
func TestANonKnowerGetsNoPerFaceBlock(t *testing.T) {
	withFaceCatalog(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	c := adventureCard(me.ID, me.ID)
	id := c.InstanceID
	g.Exile.PushTop(c)

	card := cardInSeatZone(t, ViewOfGameFor(g, opp.ID.String()).Exile, id)
	if len(card.Faces) != 0 {
		t.Errorf("a non-knower received %d faces: %+v", len(card.Faces), card.Faces)
	}
}
