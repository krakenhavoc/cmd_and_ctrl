package effects

import (
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_copy_copiable_test.go — #4794: a token copy takes the copied
// object's COPIABLE values (CR 707.2), and its last-known ones when it
// has left the battlefield before the effect resolves (CR 608.2h). The
// interesting object is a Clone copying a dragon: its copiable values are
// the dragon's while it is on the battlefield, and the card it leaves in
// the graveyard is a Clone again.

const tokenCopyElectroduplicateOracle = "112f2b3f-32e7-40b1-b80e-0b99184a840c"

// tokenCopyCloneOfDragon puts a 5/5 flying Shivan-style dragon and a real
// Clone copying it onto seat 0's battlefield, and leaves the Clone able to
// attack. Returns the dragon's and the Clone's IDs.
func tokenCopyCloneOfDragon(t *testing.T, g *game.Game) (dragon, clone uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	dragon = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Dragon", TypeLine: "Creature — Dragon",
		OracleID: "oracle-test-dragon", ScryfallID: "scryfall-test-dragon",
		ManaCost: "{4}{R}{R}", Colors: []string{"R"}, Keywords: []string{"flying"},
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	clone = castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, dragon)
	if got := copyBattlefieldCard(t, g, clone); got.Name != "Test Dragon" {
		t.Fatalf("setup: the Clone is %q, want a copy of Test Dragon", got.Name)
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == clone {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	return dragon, clone
}

// tokenCopyTemplateOf reads TokenCopyTemplate under the write lock, as an
// effect does.
func tokenCopyTemplateOf(g *game.Game, id uuid.UUID) (tmpl game.Card, ok bool) {
	g.WithWriteLock(func() { tmpl, ok = TokenCopyTemplate(g, id) })
	return tmpl, ok
}

// tokensNamed is every token on the battlefield with that name.
func tokensNamed(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// assertDragonTemplate checks a template is a token copy of the dragon.
func assertDragonTemplate(t *testing.T, tmpl game.Card) {
	t.Helper()
	if tmpl.Name != "Test Dragon" || tmpl.OracleID != "oracle-test-dragon" || tmpl.ScryfallID != "scryfall-test-dragon" {
		t.Errorf("copy is %q (oracle %q, art %q), want Test Dragon", tmpl.Name, tmpl.OracleID, tmpl.ScryfallID)
	}
	if tmpl.TypeLine != "Token Creature — Dragon" {
		t.Errorf("type line %q, want Token Creature — Dragon", tmpl.TypeLine)
	}
	if tmpl.Power != 5 || tmpl.Toughness != 5 || tmpl.ManaCost != "{4}{R}{R}" {
		t.Errorf("copy is %d/%d costing %q, want a 5/5 costing {4}{R}{R}", tmpl.Power, tmpl.Toughness, tmpl.ManaCost)
	}
	if !slices.Contains(tmpl.Keywords, "flying") {
		t.Errorf("keywords %v, want flying", tmpl.Keywords)
	}
}

// TestTokenCopyOfACloneOnTheBattlefieldIsWhatItCopies: the Clone's
// copiable values are the dragon's (CR 707.2), so the token is a dragon.
func TestTokenCopyOfACloneOnTheBattlefieldIsWhatItCopies(t *testing.T) {
	g := newCatalogGame(t)
	_, clone := tokenCopyCloneOfDragon(t, g)
	tmpl, ok := tokenCopyTemplateOf(g, clone)
	if !ok {
		t.Fatal("no template for a Clone on the battlefield")
	}
	assertDragonTemplate(t, tmpl)
}

// TestImpostorSyndromeCopiesACloneThatDiedInResponseAsWhatItCopied is the
// issue's case: the Clone copying a dragon connects, dies with the trigger
// on the stack, and the token is the dragon, from its last-known copiable
// values (CR 608.2h), not the Clone card in the graveyard.
func TestImpostorSyndromeCopiesACloneThatDiedInResponseAsWhatItCopied(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Impostor Syndrome", "Enchantment", spm3ImpostorOracle, "{4}{U}{U}", 0, 0, "U")
	_, clone := tokenCopyCloneOfDragon(t, g)
	attackWith(t, g, opp.ID, clone)
	if stackFullyEmpty(g) {
		t.Fatal("no Impostor Syndrome trigger after the Clone connected")
	}
	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(clone) })
	if err != nil {
		t.Fatalf("destroy the Clone: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := tokensNamed(g, "Clone"); len(got) != 0 {
		t.Errorf("the token is a Clone: it was copied from the graveyard card, not the permanent that left")
	}
	dragons := tokensNamed(g, "Test Dragon")
	if len(dragons) != 1 {
		t.Fatalf("%d Test Dragon tokens, want 1", len(dragons))
	}
	if dragons[0].Power != 5 || dragons[0].Toughness != 5 || dragons[0].OracleID != "oracle-test-dragon" {
		t.Errorf("token is %d/%d oracle %q, want the 5/5 dragon", dragons[0].Power, dragons[0].Toughness, dragons[0].OracleID)
	}
}

// TestTokenCopyOfACardThatMovedOnAfterLeavingIsTheCard: a Clone that died
// this turn and was then exiled from the graveyard (an embalm cost, The
// Scarab God) is a new object the effect names on its own (CR 400.7), so
// the copy is of the Clone card, not of what the departed permanent was.
func TestTokenCopyOfACardThatMovedOnAfterLeavingIsTheCard(t *testing.T) {
	g := newCatalogGame(t)
	_, clone := tokenCopyCloneOfDragon(t, g)
	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(clone) })
	if err != nil {
		t.Fatalf("destroy the Clone: %v", err)
	}
	if tmpl, ok := tokenCopyTemplateOf(g, clone); !ok || tmpl.Name != "Test Dragon" {
		t.Errorf("straight after it died: %q (ok=%v), want its last-known Test Dragon", tmpl.Name, ok)
	}
	g.WithWriteLock(func() { err = g.ExileCardForEffect(clone) })
	if err != nil {
		t.Fatalf("exile the Clone card: %v", err)
	}
	tmpl, ok := tokenCopyTemplateOf(g, clone)
	if !ok || tmpl.Name != "Clone" || tmpl.OracleID != oracleClone {
		t.Errorf("after it moved on to exile: %q oracle %q (ok=%v), want the Clone card", tmpl.Name, tmpl.OracleID, ok)
	}
}

// TestTokenCopyOfAPlainCreatureIsUnchanged pins the template for an
// ordinary printed card, on the battlefield and after it has died: the
// copiable values of a creature that is no copy are its printed values.
func TestTokenCopyOfAPlainCreatureIsUnchanged(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Bear", TypeLine: "Creature — Bear",
		OracleID: "oracle-test-bear", ScryfallID: "scryfall-test-bear",
		ManaCost: "{1}{G}", Colors: []string{"G"}, Keywords: []string{"trample"},
		Power: 2, Toughness: 2, VariableToughness: true, Owner: me.ID, Controller: me.ID,
	})
	want := game.Card{
		Name: "Test Bear", ScryfallID: "scryfall-test-bear", OracleID: "oracle-test-bear",
		TypeLine: "Token Creature — Bear", Power: 2, Toughness: 2, VariableToughness: true,
		ManaCost: "{1}{G}", Colors: []string{"G"}, Keywords: []string{"trample"},
	}
	tmpl, ok := tokenCopyTemplateOf(g, bear)
	if !ok || !reflect.DeepEqual(tmpl, want) {
		t.Errorf("on the battlefield:\n got %+v\nwant %+v", tmpl, want)
	}
	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(bear) })
	if err != nil {
		t.Fatalf("destroy: %v", err)
	}
	tmpl, ok = tokenCopyTemplateOf(g, bear)
	if !ok || !reflect.DeepEqual(tmpl, want) {
		t.Errorf("after it died:\n got %+v\nwant %+v", tmpl, want)
	}
}

// TestTokenCopyOfATokenKeepsItsTokenKey: a token's catalog identity is its
// key, so the copy carries it.
func TestTokenCopyOfATokenKeepsItsTokenKey(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	key := game.TokenKey("treasure")
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure",
		TokenKey: key, Owner: me.ID, Controller: me.ID,
	})
	tmpl, ok := tokenCopyTemplateOf(g, tok)
	if !ok || tmpl.TokenKey != key || tmpl.OracleID != "" || tmpl.TypeLine != "Token Artifact — Treasure" {
		t.Errorf("copy has key %q oracle %q type %q (ok=%v), want %q on a Token Artifact — Treasure", tmpl.TokenKey, tmpl.OracleID, tmpl.TypeLine, ok, key)
	}
}

// TestElectroduplicateOfACloneIsAHastyCopyOfWhatItCopies: the card's
// "except it has haste" applies on top of the copiable values (CR 707.9).
func TestElectroduplicateOfACloneIsAHastyCopyOfWhatItCopies(t *testing.T) {
	g := newCatalogGame(t)
	_, clone := tokenCopyCloneOfDragon(t, g)
	castCatalogSpell(t, g, "Electroduplicate", "Sorcery", tokenCopyElectroduplicateOracle, b16TargetCard(clone))
	passPriorityAroundTable(t, g)
	dragons := tokensNamed(g, "Test Dragon")
	if len(dragons) != 1 {
		t.Fatalf("%d Test Dragon tokens, want 1 (Clone tokens: %d)", len(dragons), len(tokensNamed(g, "Clone")))
	}
	if !slices.Contains(dragons[0].Keywords, "haste") || !slices.Contains(dragons[0].Keywords, "flying") {
		t.Errorf("token keywords %v, want flying and haste", dragons[0].Keywords)
	}
}

// TestTokenCopyOfAFaceDownPermanentIsANamelessTwoTwo: a face-down
// permanent's copiable values are its CR 708.2 body, so the copy is a 2/2
// creature with no name, no mana cost, no text and none of the hidden
// card's art.
func TestTokenCopyOfAFaceDownPermanentIsANamelessTwoTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c := game.Card{
		InstanceID: uuid.New(), Name: "Hidden", TypeLine: "Creature — Horror",
		OracleID: "oracle-hidden", ScryfallID: "scryfall-hidden", ManaCost: "{4}{B}",
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
		KnownBy: map[uuid.UUID]bool{me.ID: true},
	}
	c.SetFaceDown(game.FaceDownMorphed)
	id := pushBattlefieldCardWithTimestamp(g, c)
	tmpl, ok := tokenCopyTemplateOf(g, id)
	if !ok {
		t.Fatal("no template for a face-down permanent")
	}
	if tmpl.Name != "" || tmpl.OracleID != "" || tmpl.ScryfallID != "" || tmpl.ManaCost != "" {
		t.Errorf("copy shows the hidden card: name %q oracle %q art %q cost %q", tmpl.Name, tmpl.OracleID, tmpl.ScryfallID, tmpl.ManaCost)
	}
	if tmpl.Power != 2 || tmpl.Toughness != 2 || tmpl.TypeLine != "Token Creature" {
		t.Errorf("copy is a %d/%d %q, want a 2/2 Token Creature", tmpl.Power, tmpl.Toughness, tmpl.TypeLine)
	}
}
