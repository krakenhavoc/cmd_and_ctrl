package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// temples_test.go — the scry-land cycle, and the CR 614
// self-replacement it needed.
//
// The assertion that earns its keep is the difference between a
// replacement and the OnETB-tap workaround it replaces. Both leave the
// permanent tapped, so "is it tapped?" cannot tell them apart. What can:
// an OnETB tap emits EventTapCard, because it really does tap an
// untapped permanent. A replacement emits none — the permanent was
// never untapped on the battlefield.

const templeOfSilenceOracle = "e6e6fce8-0f6a-4b84-865e-d4e4a4182f9f"

// playLandFromHand puts a land in the active player's hand and plays it
// through the real cast path, so the replacement pipeline runs. Pushing
// straight onto the battlefield would skip exactly what's under test.
func playLandFromHand(t *testing.T, g *game.Game, name, oracleID string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Land",
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	// #500: the engine now enforces CR 305.2's land-play allowance.
	// These are CARD-behaviour fixtures — a fastland wants three
	// lands on the board, Field of the Dead wants eight — not tests
	// of the land-drop rule, and they build their board inside a
	// single main phase. Raising this seat's base allowance buys each
	// helper call its own land play, which is the fixture saying "and
	// another turn's land drop" without spending real turns.
	active.LandDropsPerTurn++
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	return id
}

// tapEventsFor counts EventTapCard entries for one card.
func tapEventsFor(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventTapCard && ev.CardID == id {
			n++
		}
	}
	return n
}

func TestTempleEntersTappedAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Next Draw")

	id := playLandFromHand(t, g, "Temple of Silence", templeOfSilenceOracle)

	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the Temple isn't on the battlefield")
	}
	if !card.Tapped {
		t.Error("the Temple entered untapped")
	}
	// The discriminator: a replacement taps nothing, so there is no tap
	// event. The OnETB workaround this replaced would emit one.
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events for the Temple; it should have ENTERED tapped, not been tapped", n)
	}
	// And the land's own ETB trigger is on the stack — not resolved
	// yet, so nothing has been looked at (#578).
	if triggerOnStack(g, id) == nil {
		t.Fatal("playing the Temple did not put its scry trigger on the stack")
	}
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("the scry prompt arrived before the trigger resolved")
	}
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("resolving the Temple's trigger did not queue a scry")
	}
}

func TestTempleScryFeedsTheNextDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Bad", "Good")

	playLandFromHand(t, g, "Temple of Silence", templeOfSilenceOracle)
	passPriorityAroundTable(t, g)

	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("looked at %d cards, want 1 (scry 1)", len(c.ScryCards))
	}
	if err := g.ResolveScry(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if got := libraryTopNames(me, 1); got[0] != "Good" {
		t.Errorf("library top is %q after bottoming Bad, want Good", got[0])
	}
}

// TestTempleTapsForEitherColor — the dual is one pipe ability, so the
// controller picks the colour at activation rather than being handed a
// fixed pair.
func TestTempleTapsForEitherColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Filler")
	id := playLandFromHand(t, g, "Temple of Silence", templeOfSilenceOracle)
	passPriorityAroundTable(t, g)

	// Answer the scry so the prompt isn't in the way.
	if c := scryChoiceFor(g, me.ID); c != nil {
		if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
			t.Fatalf("ResolveScry: %v", err)
		}
	}
	// It entered tapped, so it can't pay a {T} cost yet — untap it the
	// way an untap step would.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})

	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	// A pipe defers the colour to a mana_pick rather than dropping a
	// token straight in.
	if len(g.PendingChoices) == 0 {
		t.Fatal("no mana_pick; a two-colour pipe must ask which colour")
	}
	pick := g.PendingChoices[len(g.PendingChoices)-1]
	if pick.Kind != game.PendingChoiceMana {
		t.Fatalf("choice kind %q, want mana_pick", pick.Kind)
	}
	got := map[string]bool{}
	for _, c := range pick.ColorOptions {
		got[c] = true
	}
	if !got["W"] || !got["B"] {
		t.Errorf("colour options %v, want both W and B", pick.ColorOptions)
	}
	if len(pick.ColorOptions) != 2 {
		t.Errorf("colour options %v, want exactly two", pick.ColorOptions)
	}
}

// TestEveryTempleIsRegistered — the cycle is written as a loop over a
// table, so a typo in one row is invisible until someone plays that
// land. Ten distinct oracle IDs, ten distinct names.
func TestEveryTempleIsRegistered(t *testing.T) {
	temples := map[string]string{
		"e6e6fce8-0f6a-4b84-865e-d4e4a4182f9f": "Temple of Silence",
		"dc55421f-dee8-4263-9df0-2365df5f14bb": "Temple of Malady",
		"79f94050-d850-41ca-b1db-5ae0cf743f0a": "Temple of Epiphany",
		"3baa8e38-ef93-435d-b63e-f781d5bfcc68": "Temple of Abandon",
		"7e26f0b7-20e6-46d5-8130-d98c14d6aa29": "Temple of Mystery",
		"89f43e27-790b-4ca1-8ba7-0882b31e0783": "Temple of Enlightenment",
		"7c439c18-31dc-41fe-b03d-3fca06e6fc0b": "Temple of Malice",
		"e521322b-0e83-458c-8936-7021a80ee279": "Temple of Plenty",
		"33b9b3bd-33ca-46f3-b8bb-a978bc3d1085": "Temple of Deceit",
		"6f0d94d9-64bb-4175-83bc-301e8f79f54f": "Temple of Triumph",
	}
	seenNames := map[string]string{}
	for oracleID, name := range temples {
		spec, ok := Lookup(oracleID)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracleID)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered under the name %q", oracleID, spec.Name)
		}
		// A shared loop variable would give every Temple the last row's
		// colours, so check each one produces its own distinct pair.
		if len(spec.ManaAbilities) != 1 {
			t.Errorf("%s has %d mana abilities, want 1", name, len(spec.ManaAbilities))
			continue
		}
		produced := spec.ManaAbilities[0].Produced
		if prev, dupe := seenNames[produced]; dupe {
			t.Errorf("%s produces %s, same as %s — the loop variable leaked",
				name, produced, prev)
		}
		seenNames[produced] = name
		if len(spec.Replacements) != 1 {
			t.Errorf("%s has %d replacements, want 1 (enters tapped)", name, len(spec.Replacements))
		}
	}
	if len(seenNames) != 10 {
		t.Errorf("%d distinct produced-mana strings across the cycle, want 10", len(seenNames))
	}
}

// TestWornPowerstoneIsNeverUntappedOnTheBattlefield strengthens the
// existing TestWornPowerstoneEntersTapped, which checks only that the
// permanent ends up tapped — true of both the replacement and the OnETB
// workaround it replaced. This one pins the difference, and covers the
// machinery through a different door: an artifact entering from the
// stack rather than a land being played.
func TestWornPowerstoneIsNeverUntappedOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Worn Powerstone", "Artifact",
		"b166b670-febc-4821-855e-f8d465644c03", nil)
	passPriorityAroundTable(t, g)

	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Worn Powerstone isn't on the battlefield")
	}
	if !card.Tapped {
		t.Error("Worn Powerstone entered untapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; it should have ENTERED tapped, not been tapped after", n)
	}
}

// TestSelfReplacementLeavesUndeclaredCardsAlone is the control. The new
// gathering block runs for EVERY entering card, so a card that declares
// no replacement must be unaffected — otherwise the code would be
// tapping everything and every assertion above would pass for the wrong
// reason.
func TestSelfReplacementLeavesUndeclaredCardsAlone(t *testing.T) {
	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Sol Ring", "Artifact",
		"6ad8011d-3471-4369-9d68-b264cc027487", nil)
	passPriorityAroundTable(t, g)

	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Sol Ring isn't on the battlefield")
	}
	if card.Tapped {
		t.Error("Sol Ring entered tapped; the self-replacement is firing for cards that don't declare it")
	}
}
