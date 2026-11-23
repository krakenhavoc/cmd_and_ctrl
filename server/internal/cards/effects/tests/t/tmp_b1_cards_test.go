package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmp_b1_cards_test.go — Tempest, batch 1 (#5628). Each test plays the
// real catalog entry and asserts what its text says.

const (
	tmpB1FootSoldierOracle = "6ccb3ac8-0509-431b-a196-d3cbf4fa80ae"
	tmpB1MarauderOracle    = "cb754749-a769-4c9c-b62b-4972dfc5069f"
	tmpB1SentryOracle      = "ca65b908-e6fe-4ca3-8ecd-247df957aa0f"
	tmpB1MonkOracle        = "5fed5377-07cf-4ad9-86fd-1efe1f182088"
	tmpB1PriestOracle      = "f8c9cb9f-3a37-4004-9cde-f51ca5579e1a"
	tmpB1MercenaryOracle   = "1b57c480-e2c5-44cd-820f-e208041bfd18"
	tmpB1CrusaderOracle    = "39e01990-9928-440c-a5c3-82aea29e2cd0"
	tmpB1TrooperOracle     = "501109bb-3406-41d8-9433-a100e494f938"
	tmpB1SeerOracle        = "1baec770-0785-4cbc-be18-83572fd8e9c1"
	tmpB1FireflyOracle     = "6b151418-092f-4a43-8b21-3de64ea80cab"
	tmpB1WhispersOracle    = "a4a8af4e-8686-45ba-94fd-ccf6cc8adaae"
	tmpB1SearingOracle     = "1cd9a3ba-a115-4e59-983d-605e4435f3a8"
	tmpB1ElvishFuryOracle  = "557d2af0-04c5-461d-826b-7208c4535dc2"
	tmpB1BurialOracle      = "cdbaa21d-5896-4b8b-9a08-f0839ccfddfc"
)

func TestTmpB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmpB1FootSoldierOracle: "Soltari Foot Soldier",
		tmpB1MarauderOracle:    "Dauthi Marauder",
		tmpB1SentryOracle:      "Thalakos Sentry",
		tmpB1MonkOracle:        "Soltari Monk",
		tmpB1PriestOracle:      "Soltari Priest",
		tmpB1MercenaryOracle:   "Dauthi Mercenary",
		tmpB1CrusaderOracle:    "Soltari Crusader",
		tmpB1TrooperOracle:     "Soltari Trooper",
		tmpB1SeerOracle:        "Thalakos Seer",
		tmpB1FireflyOracle:     "Firefly",
		tmpB1WhispersOracle:    "Whispers of the Muse",
		tmpB1SearingOracle:     "Searing Touch",
		tmpB1ElvishFuryOracle:  "Elvish Fury",
		tmpB1BurialOracle:      "Disturbed Burial",
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

// --- shadow -----------------------------------------------------------------------

var tmpB1ShadowCards = []struct{ name, oracle, typeLine string }{
	{"Soltari Foot Soldier", tmpB1FootSoldierOracle, "Creature — Soltari Soldier"},
	{"Dauthi Marauder", tmpB1MarauderOracle, "Creature — Dauthi Minion"},
	{"Thalakos Sentry", tmpB1SentryOracle, "Creature — Thalakos Soldier"},
	{"Soltari Monk", tmpB1MonkOracle, "Creature — Soltari Monk Cleric"},
	{"Soltari Priest", tmpB1PriestOracle, "Creature — Soltari Cleric"},
	{"Dauthi Mercenary", tmpB1MercenaryOracle, "Creature — Dauthi Knight Mercenary"},
	{"Soltari Crusader", tmpB1CrusaderOracle, "Creature — Soltari Knight"},
	{"Soltari Trooper", tmpB1TrooperOracle, "Creature — Soltari Soldier"},
	{"Thalakos Seer", tmpB1SeerOracle, "Creature — Thalakos Wizard"},
}

func TestTmpB1ShadowAttackersAreBlockedOnlyByShadow(t *testing.T) {
	for _, tc := range tmpB1ShadowCards {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := effectstest.SeatsForCombat(g)
			attacker := effectstest.TmpB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 3)
			if !effectstest.HasEffectiveKeyword(t, g, attacker, "shadow") {
				t.Fatalf("%s does not have shadow", tc.name)
			}
			shadow := effectstest.TmpB1Shadow(g, them.ID, "Shadow blocker", "U")
			plain := effectstest.TmpB1Push(g, them.ID, "Plain blocker", "Creature — Bear", "", 2, 2, "U")
			effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
			if got := effectstest.BlockRefusal(t, g, shadow, attacker); got != "" {
				t.Errorf("a blocker with shadow was refused: %q", got)
			}
			if effectstest.BlockRefusal(t, g, plain, attacker) == "" {
				t.Error("a blocker without shadow was accepted")
			}
		})
	}
}

func TestTmpB1ThalakosSentryBlocksOnlyShadowAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	plain := effectstest.TmpB1Push(g, me.ID, "Plain attacker", "Creature — Bear", "", 2, 2)
	shadow := effectstest.TmpB1Shadow(g, me.ID, "Shadow attacker")
	sentry := effectstest.TmpB1Push(g, them.ID, "Thalakos Sentry", "Creature — Thalakos Soldier", tmpB1SentryOracle, 1, 2)
	effectstest.AttackIntoBlocks(t, g, them.ID, plain, shadow)
	if effectstest.BlockRefusal(t, g, sentry, plain) == "" {
		t.Error("Thalakos Sentry was allowed to block a creature without shadow")
	}
	if got := effectstest.BlockRefusal(t, g, sentry, shadow); got != "" {
		t.Errorf("Thalakos Sentry was refused a block on a shadow attacker: %q", got)
	}
}

// --- protection ----------------------------------------------------------------------

func TestTmpB1SoltariMonkAndPriestCantBeBlockedByTheirColour(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, keyword, colour string
	}{
		{"Soltari Monk", tmpB1MonkOracle, "Creature — Soltari Monk Cleric", "protection from black", "B"},
		{"Soltari Priest", tmpB1PriestOracle, "Creature — Soltari Cleric", "protection from red", "R"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := effectstest.SeatsForCombat(g)
			attacker := effectstest.TmpB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 1, "W")
			if !effectstest.HasEffectiveKeyword(t, g, attacker, tc.keyword) {
				t.Fatalf("%s does not have %s", tc.name, tc.keyword)
			}
			coloured := effectstest.TmpB1Shadow(g, them.ID, "Coloured shadow blocker", tc.colour)
			green := effectstest.TmpB1Shadow(g, them.ID, "Green shadow blocker", "G")
			effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
			if effectstest.BlockRefusal(t, g, coloured, attacker) == "" {
				t.Errorf("a %s creature with shadow was allowed to block %s", tc.colour, tc.name)
			}
			if got := effectstest.BlockRefusal(t, g, green, attacker); got != "" {
				t.Errorf("a green creature with shadow was refused: %q", got)
			}
		})
	}
}

// --- pumps ------------------------------------------------------------------------

func TestTmpB1PumpsGivePlusOnePlusZero(t *testing.T) {
	for _, tc := range []struct {
		// pay is mana that funds the cost: {B}{B} pays {1}{B}.
		name, oracle, typeLine, pay string
		power, toughness            int
	}{
		{"Dauthi Mercenary", tmpB1MercenaryOracle, "Creature — Dauthi Knight Mercenary", "{B}{B}", 2, 1},
		{"Soltari Crusader", tmpB1CrusaderOracle, "Creature — Soltari Knight", "{W}{W}", 2, 1},
		{"Firefly", tmpB1FireflyOracle, "Creature — Insect", "{R}", 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.TmpB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			for i := 1; i <= 2; i++ {
				effectstest.DcMana(t, g, me.ID, tc.pay)
				if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
					t.Fatalf("activation %d: %v", i, err)
				}
				effectstest.PassPriorityAroundTable(t, g)
				if p, tough := effectstest.PtOf(t, g, id); p != tc.power+i || tough != tc.toughness {
					t.Fatalf("%s is %d/%d after %d activations, want %d/%d", tc.name, p, tough, i, tc.power+i, tc.toughness)
				}
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Error("the pump was activated with an empty mana pool")
			}
		})
	}
}

func TestTmpB1FireflyHasFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.TmpB1Push(g, g.Seats[0].ID, "Firefly", "Creature — Insect", tmpB1FireflyOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Firefly does not have flying")
	}
}

// --- Soltari Trooper --------------------------------------------------------------------

func TestTmpB1SoltariTrooperGetsPlusOnePlusOneWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	trooper := effectstest.TmpB1Push(g, me.ID, "Soltari Trooper", "Creature — Soltari Soldier", tmpB1TrooperOracle, 1, 1)
	if p, tough := effectstest.PtOf(t, g, trooper); p != 1 || tough != 1 {
		t.Fatalf("Soltari Trooper is %d/%d before attacking, want 1/1", p, tough)
	}
	effectstest.DeclareAttack(t, g, them.ID, trooper)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, trooper); p != 2 || tough != 2 {
		t.Errorf("Soltari Trooper is %d/%d while attacking, want 2/2", p, tough)
	}
}

// --- Thalakos Seer ------------------------------------------------------------------------

func TestTmpB1ThalakosSeerDrawsWhenItLeavesTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seer := effectstest.TmpB1Push(g, me.ID, "Thalakos Seer", "Creature — Thalakos Wizard", tmpB1SeerOracle, 1, 1)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(seer) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn as the Seer left", hand, got)
	}
}

func TestTmpB1WhispersOfTheMuseDrawsACard(t *testing.T) {
	for _, tc := range effectstest.TmpB1BuybackCases {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			caster := g.Seats[g.Turn.ActiveSeat]
			hand := caster.Hand.Size()
			id, err := effectstest.CastWithOptionalCosts(t, g, "Whispers of the Muse", "Instant", tmpB1WhispersOracle, nil, tc.Optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			want := hand + 1
			if tc.WantHand {
				want++
			}
			if got := caster.Hand.Size(); got != want {
				t.Errorf("hand %d → %d, want %d", hand, got, want)
			}
			effectstest.TmpB1AssertBuyback(t, caster, id, tc.WantHand)
		})
	}
}

func TestTmpB1SearingTouchDealsOneDamage(t *testing.T) {
	for _, tc := range effectstest.TmpB1BuybackCases {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster, victim := effectstest.SeatsForCombat(g)
			life := victim.Life
			id, err := effectstest.CastWithOptionalCosts(t, g, "Searing Touch", "Instant", tmpB1SearingOracle,
				[]game.TargetRef{effectstest.PlayerTarget(victim.ID)}, tc.Optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := life - victim.Life; got != 1 {
				t.Errorf("the target lost %d life, want 1", got)
			}
			effectstest.TmpB1AssertBuyback(t, caster, id, tc.WantHand)
		})
	}
}

func TestTmpB1ElvishFuryGivesPlusTwoPlusTwo(t *testing.T) {
	for _, tc := range effectstest.TmpB1BuybackCases {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.SeedCreature(g, "Grizzly Bears", caster.ID)
			id, err := effectstest.CastWithOptionalCosts(t, g, "Elvish Fury", "Instant", tmpB1ElvishFuryOracle,
				effectstest.CardTarget(bear), tc.Optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
				t.Errorf("the target is %d/%d, want 4/4", p, tough)
			}
			effectstest.TmpB1AssertBuyback(t, caster, id, tc.WantHand)
		})
	}
}

func TestTmpB1DisturbedBurialReturnsACreatureCard(t *testing.T) {
	for _, tc := range effectstest.TmpB1BuybackCases {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.PushGraveyardCardTyped(caster, "Bear", "Creature — Bear")
			id, err := effectstest.CastWithOptionalCosts(t, g, "Disturbed Burial", "Sorcery", tmpB1BurialOracle,
				effectstest.CardTarget(bear), tc.Optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !caster.Hand.Contains(bear) {
				t.Error("the creature card did not come back to hand")
			}
			effectstest.TmpB1AssertBuyback(t, caster, id, tc.WantHand)
		})
	}
}

func TestTmpB1DisturbedBurialRefusesANoncreatureOrAnotherPlayersCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := effectstest.SeatsForCombat(g)
	rite := effectstest.PushGraveyardCardTyped(caster, "Rite", "Sorcery")
	theirs := effectstest.PushGraveyardCardTyped(opp, "Their Bear", "Creature — Bear")
	for _, bad := range []uuid.UUID{rite, theirs} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Disturbed Burial", "Sorcery", tmpB1BurialOracle, effectstest.CardTarget(bad)); err == nil {
			t.Error("a noncreature card, or a card in another graveyard, was accepted as the target")
		}
	}
}
