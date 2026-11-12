package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prototype_test.go — CR 718 / CR 702.160, prototype (#5656). The rules
// side is game/prototype.go; these drive it through real casts of the
// eleven Brothers' War cards.

// prototypeCard builds a prototype card the way deck.toGameCard does:
// the printed card, the inset as a second face, then face 0.
func prototypeCard(t *testing.T, owner uuid.UUID, oracleID, name, typeLine, cost string, power, toughness int, text string) game.Card {
	t.Helper()
	faces := game.PrototypeFaces(game.Face{
		Name:       name,
		TypeLine:   typeLine,
		ManaCost:   cost,
		Power:      power,
		Toughness:  toughness,
		OracleText: text,
	})
	if len(faces) != 2 {
		t.Fatalf("%s: oracle text has no prototype inset", name)
	}
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracleID,
		Layout:     game.LayoutPrototype,
		Owner:      owner,
		Controller: owner,
		Faces:      faces,
	}
	c.SettleImported()
	return c
}

func blitzAutomaton(t *testing.T, owner uuid.UUID) game.Card {
	return prototypeCard(t, owner, "19f344ee-031d-4346-9553-c5d32b02df43", "Blitz Automaton",
		"Artifact Creature — Construct", "{7}", 6, 4,
		"Prototype {2}{R} — 3/2 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\nHaste")
}

// prototypeGame seats the active player at a main phase with the card
// in hand.
func prototypeGame(t *testing.T, build func(*testing.T, uuid.UUID) game.Card) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := build(t, me.ID)
	me.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g, me, c.InstanceID
}

// CR 718.3b: the prototyped spell has the inset's mana cost, colour,
// power and toughness, and so does the permanent it becomes. CR 718.5:
// it keeps its name, types and abilities.
func TestBlitzAutomatonPrototypedIsARedThreeTwoForThree(t *testing.T) {
	g, me, id := prototypeGame(t, blitzAutomaton)
	floatMana(t, g, me, "{R}{C}{C}")

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: game.PrototypeFace, Strict: true}); err != nil {
		t.Fatalf("cast prototyped: %v", err)
	}
	spell, ok := g.LookupCardForEffect(id)
	if !ok || !g.Stack.Contains(id) {
		t.Fatal("the prototyped spell is not on the stack")
	}
	if !spell.Prototyped() {
		t.Error("the spell is not prototyped")
	}
	if got := spell.ManaValue(); got != 3 {
		t.Errorf("spell mana value = %d, want 3 (CR 718.3b)", got)
	}
	if eff := spell.Effective(); eff.Power != 3 || eff.Toughness != 2 || len(eff.Colors) != 1 || eff.Colors[0] != "R" {
		t.Errorf("spell = %d/%d %v, want a red 3/2", eff.Power, eff.Toughness, eff.Colors)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %+v, want it spent on {2}{R}", me.ManaPool)
	}
	passPriorityAroundTable(t, g)

	perm, ok := g.LookupCardForEffect(id)
	if !ok || !g.Battlefield.Contains(id) {
		t.Fatal("the Automaton did not resolve onto the battlefield")
	}
	eff := perm.Effective()
	if eff.Power != 3 || eff.Toughness != 2 {
		t.Errorf("permanent = %d/%d, want 3/2", eff.Power, eff.Toughness)
	}
	if len(eff.Colors) != 1 || eff.Colors[0] != "R" {
		t.Errorf("permanent colours = %v, want red", eff.Colors)
	}
	if perm.ManaValue() != 3 {
		t.Errorf("permanent mana value = %d, want 3", perm.ManaValue())
	}
	if !perm.IsArtifact() || !perm.IsCreature() || !perm.HasSubtype("Construct") || perm.Name != "Blitz Automaton" {
		t.Errorf("CR 718.5: %q %q lost its name or types", perm.Name, perm.TypeLine)
	}
	if !game.HasKeyword(&perm, "haste") {
		t.Error("CR 718.5: the prototyped Automaton lost haste")
	}
}

// Cast normally it is the colourless 6/4 for {7}, and the prototype
// cost does not pay for it.
func TestBlitzAutomatonCastNormally(t *testing.T) {
	g, me, id := prototypeGame(t, blitzAutomaton)
	floatMana(t, g, me, "{R}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast the {7} Automaton with three mana")
	}
	me.ManaPool = nil
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast normally: %v", err)
	}
	passPriorityAroundTable(t, g)
	perm, _ := g.LookupCardForEffect(id)
	eff := perm.Effective()
	if !g.Battlefield.Contains(id) || perm.Prototyped() || eff.Power != 6 || eff.Toughness != 4 || len(eff.Colors) != 0 || perm.ManaValue() != 7 {
		t.Errorf("permanent = %d/%d %v mv %d prototyped=%v, want a colourless 6/4 with mana value 7",
			eff.Power, eff.Toughness, eff.Colors, perm.ManaValue(), perm.Prototyped())
	}
	if !game.HasKeyword(&perm, "haste") {
		t.Error("no haste")
	}
}

// CR 718.4: in every zone but the stack and the battlefield the card
// has only its normal characteristics.
func TestPrototypedPermanentThatLeavesIsTheNormalCard(t *testing.T) {
	g, me, id := prototypeGame(t, blitzAutomaton)
	floatMana(t, g, me, "{R}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: game.PrototypeFace, Strict: true}); err != nil {
		t.Fatalf("cast prototyped: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	card, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatal("card lost")
	}
	eff := card.Effective()
	if card.Prototyped() || card.ManaValue() != 7 || eff.Power != 6 || eff.Toughness != 4 || len(eff.Colors) != 0 {
		t.Errorf("graveyard card = %d/%d %v mv %d, want the colourless 6/4 with mana value 7",
			eff.Power, eff.Toughness, eff.Colors, card.ManaValue())
	}
}

// castPrototyped casts the card prototyped with exactly its inset cost
// floating, and leaves the spell on the stack.
func castPrototyped(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, mana string) {
	t.Helper()
	me.ManaPool = nil
	floatMana(t, g, me, mana)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: game.PrototypeFace, Strict: true}); err != nil {
		t.Fatalf("cast prototyped: %v", err)
	}
}

// wantPrototypedAutomaton checks one object is the red 3/2 with mana
// value 3.
func wantPrototypedAutomaton(t *testing.T, what string, c game.Card) {
	t.Helper()
	eff := c.Effective()
	if c.Name != "Blitz Automaton" || eff.Power != 3 || eff.Toughness != 2 || c.ManaValue() != 3 ||
		len(eff.Colors) != 1 || eff.Colors[0] != "R" || !c.Prototyped() {
		t.Errorf("%s = %q %d/%d %v mv %d prototyped=%v, want a prototyped red 3/2 Blitz Automaton with mana value 3",
			what, c.Name, eff.Power, eff.Toughness, eff.Colors, c.ManaValue(), c.Prototyped())
	}
}

// CR 718.3d: a copy of a permanent that was a prototyped spell has the
// alternative characteristics. CR 400.7: the Clone that leaves is a
// Clone again.
func TestCloneOfAPrototypedPermanentCopiesTheInset(t *testing.T) {
	g, me, id := prototypeGame(t, blitzAutomaton)
	castPrototyped(t, g, me, id, "{R}{C}{C}")
	passPriorityAroundTable(t, g)

	cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	for i := 0; i < 16 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := copyPrompt(g)
	if prompt == nil {
		t.Fatal("no copy-target prompt")
	}
	if err := g.ResolveCopyTarget(prompt.ID, prompt.Chooser, id); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	clone, ok := g.LookupCardForEffect(cloneID)
	if !ok || !g.Battlefield.Contains(cloneID) {
		t.Fatal("the Clone is not on the battlefield")
	}
	wantPrototypedAutomaton(t, "the Clone", clone)
	if !game.HasKeyword(&clone, "haste") {
		t.Error("the Clone has no haste")
	}

	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, cloneID); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if c, _ := g.LookupCardForEffect(cloneID); c.Name != "Clone" || c.Prototyped() {
		t.Errorf("the Clone in the graveyard is %q (prototyped=%v), want Clone", c.Name, c.Prototyped())
	}
}

// CR 718.3c: a copy of a prototyped spell is a prototyped spell, and
// the token it becomes has the alternative characteristics.
func TestCopyOfAPrototypedSpellIsPrototyped(t *testing.T) {
	g, me, id := prototypeGame(t, blitzAutomaton)
	castPrototyped(t, g, me, id, "{R}{C}{C}")
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(id, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
	var copyID uuid.UUID
	for sid, item := range g.StackMeta {
		if sid != id && item != nil && item.Kind == game.StackItemSpell {
			copyID = sid
		}
	}
	if copyID == uuid.Nil {
		t.Fatal("no copy on the stack")
	}
	spellCopy, _ := g.LookupCardForEffect(copyID)
	wantPrototypedAutomaton(t, "the spell copy", spellCopy)
	passPriorityAroundTable(t, g)

	var token *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.IsToken() && c.Name == "Blitz Automaton" {
			token = c
		}
	}
	if token == nil {
		t.Fatal("the copy did not become a token")
	}
	wantPrototypedAutomaton(t, "the token", *token)
	orig, _ := g.LookupCardForEffect(id)
	wantPrototypedAutomaton(t, "the original", orig)
}

// A prototyped spell and the permanent it becomes both survive a
// restore: the face that is up is carried by the snapshot.
func TestPrototypedSpellAndPermanentSurviveARestore(t *testing.T) {
	g, me, id := prototypeGame(t, blitzAutomaton)
	castPrototyped(t, g, me, id, "{R}{C}{C}")

	g = restoreThroughJSON(t, g)
	spell, ok := g.LookupCardForEffect(id)
	if !ok || !g.Stack.Contains(id) {
		t.Fatal("the restored spell is not on the stack")
	}
	wantPrototypedAutomaton(t, "the restored spell", spell)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the restored spell did not resolve")
	}

	g = restoreThroughJSON(t, g)
	perm, _ := g.LookupCardForEffect(id)
	wantPrototypedAutomaton(t, "the restored permanent", perm)
	if !game.HasKeyword(&perm, "haste") {
		t.Error("the restored permanent has no haste")
	}
}

// CR 718.4: a prototyped spell that is countered is the normal card in
// the graveyard.
func TestCounteredPrototypedSpellIsTheNormalCard(t *testing.T) {
	g, me, id := prototypeGame(t, blitzAutomaton)
	castPrototyped(t, g, me, id, "{R}{C}{C}")
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(id); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	card, _ := g.LookupCardForEffect(id)
	if !me.Graveyard.Contains(id) || card.Prototyped() || card.ManaValue() != 7 || card.Effective().Power != 6 {
		t.Errorf("countered card: in graveyard=%v prototyped=%v mv=%d power=%d, want the normal 6/4 with mana value 7",
			me.Graveyard.Contains(id), card.Prototyped(), card.ManaValue(), card.Effective().Power)
	}
}
