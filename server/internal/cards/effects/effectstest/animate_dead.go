package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// animate_dead_test.go — ADR 0154 (#1780): Animate Dead and Dance of the
// Dead end to end, through the cast, the Aura resolving onto a card in a
// graveyard, the enters trigger, the delayed sacrifice, and the rulings
// the ADR lists.

const (
	AnimateDeadOracle    = "c0d8fef4-65f4-4769-982d-b397d2b7e977"
	DanceOfTheDeadOracle = "15d5b198-edbe-40f3-87e2-d72f7f0c58ab"
)

// adGraveyardCreature puts a 3/3 creature card into p's graveyard.
// `keywords` lets a test give it protection.
func AdGraveyardCreature(p *game.Player, name string, keywords ...string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Zombie",
		ManaCost: "{2}{B}", Colors: []string{"B"}, Power: 3, Toughness: 3,
		Owner: p.ID, Controller: p.ID, Keywords: keywords,
	})
	return id
}

// adCast puts a black {1}{B} Aura into the active seat's hand, pays for
// it from the pool and casts it at `target`. The colour matters: a
// creature with protection from black must see a black Aura.
func AdCast(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Enchantment — Aura", OracleID: oracle,
		ManaCost: "{1}{B}", Colors: []string{"B"}, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	FillPoolColored(me, "B", 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// adResolveSpell passes priority until the Aura spell has left the stack:
// it is on the battlefield with its enters trigger waiting, or it was
// countered.
func AdResolveSpell(t *testing.T, g *game.Game, spell uuid.UUID) {
	t.Helper()
	for i := 0; i < 16; i++ {
		if !g.Stack.Contains(spell) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the Aura spell never left the stack")
}

// adAnimate casts Animate Dead (or Dance of the Dead) at `target` and
// settles everything.
func AdAnimate(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID) uuid.UUID {
	t.Helper()
	aura := AdCast(t, g, name, oracle, target)
	PassPriorityAroundTable(t, g)
	return aura
}

func AdCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.WithWriteLock(func() { c, ok = g.LookupCardForEffect(id) })
	if !ok {
		t.Fatalf("card %s is in no zone", id)
	}
	return c
}

func AdZone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var k game.ZoneKind
	g.WithWriteLock(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			k = z.Kind
		}
	})
	return k
}

func AdSacrificedBy(g *game.Game, id uuid.UUID) (uuid.UUID, bool) {
	for _, ev := range g.Events {
		if ev.Kind == game.EventSacrifice && ev.CardID == id {
			return ev.Actor, true
		}
	}
	return uuid.Nil, false
}
