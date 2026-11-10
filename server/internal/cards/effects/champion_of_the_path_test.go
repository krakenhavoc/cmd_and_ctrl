package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const championOfThePathOracle = "b326ccd5-8ad2-441c-9d73-82ee6390dd36"

func TestChampionOfThePathIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(championOfThePathOracle)
	if !ok || spec.Name != "Champion of the Path" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// castChampionOfThePath casts it strictly for {3}{R}, beholding `reveal`.
func castChampionOfThePath(t *testing.T, g *game.Game, reveal ...uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	b06AddMana(me, "R", "C", "C", "C")
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Champion of the Path", TypeLine: "Creature — Elemental Sorcerer",
		ManaCost: "{3}{R}", OracleID: championOfThePathOracle, Power: 7, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, RevealIDs: reveal})
}

func championOfThePathOpponentLives(g *game.Game, me uuid.UUID) []int {
	var out []int
	for _, p := range g.Seats {
		if p.ID != me {
			out = append(out, p.Life)
		}
	}
	return out
}

// Beholding an Elemental you control exiles it. The Champion's own entry
// deals nothing; another Elemental entering deals its power to each
// opponent, a Goblin entering deals nothing. When the Champion leaves,
// the exiled Elemental returns to its owner's hand.
func TestChampionOfThePathElementalsDealTheirPowerToEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	beheld := pushVanillaCreature(g, me.ID, "Old Elemental", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == beheld {
				g.Battlefield.Cards[i].TypeLine = "Creature — Elemental"
			}
		}
		g.BumpLayerVersionForTest()
	})

	start := championOfThePathOpponentLives(g, me.ID)
	champ, err := castChampionOfThePath(t, g, beheld)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !inExile(g, beheld) {
		t.Fatal("the beheld Elemental was not exiled")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(champ) {
		t.Fatal("the Champion did not resolve")
	}
	for i, l := range championOfThePathOpponentLives(g, me.ID) {
		if l != start[i] {
			t.Fatalf("the Champion's own entry changed an opponent's life to %d", l)
		}
	}

	elemental := rfdCast(t, g, "Flamekin Card", "Creature — Elemental Warrior", "", 3, 1)
	for i, l := range championOfThePathOpponentLives(g, me.ID) {
		if l != start[i]-3 {
			t.Errorf("opponent %d life %d after a 3-power Elemental entered, want %d", i, l, start[i]-3)
		}
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	if ev, ok := lastDamageTo(g, opp); !ok || ev.Source != elemental {
		t.Errorf("damage event %+v, want one from the entering Elemental", ev)
	}

	rfdCast(t, g, "Goblin Card", "Creature — Goblin", "", 4, 4)
	for i, l := range championOfThePathOpponentLives(g, me.ID) {
		if l != start[i]-3 {
			t.Errorf("a Goblin entering changed opponent %d's life to %d", i, l)
		}
	}

	b27Kill(g, champ)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(beheld) {
		t.Error("the exiled Elemental did not return to its owner's hand")
	}
}

// The behold is mandatory and takes an Elemental only.
func TestChampionOfThePathRefusesANonElemental(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	goblin := handCard(me, "Goblin Card", "Creature — Goblin")
	if _, err := castChampionOfThePath(t, g); err == nil {
		t.Error("cast without beholding an Elemental")
	}
	if _, err := castChampionOfThePath(t, g, goblin); err == nil {
		t.Error("beheld a Goblin")
	}
}
