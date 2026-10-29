package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// key_to_the_side_door_test.go — #4487: a discard cost that reads the
// board. Key to the Side-Door discards "a legendary card with the same
// name as a legendary permanent you control", so the engine, not the
// card's func(Card) bool predicate, has to see the battlefield.

const keySideDoorOracle = "76c036df-44f5-4021-b7fc-656147351341"

// keySideDoorRummage is the draw-two ability's index.
const keySideDoorRummage = 1

type keySideDoorTable struct {
	g   *game.Game
	me  *game.Player
	opp *game.Player
	key uuid.UUID
}

func newKeySideDoorTable(t *testing.T) keySideDoorTable {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	key := b12Push(g, me.ID, "Key to the Side-Door", "Artifact", keySideDoorOracle, 0, 0)
	advanceToMain(t, g)
	return keySideDoorTable{g: g, me: me, opp: opp, key: key}
}

func keySideDoorHand(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// keySideDoorTry attempts the draw-two activation with `pick` named as
// the discard and returns the engine's verdict, floating the {1} first.
func keySideDoorTry(t *testing.T, tb keySideDoorTable, pick uuid.UUID) error {
	t.Helper()
	floatMana(t, tb.g, tb.me, "{C}")
	return tb.g.ActivateCatalogAbility(tb.me.ID, tb.key, keySideDoorRummage,
		game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pick}})
}

// keySideDoorRefused pins a refusal AND that nothing was paid: the card
// stays in the hand, the key stays untapped, the stack stays empty.
func keySideDoorRefused(t *testing.T, tb keySideDoorTable, pick uuid.UUID) {
	t.Helper()
	handBefore := handCount(tb.g, tb.me)
	err := keySideDoorTry(t, tb, pick)
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("activation = %v, want ErrInvalidParam", err)
	}
	if inGraveyard(tb.me, pick) || !tb.me.Hand.Contains(pick) {
		t.Error("the refused discard left the hand anyway")
	}
	if b16Tapped(t, tb.g, tb.key) {
		t.Error("the refused activation tapped the key")
	}
	if got := handCount(tb.g, tb.me); got != handBefore {
		t.Errorf("hand %d, want %d unchanged", got, handBefore)
	}
	if n := tb.g.Stack.Size(); n != 0 {
		t.Errorf("stack has %d items after a refusal", n)
	}
}

func TestKeyToTheSideDoorRefusesWithoutALegendaryPermanentOfThatName(t *testing.T) {
	tb := newKeySideDoorTable(t)
	pick := keySideDoorHand(tb.me, "Gandalf the Grey", "Legendary Creature — Wizard")
	keySideDoorRefused(t, tb, pick)
}

func TestKeyToTheSideDoorRefusesWhenTheSameNameCardIsNotLegendary(t *testing.T) {
	tb := newKeySideDoorTable(t)
	b12Push(tb.g, tb.me.ID, "Gandalf the Grey", "Legendary Creature — Wizard", "", 3, 4)
	pick := keySideDoorHand(tb.me, "Gandalf the Grey", "Creature — Wizard")
	keySideDoorRefused(t, tb, pick)
}

func TestKeyToTheSideDoorRefusesWhenOnlyTheSameNamePermanentIsNotLegendary(t *testing.T) {
	tb := newKeySideDoorTable(t)
	b12Push(tb.g, tb.me.ID, "Gandalf the Grey", "Creature — Wizard", "", 3, 4)
	pick := keySideDoorHand(tb.me, "Gandalf the Grey", "Legendary Creature — Wizard")
	keySideDoorRefused(t, tb, pick)
}

func TestKeyToTheSideDoorRefusesAnotherPlayersLegendaryPermanent(t *testing.T) {
	tb := newKeySideDoorTable(t)
	b12Push(tb.g, tb.opp.ID, "Gandalf the Grey", "Legendary Creature — Wizard", "", 3, 4)
	pick := keySideDoorHand(tb.me, "Gandalf the Grey", "Legendary Creature — Wizard")
	keySideDoorRefused(t, tb, pick)
}

func TestKeyToTheSideDoorRefusesALegendaryCardWithADifferentName(t *testing.T) {
	tb := newKeySideDoorTable(t)
	b12Push(tb.g, tb.me.ID, "Gandalf the Grey", "Legendary Creature — Wizard", "", 3, 4)
	pick := keySideDoorHand(tb.me, "Gollum", "Legendary Creature — Halfling")
	keySideDoorRefused(t, tb, pick)
}

func TestKeyToTheSideDoorDiscardsTheMatchingLegendAndDrawsTwo(t *testing.T) {
	tb := newKeySideDoorTable(t)
	b12Push(tb.g, tb.me.ID, "Gandalf the Grey", "Legendary Creature — Wizard", "", 3, 4)
	// A decoy that matches the predicate but not the board.
	decoy := keySideDoorHand(tb.me, "Gollum", "Legendary Creature — Halfling")
	pick := keySideDoorHand(tb.me, "Gandalf the Grey", "Legendary Creature — Wizard")
	before := handCount(tb.g, tb.me)

	if err := keySideDoorTry(t, tb, decoy); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("the off-name legend was accepted: %v", err)
	}
	if err := keySideDoorTry(t, tb, pick); err != nil {
		t.Fatalf("the matching legend was refused: %v", err)
	}
	passPriorityAroundTable(t, tb.g)
	if !inGraveyard(tb.me, pick) {
		t.Error("the matching legend was not discarded as the cost")
	}
	if !tb.me.Hand.Contains(decoy) {
		t.Error("the decoy left the hand")
	}
	if !b16Tapped(t, tb.g, tb.key) {
		t.Error("the key was not tapped")
	}
	if got := handCount(tb.g, tb.me); got != before-1+2 {
		t.Errorf("hand %d, want %d (discard one, draw two)", got, before-1+2)
	}
}

// A legend the opponent owns but I control counts: the clause reads
// who controls it, not whose card it is.
func TestKeyToTheSideDoorCountsAStolenLegend(t *testing.T) {
	tb := newKeySideDoorTable(t)
	id := uuid.New()
	tb.g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Gandalf the Grey", TypeLine: "Legendary Creature — Wizard",
		Owner: tb.opp.ID, Controller: tb.me.ID, Power: 3, Toughness: 4})
	pick := keySideDoorHand(tb.me, "Gandalf the Grey", "Legendary Creature — Wizard")
	if err := keySideDoorTry(t, tb, pick); err != nil {
		t.Fatalf("a legend I control (but do not own) was refused: %v", err)
	}
}

// The cards the client is offered are the cards the engine accepts.
func TestKeyToTheSideDoorOffersOnlyThePayableCards(t *testing.T) {
	tb := newKeySideDoorTable(t)
	b12Push(tb.g, tb.me.ID, "Gandalf the Grey", "Legendary Creature — Wizard", "", 3, 4)
	keySideDoorHand(tb.me, "Gollum", "Legendary Creature — Halfling")
	keySideDoorHand(tb.me, "Gandalf the Grey", "Creature — Wizard")
	want := keySideDoorHand(tb.me, "Gandalf the Grey", "Legendary Creature — Wizard")

	spec, ok := Lookup(keySideDoorOracle)
	if !ok {
		t.Fatal("Key to the Side-Door is not registered")
	}
	cost := spec.Activated[keySideDoorRummage].Cost.DiscardCards
	if cost == nil {
		t.Fatal("the rummage ability has no discard cost")
	}
	var got []uuid.UUID
	tb.g.ReadSnapshot(func() { got = tb.g.DiscardCostOptionsForEffect(tb.me.ID, tb.key, cost) })
	if len(got) != 1 || got[0] != want {
		t.Errorf("options = %v, want only %v", got, want)
	}
}

// keySideDoorRummageMoves lists the enumerator's draw-two moves.
func keySideDoorRummageMoves(t *testing.T, tb keySideDoorTable) []legal.Move {
	t.Helper()
	var out []legal.Move
	for _, m := range legal.EnumerateFor(tb.g, tb.me.ID) {
		if m.Type != "activate_ability" || m.Source != tb.key {
			continue
		}
		var p struct {
			AbilityIndex int      `json:"ability_index"`
			DiscardIDs   []string `json:"discard_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		if p.AbilityIndex == keySideDoorRummage {
			out = append(out, m)
		}
	}
	return out
}

func TestKeyToTheSideDoorIsOnlyEnumeratedWhileItIsPayable(t *testing.T) {
	tb := newKeySideDoorTable(t)
	floatMana(t, tb.g, tb.me, "{C}{C}")
	pick := keySideDoorHand(tb.me, "Gandalf the Grey", "Legendary Creature — Wizard")
	keySideDoorHand(tb.me, "Gollum", "Legendary Creature — Halfling")

	if moves := keySideDoorRummageMoves(t, tb); len(moves) != 0 {
		t.Fatalf("the enumerator offered the draw with no legendary permanent: %d moves", len(moves))
	}

	b12Push(tb.g, tb.me.ID, "Gandalf the Grey", "Legendary Creature — Wizard", "", 3, 4)
	moves := keySideDoorRummageMoves(t, tb)
	if len(moves) != 1 {
		t.Fatalf("the enumerator offered %d draw moves with a payable discard, want 1", len(moves))
	}
	var p struct {
		DiscardIDs []string `json:"discard_ids"`
	}
	if err := json.Unmarshal(moves[0].Params, &p); err != nil {
		t.Fatal(err)
	}
	if len(p.DiscardIDs) != 1 || p.DiscardIDs[0] != pick.String() {
		t.Errorf("discard_ids = %v, want the matching legend %v", p.DiscardIDs, pick)
	}
}

func TestKeyToTheSideDoorBoardClauseIsRefusedBesideARandomOrHandForm(t *testing.T) {
	for name, dc := range map[string]*game.DiscardCost{
		"random":       {N: 1, Random: true, SameNameAsLegendaryPermanent: true},
		"hand":         {Hand: true, SameNameAsLegendaryPermanent: true},
		"mana value x": {N: 1, ManaValueX: true, SameNameAsLegendaryPermanent: true},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: Register accepted the board clause beside it", name)
				}
			}()
			checkDiscardClause("Fixture", "ability 0", dc)
		}()
	}
	// And the printed form is accepted.
	checkDiscardClause("Fixture", "ability 0", DiscardLegendaryNamedLikeYourPermanent().DiscardCards)
}
