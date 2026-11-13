package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sth_b1_cards_test.go — Stronghold, batch 1 (#5627). Each test plays the
// real catalog entry and asserts what its text says. The seeding helpers
// are Tempest batch 1's (tmp_b1_cards_test.go).

const (
	sthB1HonorGuardOracle  = "4ba82d8e-1857-40a9-8b42-5d8348c68859"
	sthB1ShadeOracle       = "953a1a07-1bf2-480e-81a3-f2d5f19f6df5"
	sthB1FurnaceOracle     = "86b33d25-6123-439c-9026-a4e835b48562"
	sthB1ShamblerOracle    = "e72ae9ab-7405-4ea7-b075-42d0ddc352db"
	sthB1MaulerOracle      = "ea205569-b5a0-4c86-b825-67cd48d3b2e8"
	sthB1HellionOracle     = "69ea4ef2-794d-4c76-904f-0435b0efd6a0"
	sthB1TroopersOracle    = "b5a40cd8-e1f9-4294-99ef-26fb455d4f95"
	sthB1FoulImpOracle     = "25c04be3-f2e0-4f45-a1ca-9e75ffffa9b5"
	sthB1DeathStrokeOracle = "3ebaa91f-5cbf-4a82-9759-9bd4d93a3e87"
	sthB1SmiteOracle       = "367379f2-e1e4-48e7-b2c0-1003de55adbe"
	sthB1SeethingOracle    = "2effcd1c-5833-4e0e-bb67-2d2ffb22a014"
	sthB1BrushOracle       = "d87e580f-eca1-4f91-962b-82d684cba5e7"
	sthB1MindPeelOracle    = "4f424d94-1544-4086-a57d-2add17332094"
	sthB1MindGamesOracle   = "ea43372f-b0a7-4d1b-be3e-a4a738b06d59"
)

func TestSthB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sthB1HonorGuardOracle:  "Honor Guard",
		sthB1ShadeOracle:       "Dungeon Shade",
		sthB1FurnaceOracle:     "Furnace Spirit",
		sthB1ShamblerOracle:    "Flowstone Shambler",
		sthB1MaulerOracle:      "Flowstone Mauler",
		sthB1HellionOracle:     "Flowstone Hellion",
		sthB1TroopersOracle:    "Skyshroud Troopers",
		sthB1FoulImpOracle:     "Foul Imp",
		sthB1DeathStrokeOracle: "Death Stroke",
		sthB1SmiteOracle:       "Smite",
		sthB1SeethingOracle:    "Seething Anger",
		sthB1BrushOracle:       "Brush with Death",
		sthB1MindPeelOracle:    "Mind Peel",
		sthB1MindGamesOracle:   "Mind Games",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- keywords -----------------------------------------------------------------------

func TestSthB1PrintedKeywords(t *testing.T) {
	for _, tc := range []struct{ name, oracle, typeLine, keyword string }{
		{"Dungeon Shade", sthB1ShadeOracle, "Creature — Shade Spirit", "flying"},
		{"Furnace Spirit", sthB1FurnaceOracle, "Creature — Spirit", "haste"},
		{"Flowstone Mauler", sthB1MaulerOracle, "Creature — Beast", "trample"},
		{"Flowstone Hellion", sthB1HellionOracle, "Creature — Hellion Beast", "haste"},
		{"Foul Imp", sthB1FoulImpOracle, "Creature — Imp", "flying"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := tmpB1Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s does not have %s", tc.name, tc.keyword)
			}
		})
	}
}

// --- pumps ------------------------------------------------------------------------

func TestSthB1PumpsChangeTheCreatureUntilEndOfTurn(t *testing.T) {
	for _, tc := range []struct {
		// pay funds one activation; "" is the {0} cost. The Shambler is
		// seeded 2/3 so it survives two; its 2/2 death is the next test.
		name, oracle, typeLine, pay string
		power, toughness, dp, dt    int
	}{
		{"Honor Guard", sthB1HonorGuardOracle, "Creature — Human Soldier", "{W}", 1, 1, 0, 1},
		{"Dungeon Shade", sthB1ShadeOracle, "Creature — Shade Spirit", "{B}", 1, 1, 1, 1},
		{"Furnace Spirit", sthB1FurnaceOracle, "Creature — Spirit", "{R}", 1, 1, 1, 0},
		{"Flowstone Shambler", sthB1ShamblerOracle, "Creature — Beast", "{R}", 2, 3, 1, -1},
		{"Flowstone Mauler", sthB1MaulerOracle, "Creature — Beast", "{R}", 4, 5, 1, -1},
		{"Flowstone Hellion", sthB1HellionOracle, "Creature — Hellion Beast", "", 3, 3, 1, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := tmpB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			for i := 1; i <= 2; i++ {
				if tc.pay != "" {
					dcMana(t, g, me.ID, tc.pay)
				}
				if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
					t.Fatalf("activation %d: %v", i, err)
				}
				passPriorityAroundTable(t, g)
				wantP, wantT := tc.power+i*tc.dp, tc.toughness+i*tc.dt
				if p, tough := ptOf(t, g, id); p != wantP || tough != wantT {
					t.Fatalf("%s is %d/%d after %d activations, want %d/%d", tc.name, p, tough, i, wantP, wantT)
				}
			}
			if tc.pay == "" {
				return
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Error("the pump was activated with an empty mana pool")
			}
		})
	}
}

func TestSthB1FlowstoneShamblerDiesAtZeroToughness(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := tmpB1Push(g, me.ID, "Flowstone Shambler", "Creature — Beast", sthB1ShamblerOracle, 2, 2)
	for i := 1; i <= 2; i++ {
		dcMana(t, g, me.ID, "{R}")
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(id) {
		t.Error("Flowstone Shambler survived at 4/0")
	}
}

// --- Skyshroud Troopers -------------------------------------------------------------

func TestSthB1SkyshroudTroopersTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := tmpB1Push(g, me.ID, "Skyshroud Troopers", "Creature — Elf Druid Warrior", sthB1TroopersOracle, 3, 3, "G")
	before := me.ManaPool.CountColor("G")
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.CountColor("G") - before; got != 1 {
		t.Errorf("the pool gained %d {G}, want 1", got)
	}
	if !isTapped(g, id) {
		t.Error("Skyshroud Troopers did not tap")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the tapped Troopers made mana again")
	}
}

// --- Foul Imp ------------------------------------------------------------------------

func TestSthB1FoulImpCostsTwoLifeAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castCatalogSpell(t, g, "Foul Imp", "Creature — Imp", sthB1FoulImpOracle, nil)
	passPriorityAroundTable(t, g)
	if got := life - me.Life; got != 2 {
		t.Errorf("the caster lost %d life, want 2", got)
	}
}

// --- Death Stroke and Smite ----------------------------------------------------------

func TestSthB1DeathStrokeDestroysOnlyATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	untapped := b12Creature(g, opp.ID, "Awake", "Creature — Bear", 2, 2)
	tapped := b12Creature(g, opp.ID, "Asleep", "Creature — Bear", 2, 2)
	tapCard(g, tapped)
	if err := castCatalogSpellErr(t, g, "Death Stroke", "Sorcery", sthB1DeathStrokeOracle, b16TargetCard(untapped)); err == nil {
		t.Fatal("an untapped creature was accepted")
	}
	castCatalogSpell(t, g, "Death Stroke", "Sorcery", sthB1DeathStrokeOracle, b16TargetCard(tapped))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature was destroyed")
	}
}

func TestSthB1SmiteDestroysOnlyABlockedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	blocked := b12Creature(g, me.ID, "Blocked Bear", "Creature — Bear", 2, 2)
	unblocked := b12Creature(g, me.ID, "Free Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, blocked, unblocked)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, blocked); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	smite := game.Card{Name: "Smite", TypeLine: "Instant", OracleID: sthB1SmiteOracle}
	tmcB2Pool(t, g, me, "{W}")
	if _, err := tmcB2Cast(g, me, smite, game.CastSpellParams{Targets: cardRefs(unblocked)}); err == nil {
		t.Fatal("an unblocked attacker was accepted")
	}
	if _, err := tmcB2Cast(g, me, smite, game.CastSpellParams{Targets: cardRefs(wall)}); err == nil {
		t.Fatal("the blocking creature was accepted")
	}
	if _, err := tmcB2Cast(g, me, smite, game.CastSpellParams{Targets: cardRefs(blocked)}); err != nil {
		t.Fatalf("cast on the blocked attacker: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocked) {
		t.Error("the blocked creature survived")
	}
	if !g.Battlefield.Contains(unblocked) || !g.Battlefield.Contains(wall) {
		t.Error("a creature other than the target was destroyed")
	}
}

// --- buyback --------------------------------------------------------------------------

func TestSthB1SeethingAngerGivesPlusThreePlusZero(t *testing.T) {
	for _, tc := range tmpB1BuybackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			bear := seedCreature(g, "Grizzly Bears", caster.ID)
			id, err := castWithOptionalCosts(t, g, "Seething Anger", "Sorcery", sthB1SeethingOracle,
				cardTarget(bear), tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, bear); p != 5 || tough != 2 {
				t.Errorf("the target is %d/%d, want 5/2", p, tough)
			}
			tmpB1AssertBuyback(t, caster, id, tc.wantHand)
		})
	}
}

func TestSthB1BrushWithDeathDrainsTwo(t *testing.T) {
	for _, tc := range tmpB1BuybackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster, victim := seatsForCombat(g)
			mine, theirs := caster.Life, victim.Life
			id, err := castWithOptionalCosts(t, g, "Brush with Death", "Sorcery", sthB1BrushOracle,
				[]game.TargetRef{playerTarget(victim.ID)}, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			if theirs-victim.Life != 2 || caster.Life-mine != 2 {
				t.Errorf("life %d → %d (target) and %d → %d (caster), want -2 and +2", theirs, victim.Life, mine, caster.Life)
			}
			tmpB1AssertBuyback(t, caster, id, tc.wantHand)
		})
	}
}

func TestSthB1BrushWithDeathTargetsOnlyAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	if err := castCatalogSpellErr(t, g, "Brush with Death", "Sorcery", sthB1BrushOracle,
		[]game.TargetRef{playerTarget(caster.ID)}); err == nil {
		t.Error("the caster was accepted as the target opponent")
	}
}

func TestSthB1MindPeelMakesTheTargetDiscard(t *testing.T) {
	for _, tc := range tmpB1BuybackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster, victim := seatsForCombat(g)
			victim.Hand.Cards = nil
			card := uuid.New()
			victim.Hand.PushTop(game.Card{InstanceID: card, Name: "Victim", TypeLine: "Instant", Owner: victim.ID, Controller: victim.ID})
			victim.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Kept", TypeLine: "Instant", Owner: victim.ID, Controller: victim.ID})
			id, err := castWithOptionalCosts(t, g, "Mind Peel", "Sorcery", sthB1MindPeelOracle,
				[]game.TargetRef{playerTarget(victim.ID)}, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			answerDiscard(t, g, victim.ID, card)
			if !victim.Graveyard.Contains(card) || victim.Hand.Size() != 1 {
				t.Errorf("hand %d, chosen card in graveyard %v; want one card discarded, the chosen one",
					victim.Hand.Size(), victim.Graveyard.Contains(card))
			}
			tmpB1AssertBuyback(t, caster, id, tc.wantHand)
		})
	}
}

func TestSthB1MindGamesTapsTheTarget(t *testing.T) {
	for _, tc := range tmpB1BuybackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster, opp := seatsForCombat(g)
			bear := seedCreature(g, "Grizzly Bears", opp.ID)
			id, err := castWithOptionalCosts(t, g, "Mind Games", "Instant", sthB1MindGamesOracle,
				cardTarget(bear), tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			if !isTapped(g, bear) {
				t.Error("the target is untapped")
			}
			tmpB1AssertBuyback(t, caster, id, tc.wantHand)
		})
	}
}

func TestSthB1MindGamesRefusesAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	aura := tmpB1Push(g, opp.ID, "Some Enchantment", "Enchantment", "", 0, 0)
	if err := castCatalogSpellErr(t, g, "Mind Games", "Instant", sthB1MindGamesOracle, cardTarget(aura)); err == nil {
		t.Error("an enchantment was accepted as the target")
	}
}
