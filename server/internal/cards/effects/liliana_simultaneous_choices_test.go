package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestLilianaMinusFourCollectsEveryChoiceBeforeOneSimultaneousSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Turn.ActiveSeat
	me := g.Seats[active]
	liliana := pushWalkerForTest(g, me.ID, "Liliana, Dreadhorde General", wrLilianaDGOracle, 6)
	artist := pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)

	chosen := make([][]uuid.UUID, len(g.Seats))
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		p := g.Seats[seat]
		chosen[seat] = make([]uuid.UUID, 0, 3)
		if seat == active {
			chosen[seat] = append(chosen[seat], artist)
		}
		for len(chosen[seat]) < 3 {
			chosen[seat] = append(chosen[seat], pushVanillaCreature(g, p.ID, "Fodder", 1, 1))
		}
	}

	// One selected object is a token; Blood Artist must observe it in the
	// same exit as the seven nontoken creatures before it ceases to exist.
	tokenID := chosen[(active+1)%len(g.Seats)][0]
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == tokenID {
			g.Battlefield.Cards[i].TypeLine = "Token Creature — Soldier"
		}
	}

	advanceToMainOf(t, g, active)
	b16Activate(t, g, me.ID, liliana, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	selected := make(map[uuid.UUID]bool, 2*len(g.Seats))
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		p := g.Seats[seat]
		prompt := latestChooseCardsFor(g, p.ID)
		if prompt == nil {
			t.Fatalf("APNAP choice %d: seat %s was not asked", offset, p.Name)
		}
		if prompt.ChooseMin != 2 || prompt.ChooseMax != 2 {
			t.Fatalf("seat %s choice bounds = %d..%d, want exactly 2", p.Name, prompt.ChooseMin, prompt.ChooseMax)
		}
		if len(prompt.ChooseCards) != 3 {
			t.Fatalf("seat %s was offered %d creatures, want its own 3", p.Name, len(prompt.ChooseCards))
		}
		picks := chosen[seat][:2]
		for _, id := range picks {
			selected[id] = true
		}
		answerChooseCards(t, g, p.ID, picks...)

		// Each answer only records a choice. Until the final seat answers,
		// every chosen permanent is still on the battlefield.
		if offset < len(g.Seats)-1 {
			for id := range selected {
				if !g.Battlefield.Contains(id) {
					t.Fatalf("permanent %s moved after seat %s chose; all choices must precede sacrifice", id, p.Name)
				}
			}
			if next := latestChooseCardsFor(g, g.Seats[(active+offset+1)%len(g.Seats)].ID); next == nil {
				t.Fatalf("after seat %s chose, the next APNAP seat was not asked", p.Name)
			}
		}
	}

	if len(selected) != 2*len(g.Seats) {
		t.Fatalf("collected %d selections, want %d", len(selected), 2*len(g.Seats))
	}
	for id := range selected {
		if g.Battlefield.Contains(id) {
			t.Errorf("chosen creature %s survived the one final sacrifice", id)
		}
	}
	for seat, ids := range chosen {
		if !g.Battlefield.Contains(ids[2]) {
			t.Errorf("seat %d's unchosen creature %s was sacrificed", seat, ids[2])
		}
	}

	ltb := make(map[uuid.UUID]uint64, len(selected))
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventLTB && selected[ev.CardID] {
			ltb[ev.CardID] = ev.Batch
		}
	}
	if len(ltb) != len(selected) {
		t.Fatalf("observed %d selected-creature exits, want %d", len(ltb), len(selected))
	}
	var exitBatch uint64
	for id, batch := range ltb {
		if batch == 0 {
			t.Fatalf("creature %s has no event batch", id)
		}
		if exitBatch == 0 {
			exitBatch = batch
		} else if batch != exitBatch {
			t.Fatalf("selected creature %s exited in batch %d, want shared batch %d", id, batch, exitBatch)
		}
	}
	if artistTriggers := lilianaPendingTriggersFrom(g, artist); artistTriggers != len(selected) {
		t.Errorf("Blood Artist triggers = %d, want %d for all simultaneous deaths", artistTriggers, len(selected))
	}
}

func TestLilianaMinusFourUsesExactCountsAndSkipsProtectedOrEmptySeats(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Turn.ActiveSeat
	me := g.Seats[active]
	protected := g.Seats[(active+1)%len(g.Seats)]
	empty := g.Seats[(active+2)%len(g.Seats)]
	oneCreature := g.Seats[(active+3)%len(g.Seats)]
	liliana := pushWalkerForTest(g, me.ID, "Liliana, Dreadhorde General", wrLilianaDGOracle, 6)
	first := []uuid.UUID{
		pushVanillaCreature(g, me.ID, "Me A", 1, 1),
		pushVanillaCreature(g, me.ID, "Me B", 1, 1),
		pushVanillaCreature(g, me.ID, "Me C", 1, 1),
	}
	sigarda := pushCatalogPermanent(g, protected.ID, "Sigarda, Host of Herons", "Legendary Creature — Angel", sigardaHostOracle, false)
	protectedCreatures := []uuid.UUID{sigarda,
		pushVanillaCreature(g, protected.ID, "Protected A", 1, 1),
		pushVanillaCreature(g, protected.ID, "Protected B", 1, 1),
	}
	one := pushVanillaCreature(g, oneCreature.ID, "Only Creature", 1, 1)

	advanceToMainOf(t, g, active)
	b16Activate(t, g, me.ID, liliana, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	firstPrompt := latestChooseCardsFor(g, me.ID)
	if firstPrompt == nil || firstPrompt.ChooseMin != 2 || firstPrompt.ChooseMax != 2 {
		t.Fatalf("controller prompt = %+v, want exactly two from three creatures", firstPrompt)
	}
	answerChooseCards(t, g, me.ID, first[0], first[1])
	lastPrompt := latestChooseCardsFor(g, oneCreature.ID)
	if lastPrompt != nil {
		t.Fatalf("sole-creature seat has a forced set and should not be prompted, got %+v", lastPrompt)
	}
	if latestChooseCardsFor(g, protected.ID) != nil {
		t.Fatal("Sigarda's player was asked to sacrifice to an opponent's effect")
	}
	if latestChooseCardsFor(g, empty.ID) != nil {
		t.Fatal("a seat with no creatures was asked to sacrifice")
	}
	for _, id := range first[:2] {
		if g.Battlefield.Contains(id) {
			t.Errorf("chosen creature %s survived", id)
		}
	}
	if g.Battlefield.Contains(one) {
		t.Error("the sole creature was not sacrificed")
	}
	for _, id := range append(protectedCreatures, first[2]) {
		if !g.Battlefield.Contains(id) {
			t.Errorf("unaffected creature %s left the battlefield", id)
		}
	}
}

func TestLilianaMinusNineCollectsOpponentChoicesInAPNAPOrder(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Turn.ActiveSeat
	me := g.Seats[active]
	firstSeat := (active + 1) % len(g.Seats)
	secondSeat := (active + 2) % len(g.Seats)
	firstPlayer, secondPlayer := g.Seats[firstSeat], g.Seats[secondSeat]
	liliana := pushWalkerForTest(g, me.ID, "Liliana, Dreadhorde General", wrLilianaDGOracle, 9)
	firstKeep := pushVanillaCreature(g, firstPlayer.ID, "First Keep", 1, 1)
	firstSacrifice := pushVanillaCreature(g, firstPlayer.ID, "First Sacrifice", 1, 1)
	secondKeep := pushVanillaCreature(g, secondPlayer.ID, "Second Keep", 1, 1)
	secondSacrifice := pushVanillaCreature(g, secondPlayer.ID, "Second Sacrifice", 1, 1)

	advanceToMainOf(t, g, active)
	b16Activate(t, g, me.ID, liliana, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	firstPrompt := latestChooseCardsFor(g, firstPlayer.ID)
	if firstPrompt == nil || firstPrompt.ChooseMin != 1 || firstPrompt.ChooseMax != 1 {
		t.Fatalf("first APNAP opponent's keep prompt = %+v, want one permanent", firstPrompt)
	}
	if got := latestChooseCardsFor(g, secondPlayer.ID); got != nil {
		t.Fatalf("second opponent was asked before the first answered: %+v", got)
	}
	answerChooseCards(t, g, firstPlayer.ID, firstKeep)
	if !g.Battlefield.Contains(firstSacrifice) || !g.Battlefield.Contains(secondKeep) || !g.Battlefield.Contains(secondSacrifice) {
		t.Fatal("a permanent moved before every opponent chose")
	}
	secondPrompt := latestChooseCardsFor(g, secondPlayer.ID)
	if secondPrompt == nil || secondPrompt.ChooseMin != 1 || secondPrompt.ChooseMax != 1 {
		t.Fatalf("second APNAP opponent's keep prompt = %+v, want one permanent", secondPrompt)
	}
	answerChooseCards(t, g, secondPlayer.ID, secondKeep)
	if !g.Battlefield.Contains(firstKeep) || !g.Battlefield.Contains(secondKeep) {
		t.Fatal("an opponent's chosen permanent was sacrificed")
	}
	if g.Battlefield.Contains(firstSacrifice) || g.Battlefield.Contains(secondSacrifice) {
		t.Fatal("the unchosen opponent permanents were not sacrificed")
	}
}

func lilianaPendingTriggersFrom(g *game.Game, source uuid.UUID) int {
	n := 0
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source {
			n++
		}
	}
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source {
			n++
		}
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && c.Source == source {
			n++
		}
	}
	return n
}
