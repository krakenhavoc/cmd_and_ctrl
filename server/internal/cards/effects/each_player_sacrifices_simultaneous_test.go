package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestPlaguecrafterCollectsEveryChoiceInAPNAPOrderThenSacrificesTogether
// is #3664's reproduction. Plaguecrafter's "each player sacrifices a
// creature or planeswalker" used to queue one prompt per seat and
// sacrifice on each answer, so a non-active seat could answer first, its
// creature died alone, and its death triggers went on the stack while
// other players were still choosing.
//
// CR 101.4 and 608.2e: the active player chooses first, then each other
// player in turn order, and then every chosen permanent is sacrificed at
// the same time — one exit, one Blood Artist trigger per death in one
// batch.
func TestPlaguecrafterCollectsEveryChoiceInAPNAPOrderThenSacrificesTogether(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Turn.ActiveSeat
	me := g.Seats[active]
	artist := pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)

	picks := make([]uuid.UUID, len(g.Seats))
	kept := make([]uuid.UUID, len(g.Seats))
	for seat, p := range g.Seats {
		picks[seat] = pushVanillaCreature(g, p.ID, "Fodder", 1, 1)
		if seat != active {
			kept[seat] = pushVanillaCreature(g, p.ID, "Keeper", 1, 1)
		}
	}
	kept[active] = artist

	advanceToMainOf(t, g, active)
	crafter := castCatalogSpell(t, g, "Plaguecrafter", "Creature — Human Shaman", b05PlaguecrafterOracle, nil)
	passPriorityAroundTable(t, g)

	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		p := g.Seats[seat]
		prompt := sacrificeChoiceFor(g, p.ID)
		if prompt == nil {
			t.Fatalf("APNAP choice %d: seat %s was not asked", offset, p.Name)
		}
		if prompt.Kind != game.PendingChoiceChooseCards || prompt.ChooseDestination != game.ChooseSacrifice {
			t.Fatalf("seat %s got a %s (destination %q), want a choose_cards sacrifice pick",
				p.Name, prompt.Kind, prompt.ChooseDestination)
		}
		if prompt.ChooseMin != 1 || prompt.ChooseMax != 1 {
			t.Fatalf("seat %s choice bounds = %d..%d, want exactly 1", p.Name, prompt.ChooseMin, prompt.ChooseMax)
		}
		// Only this seat may answer now: a later seat has nothing to
		// answer, so it cannot sacrifice before an earlier one has chosen.
		for later := offset + 1; later < len(g.Seats); later++ {
			q := g.Seats[(active+later)%len(g.Seats)]
			if sacrificeChoiceFor(g, q.ID) != nil {
				t.Fatalf("seat %s may answer while %s (earlier in APNAP order) is still choosing", q.Name, p.Name)
			}
		}
		if err := g.ResolveChooseCards(prompt.ID, g.Seats[(seat+1)%len(g.Seats)].ID, []uuid.UUID{picks[seat]}); err == nil {
			t.Fatalf("another seat answered %s's sacrifice pick", p.Name)
		}

		answerSacrifice(t, g, p.ID, picks[seat])

		if offset < len(g.Seats)-1 {
			for _, id := range picks {
				if !g.Battlefield.Contains(id) {
					t.Fatalf("a chosen creature died after seat %s chose; nothing moves until every player has", p.Name)
				}
			}
			if n := lilianaPendingTriggersFrom(g, artist); n != 0 {
				t.Fatalf("Blood Artist triggered %d times while players were still choosing", n)
			}
		}
	}

	for seat, id := range picks {
		if g.Battlefield.Contains(id) {
			t.Errorf("seat %d's chosen creature survived", seat)
		}
		if !g.Battlefield.Contains(kept[seat]) {
			t.Errorf("seat %d's unchosen creature was sacrificed", seat)
		}
	}
	if !g.Battlefield.Contains(crafter) {
		t.Error("the Plaguecrafter was not chosen and must stay")
	}

	chosen := make(map[uuid.UUID]bool, len(picks))
	for _, id := range picks {
		chosen[id] = true
	}
	var batch uint64
	exits := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventLTB || !chosen[ev.CardID] {
			continue
		}
		exits++
		if ev.Batch == 0 {
			t.Fatalf("creature %s left with no event batch", ev.CardID)
		}
		if batch == 0 {
			batch = ev.Batch
		} else if ev.Batch != batch {
			t.Errorf("creature %s left in batch %d, want the shared batch %d", ev.CardID, ev.Batch, batch)
		}
	}
	if exits != len(picks) {
		t.Fatalf("observed %d exits, want %d", exits, len(picks))
	}
	if n := lilianaPendingTriggersFrom(g, artist); n != len(picks) {
		t.Errorf("Blood Artist triggers = %d, want %d — one per death, all in one batch", n, len(picks))
	}
}
