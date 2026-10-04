package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S58 PR 10, "Mill Me Mad (1)": the Lhurgoyfs and the sacrifice
// outlets.

const (
	mortivoreOracle         = "2f554218-2072-4d9b-b4d4-34f161f1f2a9"
	necrogoyfOracle         = "e034a02d-3482-44eb-90dc-c38be104197f"
	beastOfBurdenOracle     = "67389ccc-dacc-4ce3-a9a2-f551a0d6e9d3"
	urborgLhurgoyfOracle    = "f84d8217-51a4-49ba-9df6-d080ed38f37d"
	altarOfDementiaOracle   = "d64e9152-ef24-4394-aeb0-9c3befc56549"
	bushmeatPoacherOracle   = "0287d541-3f73-4c8b-9a77-87f99898758d"
	discipleGriselbrandOra  = "2d92a035-dd7a-4426-a8c0-f04e0b836dad"
	splinterfrightOracle    = "e54c6fa9-59ca-47dc-9354-91681228371e"
	apocalypseDemonOracle   = "5b8084b2-0c3c-4a54-86dc-41f4fe513e99"
	cruelSomnophageTestOrac = "997bdec5-f67b-4822-a3ba-c636e2685e8a"
)

func putCreatureCardsInGraveyard(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Graveyard.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear",
			Owner: p.ID, Controller: p.ID,
		})
	}
}

func TestMillMeMadLhurgoyfsCountTheirGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, other := g.Seats[0], g.Seats[1]
	put := func(name, typeLine, oracle string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
			Owner: me.ID, Controller: me.ID,
		})
	}
	mortivore := put("Mortivore", "Creature — Lhurgoyf", mortivoreOracle)
	necro := put("Necrogoyf", "Creature — Lhurgoyf", necrogoyfOracle)
	somno := put("Cruel Somnophage", "Creature — Nightmare", cruelSomnophageTestOrac)
	splinter := put("Splinterfright", "Creature — Elemental", splinterfrightOracle)
	urborg := put("Urborg Lhurgoyf", "Creature — Lhurgoyf", urborgLhurgoyfOracle)
	demon := put("Apocalypse Demon", "Creature — Demon", apocalypseDemonOracle)
	beast := put("Beast of Burden", "Artifact Creature — Golem", beastOfBurdenOracle)

	putCreatureCardsInGraveyard(me, 2)
	putCreatureCardsInGraveyard(other, 3)
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
			if got := effectivePower(t, g, tc.id); got != tc.p {
				t.Errorf("power = %d, want %d", got, tc.p)
			}
			if tc.t >= 0 {
				if got := effectiveToughness(t, g, tc.id); got != tc.t {
					t.Errorf("toughness = %d, want %d", got, tc.t)
				}
			}
		})
	}
}

func TestAltarOfDementiaMillsForTheSacrificedPower(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	altar := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Altar of Dementia", TypeLine: "Artifact",
		OracleID: altarOfDementiaOracle, Owner: me.ID, Controller: me.ID,
	})
	beast := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Beast", TypeLine: "Creature — Beast",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	before := victim.Library.Size()
	b16Activate(t, g, me.ID, altar, 0, game.ActivateAbilityParams{
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	poacher := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bushmeat Poacher", TypeLine: "Creature — Human Soldier",
		OracleID: bushmeatPoacherOracle, Power: 2, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	wall := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	life, hand := me.Life, me.Hand.Size()
	b16Activate(t, g, me.ID, poacher, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{wall}})
	if me.Life != life+5 {
		t.Errorf("life = %d, want %d", me.Life, life+5)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestDiscipleOfGriselbrandGainsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	disciple := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Disciple of Griselbrand", TypeLine: "Creature — Human Cleric",
		OracleID: discipleGriselbrandOra, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	wall := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	life := me.Life
	b16Activate(t, g, me.ID, disciple, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{wall}})
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
			g := newCatalogGame(t)
			active := g.Seats[g.Turn.ActiveSeat]
			if _, err := castWithOptionalCosts(t, g, "Urborg Lhurgoyf",
				"Creature — Lhurgoyf", urborgLhurgoyfOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			before := active.Library.Size()
			passPriorityAroundTable(t, g)
			if got := before - active.Library.Size(); got != tc.want {
				t.Errorf("milled %d, want %d", got, tc.want)
			}
		})
	}
}
