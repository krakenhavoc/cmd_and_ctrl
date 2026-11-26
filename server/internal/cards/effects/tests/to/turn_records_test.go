package teststo

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #6217: the turn records — names of the spells each player cast, the
// colours of the spell most recently cast, lands tapped for mana (CR
// 106.12), the last card drawn, and the permanents tapped to pay a
// source's costs.

func trHandSpell(g *game.Game, p *game.Player, name, cost string, colors ...string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: cost,
			Colors: colors, Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true}})
	})
	return id
}

func trCast(g *game.Game, p *game.Player, id uuid.UUID, color string) error {
	g.WithWriteLock(func() { effectstest.FillPoolColored(p, color, 1) })
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true})
}

// Mana Maze: a red spell right after a red spell is refused; a blue one is
// not, and then a red one is open again.
func TestManaMazeForbidsASpellSharingAColorWithTheLastSpell(t *testing.T) {
	const oracle = "test-6217-mana-maze"
	effectstest.RegisterForTest(t, effects.Spec{OracleID: oracle, Name: oracle,
		CastRestrictions: []game.CastRestriction{effects.CantCastSharingAColorWithTheLastSpell("Players can't cast spells that share a color with the spell most recently cast this turn.")},
	})
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Maze", "Enchantment", oracle, false)
	bolt := trHandSpell(g, me, "Bolt", "{R}", "R")
	shock := trHandSpell(g, me, "Shock", "{R}", "R")
	opt := trHandSpell(g, me, "Opt", "{U}", "U")
	burn := trHandSpell(g, me, "Burn", "{R}", "R")

	if err := trCast(g, me, bolt, "R"); err != nil {
		t.Fatalf("the turn's first spell: %v", err)
	}
	if err := trCast(g, me, shock, "R"); !errors.Is(err, game.ErrCantCast) {
		t.Fatalf("red after red: err = %v, want ErrCantCast", err)
	}
	if err := trCast(g, me, opt, "U"); err != nil {
		t.Fatalf("blue after red: %v", err)
	}
	if err := trCast(g, me, burn, "R"); err != nil {
		t.Fatalf("red after blue: %v", err)
	}
}

// Grim Reminder's reader: the caster cast a spell with that name this
// turn, and nobody else did.
func TestCastSpellNamesAreRecordedPerPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := trHandSpell(g, me, "Lightning Bolt", "{R}", "R")
	if err := trCast(g, me, bolt, "R"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	named := game.Card{Name: "Lightning Bolt"}
	other := game.Card{Name: "Shock"}
	g.WithWriteLock(func() {
		if !g.CastSpellNamedThisTurnForEffect(me.ID, named) {
			t.Error("the caster's Lightning Bolt is not recorded")
		}
		if g.CastSpellNamedThisTurnForEffect(opp.ID, named) {
			t.Error("a player who cast nothing is recorded as casting it")
		}
		if g.CastSpellNamedThisTurnForEffect(me.ID, other) {
			t.Error("a name nobody cast matched")
		}
		if cols, ok := g.LastSpellCastColorsThisTurnForEffect(); !ok || len(cols) != 1 || cols[0] != "R" {
			t.Errorf("last spell's colours = %v (%v), want [R]", cols, ok)
		}
	})
}

// Desolation's reader: tapping a land for mana marks its activator (CR
// 106.12); nobody else is marked.
func TestTappingALandForManaIsRecorded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var land uuid.UUID
	g.WithWriteLock(func() { land = effectstest.CcLand(g, me.ID, "Mountain", "Basic Land — Mountain") })
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Mountain: %v", err)
	}
	g.WithWriteLock(func() {
		if !g.TappedLandForManaThisTurnForEffect(me.ID) {
			t.Error("tapping a Mountain for mana was not recorded")
		}
		if g.TappedLandForManaThisTurnForEffect(opp.ID) {
			t.Error("a player who tapped nothing is recorded")
		}
	})
}

// Jandor's Ring's cost: only the last card drawn this turn, while it is in
// hand.
func TestDiscardTheLastCardDrawnThisTurn(t *testing.T) {
	const oracle = "test-6217-jandors-ring"
	effectstest.RegisterForTest(t, effects.Spec{OracleID: oracle, Name: oracle,
		Activated: []effects.ActivatedAbility{{
			Label:  "{T}, Discard the last card you drew this turn: Draw a card.",
			Cost:   effects.Plus(effects.TapCost(), effects.DiscardTheLastCardYouDrewThisTurn()),
			Effect: func(g *game.Game, item *game.StackItem) error { return g.DrawNForEffect(item.Controller, 1) },
		}},
	})
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ring := effectstest.PushCatalogPermanent(g, me.ID, "Ring", "Artifact", oracle, false)
	old := trHandSpell(g, me, "Kept", "{R}", "R")
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{old}}); err == nil {
		t.Fatal("a card not drawn this turn paid the cost")
	}
	var drawn uuid.UUID
	g.WithWriteLock(func() {
		_ = g.DrawNForEffect(me.ID, 1)
		top, _ := me.Hand.Top()
		drawn = top.InstanceID
	})
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{old}}); err == nil {
		t.Fatal("an older card paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{drawn}}); err != nil {
		t.Fatalf("the last card drawn: %v", err)
	}
	if !me.Graveyard.Contains(drawn) {
		t.Error("the last card drawn was not discarded")
	}
}

// Vodalian War Machine's reader: the Merfolk tapped to pay its abilities
// this turn, by the source object.
func TestTappedToPayASourceIsRecorded(t *testing.T) {
	const oracle = "test-6217-war-machine"
	effectstest.RegisterForTest(t, effects.Spec{OracleID: oracle, Name: oracle,
		Activated: []effects.ActivatedAbility{{
			Label:  "Tap an untapped Merfolk you control: nothing",
			Cost:   effects.TapAnUntapped("an untapped Merfolk you control", effects.HasSubtype("Merfolk")),
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	machine := effectstest.PushCatalogPermanent(g, me.ID, "Machine", "Creature — Wall", oracle, false)
	var fish, other uuid.UUID
	g.WithWriteLock(func() {
		fish = effectstest.CcLand(g, me.ID, "Fish", "Creature — Merfolk")
		other = effectstest.CcLand(g, me.ID, "Fish Two", "Creature — Merfolk")
	})
	if err := g.ActivateCatalogAbility(me.ID, machine, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{fish}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		ref, _ := g.PermanentRefForEffect(machine)
		got := g.TappedToPaySourceThisTurnForEffect(ref)
		if len(got) != 1 || got[0].ID != fish {
			t.Errorf("tapped to pay = %+v, want the one Merfolk", got)
		}
		if len(g.TappedToPaySourceThisTurnForEffect(game.ObjectRef{ID: other})) != 0 {
			t.Error("another source has a record")
		}
	})
}
