package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
			g := effectstest.NewCatalogGame(t)
			id := effectstest.TmpB1Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			if !effectstest.HasEffectiveKeyword(t, g, id, tc.keyword) {
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
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.TmpB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			for i := 1; i <= 2; i++ {
				if tc.pay != "" {
					effectstest.DcMana(t, g, me.ID, tc.pay)
				}
				if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
					t.Fatalf("activation %d: %v", i, err)
				}
				effectstest.PassPriorityAroundTable(t, g)
				wantP, wantT := tc.power+i*tc.dp, tc.toughness+i*tc.dt
				if p, tough := effectstest.PtOf(t, g, id); p != wantP || tough != wantT {
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
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.TmpB1Push(g, me.ID, "Flowstone Shambler", "Creature — Beast", sthB1ShamblerOracle, 2, 2)
	for i := 1; i <= 2; i++ {
		effectstest.DcMana(t, g, me.ID, "{R}")
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(id) {
		t.Error("Flowstone Shambler survived at 4/0")
	}
}

// --- Skyshroud Troopers -------------------------------------------------------------

func TestSthB1SkyshroudTroopersTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.TmpB1Push(g, me.ID, "Skyshroud Troopers", "Creature — Elf Druid Warrior", sthB1TroopersOracle, 3, 3, "G")
	before := me.ManaPool.CountColor("G")
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.CountColor("G") - before; got != 1 {
		t.Errorf("the pool gained %d {G}, want 1", got)
	}
	if !effectstest.IsTapped(g, id) {
		t.Error("Skyshroud Troopers did not tap")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the tapped Troopers made mana again")
	}
}

// --- Foul Imp ------------------------------------------------------------------------

func TestSthB1FoulImpCostsTwoLifeAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Foul Imp", "Creature — Imp", sthB1FoulImpOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - me.Life; got != 2 {
		t.Errorf("the caster lost %d life, want 2", got)
	}
}

// --- Death Stroke and Smite ----------------------------------------------------------

func TestSthB1DeathStrokeDestroysOnlyATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	untapped := effectstest.B12Creature(g, opp.ID, "Awake", "Creature — Bear", 2, 2)
	tapped := effectstest.B12Creature(g, opp.ID, "Asleep", "Creature — Bear", 2, 2)
	effectstest.TapCard(g, tapped)
	if err := effectstest.CastCatalogSpellErr(t, g, "Death Stroke", "Sorcery", sthB1DeathStrokeOracle, effectstest.B16TargetCard(untapped)); err == nil {
		t.Fatal("an untapped creature was accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Death Stroke", "Sorcery", sthB1DeathStrokeOracle, effectstest.B16TargetCard(tapped))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature was destroyed")
	}
}

func TestSthB1SmiteDestroysOnlyABlockedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	blocked := effectstest.B12Creature(g, me.ID, "Blocked Bear", "Creature — Bear", 2, 2)
	unblocked := effectstest.B12Creature(g, me.ID, "Free Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.DeclareAttack(t, g, opp.ID, blocked, unblocked)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, blocked); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	smite := game.Card{Name: "Smite", TypeLine: "Instant", OracleID: sthB1SmiteOracle}
	effectstest.TmcB2Pool(t, g, me, "{W}")
	if _, err := effectstest.TmcB2Cast(g, me, smite, game.CastSpellParams{Targets: effectstest.CardRefs(unblocked)}); err == nil {
		t.Fatal("an unblocked attacker was accepted")
	}
	if _, err := effectstest.TmcB2Cast(g, me, smite, game.CastSpellParams{Targets: effectstest.CardRefs(wall)}); err == nil {
		t.Fatal("the blocking creature was accepted")
	}
	if _, err := effectstest.TmcB2Cast(g, me, smite, game.CastSpellParams{Targets: effectstest.CardRefs(blocked)}); err != nil {
		t.Fatalf("cast on the blocked attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocked) {
		t.Error("the blocked creature survived")
	}
	if !g.Battlefield.Contains(unblocked) || !g.Battlefield.Contains(wall) {
		t.Error("a creature other than the target was destroyed")
	}
}

// --- buyback --------------------------------------------------------------------------

func TestSthB1SeethingAngerGivesPlusThreePlusZero(t *testing.T) {
	for _, tc := range effectstest.TmpB1BuybackCases {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.SeedCreature(g, "Grizzly Bears", caster.ID)
			id, err := effectstest.CastWithOptionalCosts(t, g, "Seething Anger", "Sorcery", sthB1SeethingOracle,
				effectstest.CardTarget(bear), tc.Optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 2 {
				t.Errorf("the target is %d/%d, want 5/2", p, tough)
			}
			effectstest.TmpB1AssertBuyback(t, caster, id, tc.WantHand)
		})
	}
}

func TestSthB1BrushWithDeathDrainsTwo(t *testing.T) {
	for _, tc := range effectstest.TmpB1BuybackCases {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster, victim := effectstest.SeatsForCombat(g)
			mine, theirs := caster.Life, victim.Life
			id, err := effectstest.CastWithOptionalCosts(t, g, "Brush with Death", "Sorcery", sthB1BrushOracle,
				[]game.TargetRef{effectstest.PlayerTarget(victim.ID)}, tc.Optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if theirs-victim.Life != 2 || caster.Life-mine != 2 {
				t.Errorf("life %d → %d (target) and %d → %d (caster), want -2 and +2", theirs, victim.Life, mine, caster.Life)
			}
			effectstest.TmpB1AssertBuyback(t, caster, id, tc.WantHand)
		})
	}
}

func TestSthB1BrushWithDeathTargetsOnlyAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	if err := effectstest.CastCatalogSpellErr(t, g, "Brush with Death", "Sorcery", sthB1BrushOracle,
		[]game.TargetRef{effectstest.PlayerTarget(caster.ID)}); err == nil {
		t.Error("the caster was accepted as the target opponent")
	}
}

func TestSthB1MindPeelMakesTheTargetDiscard(t *testing.T) {
	for _, tc := range effectstest.TmpB1BuybackCases {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster, victim := effectstest.SeatsForCombat(g)
			victim.Hand.Cards = nil
			card := uuid.New()
			victim.Hand.PushTop(game.Card{InstanceID: card, Name: "Victim", TypeLine: "Instant", Owner: victim.ID, Controller: victim.ID})
			victim.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Kept", TypeLine: "Instant", Owner: victim.ID, Controller: victim.ID})
			id, err := effectstest.CastWithOptionalCosts(t, g, "Mind Peel", "Sorcery", sthB1MindPeelOracle,
				[]game.TargetRef{effectstest.PlayerTarget(victim.ID)}, tc.Optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerDiscard(t, g, victim.ID, card)
			if !victim.Graveyard.Contains(card) || victim.Hand.Size() != 1 {
				t.Errorf("hand %d, chosen card in graveyard %v; want one card discarded, the chosen one",
					victim.Hand.Size(), victim.Graveyard.Contains(card))
			}
			effectstest.TmpB1AssertBuyback(t, caster, id, tc.WantHand)
		})
	}
}

func TestSthB1MindGamesTapsTheTarget(t *testing.T) {
	for _, tc := range effectstest.TmpB1BuybackCases {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster, opp := effectstest.SeatsForCombat(g)
			bear := effectstest.SeedCreature(g, "Grizzly Bears", opp.ID)
			id, err := effectstest.CastWithOptionalCosts(t, g, "Mind Games", "Instant", sthB1MindGamesOracle,
				effectstest.CardTarget(bear), tc.Optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.IsTapped(g, bear) {
				t.Error("the target is untapped")
			}
			effectstest.TmpB1AssertBuyback(t, caster, id, tc.WantHand)
		})
	}
}

func TestSthB1MindGamesRefusesAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	aura := effectstest.TmpB1Push(g, opp.ID, "Some Enchantment", "Enchantment", "", 0, 0)
	if err := effectstest.CastCatalogSpellErr(t, g, "Mind Games", "Instant", sthB1MindGamesOracle, effectstest.CardTarget(aura)); err == nil {
		t.Error("an enchantment was accepted as the target")
	}
}
