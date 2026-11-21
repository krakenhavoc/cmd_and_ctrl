package testsm

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardslu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/lu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// memory_jar_test.go — #5841: each player exiles their hand face down
// (hidden from every seat) and draws seven; at the beginning of the next end
// step each player discards their hand and gets back exactly the cards they
// exiled this way that are still the objects they were (CR 400.7).

func TestMemoryJarIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(cardslu.MemoryJarOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Memory Jar: %+v %v", spec, ok)
	}
}

// jarTable is a 4-seat table in the active player's main phase where every
// player holds a named card on top of their opening hand, with Memory Jar
// activated by the active player and its ability resolved.
type jarTable struct {
	g      *game.Game
	me     *game.Player
	jar    uuid.UUID
	secret map[uuid.UUID]string      // player -> the name of the card they added
	old    map[uuid.UUID][]uuid.UUID // player -> their hand before the activation
	drawn  map[uuid.UUID][]uuid.UUID // player -> their hand right after it
}

func jarHandIDs(p *game.Player) []uuid.UUID {
	out := make([]uuid.UUID, 0, p.Hand.Size())
	for _, c := range p.Hand.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}

func newJarTable(t *testing.T, step game.Step) jarTable {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, step)
	jt := jarTable{
		g:      g,
		me:     g.Seats[g.Turn.ActiveSeat],
		secret: map[uuid.UUID]string{},
		old:    map[uuid.UUID][]uuid.UUID{},
		drawn:  map[uuid.UUID][]uuid.UUID{},
	}
	if len(g.Seats) != 4 {
		t.Fatalf("want a 4-seat table, have %d", len(g.Seats))
	}
	for i, p := range g.Seats {
		name := fmt.Sprintf("Jar Secret %c", 'A'+i)
		effectstest.HandCardOf(g, p, name, "Sorcery", "{1}", "")
		jt.secret[p.ID] = name
		jt.old[p.ID] = jarHandIDs(p)
	}
	jt.jar = effectstest.PushCatalogPermanent(g, jt.me.ID, "Memory Jar", "Artifact", cardslu.MemoryJarOracle, false)
	if err := g.ActivateCatalogAbility(jt.me.ID, jt.jar, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Memory Jar: %v", err)
	}
	if z := effectstest.ZoneOfCard(g, jt.jar); z != game.ZoneGraveyard {
		t.Fatalf("Memory Jar is in %v after its cost, want the graveyard", z)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		jt.drawn[p.ID] = jarHandIDs(p)
	}
	return jt
}

// jarSameIDs reports whether a and b hold the same IDs, in any order.
func jarSameIDs(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[uuid.UUID]int{}
	for _, id := range a {
		seen[id]++
	}
	for _, id := range b {
		seen[id]--
	}
	for _, n := range seen {
		if n != 0 {
			return false
		}
	}
	return true
}

func jarExiledCard(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

// checkJarExiled asserts every player's old hand is in exile face down and
// their hand is seven new cards.
func checkJarExiled(t *testing.T, jt jarTable) {
	t.Helper()
	g := jt.g
	for _, p := range g.Seats {
		for _, id := range jt.old[p.ID] {
			c, ok := jarExiledCard(g, id)
			if !ok {
				t.Fatalf("seat %s: card %v is not in exile", p.Name, id)
			}
			if !c.FaceDown || c.FaceDownKind != game.FaceDownExiled {
				t.Errorf("seat %s: exiled card %v is face down=%v kind=%q, want face down %q",
					p.Name, id, c.FaceDown, c.FaceDownKind, game.FaceDownExiled)
			}
			if len(c.KnownBy) != 0 {
				t.Errorf("seat %s: exiled card %v is known by %v, want nobody (CR 406.3)", p.Name, id, c.KnownBy)
			}
		}
		if got := p.Hand.Size(); got != 7 {
			t.Errorf("seat %s holds %d cards, want the seven drawn", p.Name, got)
		}
		for _, id := range jt.old[p.ID] {
			if p.Hand.Contains(id) {
				t.Errorf("seat %s still holds old card %v", p.Name, id)
			}
		}
	}
}

func TestMemoryJarEachPlayerExilesTheirHandFaceDownAndDrawsSeven(t *testing.T) {
	jt := newJarTable(t, game.StepPrecombatMain)
	checkJarExiled(t, jt)
	if n := len(jt.g.DelayedTriggers); n != 1 {
		t.Fatalf("%d delayed triggers queued, want 1", n)
	}
}

func TestMemoryJarExiledCardsAreHiddenFromEverySeat(t *testing.T) {
	jt := newJarTable(t, game.StepPrecombatMain)
	g := jt.g
	for _, viewer := range g.Seats {
		v := protocol.ViewOfGameFor(g, viewer.ID.String())
		exiled := map[string]bool{}
		for _, p := range g.Seats {
			for _, id := range jt.old[p.ID] {
				exiled[id.String()] = true
			}
		}
		found := 0
		for _, c := range v.Exile.Cards {
			if !exiled[c.InstanceID] {
				continue
			}
			found++
			// CR 406.3: nobody may look, the owner included.
			if c.Name != "" {
				t.Errorf("viewer %s sees exiled card %s as %q", viewer.Name, c.InstanceID, c.Name)
			}
		}
		if found != len(exiled) {
			t.Errorf("viewer %s sees %d of the %d exiled cards", viewer.Name, found, len(exiled))
		}
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal view: %v", err)
		}
		for _, p := range g.Seats {
			if p.ID == viewer.ID {
				continue
			}
			if strings.Contains(string(raw), jt.secret[p.ID]) {
				t.Errorf("viewer %s's view names %q, a card %s exiled face down", viewer.Name, jt.secret[p.ID], p.Name)
			}
		}
	}
}

// checkJarReturned asserts every player discarded the seven they drew and
// holds exactly their old hand, less `gone`.
func checkJarReturned(t *testing.T, g *game.Game, jt jarTable, gone map[uuid.UUID]bool) {
	t.Helper()
	if n := len(g.DelayedTriggers); n != 0 {
		t.Errorf("%d delayed triggers still queued after the end step", n)
	}
	for _, p := range g.Seats {
		var want []uuid.UUID
		for _, id := range jt.old[p.ID] {
			if !gone[id] {
				want = append(want, id)
			}
		}
		if got := jarHandIDs(p); !jarSameIDs(got, want) {
			t.Errorf("seat %s holds %v, want their exiled cards %v", p.Name, got, want)
		}
		for _, id := range jt.drawn[p.ID] {
			if !p.Graveyard.Contains(id) {
				t.Errorf("seat %s: drawn card %v was not discarded to their graveyard", p.Name, id)
			}
		}
		for _, id := range want {
			for _, c := range p.Hand.Cards {
				if c.InstanceID == id && (c.FaceDown || !c.KnownBy[p.ID]) {
					t.Errorf("seat %s: returned card %v is face down=%v or unknown to its owner", p.Name, id, c.FaceDown)
				}
			}
		}
	}
}

func TestMemoryJarEndStepDiscardsAndReturnsEachPlayersOwnCards(t *testing.T) {
	jt := newJarTable(t, game.StepPrecombatMain)
	g := jt.g
	mark := len(g.Events)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if !effectstest.TriggerOnStackAny(g) {
		t.Fatal("the delayed trigger did not fire at the beginning of the end step")
	}
	effectstest.PassPriorityAroundTable(t, g)
	checkJarReturned(t, g, jt, nil)
	discards := 0
	for _, ev := range g.Events[mark:] {
		if ev.Kind == game.EventDiscardCard {
			discards++
		}
	}
	if discards != 7*len(g.Seats) {
		t.Errorf("%d discard events, want %d (a real discard of every drawn card)", discards, 7*len(g.Seats))
	}
}

func TestMemoryJarDoesNotReturnACardThatLeftExile(t *testing.T) {
	jt := newJarTable(t, game.StepPrecombatMain)
	g := jt.g
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	left, again := jt.old[opp.ID][0], jt.old[opp.ID][1]
	exile := game.ZoneRef{Kind: game.ZoneExile}
	yard := game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}
	if err := g.MoveCardByID(exile, yard, left); err != nil {
		t.Fatalf("move to graveyard: %v", err)
	}
	// Out and back in: a new object (CR 400.7), not one exiled this way.
	if err := g.MoveCardByID(exile, yard, again); err != nil {
		t.Fatalf("move to graveyard: %v", err)
	}
	if err := g.MoveCardByID(yard, exile, again); err != nil {
		t.Fatalf("move back to exile: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	checkJarReturned(t, g, jt, map[uuid.UUID]bool{left: true, again: true})
	if z := effectstest.ZoneOfCard(g, left); z != game.ZoneGraveyard {
		t.Errorf("the card that left exile is in %v, want the graveyard it went to", z)
	}
	if z := effectstest.ZoneOfCard(g, again); z != game.ZoneExile {
		t.Errorf("the card exiled again is in %v, want it left in exile", z)
	}
}

func TestMemoryJarReturnSurvivesASnapshotRoundTrip(t *testing.T) {
	jt := newJarTable(t, game.StepPrecombatMain)
	g := effectstest.RestoreRoundTrip(t, jt.g, true)
	checkJarExiled(t, jarTable{g: g, old: jt.old})
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if !effectstest.TriggerOnStackAny(g) {
		t.Fatal("restored: the delayed trigger did not fire at the beginning of the end step")
	}
	// And again with the trigger on the stack.
	g = effectstest.RestoreRoundTrip(t, g, true)
	effectstest.PassPriorityAroundTable(t, g)
	checkJarReturned(t, g, jt, nil)
}

func TestMemoryJarActivatedInAnEndStepWaitsForTheNextOne(t *testing.T) {
	jt := newJarTable(t, game.StepEnd)
	g := jt.g
	seq := g.Turn.Seq
	if n := len(g.DelayedTriggers); n != 1 {
		t.Fatalf("%d delayed triggers queued, want 1 for the next end step", n)
	}
	checkJarExiled(t, jt)
	// Out of this end step: the hands are untouched, and the trigger waits.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if len(g.DelayedTriggers) != 1 || effectstest.TriggerOnStackAny(g) {
		t.Fatal("the delayed trigger fired in the end step Memory Jar was activated in")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if g.Turn.Seq != seq+1 {
		t.Fatalf("at turn %d's end step, want the next turn's (%d)", g.Turn.Seq, seq+1)
	}
	if !effectstest.TriggerOnStackAny(g) {
		t.Fatal("the delayed trigger did not fire at the next turn's end step")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		for _, id := range jt.old[p.ID] {
			if !p.Hand.Contains(id) {
				t.Errorf("seat %s did not get back %v", p.Name, id)
			}
		}
	}
}
