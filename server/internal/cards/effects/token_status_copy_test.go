package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_status_copy_test.go — #4296. Being a token is a status of the
// object (CR 111.1, CR 111.6), not one of the copiable values CR 707.2
// lists, so no copy effect may hand it over or take it away: a token
// that becomes a copy of a nontoken stays a token, and a nontoken that
// becomes a copy of a token stays a card. The engine reads token status
// off the printed type line (Card.IsToken), so a copy that overwrites
// the type line with the copied object's has to carry the copier's
// status across it.

// tokenCloneCopying makes one token copy of a battlefield Clone, answers
// its "enter as a copy?" question with `pick`, and returns the token.
func tokenCloneCopying(t *testing.T, g *game.Game, controller, clone, pick uuid.UUID) uuid.UUID {
	t.Helper()
	before := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		before[c.InstanceID] = true
	}
	g.WithWriteLock(func() {
		tmpl, ok := TokenCopyTemplate(g, clone)
		if !ok {
			t.Fatal("TokenCopyTemplate failed")
		}
		if err := g.CreateTokenForEffect(controller, tmpl, 1); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
	c := copyPrompt(g)
	if c == nil {
		t.Fatal("the token Clone did not ask what to copy")
	}
	if err := g.ResolveCopyTarget(c.ID, c.Chooser, pick); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if !before[c.InstanceID] {
			return c.InstanceID
		}
	}
	t.Fatal("no new permanent entered")
	return uuid.Nil
}

func seedGoblinToken(g *game.Game, controller uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: controller, Controller: controller,
	})
}

func tokenSupertypes(typeLine string) int {
	n := 0
	for _, f := range strings.Fields(typeLine) {
		if f == "Token" {
			n++
		}
	}
	return n
}

// The report: a token copy of Clone that copies a nontoken creature.
func TestTokenStatusSurvivesCopyTokenCloneCopiesNontoken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	clone := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: oracleClone,
		Owner: me.ID, Controller: me.ID,
	})
	tok := tokenCloneCopying(t, g, me.ID, clone, bears)

	got := copyBattlefieldCard(t, g, tok)
	if got.Name != "Grizzly Bears" || !got.IsCopy() {
		t.Fatalf("setup: the token is %q (copy %v), want a copy of Grizzly Bears", got.Name, got.IsCopy())
	}
	if !got.IsToken() {
		t.Errorf("the token copy of Clone lost its token status: type line %q", got.TypeLine)
	}
	if !got.IsCreature() {
		t.Errorf("the token is not a creature: %q", got.TypeLine)
	}
	if bc := copyBattlefieldCard(t, g, bears); bc.IsToken() {
		t.Errorf("the original Grizzly Bears became a token: %q", bc.TypeLine)
	}
	if cc := copyBattlefieldCard(t, g, clone); cc.IsToken() {
		t.Errorf("the real Clone became a token: %q", cc.TypeLine)
	}

	// CR 704.5d: a token in a zone other than the battlefield ceases to
	// exist. The copy that dies must not linger in the graveyard as a card.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tok) })
	g.RunStateChecksForTest()
	for _, c := range g.Seats[0].Graveyard.Cards {
		if c.InstanceID == tok {
			t.Errorf("the dead token copy of Clone is a card in the graveyard: %q", c.TypeLine)
		}
	}
}

// A token Clone copying a token is one Token, not two.
func TestTokenStatusSurvivesCopyTokenCloneCopiesToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gob := seedGoblinToken(g, me.ID)
	clone := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: oracleClone,
		Owner: me.ID, Controller: me.ID,
	})
	tok := tokenCloneCopying(t, g, me.ID, clone, gob)
	got := copyBattlefieldCard(t, g, tok)
	if !got.IsToken() || tokenSupertypes(got.TypeLine) != 1 {
		t.Errorf("type line = %q, want exactly one Token", got.TypeLine)
	}
}

// The reverse: a nontoken Clone that copies a token is still a card.
func TestTokenStatusSurvivesCopyNontokenCloneCopiesToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gob := seedGoblinToken(g, me.ID)

	id := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, gob)

	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Goblin" || !got.IsCopy() {
		t.Fatalf("setup: the Clone is %q (copy %v), want a copy of the Goblin token", got.Name, got.IsCopy())
	}
	if got.IsToken() {
		t.Errorf("a Clone that copied a token became a token: type line %q", got.TypeLine)
	}
	// It is a card: it goes to the graveyard and stays there as the
	// Clone card, the copy undone (CR 400.7, CR 704.5d reads tokens only).
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	found := false
	for _, c := range g.Seats[g.Turn.ActiveSeat].Graveyard.Cards {
		if c.InstanceID == id {
			found = true
			if c.Name != "Clone" || c.IsToken() {
				t.Errorf("the graveyard card is %q (%q), want the Clone card", c.Name, c.TypeLine)
			}
		}
	}
	if !found {
		t.Error("the destroyed Clone is not in the graveyard")
	}
}

// Another entry copy, with an except clause on top: a Phyrexian Metamorph
// copying a Treasure token is an artifact card, not a token.
func TestTokenStatusSurvivesCopyMetamorphCopiesToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tr := TreasureToken()
	tr.InstanceID = uuid.New()
	tr.Owner, tr.Controller = me.ID, me.ID
	treasure := pushBattlefieldCardWithTimestamp(g, tr)

	id := castCatalogSpell(t, g, "Phyrexian Metamorph",
		"Artifact Creature — Phyrexian Shapeshifter", oracleMetamorph, nil)
	resolveWithCopyChoice(t, g, treasure)

	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Treasure" || !got.IsArtifact() {
		t.Fatalf("setup: the Metamorph is %q %q, want a Treasure artifact", got.Name, got.TypeLine)
	}
	if got.IsToken() {
		t.Errorf("a Metamorph that copied a Treasure token became a token: %q", got.TypeLine)
	}
}

// The on-battlefield copy (Cytoshape, Mirage Mirror, Unstable Shapeshifter
// shape): a token that becomes a copy of a card stays a token, a card that
// becomes a copy of a token stays a card, and both are themselves again
// when the copy ends.
func TestTokenStatusSurvivesCopyDurationCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	gob := seedGoblinToken(g, me.ID)
	ogre := seedCopyableCreature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)

	g.WithWriteLock(func() {
		// Both sets of values are read before either copy applies.
		ofBears, _ := g.CopiableValuesForEffect(bears)
		ofGoblin, _ := g.CopiableValuesForEffect(gob)
		g.BecomeCopyForEffect(uuid.Nil, bears, []uuid.UUID{gob}, ofBears, g.UntilEndOfTurnDuration(), "probe")
		g.BecomeCopyForEffect(uuid.Nil, gob, []uuid.UUID{ogre}, ofGoblin, g.UntilEndOfTurnDuration(), "probe")
	})
	if c := dcCard(t, g, gob); c.Name != "Grizzly Bears" || !c.IsToken() {
		t.Errorf("the token as a Grizzly Bears is %q (%q), want a token copy", c.Name, c.TypeLine)
	}
	if c := dcCard(t, g, ogre); c.Name != "Goblin" || c.IsToken() {
		t.Errorf("the card as a Goblin is %q (%q), want a nontoken copy", c.Name, c.TypeLine)
	}
	dcCleanup(g)
	if c := dcCard(t, g, gob); c.Name != "Goblin" || !c.IsToken() || tokenSupertypes(c.TypeLine) != 1 {
		t.Errorf("after cleanup the token is %q (%q), want the Goblin token", c.Name, c.TypeLine)
	}
	if c := dcCard(t, g, ogre); c.Name != "Hill Giant" || c.IsToken() {
		t.Errorf("after cleanup the card is %q (%q), want the Hill Giant card", c.Name, c.TypeLine)
	}
}
