package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	radiantPerformerOracle = "ad16f918-5603-437b-b143-49c69ceae17d"
	radiantShockOracle     = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"
)

// radiantBig is a 2/5 creature, which survives a Shock.
func radiantBig(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Test",
		Power: 2, Toughness: 5, Owner: owner, Controller: owner})
	return id
}

// radiantSetUp: an opponent's Shock aimed at creature A is on the stack,
// and I flash in Radiant Performer from my hand in response; its trigger
// is put on the stack targeting the Shock.
func radiantSetUp(t *testing.T) (g *game.Game, me, opp *game.Player, shock, performer uuid.UUID, creatures []uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	opp = g.Seats[g.Turn.ActiveSeat]
	me = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	creatures = []uuid.UUID{
		radiantBig(g, opp.ID, "A"), radiantBig(g, opp.ID, "B"),
		radiantBig(g, me.ID, "C"), radiantBig(g, me.ID, "D"),
	}
	shock = uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: shock, Name: "Shock", TypeLine: "Instant", OracleID: radiantShockOracle,
		Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, shock, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creatures[0]}}}); err != nil {
		t.Fatalf("cast Shock: %v", err)
	}
	if err := g.PassPriority(); err != nil { // the caster passes; I hold priority now
		t.Fatalf("pass: %v", err)
	}
	performer = uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: performer, Name: "Radiant Performer", TypeLine: "Creature — Human Wizard",
		OracleID: radiantPerformerOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, performer, game.CastSpellParams{}); err != nil {
		t.Fatalf("flash in Radiant Performer: %v", err)
	}
	// Resolve the Performer; its trigger targets the Shock.
	for i := 0; i < 16; i++ {
		if _, onBattlefield := effectstest.BattlefieldCard(g, performer); onBattlefield {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	if _, ok := effectstest.BattlefieldCard(g, performer); !ok {
		t.Fatal("Radiant Performer did not resolve")
	}
	return g, me, opp, shock, performer, creatures
}

// radiantResolveTriggerTargeting answers the trigger's target walk with the
// Shock if it asks, and passes until the order prompt opens.
func radiantResolveTriggerTargeting(t *testing.T, g *game.Game, me *game.Player, shock uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 32; i++ {
		if c := effectstest.LatestOptionPickFor(g, me.ID); c != nil {
			return c
		}
		answered := false
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == me.ID {
				if err := g.ResolvePickTarget(c.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: shock}); err != nil {
					t.Fatalf("target the Shock: %v", err)
				}
				answered = true
				break
			}
		}
		if answered {
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	t.Fatal("no order prompt")
	return nil
}

// The headline (CR 707.10d): copying a Shock aimed at one of four
// creatures makes a copy for every other legal player and creature (the
// Performer included), each targeting a different one, put on the stack in
// the order I chose, and each resolves dealing 2.
func TestRadiantPerformerCopiesAShockForEveryOtherTarget(t *testing.T) {
	g, me, _, shock, performer, creatures := radiantSetUp(t)
	c := radiantResolveTriggerTargeting(t, g, me, shock)

	// Every player and every other creature, Performer included: 4 + 4.
	want := map[uuid.UUID]bool{performer: true, creatures[1]: true, creatures[2]: true, creatures[3]: true}
	for _, p := range g.Seats {
		want[p.ID] = true
	}
	if len(c.PickOptions) != len(want) {
		t.Fatalf("the order prompt offers %d targets, want %d", len(c.PickOptions), len(want))
	}
	// Place them last-offered first, to prove the order is mine.
	var order []uuid.UUID
	for c != nil {
		o := c.PickOptions[len(c.PickOptions)-1]
		id := o.Player
		if len(o.Cards) == 1 {
			id = o.Cards[0]
		}
		order = append(order, id)
		if err := g.ResolveOptionPick(c.ID, me.ID, len(c.PickOptions)-1); err != nil {
			t.Fatalf("order pick: %v", err)
		}
		if len(order) == 1 {
			// The order prompt is a restore point.
			seat := 0
			for i, p := range g.Seats {
				if p.ID == me.ID {
					seat = i
				}
			}
			g = effectstest.RestoreRoundTrip(t, g, true)
			me = g.Seats[seat]
		}
		c = effectstest.LatestOptionPickFor(g, me.ID)
	}
	// The last one is placed without a question.
	for id := range want {
		found := false
		for _, o := range order {
			found = found || o == id
		}
		if !found {
			order = append(order, id)
		}
	}
	// The copies are on the stack above the Shock, first placed lowest.
	var copies []uuid.UUID
	for _, sc := range g.Stack.Cards {
		if meta := g.StackMeta[sc.InstanceID]; meta != nil && meta.IsCopy {
			if meta.Controller != me.ID {
				t.Errorf("a copy is controlled by %s, want me", meta.Controller)
			}
			if len(meta.Targets) != 1 {
				t.Fatalf("a copy has %d targets", len(meta.Targets))
			}
			copies = append(copies, meta.Targets[0].ID)
		}
	}
	if len(copies) != len(order) {
		t.Fatalf("%d copies on the stack, want %d", len(copies), len(order))
	}
	for i := range order {
		if copies[i] != order[i] {
			t.Fatalf("copy %d targets %s, want %s (the order I chose)", i, copies[i], order[i])
		}
	}

	lifeBefore := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lifeBefore[p.ID] = p.Life
	}
	for i := 0; i < 400 && !effectstest.StackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	if !effectstest.StackFullyEmpty(g) {
		t.Fatal("the stack did not empty")
	}
	for _, p := range g.Seats {
		if got := lifeBefore[p.ID] - p.Life; got != 2 {
			t.Errorf("%s lost %d life, want 2", p.Name, got)
		}
	}
	for _, id := range creatures {
		card, ok := effectstest.BattlefieldCard(g, id)
		if !ok || card.DamageMarked != 2 {
			t.Errorf("creature %s has %d damage (on battlefield %v), want 2", id, card.DamageMarked, ok)
		}
	}
}

// A spell with two different targets is not a legal target for the
// trigger: "targets only a single permanent or player".
func TestRadiantPerformerRefusesASpellWithTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := radiantBig(g, me.ID, "A")
	b := radiantBig(g, me.ID, "B")
	pred := effects.ItemTargetsOnlyASinglePermanentOrPlayer()
	two := &game.StackItem{ID: uuid.New(), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}
	one := &game.StackItem{ID: uuid.New(), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}}}
	player := &game.StackItem{ID: uuid.New(), Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}
	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*game.StackItem{}
		}
		for _, it := range []*game.StackItem{two, one, player} {
			g.StackMeta[it.ID] = it
		}
	})
	if pred(g, me.ID, two) {
		t.Error("a two-target item reads as targeting a single permanent")
	}
	if !pred(g, me.ID, one) {
		t.Error("a one-permanent item does not read as targeting a single permanent")
	}
	if !pred(g, me.ID, player) {
		t.Error("a one-player item does not read as targeting a single player")
	}
}
