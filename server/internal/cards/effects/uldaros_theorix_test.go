package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// uldaros_theorix_test.go — Uldaros Theorix (#2017): one target clause per
// card type, and copies cast against a running total mana value of 6.

const uldarosOracle = "020d561b-2e5d-439e-9e6a-318d23e73542"

// uldarosCast casts Theorix for me and answers each "up to one target
// <type>" prompt from picks, keyed by the type word in the prompt's
// label; a type not named is answered with nothing. It returns the
// labels asked, in order, and leaves the trigger on the stack.
func uldarosCast(t *testing.T, g *game.Game, me *game.Player, picks map[string]uuid.UUID) []string {
	t.Helper()
	castCatalogSpell(t, g, "Uldaros Theorix", "Legendary Creature — Elder Sphinx", uldarosOracle, nil)
	passPriorityAroundTable(t, g)
	var asked []string
	for i := 0; i < 12; i++ {
		p := latestPickTarget(g, me.ID)
		if p == nil {
			break
		}
		asked = append(asked, p.Reason)
		var refs []game.TargetRef
		for word, id := range picks {
			if strings.Contains(p.Reason, " "+word+" card") && hasID(p.PickTargetCards, id) {
				refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
				break
			}
		}
		if err := g.ResolvePickTargets(p.ID, me.ID, refs); err != nil {
			t.Fatalf("answer %q: %v", p.Reason, err)
		}
	}
	return asked
}

// uldarosResolve runs the trigger to resolution, leaving the copies offered.
func uldarosResolve(t *testing.T, g *game.Game) {
	t.Helper()
	passPriorityAroundTable(t, g)
}

// castCopyNamed casts the copy named name; the error is the cast's.
func castCopyNamed(t *testing.T, g *game.Game, me *game.Player, name string) error {
	t.Helper()
	return castCopy(g, me.ID, copyOf(t, g, name).InstanceID, true)
}

// uldarosGraveyard seeds three nonland cards costing 3, 2 and 4 (an
// artifact creature, an instant and a sorcery) and returns their ids.
func uldarosGraveyard(me *game.Player) (golem, bolt, ritual uuid.UUID) {
	golem = gyCard(me, "Test Golem", "Artifact Creature — Golem", "{3}", "")
	bolt = gyCard(me, "Test Bolt", "Instant", "{1}{R}", "")
	ritual = gyCard(me, "Test Ritual", "Sorcery", "{2}{B}{B}", "")
	return
}

func TestUldarosTheorixOffersTotalManaValueSixAndNoMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	golem, bolt, ritual := uldarosGraveyard(me)

	uldarosCast(t, g, me, map[string]uuid.UUID{"artifact": golem, "instant": bolt, "sorcery": ritual})
	uldarosResolve(t, g)
	for _, id := range []uuid.UUID{golem, bolt, ritual} {
		if !inExile(g, id) {
			t.Fatalf("%v was picked and not exiled", id)
		}
	}
	if n := len(cardCopies(g)); n != 3 {
		t.Fatalf("%d copies made, want 3 (one per exiled card)", n)
	}

	// 3 + 2 = 5 fits; the 4 would make 9.
	if err := castCopyNamed(t, g, me, "Test Golem"); err != nil {
		t.Fatalf("cast the mana value 3 copy: %v", err)
	}
	if err := castCopyNamed(t, g, me, "Test Bolt"); err != nil {
		t.Fatalf("cast the mana value 2 copy (total 5): %v", err)
	}
	err := castCopyNamed(t, g, me, "Test Ritual")
	if !errors.Is(err, game.ErrSpellManaValueTooHigh) {
		t.Fatalf("cast the mana value 4 copy at total 5: err = %v, want ErrSpellManaValueTooHigh", err)
	}
}

func TestUldarosTheorixTotalIsOrderIndependent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	golem, bolt, ritual := uldarosGraveyard(me)

	uldarosCast(t, g, me, map[string]uuid.UUID{"artifact": golem, "instant": bolt, "sorcery": ritual})
	uldarosResolve(t, g)

	// 4 + 2 = 6 is exactly the budget, in the other order.
	if err := castCopyNamed(t, g, me, "Test Ritual"); err != nil {
		t.Fatalf("cast the mana value 4 copy first: %v", err)
	}
	if err := castCopyNamed(t, g, me, "Test Bolt"); err != nil {
		t.Fatalf("cast the mana value 2 copy (total 6): %v", err)
	}
	if err := castCopyNamed(t, g, me, "Test Golem"); !errors.Is(err, game.ErrSpellManaValueTooHigh) {
		t.Fatalf("a third copy past the total: err = %v, want ErrSpellManaValueTooHigh", err)
	}
}

func TestUldarosTheorixPermanentCopyBecomesAToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	golem, _, _ := uldarosGraveyard(me)

	uldarosCast(t, g, me, map[string]uuid.UUID{"artifact": golem})
	uldarosResolve(t, g)
	if err := castCopyNamed(t, g, me, "Test Golem"); err != nil {
		t.Fatalf("cast the copy: %v", err)
	}
	passPriorityAroundTable(t, g)
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Test Golem" {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil || !token.IsToken() {
		t.Fatalf("the permanent copy did not resolve as a token: %+v", token)
	}
	if !inExile(g, golem) {
		t.Error("the original stays in exile")
	}
}

func TestUldarosTheorixUncastCopiesCeaseWhenTheCasterDeclines(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	golem, bolt, _ := uldarosGraveyard(me)

	uldarosCast(t, g, me, map[string]uuid.UUID{"artifact": golem, "instant": bolt})
	uldarosResolve(t, g)
	if n := len(cardCopies(g)); n != 2 {
		t.Fatalf("%d copies, want 2", n)
	}
	if err := castCopyNamed(t, g, me, "Test Bolt"); err != nil {
		t.Fatalf("cast one copy: %v", err)
	}
	declineResolutionCasts(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left after the caster declined: an uncast copy ceases to exist", n)
	}
}

// An artifact creature card fills ONE slot: after it is picked for the
// artifact slot the creature slot does not offer it again, and a creature
// card that is not an artifact is still offered there.
func TestUldarosTheorixACardWithTwoTypesFillsOneSlot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	golem := gyCard(me, "Test Golem", "Artifact Creature — Golem", "{3}", "")
	bear := gyCard(me, "Test Bear", "Creature — Bear", "{1}{G}", "")
	forest := gyCard(me, "Test Forest", "Basic Land — Forest", "", "")

	castCatalogSpell(t, g, "Uldaros Theorix", "Legendary Creature — Elder Sphinx", uldarosOracle, nil)
	passPriorityAroundTable(t, g)
	art := latestPickTarget(g, me.ID)
	if art == nil || !strings.Contains(art.Reason, "artifact") {
		t.Fatalf("the first slot is the artifact one: %+v", art)
	}
	if !hasID(art.PickTargetCards, golem) || hasID(art.PickTargetCards, bear) || hasID(art.PickTargetCards, forest) {
		t.Fatalf("artifact slot offers %v, want only the golem", art.PickTargetCards)
	}
	if err := g.ResolvePickTargets(art.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: golem}}); err != nil {
		t.Fatal(err)
	}
	cre := latestPickTarget(g, me.ID)
	if cre == nil || !strings.Contains(cre.Reason, "creature") {
		t.Fatalf("the next slot with a candidate is the creature one: %+v", cre)
	}
	if hasID(cre.PickTargetCards, golem) {
		t.Error("the golem was picked for the artifact slot and is offered again for creature")
	}
	if !hasID(cre.PickTargetCards, bear) || hasID(cre.PickTargetCards, forest) {
		t.Errorf("creature slot offers %v, want only the bear (a land is never a nonland card)", cre.PickTargetCards)
	}
}

// A slot with nothing to pick is not asked, and a graveyard with no
// nonland card asks nothing at all.
func TestUldarosTheorixSkipsSlotsWithNoCandidate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	_, bolt, _ := uldarosGraveyard(me)
	asked := uldarosCast(t, g, me, map[string]uuid.UUID{"instant": bolt})
	// The golem answers artifact AND creature; the bolt, instant; the
	// ritual, sorcery. No battle, enchantment, kindred or planeswalker.
	for _, label := range asked {
		for _, none := range []string{"battle", "enchantment", "kindred", "planeswalker"} {
			if strings.Contains(label, " "+none+" card") {
				t.Errorf("asked %q with no such card in the graveyard", label)
			}
		}
	}
	if len(asked) != 4 {
		t.Errorf("asked %d slots (%v), want artifact, creature, instant, sorcery", len(asked), asked)
	}
}

func TestUldarosTheorixNeedsToHaveBeenCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	uldarosGraveyard(me)
	card := game.Card{InstanceID: uuid.New(), Name: "Uldaros Theorix", TypeLine: "Legendary Creature — Elder Sphinx", OracleID: uldarosOracle, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(card.InstanceID, game.HandEntryOptions{}); err != nil {
			t.Fatalf("put onto the battlefield: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil || len(g.PendingTriggers) != 0 {
		t.Errorf("a Theorix that was not cast triggered: %+v", p)
	}
}

func TestUldarosTheorixWithNoTargetsDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	golem, _, _ := uldarosGraveyard(me)
	uldarosCast(t, g, me, nil)
	uldarosResolve(t, g)
	if inExile(g, golem) || len(cardCopies(g)) != 0 {
		t.Error("\"up to one\" answered with none exiles and copies nothing")
	}
}

// The enumerator (what a bot picks from) offers exactly the casts the gate
// accepts, and tracks the running total as casts are made.
func TestUldarosTheorixEnumeratorFollowsTheRunningTotal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	golem, bolt, ritual := uldarosGraveyard(me)
	uldarosCast(t, g, me, map[string]uuid.UUID{"artifact": golem, "instant": bolt, "sorcery": ritual})
	uldarosResolve(t, g)

	offered := func(name string) bool {
		id := copyOf(t, g, name).InstanceID
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Kind == legal.KindCast && strings.Contains(string(m.Params), id.String()) {
				return true
			}
		}
		return false
	}
	for _, name := range []string{"Test Golem", "Test Bolt", "Test Ritual"} {
		if !offered(name) {
			t.Errorf("%s is under the total of 6 and not offered", name)
		}
	}
	if err := castCopyNamed(t, g, me, "Test Ritual"); err != nil {
		t.Fatal(err)
	}
	if offered("Test Golem") {
		t.Error("the mana value 3 copy is offered with 2 left")
	}
	if !offered("Test Bolt") {
		t.Error("the mana value 2 copy is not offered with 2 left")
	}
}

// The running total is part of the snapshot: a restored game still
// refuses a copy over what is left.
func TestUldarosTheorixTotalSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	golem, bolt, ritual := uldarosGraveyard(me)
	uldarosCast(t, g, me, map[string]uuid.UUID{"artifact": golem, "instant": bolt, "sorcery": ritual})
	uldarosResolve(t, g)
	if err := castCopyNamed(t, g, me, "Test Ritual"); err != nil {
		t.Fatal(err)
	}

	r := restoreRoundTrip(t, g, false)
	rme := r.PlayerByIDForEffect(me.ID)
	if err := castCopyNamed(t, r, rme, "Test Golem"); !errors.Is(err, game.ErrSpellManaValueTooHigh) {
		t.Fatalf("restored: a 3 with 2 left: err = %v, want ErrSpellManaValueTooHigh", err)
	}
	if err := castCopyNamed(t, r, rme, "Test Bolt"); err != nil {
		t.Fatalf("restored: a 2 with 2 left: %v", err)
	}
}

func TestUldarosTheorixIsComplete(t *testing.T) {
	spec, ok := Lookup(uldarosOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Uldaros Theorix ships in full: %+v", spec)
	}
}
