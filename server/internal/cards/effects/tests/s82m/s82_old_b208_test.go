package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b208_test.go — Dominaria kicker cards, batch 208 (#5530): each test casts
// the real catalog entry, kicked or not, through the real cast and resolution paths.

const (
	oldB208BarrageOracle  = "0071723e-388d-4ae7-bc99-c143dd356c63"
	oldB208SkizzikOracle  = "0c77d37e-791c-415e-b369-55e93147ec34"
	oldB208OverseerOracle = "1b1dc78b-0102-49d6-8e54-89bc18a1628f"
	oldB208FightOracle    = "202b275e-e0c4-482d-abe9-c5e6360208f3"
	oldB208OnslaughtOrcl  = "7ca9ef3e-e5b2-4eb5-87b3-0d97637428a0"
	oldB208AllosaurOracle = "7fe39c74-cb7c-4eee-a657-8a99a4ab383a"
	oldB208GrunnOracle    = "e3d6d146-47f4-46f9-9a15-2ea649093668"
	oldB208GiftOracle     = "edbe4fdf-9e55-44fa-b3ec-5b1ac615f3a7"
)

func TestOldB208CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB208BarrageOracle:  "Goblin Barrage",
		oldB208SkizzikOracle:  "Skizzik",
		oldB208OverseerOracle: "Keldon Overseer",
		oldB208FightOracle:    "Fight with Fire",
		oldB208OnslaughtOrcl:  "Wild Onslaught",
		oldB208AllosaurOracle: "Territorial Allosaurus",
		oldB208GrunnOracle:    "Grunn, the Lonely King",
		oldB208GiftOracle:     "Gift of Growth",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB208Card(g *game.Game, who *game.Player, c game.Card) uuid.UUID {
	id := uuid.New()
	c.InstanceID, c.Owner, c.Controller = id, who.ID, who.ID
	who.Hand.PushTop(c)
	return id
}

// oldB208Cast casts the card from the active seat's hand in a main phase, with the
// given optional costs claimed and permanents sacrificed to them.
func oldB208Cast(t *testing.T, g *game.Game, c game.Card, targets []game.TargetRef, optional []int, sacrifices []uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := oldB208Card(g, g.Seats[g.Turn.ActiveSeat], c)
	return id, g.CastSpell(g.Seats[g.Turn.ActiveSeat].ID, id, game.CastSpellParams{
		Targets: targets, OptionalCosts: optional, SacrificeIDs: sacrifices, Strict: false,
	})
}

func oldB208CardTarget(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}
func oldB208PlayerTarget(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id}
}

// --- Goblin Barrage ---

func TestOldB208GoblinBarrageDamagesOnlyTheCreatureWhenUnkicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 4)
	life := opp.Life
	if _, err := oldB208Cast(t, g, game.Card{Name: "Goblin Barrage", TypeLine: "Sorcery", OracleID: oldB208BarrageOracle, ManaCost: "{3}{R}"},
		[]game.TargetRef{oldB208CardTarget(victim)}, nil, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature survived 4 damage")
	}
	if opp.Life != life {
		t.Errorf("a player was damaged by the unkicked spell")
	}
}

func TestOldB208GoblinBarrageKickedAlsoHitsAPlayerAndNeedsAGoblinOrArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 4)
	goblin := effectstest.B12Creature(g, me.ID, "Fodder Goblin", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	card := game.Card{Name: "Goblin Barrage", TypeLine: "Sorcery", OracleID: oldB208BarrageOracle, ManaCost: "{3}{R}"}
	both := []game.TargetRef{oldB208CardTarget(victim), oldB208PlayerTarget(opp.ID)}
	if _, err := oldB208Cast(t, g, card, both, []int{0}, []uuid.UUID{bear}); err == nil {
		t.Error("a Bear was accepted as the sacrifice for an artifact or Goblin")
	}
	if _, err := oldB208Cast(t, g, card, []game.TargetRef{oldB208CardTarget(victim)}, []int{0}, []uuid.UUID{goblin}); err == nil {
		t.Error("a kicked cast was accepted without its second target")
	}
	life := opp.Life
	if _, err := oldB208Cast(t, g, card, both, []int{0}, []uuid.UUID{goblin}); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(goblin) {
		t.Error("the Goblin was not sacrificed")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the creature survived 4 damage")
	}
	if opp.Life != life-4 {
		t.Errorf("the player took %d, want 4", life-opp.Life)
	}
}

// --- Skizzik ---

func TestOldB208SkizzikIsSacrificedAtTheEndStepUnlessKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		var optional []int
		if kicked {
			optional = []int{0}
		}
		id, err := oldB208Cast(t, g, game.Card{Name: "Skizzik", TypeLine: "Creature — Elemental", OracleID: oldB208SkizzikOracle, ManaCost: "{3}{R}", Power: 5, Toughness: 3},
			nil, optional, nil)
		if err != nil {
			t.Fatalf("cast (kicked=%v): %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(id) {
			t.Fatalf("kicked=%v: Skizzik did not enter", kicked)
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(id); got != kicked {
			t.Errorf("kicked=%v: Skizzik on the battlefield after the end step = %v", kicked, got)
		}
	}
}

func TestOldB208SkizzikThatWasNeverCastIsNotKickedAndDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Skizzik", "Creature — Elemental", oldB208SkizzikOracle, 5, 3)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("a Skizzik put onto the battlefield was treated as kicked")
	}
}

// --- Keldon Overseer ---

func TestOldB208KeldonOverseerKickedStealsUntapsAndHastesUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if _, err := oldB208Cast(t, g, game.Card{Name: "Keldon Overseer", TypeLine: "Creature — Human Warrior", OracleID: oldB208OverseerOracle, ManaCost: "{2}{R}", Power: 3, Toughness: 1},
		nil, []int{0}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, theirs)
	if !ok || c.Controller != me.ID {
		t.Fatal("the creature is not under my control")
	}
	if c.Tapped {
		t.Error("the creature is still tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("the creature lacks haste")
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	effectstest.AdvanceToMainOf(t, g, 1)
	if c, ok := effectstest.BattlefieldCard(g, theirs); !ok || c.Controller != opp.ID {
		t.Error("the creature did not go back at end of turn")
	}
}

func TestOldB208KeldonOverseerUnkickedStealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if _, err := oldB208Cast(t, g, game.Card{Name: "Keldon Overseer", TypeLine: "Creature — Human Warrior", OracleID: oldB208OverseerOracle, ManaCost: "{2}{R}", Power: 3, Toughness: 1},
		nil, nil, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, theirs); !ok || c.Controller != opp.ID {
		t.Error("an unkicked Overseer took a creature")
	}
	if effectstest.LatestPickTarget(g, g.Seats[0].ID) != nil {
		t.Error("an unkicked Overseer asked for a target")
	}
}

// --- Fight with Fire ---

func TestOldB208FightWithFireUnkickedDealsFiveToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Wall", 0, 5)
	if _, err := oldB208Cast(t, g, game.Card{Name: "Fight with Fire", TypeLine: "Sorcery", OracleID: oldB208FightOracle, ManaCost: "{2}{R}"},
		[]game.TargetRef{oldB208CardTarget(big)}, nil, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the creature survived 5 damage")
	}
}

func TestOldB208FightWithFireKickedDividesTenAmongAnyTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 4)
	b := effectstest.B12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 3)
	life := opp.Life
	id := uuid.New()
	g.Seats[0].Hand.PushTop(game.Card{InstanceID: id, Name: "Fight with Fire", TypeLine: "Sorcery", OracleID: oldB208FightOracle,
		ManaCost: "{2}{R}", Owner: g.Seats[0].ID, Controller: g.Seats[0].ID})
	effectstest.AdvanceToMain(t, g)
	err := g.CastSpell(g.Seats[0].ID, id, game.CastSpellParams{
		Targets:       []game.TargetRef{oldB208CardTarget(a), oldB208CardTarget(b), oldB208PlayerTarget(opp.ID)},
		OptionalCosts: []int{0},
		Distribution:  map[uuid.UUID]int{a: 4, b: 3, opp.ID: 3},
	})
	if err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a creature survived the damage assigned to it")
	}
	if opp.Life != life-3 {
		t.Errorf("the player took %d, want 3", life-opp.Life)
	}
}

// --- Wild Onslaught ---

func TestOldB208WildOnslaughtPutsOneOrTwoCountersOnEachCreature(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		var optional []int
		want := 1
		if kicked {
			optional, want = []int{0}, 2
		}
		if _, err := oldB208Cast(t, g, game.Card{Name: "Wild Onslaught", TypeLine: "Instant", OracleID: oldB208OnslaughtOrcl, ManaCost: "{3}{G}"},
			nil, optional, nil); err != nil {
			t.Fatalf("cast (kicked=%v): %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.PlusOnes(g, mine); got != want {
			t.Errorf("kicked=%v: %d counters on my creature, want %d", kicked, got, want)
		}
		if got := effectstest.PlusOnes(g, theirs); got != 0 {
			t.Errorf("kicked=%v: an opponent's creature got %d counters", kicked, got)
		}
	}
}

// --- Territorial Allosaurus ---

func TestOldB208TerritorialAllosaurusKickedFightsAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	prey := effectstest.B12Creature(g, opp.ID, "Prey", "Creature — Bear", 2, 3)
	id, err := oldB208Cast(t, g, game.Card{Name: "Territorial Allosaurus", TypeLine: "Creature — Dinosaur", OracleID: oldB208AllosaurOracle, ManaCost: "{2}{G}{G}", Power: 5, Toughness: 5},
		nil, []int{0}, nil)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.AnswerPickTarget(t, g, prey)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(prey) {
		t.Error("the creature survived the fight")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the Allosaurus died fighting a 2/3")
	}
}

func TestOldB208TerritorialAllosaurusUnkickedDoesNotFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	prey := effectstest.B12Creature(g, g.Seats[1].ID, "Prey", "Creature — Bear", 2, 3)
	if _, err := oldB208Cast(t, g, game.Card{Name: "Territorial Allosaurus", TypeLine: "Creature — Dinosaur", OracleID: oldB208AllosaurOracle, ManaCost: "{2}{G}{G}", Power: 5, Toughness: 5},
		nil, nil, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(prey) {
		t.Error("an unkicked Allosaurus fought")
	}
}

// --- Grunn, the Lonely King ---

func TestOldB208GrunnEntersWithFiveCountersOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		var optional []int
		want := 0
		if kicked {
			optional, want = []int{0}, 5
		}
		id, err := oldB208Cast(t, g, game.Card{Name: "Grunn, the Lonely King", TypeLine: "Legendary Creature — Ape Warrior", OracleID: oldB208GrunnOracle, ManaCost: "{4}{G}{G}", Power: 5, Toughness: 5},
			nil, optional, nil)
		if err != nil {
			t.Fatalf("cast (kicked=%v): %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.PlusOnes(g, id); got != want {
			t.Errorf("kicked=%v: %d counters, want %d", kicked, got, want)
		}
	}
}

func TestOldB208GrunnDoublesOnlyWhenItAttacksAlone(t *testing.T) {
	for _, alone := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		grunn := effectstest.B12Push(g, me.ID, "Grunn, the Lonely King", "Legendary Creature — Ape Warrior", oldB208GrunnOracle, 5, 5)
		ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		attackers := []uuid.UUID{grunn}
		if !alone {
			attackers = append(attackers, ally)
		}
		for _, id := range attackers {
			if err := g.DeclareAttacker(id, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
		}
		effectstest.LockInAttacks(t, g)
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
		p, tt := effectstest.PtOf(t, g, grunn)
		if alone && (p != 10 || tt != 10) {
			t.Errorf("alone: Grunn is %d/%d, want 10/10", p, tt)
		}
		if !alone && (p != 5 || tt != 5) {
			t.Errorf("with another attacker: Grunn is %d/%d, want 5/5", p, tt)
		}
	}
}

// --- Gift of Growth ---

func TestOldB208GiftOfGrowthUntapsAndPumpsTwoOrFour(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == bear {
					g.Battlefield.Cards[i].Tapped = true
				}
			}
		})
		var optional []int
		want := 4
		if kicked {
			optional, want = []int{0}, 6
		}
		if _, err := oldB208Cast(t, g, game.Card{Name: "Gift of Growth", TypeLine: "Instant", OracleID: oldB208GiftOracle, ManaCost: "{1}{G}"},
			[]game.TargetRef{oldB208CardTarget(bear)}, optional, nil); err != nil {
			t.Fatalf("cast (kicked=%v): %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AssertTapped(t, g, false, bear)
		if p, tt := effectstest.PtOf(t, g, bear); p != want || tt != want {
			t.Errorf("kicked=%v: the creature is %d/%d, want %d/%d", kicked, p, tt, want, want)
		}
	}
}
