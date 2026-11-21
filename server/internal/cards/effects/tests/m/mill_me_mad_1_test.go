package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMillMeMadLhurgoyfsCountTheirGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, other := g.Seats[0], g.Seats[1]
	put := func(name, typeLine, oracle string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
			Owner: me.ID, Controller: me.ID,
		})
	}
	mortivore := put("Mortivore", "Creature — Lhurgoyf", effectstest.MortivoreOracle)
	necro := put("Necrogoyf", "Creature — Lhurgoyf", effectstest.NecrogoyfOracle)
	somno := put("Cruel Somnophage", "Creature — Nightmare", effectstest.CruelSomnophageTestOrac)
	splinter := put("Splinterfright", "Creature — Elemental", effectstest.SplinterfrightOracle)
	urborg := put("Urborg Lhurgoyf", "Creature — Lhurgoyf", effectstest.UrborgLhurgoyfOracle)
	demon := put("Apocalypse Demon", "Creature — Demon", effectstest.ApocalypseDemonOracle)
	beast := put("Beast of Burden", "Artifact Creature — Golem", effectstest.BeastOfBurdenOracle)

	effectstest.PutCreatureCardsInGraveyard(me, 2)
	effectstest.PutCreatureCardsInGraveyard(other, 3)
	other.Graveyard.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island",
		Owner: other.ID, Controller: other.ID,
	})
	// Tokens never reach a graveyard as cards, and a land is not a
	// creature card: the all-graveyards count is 5, mine is 2, and my
	// whole graveyard is 2.

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		p, t int
	}{
		{"Mortivore counts every creature card", mortivore, 5, 5},
		{"Cruel Somnophage counts every creature card", somno, 5, 5},
		{"Necrogoyf's power counts every creature card, toughness stays 0 here", necro, 5, -1},
		{"Splinterfright counts only my creature cards", splinter, 2, 2},
		{"Urborg Lhurgoyf is n and n+1", urborg, 2, 3},
		{"Apocalypse Demon counts every card in my graveyard", demon, 2, 2},
		{"Beast of Burden counts every creature on the battlefield", beast, 7, 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := effectstest.EffectivePower(t, g, tc.id); got != tc.p {
				t.Errorf("power = %d, want %d", got, tc.p)
			}
			if tc.t >= 0 {
				if got := effectstest.EffectiveToughness(t, g, tc.id); got != tc.t {
					t.Errorf("toughness = %d, want %d", got, tc.t)
				}
			}
		})
	}
}

func TestAltarOfDementiaMillsForTheSacrificedPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	altar := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Altar of Dementia", TypeLine: "Artifact",
		OracleID: effectstest.AltarOfDementiaOracle, Owner: me.ID, Controller: me.ID,
	})
	beast := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Beast", TypeLine: "Creature — Beast",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	before := victim.Library.Size()
	effectstest.B16Activate(t, g, me.ID, altar, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{beast},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
	})
	if got := before - victim.Library.Size(); got != 3 {
		t.Errorf("milled %d cards, want 3", got)
	}
	if !me.Graveyard.Contains(beast) {
		t.Error("the sacrificed creature is in the graveyard")
	}
}

func TestBushmeatPoacherGainsToughnessAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	poacher := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bushmeat Poacher", TypeLine: "Creature — Human Soldier",
		OracleID: effectstest.BushmeatPoacherOracle, Power: 2, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	wall := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	life, hand := me.Life, me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, poacher, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{wall}})
	if me.Life != life+5 {
		t.Errorf("life = %d, want %d", me.Life, life+5)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestDiscipleOfGriselbrandGainsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	disciple := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Disciple of Griselbrand", TypeLine: "Creature — Human Cleric",
		OracleID: effectstest.DiscipleGriselbrandOra, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	wall := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, disciple, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{wall}})
	if me.Life != life+6 {
		t.Errorf("life = %d, want %d", me.Life, life+6)
	}
}

func TestUrborgLhurgoyfMillsThreePerKick(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{
		{"unkicked", nil, 0},
		{"kicked with {U}", []int{0}, 3},
		{"kicked with {B}", []int{1}, 3},
		{"kicked with both", []int{0, 1}, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			active := g.Seats[g.Turn.ActiveSeat]
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Urborg Lhurgoyf",
				"Creature — Lhurgoyf", effectstest.UrborgLhurgoyfOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			before := active.Library.Size()
			effectstest.PassPriorityAroundTable(t, g)
			if got := before - active.Library.Size(); got != tc.want {
				t.Errorf("milled %d, want %d", got, tc.want)
			}
		})
	}
}
