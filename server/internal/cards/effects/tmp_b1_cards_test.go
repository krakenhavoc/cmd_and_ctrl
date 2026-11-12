package effects

import (
	"testing"

	"github.com/google/uuid"

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

// tmpB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func tmpB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Colors: colors, Owner: owner, Controller: owner,
	})
}

// tmpB1Shadow seeds a vanilla creature with shadow.
func tmpB1Shadow(g *game.Game, owner uuid.UUID, name string, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Spirit",
		Power: 2, Toughness: 2, Keywords: []string{"shadow"}, Colors: colors, Owner: owner, Controller: owner,
	})
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
			g := newCatalogGame(t)
			me, them := seatsForCombat(g)
			attacker := tmpB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 3)
			if !hasEffectiveKeyword(t, g, attacker, "shadow") {
				t.Fatalf("%s does not have shadow", tc.name)
			}
			shadow := tmpB1Shadow(g, them.ID, "Shadow blocker", "U")
			plain := tmpB1Push(g, them.ID, "Plain blocker", "Creature — Bear", "", 2, 2, "U")
			attackIntoBlocks(t, g, them.ID, attacker)
			if got := blockRefusal(t, g, shadow, attacker); got != "" {
				t.Errorf("a blocker with shadow was refused: %q", got)
			}
			if blockRefusal(t, g, plain, attacker) == "" {
				t.Error("a blocker without shadow was accepted")
			}
		})
	}
}

func TestTmpB1ThalakosSentryBlocksOnlyShadowAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	plain := tmpB1Push(g, me.ID, "Plain attacker", "Creature — Bear", "", 2, 2)
	shadow := tmpB1Shadow(g, me.ID, "Shadow attacker")
	sentry := tmpB1Push(g, them.ID, "Thalakos Sentry", "Creature — Thalakos Soldier", tmpB1SentryOracle, 1, 2)
	attackIntoBlocks(t, g, them.ID, plain, shadow)
	if blockRefusal(t, g, sentry, plain) == "" {
		t.Error("Thalakos Sentry was allowed to block a creature without shadow")
	}
	if got := blockRefusal(t, g, sentry, shadow); got != "" {
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
			g := newCatalogGame(t)
			me, them := seatsForCombat(g)
			attacker := tmpB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 1, "W")
			if !hasEffectiveKeyword(t, g, attacker, tc.keyword) {
				t.Fatalf("%s does not have %s", tc.name, tc.keyword)
			}
			coloured := tmpB1Shadow(g, them.ID, "Coloured shadow blocker", tc.colour)
			green := tmpB1Shadow(g, them.ID, "Green shadow blocker", "G")
			attackIntoBlocks(t, g, them.ID, attacker)
			if blockRefusal(t, g, coloured, attacker) == "" {
				t.Errorf("a %s creature with shadow was allowed to block %s", tc.colour, tc.name)
			}
			if got := blockRefusal(t, g, green, attacker); got != "" {
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
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := tmpB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			for i := 1; i <= 2; i++ {
				dcMana(t, g, me.ID, tc.pay)
				if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
					t.Fatalf("activation %d: %v", i, err)
				}
				passPriorityAroundTable(t, g)
				if p, tough := ptOf(t, g, id); p != tc.power+i || tough != tc.toughness {
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
	g := newCatalogGame(t)
	id := tmpB1Push(g, g.Seats[0].ID, "Firefly", "Creature — Insect", tmpB1FireflyOracle, 1, 1)
	if !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Firefly does not have flying")
	}
}

// --- Soltari Trooper --------------------------------------------------------------------

func TestTmpB1SoltariTrooperGetsPlusOnePlusOneWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	trooper := tmpB1Push(g, me.ID, "Soltari Trooper", "Creature — Soltari Soldier", tmpB1TrooperOracle, 1, 1)
	if p, tough := ptOf(t, g, trooper); p != 1 || tough != 1 {
		t.Fatalf("Soltari Trooper is %d/%d before attacking, want 1/1", p, tough)
	}
	declareAttack(t, g, them.ID, trooper)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, trooper); p != 2 || tough != 2 {
		t.Errorf("Soltari Trooper is %d/%d while attacking, want 2/2", p, tough)
	}
}

// --- Thalakos Seer ------------------------------------------------------------------------

func TestTmpB1ThalakosSeerDrawsWhenItLeavesTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seer := tmpB1Push(g, me.ID, "Thalakos Seer", "Creature — Thalakos Wizard", tmpB1SeerOracle, 1, 1)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(seer) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn as the Seer left", hand, got)
	}
}

// --- buyback ------------------------------------------------------------------------------

// tmpB1BuybackCases is each buyback spell cast twice: without buyback it
// ends in the graveyard, with it back in its owner's hand.
var tmpB1BuybackCases = []struct {
	name     string
	optional []int
	wantHand bool
}{
	{"without buyback", nil, false},
	{"with buyback", []int{0}, true},
}

// tmpB1AssertBuyback checks where the resolved spell went.
func tmpB1AssertBuyback(t *testing.T, caster *game.Player, id uuid.UUID, wantHand bool) {
	t.Helper()
	inHand, inGrave := caster.Hand.Contains(id), caster.Graveyard.Contains(id)
	if inHand != wantHand || inGrave == wantHand {
		t.Errorf("after resolution: hand=%v graveyard=%v, want hand=%v", inHand, inGrave, wantHand)
	}
}

func TestTmpB1WhispersOfTheMuseDrawsACard(t *testing.T) {
	for _, tc := range tmpB1BuybackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			caster := g.Seats[g.Turn.ActiveSeat]
			hand := caster.Hand.Size()
			id, err := castWithOptionalCosts(t, g, "Whispers of the Muse", "Instant", tmpB1WhispersOracle, nil, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			want := hand + 1
			if tc.wantHand {
				want++
			}
			if got := caster.Hand.Size(); got != want {
				t.Errorf("hand %d → %d, want %d", hand, got, want)
			}
			tmpB1AssertBuyback(t, caster, id, tc.wantHand)
		})
	}
}

func TestTmpB1SearingTouchDealsOneDamage(t *testing.T) {
	for _, tc := range tmpB1BuybackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster, victim := seatsForCombat(g)
			life := victim.Life
			id, err := castWithOptionalCosts(t, g, "Searing Touch", "Instant", tmpB1SearingOracle,
				[]game.TargetRef{playerTarget(victim.ID)}, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			if got := life - victim.Life; got != 1 {
				t.Errorf("the target lost %d life, want 1", got)
			}
			tmpB1AssertBuyback(t, caster, id, tc.wantHand)
		})
	}
}

func TestTmpB1ElvishFuryGivesPlusTwoPlusTwo(t *testing.T) {
	for _, tc := range tmpB1BuybackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			bear := seedCreature(g, "Grizzly Bears", caster.ID)
			id, err := castWithOptionalCosts(t, g, "Elvish Fury", "Instant", tmpB1ElvishFuryOracle,
				cardTarget(bear), tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
				t.Errorf("the target is %d/%d, want 4/4", p, tough)
			}
			tmpB1AssertBuyback(t, caster, id, tc.wantHand)
		})
	}
}

func TestTmpB1DisturbedBurialReturnsACreatureCard(t *testing.T) {
	for _, tc := range tmpB1BuybackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			bear := pushGraveyardCardTyped(caster, "Bear", "Creature — Bear")
			id, err := castWithOptionalCosts(t, g, "Disturbed Burial", "Sorcery", tmpB1BurialOracle,
				cardTarget(bear), tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			if !caster.Hand.Contains(bear) {
				t.Error("the creature card did not come back to hand")
			}
			tmpB1AssertBuyback(t, caster, id, tc.wantHand)
		})
	}
}

func TestTmpB1DisturbedBurialRefusesANoncreatureOrAnotherPlayersCard(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := seatsForCombat(g)
	rite := pushGraveyardCardTyped(caster, "Rite", "Sorcery")
	theirs := pushGraveyardCardTyped(opp, "Their Bear", "Creature — Bear")
	for _, bad := range []uuid.UUID{rite, theirs} {
		if err := castCatalogSpellErr(t, g, "Disturbed Burial", "Sorcery", tmpB1BurialOracle, cardTarget(bad)); err == nil {
			t.Error("a noncreature card, or a card in another graveyard, was accepted as the target")
		}
	}
}
