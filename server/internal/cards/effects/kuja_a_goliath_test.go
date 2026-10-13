package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_a_goliath_test.go — Goliath Daydreamer (#3213) and the engine
// piece under it, a spell exiled with a counter as it resolves.

func goliathGame(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	return g, pushDiesCreatureForTest(g, me.ID, "Goliath Daydreamer", goliathOracle, "Creature — Giant Wizard", 4, 4)
}

func exiledCard(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == id {
			return &g.Exile.Cards[i]
		}
	}
	return nil
}

func TestGoliathExilesAResolvedSpellCastFromHandWithADreamCounter(t *testing.T) {
	g, _ := goliathGame(t)
	me := g.Seats[0]
	div := castCatalogSpell(t, g, "Divination", "Sorcery", spellControlDivinationOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("Divination should still resolve and draw two: hand %d -> %d", hand, got)
	}
	c := exiledCard(g, div)
	if c == nil {
		t.Fatal("the resolved spell should be in exile")
	}
	if c.Counters[game.CounterDream] != 1 {
		t.Errorf("counters = %v, want one dream counter", c.Counters)
	}
	if me.Graveyard.Contains(div) {
		t.Error("it must not also be in the graveyard")
	}
}

func TestGoliathLeavesACounteredSpellInTheGraveyard(t *testing.T) {
	g, _ := goliathGame(t)
	me := g.Seats[0]
	div := castCatalogSpell(t, g, "Divination", "Sorcery", spellControlDivinationOracle, nil)
	// Let Goliath's trigger resolve (it marks the spell), then counter
	// the spell before it resolves.
	marked := false
	for i := 0; i < 8 && !marked; i++ {
		g.ReadSnapshot(func() {
			if it := g.StackMeta[div]; it != nil && it.ExileWithCounter != "" {
				marked = true
			}
		})
		if marked {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !marked {
		t.Fatal("Goliath's trigger never marked the spell")
	}
	castCatalogSpell(t, g, "Counterspell", "Instant", b14CounterspellOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: div}})
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(div) {
		t.Error("a countered spell goes to the graveyard: the exile replaces only a resolution")
	}
	if exiledCard(g, div) != nil {
		t.Error("a countered spell must not be exiled with a dream counter")
	}
}

func TestGoliathIgnoresASpellNotCastFromHand(t *testing.T) {
	g, _ := goliathGame(t)
	me := g.Seats[0]
	// A Divination in exile with a free-cast grant (an impulse effect).
	div := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: div, Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}",
		OracleID: spellControlDivinationOracle, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(div, game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Cost: "{0}", CastOnly: true})
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(div) {
		t.Error("a spell cast from exile goes to the graveyard as usual")
	}
	if exiledCard(g, div) != nil {
		t.Error("a spell cast from exile must not get a dream counter")
	}
}

func TestGoliathAttackLetsYouCastOneDreamSpellFreeAndNothingElse(t *testing.T) {
	g, goliath := goliathGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spellA := uuid.New()
	spellB := uuid.New()
	land := uuid.New()
	noCounter := uuid.New()
	theirs := uuid.New()
	dream := map[string]int{game.CounterDream: 1}
	push := func(id uuid.UUID, name, typ string, owner uuid.UUID, counters map[string]int) {
		g.Exile.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typ, ManaCost: "{2}{U}",
			OracleID: spellControlDivinationOracle, Owner: owner, Controller: owner, Counters: counters})
	}
	push(spellA, "Dreamt Spell A", "Sorcery", me.ID, dream)
	push(spellB, "Dreamt Spell B", "Sorcery", me.ID, dream)
	push(land, "Dreamt Land", "Basic Land — Island", me.ID, dream)
	push(noCounter, "Plain Exiled Spell", "Sorcery", me.ID, nil)
	push(theirs, "Their Dreamt Spell", "Sorcery", opp.ID, dream)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(goliath, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"a land": land, "a card with no counter": noCounter, "an opponent's card": theirs} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"}); err == nil {
			t.Errorf("%s must not be castable through Goliath's grant", name)
		}
	}
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, spellA, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("free cast of a dream spell: %v", err)
	}
	if err := g.CastSpell(me.ID, spellB, game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("\"a spell\": only one may be cast through the grant")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("the free Divination should draw two: hand %d -> %d", hand, got)
	}
}

func TestGoliathAttackWithNothingDreamingDoesNothing(t *testing.T) {
	g, goliath := goliathGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(goliath, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	_ = me
	if len(g.PendingChoices) != 0 {
		t.Errorf("pending choices = %d, want none", len(g.PendingChoices))
	}
}
