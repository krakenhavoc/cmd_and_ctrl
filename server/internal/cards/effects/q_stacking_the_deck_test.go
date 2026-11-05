package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// q_stacking_the_deck_test.go — Q, Stacking the Deck (#4973, ADR 0219):
// an ability a resolved trigger gives a card in exile.

const qStackingTheDeckOracle = "6866f0a4-5fe6-40d9-bbfc-6f116f4e415d"

const qGrantRef = "grant:" + qStackingTheDeckGrant + ":0:0"

func TestQStackingTheDeckIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(qStackingTheDeckOracle)
	if !ok {
		t.Fatal("Q, Stacking the Deck is not registered")
	}
	if spec.Name != "Q, Stacking the Deck" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

func qCard() game.Card {
	return game.Card{Name: "Q, Stacking the Deck", TypeLine: "Legendary Creature — Q", OracleID: qStackingTheDeckOracle,
		ManaCost: "{2}{R}{R}", Power: 4, Toughness: 4}
}

func qDarkRitual() game.Card {
	return game.Card{Name: "Dark Ritual", TypeLine: "Instant", ManaCost: "{B}", OracleID: darkRitualOracle, Colors: []string{"B"}}
}

// qEnter plays Q with a library of `lands` Forests and one Dark Ritual and
// resolves the enter trigger. It returns the Ritual's id and the ids of the
// Forests.
func qEnter(t *testing.T, g *game.Game, me *game.Player, lands int) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	cards := make([]game.Card, 0, lands+1)
	for i := 0; i < lands; i++ {
		cards = append(cards, gsForest())
	}
	cards = append(cards, qDarkRitual())
	ids := gsSetLibrary(me, cards...)
	msc17Play(t, g, me, qCard())
	return ids[lands], ids[:lands]
}

// qActivate activates the granted ability of the exiled card, paying with
// the hand card `pay`.
func qActivate(g *game.Game, who, card, pay uuid.UUID) error {
	return g.ActivateCatalogAbility(who, card, 0, game.ActivateAbilityParams{
		Ref:      qGrantRef,
		ExileIDs: []uuid.UUID{pay},
		Strict:   true,
	})
}

// Q entering shuffles, exiles past the lands to the nonland card, and that
// card alone gains the ability; the lands exiled on the way gain nothing.
// The shuffle is random, so the walk is repeated until it has passed at
// least one land.
func TestQStackingTheDeckGrantsTheNonlandCardAndNoLand(t *testing.T) {
	for attempt := 0; attempt < 20; attempt++ {
		g := newCatalogGame(t)
		me := g.Seats[0]
		ritual, forests := qEnter(t, g, me, 7)

		rc := gsExiled(g, ritual)
		if rc == nil {
			t.Fatal("the nonland card is not in exile")
		}
		if len(rc.ExileGrants) != 1 || rc.ExileGrants[0].Key != game.GrantKey(qStackingTheDeckGrant) || rc.ExileGrants[0].Activator != me.ID {
			t.Fatalf("the nonland card's grants = %+v, want Q's bundle for its controller", rc.ExileGrants)
		}
		exiledLands := 0
		for _, f := range forests {
			c := gsExiled(g, f)
			if c == nil {
				if !me.Library.Contains(f) {
					t.Fatal("a Forest is neither in exile nor in the library")
				}
				continue
			}
			exiledLands++
			if len(c.ExileGrants) != 0 {
				t.Fatalf("an exiled land gained %+v", c.ExileGrants)
			}
			if rows, _ := game.ActivatedAbilitiesWithOrigins(*c); len(rows) != 0 {
				t.Fatal("an exiled land has an activated ability")
			}
		}
		if exiledLands > 0 {
			return
		}
	}
	t.Fatal("twenty shuffles never put a Forest above the Dark Ritual")
}

// The headline: on its owner's turn the exiled card's granted ability is
// activated by exiling a card from hand, and makes a copy that may be cast
// at its printed cost; the copy does what the card does, and the original
// stays in exile with the ability still on it.
func TestQStackingTheDeckCopyIsCastableOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ritual, _ := qEnter(t, g, me, 0)
	toMainForCost(t, g)

	pay := handCardForTest(me, "Fodder", "Sorcery", "")
	if err := qActivate(g, me.ID, ritual, pay); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	if !inExile(g, pay) || me.Hand.Contains(pay) {
		t.Fatal("the hand card was not exiled as the cost")
	}
	passPriorityAroundTable(t, g)

	cp := copyOf(t, g, "Dark Ritual")
	if len(cp.ExileGrants) != 0 {
		t.Errorf("the copy has the gained ability %+v: a gained ability is not a copiable value", cp.ExileGrants)
	}
	if err := castCopy(g, me.ID, cp.InstanceID, true); err == nil {
		t.Fatal("the copy was cast with no {B}: \"you may cast the copy\" pays its mana cost")
	}
	floatForTest(g, me, "B")
	if err := castCopy(g, me.ID, cp.InstanceID, true); err != nil {
		t.Fatalf("cast the copy for {B}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("mana pool holds %d, want {B}{B}{B} from the copied Dark Ritual", got)
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left after the copy resolved", n)
	}
	rc := gsExiled(g, ritual)
	if rc == nil || len(rc.ExileGrants) != 1 {
		t.Fatal("the original left exile or lost the ability by being copied")
	}

	// Again, with a second hand card: the ability has no per-turn limit.
	pay2 := handCardForTest(me, "Fodder Two", "Sorcery", "")
	if err := qActivate(g, me.ID, ritual, pay2); err != nil {
		t.Fatalf("activate the ability a second time: %v", err)
	}
}

// The cost is real: with an empty hand there is nothing to exile.
func TestQStackingTheDeckNeedsACardInHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ritual, _ := qEnter(t, g, me, 0)
	toMainForCost(t, g)
	me.Hand.Cards = nil
	if err := g.ActivateCatalogAbility(me.ID, ritual, 0, game.ActivateAbilityParams{Ref: qGrantRef, Strict: true}); err == nil {
		t.Fatal("the ability was activated with no card in hand to exile")
	}
}

// "Activate only during your turn": refused on an opponent's turn.
func TestQStackingTheDeckIsRefusedOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ritual, _ := qEnter(t, g, me, 0)
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	// Hand me priority, so the turn is the only thing that refuses it.
	g.Turn.PriorityHolder = 0

	pay := handCardForTest(me, "Fodder", "Sorcery", "")
	err := qActivate(g, me.ID, ritual, pay)
	if err == nil {
		t.Fatal("the ability was activated during an opponent's turn")
	}
	if !strings.Contains(err.Error(), "activation condition") {
		t.Errorf("refused for %v, want the activation condition (only during your turn)", err)
	}
	if !me.Hand.Contains(pay) {
		t.Error("a refused activation exiled the hand card")
	}
	if gsExiled(g, ritual) == nil || len(gsExiled(g, ritual).ExileGrants) != 1 {
		t.Fatal("the card lost the ability on an opponent's turn")
	}
}

// CR 400.7: the ability is the exiled object's. Once the card leaves
// exile it is gone, and coming back to exile does not restore it.
func TestQStackingTheDeckAbilityEndsWhenTheCardLeavesExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ritual, _ := qEnter(t, g, me, 0)
	toMainForCost(t, g)

	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, ritual); err != nil {
		t.Fatalf("move to graveyard: %v", err)
	}
	pay := handCardForTest(me, "Fodder", "Sorcery", "")
	if err := qActivate(g, me.ID, ritual, pay); err == nil {
		t.Fatal("the ability was activated from the graveyard")
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, ritual); err != nil {
		t.Fatalf("move back to exile: %v", err)
	}
	c := gsExiled(g, ritual)
	if c == nil {
		t.Fatal("the card did not return to exile")
	}
	if len(c.ExileGrants) != 0 {
		t.Fatalf("the card came back to exile with %+v: it is a new object (CR 400.7)", c.ExileGrants)
	}
	if err := qActivate(g, me.ID, ritual, pay); err == nil {
		t.Fatal("a new object in exile activated the old object's ability")
	}
}
