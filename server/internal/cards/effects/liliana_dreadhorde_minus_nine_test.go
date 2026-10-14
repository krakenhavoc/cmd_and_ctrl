package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// liliana_dreadhorde_minus_nine_test.go pins Liliana, Dreadhorde General's
// -9 against the official rulings (#3209):
//
//   - "As Liliana's last ability resolves, the next opponent in turn order
//     makes all of their choices for it, then each other opponent in turn
//     order does the same, knowing the choices made before them. Then all
//     the unchosen permanents are sacrificed at the same time."
//   - "The permanent types are artifact, creature, enchantment, land, and
//     planeswalker. Supertypes, like legendary, aren't permanent types."
//   - "While making choices for Liliana's last ability, if a permanent has
//     more than one permanent type, it can count for any of them. For
//     example, you could choose an artifact creature as the artifact
//     you're sparing, ... you could choose an enchantment creature as both
//     the creature and the enchantment that you're sparing, even if you
//     control another creature and/or another enchantment."
//
// Battle is also a permanent type (CR 110.4), so it is a slot too.

func lilianaKeepValid(cards []game.Card, keep ...game.Card) bool {
	return lilianaKeepsOneOfEachType(permanentTypesAmong(cards))(keep)
}

// TestLilianaMinusNineArtifactCreatureExample is the first ruling example:
// an artifact creature may be the spared artifact, and equally may stand
// for both the artifact and the creature.
func TestLilianaMinusNineArtifactCreatureExample(t *testing.T) {
	golem := game.Card{TypeLine: "Artifact Creature — Golem"}
	trinket := game.Card{TypeLine: "Artifact"}
	bear := game.Card{TypeLine: "Creature — Bear"}
	board := []game.Card{golem, trinket, bear}
	for _, tc := range []struct {
		name string
		keep []game.Card
		want bool
	}{
		{"golem as the artifact, bear as the creature", []game.Card{golem, bear}, true},
		{"golem alone as both artifact and creature", []game.Card{golem}, true},
		{"trinket as the artifact, bear as the creature", []game.Card{trinket, bear}, true},
		{"trinket alone leaves no creature", []game.Card{trinket}, false},
		{"bear alone leaves no artifact", []game.Card{bear}, false},
		{"all three: one more than there are types", []game.Card{golem, trinket, bear}, false},
		{"nothing kept", nil, false},
	} {
		if got := lilianaKeepValid(board, tc.keep...); got != tc.want {
			t.Errorf("%s: valid = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestLilianaMinusNineEnchantmentCreatureExample is the second example: an
// enchantment creature can be both the creature and the enchantment spared
// even with another creature and another enchantment on the board, but the
// player may instead keep the others.
func TestLilianaMinusNineEnchantmentCreatureExample(t *testing.T) {
	spirit := game.Card{TypeLine: "Enchantment Creature — Spirit"}
	bear := game.Card{TypeLine: "Creature — Bear"}
	aura := game.Card{TypeLine: "Enchantment — Aura"}
	board := []game.Card{spirit, bear, aura}
	for _, tc := range []struct {
		name string
		keep []game.Card
		want bool
	}{
		{"spirit alone as both creature and enchantment", []game.Card{spirit}, true},
		{"spirit as the enchantment, bear as the creature", []game.Card{spirit, bear}, true},
		{"spirit as the creature, aura as the enchantment", []game.Card{spirit, aura}, true},
		{"bear and aura instead of the spirit", []game.Card{bear, aura}, true},
		{"bear alone leaves no enchantment", []game.Card{bear}, false},
		{"aura alone leaves no creature", []game.Card{aura}, false},
		{"all three", []game.Card{spirit, bear, aura}, false},
	} {
		if got := lilianaKeepValid(board, tc.keep...); got != tc.want {
			t.Errorf("%s: valid = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// The same example end to end: the spirit alone is kept, the other
// creature and enchantment are sacrificed, and the answer that keeps
// everything is refused.
func TestLilianaMinusNineEnchantmentCreatureKeptForBothSlots(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Turn.ActiveSeat
	me := g.Seats[active]
	opp := g.Seats[(active+1)%len(g.Seats)]
	liliana := pushWalkerForTest(g, me.ID, "Liliana, Dreadhorde General", wrLilianaDGOracle, 9)
	spirit := pushCatalogPermanent(g, opp.ID, "Spirit", "Enchantment Creature — Spirit", "", false)
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	aura := pushCatalogPermanent(g, opp.ID, "Aura", "Enchantment — Aura", "", false)

	advanceToMainOf(t, g, active)
	b16Activate(t, g, me.ID, liliana, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	prompt := latestChooseCardsFor(g, opp.ID)
	if prompt == nil {
		t.Fatal("the opponent was not asked")
	}
	if err := g.ResolveChooseCards(prompt.ID, opp.ID, []uuid.UUID{spirit, bear, aura}); err == nil {
		t.Fatal("keeping three permanents for two types was accepted")
	}
	answerChooseCards(t, g, opp.ID, spirit)
	if !g.Battlefield.Contains(spirit) {
		t.Error("the spirit, kept as both creature and enchantment, was sacrificed")
	}
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(aura) {
		t.Error("the unchosen creature and enchantment survived")
	}
}

// Supertypes are not permanent types: a legendary or basic permanent does
// not add a slot, and an opponent whose board is one type keeps only one.
func TestLilianaMinusNineSupertypesAreNotPermanentTypes(t *testing.T) {
	legend := game.Card{TypeLine: "Legendary Creature — Human"}
	bear := game.Card{TypeLine: "Creature — Bear"}
	basic := game.Card{TypeLine: "Basic Land — Swamp"}
	snow := game.Card{TypeLine: "Snow Land"}
	if got := permanentTypesOf(legend); len(got) != 1 || got[0] != "creature" {
		t.Errorf("legendary creature types = %v, want [creature]", got)
	}
	if got := permanentTypesOf(basic); len(got) != 1 || got[0] != "land" {
		t.Errorf("basic land types = %v, want [land]", got)
	}
	if lilianaKeepValid([]game.Card{legend, bear}, legend, bear) {
		t.Error("keeping two creatures was accepted; legendary is not a second type")
	}
	if !lilianaKeepValid([]game.Card{legend, bear}, legend) {
		t.Error("keeping the legend as the one creature was refused")
	}
	if lilianaKeepValid([]game.Card{basic, snow}, basic, snow) {
		t.Error("keeping a basic and a snow land was accepted; they are both just lands")
	}
}

// An opponent with no permanent of a type has no pick for it: the keep
// set needs only the types actually present.
func TestLilianaMinusNineMissingTypesNeedNoPick(t *testing.T) {
	land := game.Card{TypeLine: "Basic Land — Island"}
	bear := game.Card{TypeLine: "Creature — Bear"}
	board := []game.Card{land, bear, bear}
	if !lilianaKeepValid(board, land, bear) {
		t.Error("a land and a creature should satisfy a board with only those types")
	}
	if lilianaKeepValid(board, land) {
		t.Error("a missing creature pick was accepted")
	}
	if lilianaKeepValid(board, bear) {
		t.Error("a missing land pick was accepted")
	}
}

// Battle is a permanent type (CR 110.4); the card's "each permanent type"
// is not limited to the five the ruling lists.
func TestLilianaMinusNineBattlesAreAPermanentType(t *testing.T) {
	siege := game.Card{TypeLine: "Battle — Siege"}
	bear := game.Card{TypeLine: "Creature — Bear"}
	board := []game.Card{siege, bear}
	if got := permanentTypesOf(siege); len(got) != 1 || got[0] != "battle" {
		t.Fatalf("battle types = %v, want [battle]", got)
	}
	if !lilianaKeepValid(board, siege, bear) {
		t.Error("keeping the battle and the creature was refused")
	}
	if lilianaKeepValid(board, bear) {
		t.Error("dropping the only battle was accepted")
	}
	if lilianaKeepValid(board, siege) {
		t.Error("dropping the only creature was accepted")
	}
}

func TestLilianaMinusNineBattleEndToEnd(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Turn.ActiveSeat
	me := g.Seats[active]
	opp := g.Seats[(active+1)%len(g.Seats)]
	liliana := pushWalkerForTest(g, me.ID, "Liliana, Dreadhorde General", wrLilianaDGOracle, 9)
	siege := pushCatalogPermanent(g, opp.ID, "Siege", "Battle — Siege", "", false)
	addCountersForTest(g, siege, game.CounterDefense, 4)
	keepBear := pushCatalogPermanent(g, opp.ID, "Keep Bear", "Creature — Bear", "", false)
	loseBear := pushCatalogPermanent(g, opp.ID, "Lose Bear", "Creature — Bear", "", false)

	advanceToMainOf(t, g, active)
	b16Activate(t, g, me.ID, liliana, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	prompt := latestChooseCardsFor(g, opp.ID)
	if prompt == nil {
		t.Fatal("the opponent with a battle and creatures was not asked")
	}
	if prompt.ChooseMax != 2 {
		t.Errorf("ChooseMax = %d, want 2 (battle and creature)", prompt.ChooseMax)
	}
	if err := g.ResolveChooseCards(prompt.ID, opp.ID, []uuid.UUID{keepBear}); err == nil {
		t.Fatal("sacrificing the only battle was accepted")
	}
	answerChooseCards(t, g, opp.ID, siege, keepBear)
	if !g.Battlefield.Contains(siege) || !g.Battlefield.Contains(keepBear) {
		t.Error("a kept permanent was sacrificed")
	}
	if g.Battlefield.Contains(loseBear) {
		t.Error("the unchosen creature survived")
	}
}

// Turn order starts with the next player after Liliana's controller, wraps
// the table, and skips eliminated players, whoever the active player is.
func TestLilianaOpponentsAreAskedInTurnOrderFromTheController(t *testing.T) {
	g := newCatalogGame(t)
	n := len(g.Seats)
	for start := 0; start < n; start++ {
		controller := g.Seats[start].ID
		got := opponentsInTurnOrder(g, controller)
		if len(got) != n-1 {
			t.Fatalf("controller seat %d: %d opponents, want %d", start, len(got), n-1)
		}
		for i, id := range got {
			if want := g.Seats[(start+1+i)%n].ID; id != want {
				t.Errorf("controller seat %d: opponent %d is not the next in turn order", start, i)
			}
		}
	}
	g.Seats[2].Eliminated = true
	if got := opponentsInTurnOrder(g, g.Seats[1].ID); len(got) != 2 || got[0] != g.Seats[3].ID || got[1] != g.Seats[0].ID {
		t.Errorf("order with seat 2 out = %v, want seats 3 then 0", got)
	}
}

// Each later opponent sees the earlier choices, and nothing is sacrificed
// until the last has chosen; then everything unchosen goes at once (one
// shared leaves-the-battlefield batch).
func TestLilianaMinusNineLaterOpponentsSeeEarlierChoicesThenSacrificeTogether(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Turn.ActiveSeat
	me := g.Seats[active]
	first := g.Seats[(active+1)%len(g.Seats)]
	second := g.Seats[(active+2)%len(g.Seats)]
	liliana := pushWalkerForTest(g, me.ID, "Liliana, Dreadhorde General", wrLilianaDGOracle, 9)
	firstKeep := pushCatalogPermanent(g, first.ID, "First Keep", "Creature — Bear", "", false)
	firstLose := pushCatalogPermanent(g, first.ID, "First Lose", "Creature — Bear", "", false)
	secondKeep := pushCatalogPermanent(g, second.ID, "Second Keep", "Creature — Bear", "", false)
	secondLose := pushCatalogPermanent(g, second.ID, "Second Lose", "Creature — Bear", "", false)

	advanceToMainOf(t, g, active)
	b16Activate(t, g, me.ID, liliana, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	if latestChooseCardsFor(g, second.ID) != nil {
		t.Fatal("the second opponent was asked before the first")
	}
	firstPrompt := latestChooseCardsFor(g, first.ID)
	if firstPrompt == nil {
		t.Fatal("the first opponent was not asked")
	}
	if got := firstPrompt.EarlierPermanentChoicesForEffect(); len(got) != 0 {
		t.Errorf("the first opponent saw earlier choices: %v", got)
	}
	answerChooseCards(t, g, first.ID, firstKeep)
	prompt := latestChooseCardsFor(g, second.ID)
	if prompt == nil {
		t.Fatal("the second opponent was not asked after the first answered")
	}
	earlier := prompt.EarlierPermanentChoicesForEffect()
	if len(earlier) != 1 || earlier[0].Seat != first.ID || len(earlier[0].Cards) != 1 || earlier[0].Cards[0] != firstKeep {
		t.Errorf("second opponent saw earlier choices %v, want the first opponent's keep", earlier)
	}
	for _, id := range []uuid.UUID{firstKeep, firstLose, secondKeep, secondLose} {
		if !g.Battlefield.Contains(id) {
			t.Fatal("a permanent left before the last opponent chose")
		}
	}
	answerChooseCards(t, g, second.ID, secondKeep)

	if g.Battlefield.Contains(firstLose) || g.Battlefield.Contains(secondLose) {
		t.Fatal("an unchosen permanent survived")
	}
	if !g.Battlefield.Contains(firstKeep) || !g.Battlefield.Contains(secondKeep) {
		t.Fatal("a chosen permanent was sacrificed")
	}
	batches := map[uint64]bool{}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventLTB && (ev.CardID == firstLose || ev.CardID == secondLose) {
			batches[ev.Batch] = true
		}
	}
	if len(batches) != 1 {
		t.Errorf("the sacrifices happened in %d batches, want 1 simultaneous", len(batches))
	}
}
