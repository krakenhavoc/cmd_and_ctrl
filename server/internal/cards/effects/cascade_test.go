package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cascade_test.go — S28 sub-PR 3. The engine's exile / bottom / prompt
// behaviour is pinned in game/cascade_test.go; these tests are about
// the DECLARATIONS: that the keyword reaches the stack from a spell,
// that "cascade, cascade" is two triggers and not one, and that the
// two granting cards hand it to the right spells.

const (
	bloodbraidElfOracle    = "3f0c9466-5ab9-4205-a84f-b4b27b5a678e"
	shardlessAgentOracle   = "2afbaa9a-c171-4a8b-90f3-5250d8498356"
	maelstromWandererOracl = "ad9b7fbc-61c8-43ee-a65c-99206fd1e4df"
	maelstromNexusOracle   = "b1c346a4-fc9b-474a-9d51-cac168e363fa"
	imotiOracle            = "eca32dcd-6845-433e-a631-ed1f0ee78f25"
)

// castWithCost seeds a spell with a real mana cost into the active
// seat's hand, walks to a main phase, and casts it.
func castWithCost(t *testing.T, g *game.Game, name, typeLine, manaCost, oracle string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		OracleID: oracle, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// seedCheapLibrary replaces the active seat's library with a deck the
// cascade will hit: `n` cheap instants under a pile of lands.
func seedCheapLibrary(g *game.Game, p *game.Player, cheap int) {
	p.Library.Cards = nil
	for i := 0; i < cheap; i++ {
		c := game.NewCard("Cheap Spell", p.ID)
		c.TypeLine = "Instant"
		c.ManaCost = "{U}"
		p.Library.PushTop(c)
	}
}

// countMayCastPrompts reports how many cascade offers are waiting.
func countMayCastPrompts(g *game.Game) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMayCast {
			n++
		}
	}
	return n
}

// answerAllMayCast answers every outstanding offer the same way,
// re-reading the queue each time because resolving one can queue the
// next (Maelstrom Wanderer's second cascade).
func answerAllMayCast(t *testing.T, g *game.Game, chooser uuid.UUID, apply bool) int {
	t.Helper()
	answered := 0
	for i := 0; i < 8; i++ {
		var id uuid.UUID
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceMayCast && c.Chooser == chooser {
				id = c.ID
				break
			}
		}
		if id == uuid.Nil {
			return answered
		}
		if err := g.ResolveMayCast(id, chooser, apply); err != nil {
			t.Fatalf("ResolveMayCast: %v", err)
		}
		answered++
	}
	return answered
}

func TestCascadeIsWired(t *testing.T) {
	for _, tc := range []struct {
		oracle string
		want   int
	}{
		{bloodbraidElfOracle, 1},
		{shardlessAgentOracle, 1},
		{maelstromWandererOracl, 2}, // "cascade, cascade"
		{maelstromNexusOracle, 1},
		{imotiOracle, 2}, // has cascade AND grants it
	} {
		if got := len(game.CatalogTriggers(tc.oracle)); got != tc.want {
			t.Errorf("%s: %d triggered abilities, want %d", tc.oracle, got, tc.want)
		}
	}
}

// Casting a cascader queues the trigger from the STACK, resolves it,
// and offers the hit.
func TestBloodbraidElfCascadesOnCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 3)

	castWithCost(t, g, "Bloodbraid Elf", "Creature — Elf Berserker", "{2}{R}{G}", bloodbraidElfOracle)
	passPriorityAroundTable(t, g)

	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("cascade offers = %d, want 1", got)
	}
	if n := answerAllMayCast(t, g, me.ID, false); n != 1 {
		t.Errorf("answered %d offers, want 1", n)
	}
}

// "Cascade, cascade" is two abilities, so the Wanderer makes two
// separate offers.
func TestMaelstromWandererCascadesTwice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 6)

	castWithCost(t, g, "Maelstrom Wanderer", "Legendary Creature — Elemental", "{5}{G}{U}{R}", maelstromWandererOracl)
	passPriorityAroundTable(t, g)

	total := answerAllMayCast(t, g, me.ID, false)
	// The two triggers resolve one at a time with a priority pass
	// between them, so drain, pass, drain again.
	passPriorityAroundTable(t, g)
	total += answerAllMayCast(t, g, me.ID, false)
	if total != 2 {
		t.Errorf("Maelstrom Wanderer made %d cascade offers, want 2", total)
	}
}

// Maelstrom Nexus grants cascade to the FIRST spell of the turn and
// nothing after it.
func TestMaelstromNexusGrantsCascadeToTheFirstSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 6)
	pushCatalogPermanent(g, me.ID, "Maelstrom Nexus", "Enchantment", maelstromNexusOracle, false)

	castWithCost(t, g, "First Spell", "Sorcery", "{3}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("first spell: cascade offers = %d, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
	// #730: the cascade offer gated the table while it was open, so
	// the spell that made it is still on the stack until it is
	// answered — and the second cast is sorcery speed.
	passPriorityAroundTable(t, g)

	castWithCost(t, g, "Second Spell", "Sorcery", "{3}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Errorf("second spell: cascade offers = %d, want 0", got)
	}
}

// Imoti grants cascade on mana value 6 or greater, measured on the
// PRINTED cost.
func TestImotiGrantsCascadeToSixDrops(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 6)
	pushCatalogPermanent(g, me.ID, "Imoti, Celebrant of Bounty", "Legendary Creature — Snake Druid", imotiOracle, false)

	castWithCost(t, g, "Five Drop", "Sorcery", "{4}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Errorf("a five-drop got %d cascade offers, want 0", got)
	}

	castWithCost(t, g, "Six Drop", "Sorcery", "{5}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Errorf("a six-drop got %d cascade offers, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
}

// Accepting an offer really does leave a castable card in exile —
// the end-to-end claim the keyword makes.
func TestCascadeHitIsCastableForFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 1)

	castWithCost(t, g, "Shardless Agent", "Artifact Creature — Human Rogue", "{1}{G}{U}", shardlessAgentOracle)
	passPriorityAroundTable(t, g)
	answerAllMayCast(t, g, me.ID, true)

	var hit uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.ExilePlay.Active(me.ID, g.Turn.Number) {
			hit = c.InstanceID
		}
	}
	if hit == uuid.Nil {
		t.Fatalf("no free-cast grant in exile after accepting")
	}
	// Strict mana with an empty pool: the only way this succeeds is
	// if the grant really made it free.
	if err := g.CastSpell(me.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("casting the cascade hit for free: %v", err)
	}
}
