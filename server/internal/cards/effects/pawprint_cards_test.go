package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pawprint_cards_test.go — #4786, CR 700.2i: pawprint spells cast through
// CastSpell with Strict set, so the mana gate and the mode gate both
// judge the announcement, then resolved by passing priority.

const (
	helasArcanumOracle      = "4332a59b-f4b5-45c9-8a59-5baa59b46153"
	seasonOfLossOracle      = "31b5d0ad-3a4d-4f43-b0a9-c9f7a362680f"
	seasonOfTheBurrowOracle = "8e013e51-1684-4f58-8fe1-3588b6fbb5b5"
)

// castPawprintSpell puts the spell in the active player's hand, floats
// exactly its cost and casts it with `modes`.
func castPawprintSpell(t *testing.T, g *game.Game, name, typeLine, cost, mana, oracle string, modes []int, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	advanceToMain(t, g)
	floatMana(t, g, me, mana)
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets, Strict: true})
}

func castHelasArcanum(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) (uuid.UUID, error) {
	return castPawprintSpell(t, g, "Hela's Arcanum", "Instant — Arcane", "{3}{B}{B}", "{C}{C}{C}{B}{B}", helasArcanumOracle, modes, targets)
}

func modeTargetRef(id uuid.UUID, occurrence int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id, Mode: occurrence}
}

func pawBattlefieldCardNamed(g *game.Game, name string) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == name {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func TestHelasArcanumSweepsAndThePumpRunsFirstWhateverTheAnnounceOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := pushVanillaCreature(g, me.ID, "Bear", 3, 3)
	elf := pushVanillaCreature(g, opp.ID, "Elf", 2, 2)
	zombie := pushDiesCreatureForTest(g, opp.ID, "Walker", "", "Creature — Zombie", 2, 2)
	giant := pushVanillaCreature(g, opp.ID, "Giant", 4, 4)

	// The sweep announced FIRST and the pump second, its target filed
	// under occurrence 1: CR 608.2c still runs the pump first.
	if _, err := castHelasArcanum(t, g, []int{2, 0}, []game.TargetRef{modeTargetRef(bear, 1)}); err != nil {
		t.Fatalf("{P}{P}{P} + {P} (4 {P}) was refused: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear was pumped to power 4 before the sweep and should have survived it")
	}
	if g.Battlefield.Contains(elf) || !opp.Graveyard.Contains(elf) {
		t.Error("the 2/2 Elf should have been destroyed")
	}
	if !g.Battlefield.Contains(zombie) {
		t.Error("the sweep spares Zombies")
	}
	if !g.Battlefield.Contains(giant) {
		t.Error("the sweep spares power 4 or greater")
	}
	if !eotHasAbility(effectiveAbilities(t, g, bear), "deathtouch") {
		t.Error("the pumped creature should have deathtouch until end of turn")
	}
}

func TestHelasArcanumsGrantReturnsTheCreatureItsSweepKills(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	elf := pushVanillaCreature(g, opp.ID, "Elf", 2, 2)

	if _, err := castHelasArcanum(t, g, []int{1, 2}, []game.TargetRef{modeTargetRef(elf, 0)}); err != nil {
		t.Fatalf("{P}{P} + {P}{P}{P} (5 {P}) was refused: %v", err)
	}
	passPriorityAroundTable(t, g) // the spell
	passPriorityAroundTable(t, g) // the granted dies trigger

	back := pawBattlefieldCardNamed(g, "Elf")
	if back == nil {
		t.Fatal("the Elf died to the sweep with the granted ability and should have returned")
	}
	if !back.Tapped || back.Controller != opp.ID {
		t.Errorf("returned Elf: tapped=%v controller=%v, want tapped under its owner's control", back.Tapped, back.Controller)
	}
}

func TestHelasArcanumRefusesMoreThanFivePawprints(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVanillaCreature(g, me.ID, "Bear", 3, 3)

	for _, modes := range [][]int{{2, 2}, {2, 1, 0}, {0, 0, 0, 0, 0, 0}} {
		var targets []game.TargetRef
		for occ, m := range modes {
			if m != 2 {
				targets = append(targets, modeTargetRef(bear, occ))
			}
		}
		id, err := castHelasArcanum(t, g, modes, targets)
		if err == nil {
			t.Fatalf("modes %v spend more than five {P} and were accepted", modes)
		}
		if !me.Hand.Contains(id) {
			t.Errorf("modes %v: a refused cast left the hand", modes)
		}
		me.Hand.Remove(id)
		me.ManaPool = nil
	}
}

func TestSeasonOfLossCountsTheCreaturesItsOwnEdictKilled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	elf := pushVanillaCreature(g, opp.ID, "Elf", 2, 2)

	_, err := castPawprintSpell(t, g, "Season of Loss", "Sorcery", "{3}{B}{B}", "{C}{C}{C}{B}{B}", seasonOfLossOracle, []int{0, 1}, nil)
	if err != nil {
		t.Fatalf("{P} + {P}{P} was refused: %v", err)
	}
	handBefore := me.Hand.Size()
	passPriorityAroundTable(t, g)

	// Each seat with one creature sacrifices it unasked (the pick is
	// forced), and the draw mode then sees the Bear among the dead.
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(elf) {
		t.Fatal("each player should have sacrificed their only creature")
	}
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("drew %d cards, want 1 (one creature died under your control)", got)
	}
}

func TestSeasonOfLossDrainsByTheCreatureCardsInYourGraveyardAfterTheEdict(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	graveyardCardOf(g, me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	graveyardCardOf(g, me, "Old Elf", "Creature — Elf", "{G}", "")
	graveyardCardOf(g, me, "Old Bolt", "Instant", "{R}", "")
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}

	// Announced drain first: printed order still sacrifices first, so the
	// Bear is the third creature card in the graveyard when X is counted.
	if _, err := castPawprintSpell(t, g, "Season of Loss", "Sorcery", "{3}{B}{B}", "{C}{C}{C}{B}{B}", seasonOfLossOracle, []int{2, 0}, nil); err != nil {
		t.Fatalf("{P}{P}{P} + {P} was refused: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		want := before[p.ID] - 3
		if p.ID == me.ID {
			want = before[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %v life = %d, want %d", p.ID, p.Life, want)
		}
	}
}

func TestSeasonOfTheBurrowExilesAndItsControllerDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	rock := pushArtifactForTest(g, opp.ID, "Mind Stone")
	oppHand := opp.Hand.Size()

	if _, err := castPawprintSpell(t, g, "Season of the Burrow", "Sorcery", "{3}{W}{W}", "{C}{C}{C}{W}{W}", seasonOfTheBurrowOracle,
		[]int{1, 0}, []game.TargetRef{modeTargetRef(rock, 0)}); err != nil {
		t.Fatalf("{P}{P} + {P} was refused: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact should have been exiled")
	}
	if got := opp.Hand.Size() - oppHand; got != 1 {
		t.Errorf("the exiled permanent's controller drew %d, want 1", got)
	}
	if n := countTokensControlled(g, me.ID, "Rabbit"); n != 1 {
		t.Errorf("Rabbits = %d, want 1", n)
	}
}

func TestSeasonOfTheBurrowReturnsAPermanentCardWithAnIndestructibleCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// An artifact, so the returned card has no 0 toughness to die of.
	cheap := graveyardCardOf(g, me, "Old Stone", "Artifact", "{2}", "")
	dear := graveyardCardOf(g, me, "Old Giant", "Creature — Giant", "{3}{R}{R}", "")

	if _, err := castPawprintSpell(t, g, "Season of the Burrow", "Sorcery", "{3}{W}{W}", "{C}{C}{C}{W}{W}", seasonOfTheBurrowOracle,
		[]int{0, 0, 2}, []game.TargetRef{modeTargetRef(dear, 2)}); err == nil {
		t.Fatal("a mana value 5 card is not a legal target for the {P}{P}{P} mode")
	}
	me.ManaPool = nil
	for _, c := range me.Hand.Cards {
		if c.Name == "Season of the Burrow" {
			me.Hand.Remove(c.InstanceID)
			break
		}
	}

	if _, err := castPawprintSpell(t, g, "Season of the Burrow", "Sorcery", "{3}{W}{W}", "{C}{C}{C}{W}{W}", seasonOfTheBurrowOracle,
		[]int{0, 0, 2}, []game.TargetRef{modeTargetRef(cheap, 2)}); err != nil {
		t.Fatalf("{P} + {P} + {P}{P}{P} was refused: %v", err)
	}
	passPriorityAroundTable(t, g)

	back := pawBattlefieldCardNamed(g, "Old Stone")
	if back == nil {
		t.Fatal("the mana value 2 card should have returned to the battlefield")
	}
	if back.Counters[game.CounterIndestructible] != 1 {
		t.Errorf("counters = %v, want one indestructible counter", back.Counters)
	}
	if n := countTokensControlled(g, me.ID, "Rabbit"); n != 2 {
		t.Errorf("Rabbits = %d, want 2 (the {P} mode chosen twice)", n)
	}
}
