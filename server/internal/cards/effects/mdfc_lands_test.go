package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mdfc_lands_test.go — the three bugs ADR 0034 set out to close,
// written as the reproduction first and the fix second.
//
//  1. Sea Gate Restoration cast as a SORCERY costs {4}{U}{U}{U}.
//     Before the face model its imported ManaCost was "" (Scryfall
//     nulls the top-level cost on every modal DFC), ParseCost("")
//     succeeded with the zero cost, and the spell was FREE.
//  2. Sea Gate, Reborn PLAYED as a land enters tapped unless you pay
//     3 life, with a real prompt and a real entry replacement.
//  3. A creature-front MDFC — Kazandu Mammoth — is no longer a free
//     land drop. Its top-level type line is "Creature — Elephant //
//     Land", Card.IsLand() is a substring scan, and CastSpell's land
//     branch RETURNS BEFORE THE COST GATE. Every ZNR MDFC in the
//     game was a one-mana-free Elephant that skipped the stack.
//
// Each fixture is imported the way a real card is — through the
// per-face data — rather than assembled by hand, so what is being
// tested is the whole chain from Scryfall shape to battlefield.

const (
	// Oracle IDs from the Scryfall default-cards dump.
	seaGateRestorationOracle = "4a8d41fe-e04d-484b-a7d1-19be311e6ca7"
	kazanduMammothOracle     = "2ac1c95c-2a9d-40bc-9cad-9cadfa3f19f7"
)

// mdfcCard builds a two-faced game.Card the way deck.toGameCard
// does: per-face printed data, then SetFace(0) to materialise the
// front. The flat fields are deliberately NOT set here — if the face
// model is not carrying them, the card arrives blank and every
// assertion below fails loudly rather than silently passing on a
// hand-stamped value.
func mdfcCard(owner uuid.UUID, oracleID, layout string, faces ...game.Face) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracleID,
		Layout:     layout,
		Faces:      faces,
		Owner:      owner,
		Controller: owner,
	}
	c.SetFace(0)
	return c
}

// seaGateRestoration is the ZNR mythic: a seven-mana sorcery whose
// back is a pay-3-life land.
func seaGateRestoration(owner uuid.UUID) game.Card {
	return mdfcCard(owner, seaGateRestorationOracle, game.LayoutModalDFC,
		game.Face{
			Name:     "Sea Gate Restoration",
			TypeLine: "Sorcery",
			ManaCost: "{4}{U}{U}{U}",
			Colors:   []string{"U"},
		},
		game.Face{
			Name:     "Sea Gate, Reborn",
			TypeLine: "Land",
			ManaCost: "",
		},
	)
}

// kazanduMammoth is the creature-front half of the problem: a
// three-mana 3/3 whose back is a tapped green land, and whose
// top-level type line therefore contains the substring "land".
func kazanduMammoth(owner uuid.UUID) game.Card {
	return mdfcCard(owner, kazanduMammothOracle, game.LayoutModalDFC,
		game.Face{
			Name:      "Kazandu Mammoth",
			TypeLine:  "Creature — Elephant",
			ManaCost:  "{1}{G}{G}",
			Colors:    []string{"G"},
			Power:     3,
			Toughness: 3,
		},
		game.Face{
			Name:     "Kazandu Valley",
			TypeLine: "Land",
		},
	)
}

// handWithMDFC seats a game in a main phase with `card` in the
// active player's hand, and returns the game, the player and the
// card's instance ID.
func handWithMDFC(t *testing.T, build func(uuid.UUID) game.Card) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := build(me.ID)
	me.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g, me, c.InstanceID
}

// --- #289 / #265: the front face is not a land ----------------------

// TestMDFCFrontFaceIsNotALand is the type-line collision itself, at
// the smallest possible scale. Before ADR 0034 the imported TypeLine
// was the concatenation "Sorcery // Land" and typeLineHas found
// "land" inside it.
func TestMDFCFrontFaceIsNotALand(t *testing.T) {
	c := seaGateRestoration(uuid.New())
	if c.Name != "Sea Gate Restoration" {
		t.Errorf("front face name = %q, want %q", c.Name, "Sea Gate Restoration")
	}
	if c.TypeLine != "Sorcery" {
		t.Errorf("front face type line = %q, want %q — the composite "+
			"%q is what made IsLand() true", c.TypeLine, "Sorcery", "Sorcery // Land")
	}
	if c.IsLand() {
		t.Error("Sea Gate Restoration's FRONT face reports IsLand(); " +
			"CastSpell's land branch would move it to the battlefield " +
			"and return before the cost gate (#289)")
	}
	if !c.IsSorcery() {
		t.Error("front face does not report IsSorcery()")
	}
	if c.ManaCost != "{4}{U}{U}{U}" {
		t.Errorf("front face cost = %q, want %q — the null top-level "+
			"cost is what made every modal DFC free", c.ManaCost, "{4}{U}{U}{U}")
	}

	// And the back face, the other way round.
	c.SetFace(1)
	if !c.IsLand() {
		t.Error("back face does not report IsLand()")
	}
	if c.IsSorcery() {
		t.Error("back face still reports IsSorcery(); the flat fields " +
			"did not re-materialise")
	}
	if c.Name != "Sea Gate, Reborn" {
		t.Errorf("back face name = %q, want %q", c.Name, "Sea Gate, Reborn")
	}
}

// TestMDFCTypeLineHasNoJunkTypes pins the ParseTypeLine half: the
// concatenated line tokenised into the literal types "//" and "—",
// wrong on all 501 DFC oracle IDs.
func TestMDFCTypeLineHasNoJunkTypes(t *testing.T) {
	c := kazanduMammoth(uuid.New())
	eff := c.Effective()
	for _, ty := range append(append([]string{}, eff.Types...), eff.Subtypes...) {
		if ty == "//" || ty == "—" || ty == "-" {
			t.Errorf("junk type %q in %v / %v", ty, eff.Types, eff.Subtypes)
		}
	}
	if len(eff.Types) != 1 || eff.Types[0] != "Creature" {
		t.Errorf("types = %v, want exactly [Creature]", eff.Types)
	}
	if len(eff.Subtypes) != 1 || eff.Subtypes[0] != "Elephant" {
		t.Errorf("subtypes = %v, want exactly [Elephant]", eff.Subtypes)
	}
}

// TestMDFCCreatureFrontIsNotAFreeLandDrop is #289's remaining half
// end to end. Cast with no face parameter — which is what every
// client that has never heard of faces sends — Kazandu Mammoth must
// go on the STACK as a creature spell, not onto the battlefield as a
// land, and must be charged for.
func TestMDFCCreatureFrontIsNotAFreeLandDrop(t *testing.T) {
	g, me, id := handWithMDFC(t, kazanduMammoth)

	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	if err == nil {
		t.Fatal("cast with an empty mana pool succeeded — the land " +
			"branch fired on a creature and skipped the cost gate (#289)")
	}
	var insufficient *game.InsufficientManaError
	if !errors.As(err, &insufficient) {
		t.Fatalf("CastSpell error = %v, want an InsufficientManaError "+
			"for {1}{G}{G}", err)
	}
	if g.Battlefield.Contains(id) {
		t.Fatal("Kazandu Mammoth reached the battlefield as a FREE LAND")
	}
	if n := g.LandsPlayedThisTurn[me.ID]; n != 0 {
		t.Errorf("land drops used = %d, want 0 — casting a creature "+
			"spell is not a land drop", n)
	}
}

// --- the picker: choosing the back face -----------------------------

// TestMDFCBackFacePlayedAsALand is the MDFC picker's whole point.
// Face 1 is a land, so it takes the special-action path: no stack, a
// land drop spent, and — because this particular back prints the
// shockland clause with a 3 — a real entry prompt.
func TestMDFCBackFacePlayedAsALand(t *testing.T) {
	g, me, id := handWithMDFC(t, seaGateRestoration)
	lifeBefore := me.Life

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("play Sea Gate, Reborn as a land: %v", err)
	}

	// The entry is REPLACED, so nothing has entered yet.
	c := entryPayLifeChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("playing an MDFC land back queued no pay-life prompt — " +
			"the back face's spec was not found under its composite " +
			"catalog key")
	}
	if c.PayCost != "3 life" {
		t.Errorf("prompt cost = %q, want %q", c.PayCost, "3 life")
	}
	if !strings.Contains(c.Reason, "Sea Gate, Reborn") {
		t.Errorf("prompt names %q; it should name the BACK face", c.Reason)
	}
	if g.Battlefield.Contains(id) {
		t.Error("the land entered before the choice was answered")
	}

	answerEntryPayLife(t, g, me.ID, false)

	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("answering the prompt did not resume the entry")
	}
	if !card.Tapped {
		t.Error("declined the 3 life and the land entered untapped")
	}
	if card.ActiveFace != 1 {
		t.Errorf("permanent's ActiveFace = %d, want 1 — an MDFC keeps "+
			"the face it was played as", card.ActiveFace)
	}
	if card.Name != "Sea Gate, Reborn" {
		t.Errorf("permanent name = %q, want the back face", card.Name)
	}
	if !card.IsLand() || card.IsSorcery() {
		t.Errorf("permanent type line = %q, want the back face's Land",
			card.TypeLine)
	}
	if me.Life != lifeBefore {
		t.Errorf("life changed on a decline: %d → %d", lifeBefore, me.Life)
	}
	// The discriminator from the shockland tests: the land ENTERED
	// tapped, so nothing tapped it.
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped", n)
	}
	if n := g.LandsPlayedThisTurn[me.ID]; n != 1 {
		t.Errorf("land drops used = %d, want 1", n)
	}
}

// TestMDFCBackFacePaidThreeLifeEntersUntapped is the other branch.
func TestMDFCBackFacePaidThreeLifeEntersUntapped(t *testing.T) {
	g, me, id := handWithMDFC(t, seaGateRestoration)
	lifeBefore := me.Life

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("play as a land: %v", err)
	}
	answerEntryPayLife(t, g, me.ID, true)

	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Sea Gate, Reborn is not on the battlefield")
	}
	if card.Tapped {
		t.Error("paid 3 life and the land still entered tapped")
	}
	if got := lifeBefore - me.Life; got != 3 {
		t.Errorf("life paid = %d, want 3", got)
	}
	if n := untapEventsFor(g, id); n != 0 {
		t.Errorf("%d untap events; it entered untapped, it was not untapped", n)
	}
}

// TestMDFCFrontFaceCostsItsRealMana is #289's first half: cast as
// the sorcery it is, Sea Gate Restoration owes {4}{U}{U}{U} and a
// player with nothing in their pool cannot pay it.
func TestMDFCFrontFaceCostsItsRealMana(t *testing.T) {
	g, me, id := handWithMDFC(t, seaGateRestoration)

	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	if err == nil {
		t.Fatal("cast Sea Gate Restoration for FREE with an empty pool")
	}
	var insufficient *game.InsufficientManaError
	if !errors.As(err, &insufficient) {
		t.Fatalf("error = %v, want InsufficientManaError", err)
	}
	if g.Battlefield.Contains(id) {
		t.Fatal("the front face reached the battlefield — the land " +
			"branch fired on a sorcery (#265)")
	}
	if n := g.LandsPlayedThisTurn[me.ID]; n != 0 {
		t.Errorf("land drops used = %d, want 0", n)
	}
}

// TestMDFCFrontFaceGoesToTheStack completes it: with the mana
// available the sorcery is a normal spell on a normal stack.
func TestMDFCFrontFaceGoesToTheStack(t *testing.T) {
	g, me, id := handWithMDFC(t, seaGateRestoration)

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Fatal("the sorcery is not on the stack")
	}
	if g.Battlefield.Contains(id) {
		t.Fatal("the sorcery is on the battlefield")
	}
	for _, c := range g.Stack.Cards {
		if c.InstanceID != id {
			continue
		}
		if c.ActiveFace != 0 {
			t.Errorf("stack card ActiveFace = %d, want 0", c.ActiveFace)
		}
		if c.Name != "Sea Gate Restoration" {
			t.Errorf("stack card name = %q, want the front face", c.Name)
		}
	}
}

// --- face validation ------------------------------------------------

// TestCastRejectsAFaceTheCardDoesNotHave pins the refusal. Clamping
// to the front instead would silently cast the wrong half.
func TestCastRejectsAFaceTheCardDoesNotHave(t *testing.T) {
	for _, face := range []int{2, -1, 7} {
		g, me, id := handWithMDFC(t, seaGateRestoration)
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: face})
		if !errors.Is(err, game.ErrInvalidFace) {
			t.Errorf("face %d: error = %v, want ErrInvalidFace", face, err)
		}
		if g.Stack.Contains(id) || g.Battlefield.Contains(id) {
			t.Errorf("face %d: the card moved anyway", face)
		}
	}
}

// TestTransformBackFaceIsNotCastable is CR 712.11. A transform card
// is always cast as its front face; its back is reached by
// transforming the permanent, which is a later PR. Offering the back
// at announce would be strictly wrong rules, so it is refused.
func TestTransformBackFaceIsNotCastable(t *testing.T) {
	g, me, id := handWithMDFC(t, func(owner uuid.UUID) game.Card {
		return mdfcCard(owner, "0de1d4dd-9d4e-4c3d-b6ff-1f0b0a4a7c5e", game.LayoutTransform,
			game.Face{
				Name: "Delver of Secrets", TypeLine: "Creature — Human Wizard",
				ManaCost: "{U}", Colors: []string{"U"}, Power: 1, Toughness: 1,
			},
			game.Face{
				Name: "Insectile Aberration", TypeLine: "Creature — Human Insect",
				Colors: []string{"U"}, Power: 3, Toughness: 2,
			},
		)
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); !errors.Is(err, game.ErrInvalidFace) {
		t.Fatalf("casting a transform card's back face: error = %v, want ErrInvalidFace", err)
	}
	// The front face still casts normally.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the front face: %v", err)
	}
}

// --- the failed-cast rollback ---------------------------------------

// TestRejectedLandPlayLeavesTheFaceAlone pins the source-zone write
// the land branch has to do. The face is stamped on the card IN HAND
// before the replacement pipeline runs, because the pipeline looks
// the card up by ID; a cast that then fails must put it back, or the
// player is left holding a card wearing the wrong face.
func TestRejectedLandPlayLeavesTheFaceAlone(t *testing.T) {
	g, me, id := handWithMDFC(t, seaGateRestoration)
	// Sorcery speed is closed while something is on the stack, so
	// put an opponent's spell there and try to play the land under it.
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.Stack.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Ambush", TypeLine: "Instant",
		Owner: them.ID, Controller: them.ID,
	})

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err == nil {
		t.Fatal("played a land with a non-empty stack")
	}
	for _, c := range me.Hand.Cards {
		if c.InstanceID != id {
			continue
		}
		if c.ActiveFace != 0 {
			t.Errorf("card in hand left on face %d after a refused "+
				"land play; it should be back on the front", c.ActiveFace)
		}
		if c.Name != "Sea Gate Restoration" {
			t.Errorf("card in hand is named %q after a refused land play", c.Name)
		}
	}
}

// --- the catalog key ------------------------------------------------

// TestBackFaceSpecsAreKeyedByFace pins the composite key from both
// ends: the cycle registered 60 back faces, none of them collided
// with a front face, and CatalogKey reproduces the string Register
// was given.
func TestBackFaceSpecsAreKeyedByFace(t *testing.T) {
	c := seaGateRestoration(uuid.New())
	if got := game.CatalogKey(c); got != seaGateRestorationOracle {
		t.Errorf("front-face key = %q, want the BARE oracle ID %q — "+
			"face 0 keeping it is what makes this change a no-op for "+
			"every single-faced card", got, seaGateRestorationOracle)
	}
	// Both faces are registered (the front in sea_gate_restoration.go),
	// and the two keys must reach two different specs.
	if front, ok := Lookup(game.CatalogKey(c)); !ok || front.Name != "Sea Gate Restoration" {
		t.Errorf("the FRONT face must resolve to the sorcery's own spec, got ok=%v name=%q",
			ok, front.Name)
	}

	c.SetFace(1)
	want := seaGateRestorationOracle + "#1"
	if got := game.CatalogKey(c); got != want {
		t.Errorf("back-face key = %q, want %q", got, want)
	}
	if !Has(want) {
		t.Fatalf("no spec registered under %q", want)
	}
	spec, _ := Lookup(want)
	if spec.Name != "Sea Gate, Reborn" {
		t.Errorf("spec name = %q, want the BACK face's", spec.Name)
	}
	if len(spec.Replacements) != 1 {
		t.Fatalf("back face has %d replacements, want 1 (the pay-3-life clause)",
			len(spec.Replacements))
	}
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{U}" {
		t.Errorf("back face mana abilities = %v, want one producing {U}",
			spec.ManaAbilities)
	}
}

// TestMDFCLandBackCycleIsComplete guards the generated table against
// a partial edit. Sixty of the hundred modal_dfc oracle IDs in Magic
// have a land back; all sixty are registered, every one of them under
// a "#1" key, and every one taps for mana.
//
// It walks mdfcLandBackKeys rather than every "#1" spec in the
// catalog. Since S32 the back-face keyspace has a second tenant — the
// Sieges, whose back faces are creatures and enchantments and tap for
// nothing — so a suffix scan would fail on cards this cycle has never
// heard of. See mdfcLandBackKeys.
func TestMDFCLandBackCycleIsComplete(t *testing.T) {
	for _, key := range mdfcLandBackKeys {
		if !strings.HasSuffix(key, "#1") {
			t.Errorf("MDFC land back %q is not keyed on face 1", key)
		}
		s, ok := Lookup(key)
		if !ok {
			t.Errorf("MDFC land back %q is listed but not registered", key)
			continue
		}
		if s.Name == "" {
			t.Errorf("back-face spec %q has no name", s.OracleID)
		}
		if len(s.ManaAbilities) != 1 {
			t.Errorf("%s (%s): %d mana abilities, want 1 — every MDFC "+
				"land back taps for mana", s.Name, s.OracleID,
				len(s.ManaAbilities))
		}
	}
	if n := len(mdfcLandBackKeys); n != 60 {
		t.Errorf("registered MDFC land backs = %d, want 60", n)
	}
}
