package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Semester's End (#4193, Guffster's Party #4184).

const semestersEndDoublingSeasonOracle = "01546b7d-a233-4176-8843-d732074dc5b6"

func semestersEndCard(t *testing.T, g *game.Game, name string) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("%s is not on the battlefield", name)
	return game.Card{}
}

func semestersEndWalker(g *game.Game, owner uuid.UUID, name, typeLine string, printed, current, p, tough int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: "oracle-" + name,
		StartingLoyalty: printed, Power: p, Toughness: tough,
		Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterLoyalty: current},
	})
}

func semestersEndCast(t *testing.T, g *game.Game, targets ...uuid.UUID) {
	t.Helper()
	refs := make([]game.TargetRef, 0, len(targets))
	for _, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	castCatalogSpell(t, g, "Semester's End", "Instant", guffaSemestersOracle, refs)
	passPriorityAroundTable(t, g)
	for _, id := range targets {
		if g.Battlefield.Contains(id) {
			t.Fatalf("target %s was not exiled", id)
		}
	}
}

// A creature gets a +1/+1 counter, a planeswalker a loyalty counter on
// top of its PRINTED loyalty (it left at 1 and comes back at 3 + 1), and
// a creature planeswalker both — in one return, so every card shares
// one entry (CR 603.6a).
func TestSemestersEndReturnsEachWithItsOwnAdditionalCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	walker := semestersEndWalker(g, me.ID, "Walker", "Legendary Planeswalker — Test", 3, 1, 0, 0)
	both := semestersEndWalker(g, me.ID, "Creature Walker", "Legendary Creature Planeswalker — Test", 4, 2, 2, 2)
	bystander := pushVanillaCreature(g, me.ID, "Bystander", 2, 2)
	pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	semestersEndCast(t, g, bear, walker, both)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	b := semestersEndCard(t, g, "Bear")
	if b.Counters[game.CounterPlusOne] != 1 || b.Counters[game.CounterLoyalty] != 0 {
		t.Errorf("Bear counters %v, want one +1/+1 and no loyalty", b.Counters)
	}
	if b.Controller != me.ID {
		t.Errorf("Bear returned under %s, want its owner", b.Controller)
	}
	w := semestersEndCard(t, g, "Walker")
	if w.Counters[game.CounterLoyalty] != 4 || w.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("Walker counters %v, want 4 loyalty (printed 3 + 1) and no +1/+1", w.Counters)
	}
	cw := semestersEndCard(t, g, "Creature Walker")
	if cw.Counters[game.CounterLoyalty] != 5 || cw.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("Creature Walker counters %v, want 5 loyalty (printed 4 + 1) and one +1/+1", cw.Counters)
	}
	if b.EntryOrdinal == 0 || b.EntryOrdinal != w.EntryOrdinal || w.EntryOrdinal != cw.EntryOrdinal {
		t.Errorf("entry ordinals %d, %d, %d: want one simultaneous return", b.EntryOrdinal, w.EntryOrdinal, cw.EntryOrdinal)
	}
	if c := semestersEndCard(t, g, "Bystander"); len(c.Counters) != 0 || c.InstanceID != bystander {
		t.Errorf("an untargeted creature was touched: %v", c.Counters)
	}
	if c := semestersEndCard(t, g, "Their Bear"); len(c.Counters) != 0 {
		t.Errorf("an opponent's creature got counters: %v", c.Counters)
	}
}

// An opponent's creature is not a legal target: "you control".
func TestSemestersEndCannotTargetAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Semester's End", TypeLine: "Instant",
		OracleID: guffaSemestersOracle, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}})
	if err == nil {
		t.Fatal("Semester's End was cast at an opponent's creature")
	}
	// The same cast at a creature you control is legal.
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}}); err != nil {
		t.Fatalf("Semester's End at your own creature: %v", err)
	}
}

// CR 614.12: the card is judged as it would exist on the battlefield,
// after the copy it is told to enter as. A Spark Double that comes back
// as a copy of a planeswalker is a planeswalker, not the Illusion
// creature card it is in exile: it gets the loyalty counter (on top of
// the copied loyalty and Spark Double's own extra one) and no +1/+1
// counter.
func TestSemestersEndJudgesASparkDoubleByWhatItCopies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	walker := semestersEndWalker(g, me.ID, "Walker", "Legendary Planeswalker — Test", 3, 3, 0, 0)

	double := castCatalogSpell(t, g, "Spark Double", "Creature — Illusion", oracleSparkDoubl, nil)
	resolveWithCopyChoice(t, g, bear)
	if got := copyBattlefieldCard(t, g, double); got.Counters[game.CounterPlusOne] != 1 {
		t.Fatalf("Spark Double as a Bears has %v, want its own +1/+1", got.Counters)
	}

	semestersEndCast(t, g, double)
	advanceToEndStepOf(t, g, 0)
	resolveWithCopyChoice(t, g, walker)

	var back *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Walker" && c.InstanceID != walker {
			back = c
		}
	}
	if back == nil {
		t.Fatal("Spark Double did not return as a copy of the walker")
	}
	if back.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("the planeswalker copy got %d +1/+1 counters, want 0", back.Counters[game.CounterPlusOne])
	}
	// Printed 3, Spark Double's additional 1, Semester's End's 1.
	if back.Counters[game.CounterLoyalty] != 5 {
		t.Errorf("the planeswalker copy has %d loyalty, want 5", back.Counters[game.CounterLoyalty])
	}
}

// CR 122.6: the counters a permanent enters with are put on it, so
// Doubling Season doubles both the printed loyalty and the additional
// counter, and the +1/+1 counter.
func TestSemestersEndCountersAreDoubledByDoublingSeason(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Doubling Season", "Enchantment", semestersEndDoublingSeasonOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	walker := semestersEndWalker(g, me.ID, "Walker", "Legendary Planeswalker — Test", 3, 1, 0, 0)

	semestersEndCast(t, g, bear, walker)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	if got := semestersEndCard(t, g, "Bear").Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("Bear has %d +1/+1 counters under Doubling Season, want 2", got)
	}
	if got := semestersEndCard(t, g, "Walker").Counters[game.CounterLoyalty]; got != 8 {
		t.Errorf("Walker has %d loyalty under Doubling Season, want 8 ((3 + 1) doubled)", got)
	}
}
