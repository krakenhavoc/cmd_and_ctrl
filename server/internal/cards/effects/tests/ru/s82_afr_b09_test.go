package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b09_test.go — Adventures in the Forgotten Realms, batch 9: the
// Treasure-mana payoffs and pack tactics.

const (
	afrb09SpoilsOracle    = "c29713ee-57e9-40b4-8ca3-a4cc6ede0f46"
	afrb09PriceOracle     = "f25b97f9-8ca2-40f7-9d55-47fa8bb2054e"
	afrb09PaladinOracle   = "d3564614-4463-4683-83f0-07561ab3431c"
	afrb09SellSwordOracle = "f666a990-bd9c-4160-9770-4758fd40ca92"
	afrb09MimicOracle     = "ba8156cd-2e78-4763-b61d-8bc85b2f3657"
	afrb09ContainOracle   = "0a5e6223-ca31-4605-9c33-9ce492aaf556"
	afrb09GnollOracle     = "9e445336-8b1e-4642-b8bc-aa460ad79a3f"
	afrb09TigerOracle     = "d5b4e9fd-3df6-4d1d-8135-d7c8615b05a7"
	afrb09TargNarOracle   = "d53ef172-8c26-4469-94c6-bf7fe4d48dbf"
	afrb09BattleCryOracle = "e74c4001-d958-4d2b-b57d-c3601f01580c"
	afrb09CaptainOracle   = "4d92b3d3-d34b-430f-af6a-db512d204d06"

	afrb09Treasure = game.ManaSourceTreasure | game.ManaSourceArtifact
)

func TestS82AfrB09EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		afrb09SpoilsOracle: "Spoils of the Hunt", afrb09PriceOracle: "Price of Loyalty",
		afrb09PaladinOracle: "Forsworn Paladin", afrb09SellSwordOracle: "Jaded Sell-Sword",
		afrb09MimicOracle: "Mimic", afrb09ContainOracle: "Minimus Containment",
		afrb09GnollOracle: "Gnoll Hunter", afrb09TigerOracle: "Tiger-Tribe Hunter",
		afrb09TargNarOracle: "Targ Nar", afrb09BattleCryOracle: "Battle Cry Goblin",
		afrb09CaptainOracle: "Hobgoblin Captain",
	} {
		effectstest.RequireFull(t, oracle, name)
	}
}

func afrb09Creature(g *game.Game, owner uuid.UUID, name, oracle, typeLine string, p, tough int) uuid.UUID {
	return effectstest.PushDiesCreatureForTest(g, owner, name, oracle, typeLine, p, tough)
}

// --- pack tactics --------------------------------------------------------

func TestS82AfrB09GnollHunterGrowsOnlyWithSixPower(t *testing.T) {
	for _, tc := range []struct {
		name            string
		other, counters int
	}{{"total power 6", 4, 1}, {"total power 5", 3, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			hunter := afrb09Creature(g, me.ID, "Gnoll Hunter", afrb09GnollOracle, "Creature — Gnoll", 2, 2)
			other := effectstest.PushVanillaCreature(g, me.ID, "Bear", tc.other, tc.other)
			effectstest.HwDeclare(t, g, opp.ID, hunter, other)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, hunter, game.CounterPlusOne); got != tc.counters {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.counters)
			}
		})
	}
}

func TestS82AfrB09HobgoblinCaptainGainsFirstStrike(t *testing.T) {
	for _, tc := range []struct {
		name  string
		other int
		first bool
	}{{"six power", 3, true}, {"five power", 2, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			cap := afrb09Creature(g, me.ID, "Hobgoblin Captain", afrb09CaptainOracle, "Creature — Goblin Barbarian", 3, 1)
			other := effectstest.PushVanillaCreature(g, me.ID, "Bear", tc.other, tc.other)
			effectstest.HwDeclare(t, g, opp.ID, cap, other)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.HwHasKeyword(effectstest.LayeredCard(t, g, cap), "first strike"); got != tc.first {
				t.Errorf("first strike = %v, want %v", got, tc.first)
			}
		})
	}
}

func TestS82AfrB09TargNarPumpsAttackersAndDoubles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nar := afrb09Creature(g, me.ID, "Targ Nar, Demon-Fang Gnoll", afrb09TargNarOracle, "Legendary Creature — Gnoll", 2, 2)
	other := effectstest.PushVanillaCreature(g, me.ID, "Bear", 4, 4)
	idle := effectstest.PushVanillaCreature(g, me.ID, "Idle", 1, 1)
	effectstest.HwDeclare(t, g, opp.ID, nar, other)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, nar); c.CurrentPower() != 3 {
		t.Errorf("Targ Nar power = %d, want 3", c.CurrentPower())
	}
	if c := effectstest.LayeredCard(t, g, other); c.CurrentPower() != 5 {
		t.Errorf("attacking Bear power = %d, want 5", c.CurrentPower())
	}
	if c := effectstest.LayeredCard(t, g, idle); c.CurrentPower() != 1 {
		t.Errorf("a creature that did not attack has power %d, want 1", c.CurrentPower())
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{G}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, nar, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, nar); c.CurrentPower() != 6 || c.CurrentToughness() != 4 {
		t.Errorf("doubled Targ Nar is %d/%d, want 6/4 (3/2 doubled)", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestS82AfrB09BattleCryGoblinMakesATappedAttackingGoblin(t *testing.T) {
	for _, tc := range []struct {
		name           string
		other, goblins int
	}{{"six power", 4, 2}, {"five power", 3, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			gob := afrb09Creature(g, me.ID, "Battle Cry Goblin", afrb09BattleCryOracle, "Creature — Goblin", 2, 2)
			other := effectstest.PushVanillaCreature(g, me.ID, "Bear", tc.other, tc.other)
			effectstest.HwDeclare(t, g, opp.ID, gob, other)
			effectstest.PassPriorityAroundTable(t, g)
			n := 0
			for _, c := range g.Battlefield.Cards {
				if c.IsToken() && c.Name == "Goblin" {
					n++
					if !c.Tapped || c.AttackingTarget != opp.ID {
						t.Errorf("token tapped=%v attacking=%v, want tapped and attacking the defender", c.Tapped, c.AttackingTarget == opp.ID)
					}
				}
			}
			if want := tc.goblins - 1; n != want {
				t.Errorf("Goblin tokens = %d, want %d", n, want)
			}
		})
	}
}

func TestS82AfrB09BattleCryGoblinPumpsYourGoblinsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gob := afrb09Creature(g, me.ID, "Battle Cry Goblin", afrb09BattleCryOracle, "Creature — Goblin", 2, 2)
	friend := afrb09Creature(g, me.ID, "Friend", "", "Creature — Goblin", 1, 1)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	foe := afrb09Creature(g, opp.ID, "Foe", "", "Creature — Goblin", 1, 1)
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, gob, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]struct {
		p     int
		haste bool
	}{gob: {3, true}, friend: {2, true}, bear: {2, false}, foe: {1, false}} {
		c := effectstest.LayeredCard(t, g, id)
		if c.CurrentPower() != want.p || effectstest.HwHasKeyword(c, "haste") != want.haste {
			t.Errorf("%s: power %d haste %v, want %d %v", c.Name, c.CurrentPower(), effectstest.HwHasKeyword(c, "haste"), want.p, want.haste)
		}
	}
}

func TestS82AfrB09TigerTribeHunterSacrificesForDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tiger := afrb09Creature(g, me.ID, "Tiger-Tribe Hunter", afrb09TigerOracle, "Creature — Human Barbarian", 4, 4)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Ogre", 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 3)
	effectstest.HwDeclare(t, g, opp.ID, tiger, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is gone")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("3 damage kills a 1/3")
	}
}

func TestS82AfrB09TigerTribeHunterNeedsSixPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tiger := afrb09Creature(g, me.ID, "Tiger-Tribe Hunter", afrb09TigerOracle, "Creature — Human Barbarian", 4, 4)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Ogre", 1, 1)
	effectstest.HwDeclare(t, g, opp.ID, tiger, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) {
		t.Error("with 5 total power the trigger never happens, so nothing is sacrificed")
	}
}

// --- Treasure mana -------------------------------------------------------

func afrb09Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle string, treasure, plain string, targets ...uuid.UUID) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.GmrFloatFrom(g, p, treasure, afrb09Treasure)
	effectstest.GmrFloatFrom(g, p, plain, 0)
	var refs []game.TargetRef
	for _, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return effectstest.CastFromHandForTest(t, g, p, name, typeLine, cost, oracle, game.CastSpellParams{Strict: true, Targets: refs})
}

func TestS82AfrB09SpoilsOfTheHuntCountsEachTreasureMana(t *testing.T) {
	for _, tc := range []struct {
		name, treasure, plain string
		power                 int
		dead                  bool
	}{
		{"no Treasure mana", "", "GGG", 2, false},
		{"two Treasure mana", "GG", "G", 4, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
			victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 1, 4)
			afrb09Cast(t, g, me, "Spoils of the Hunt", "Instant", "{2}{G}", afrb09SpoilsOracle, tc.treasure, tc.plain, mine, victim)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.LayeredCard(t, g, mine).CurrentPower(); got != tc.power {
				t.Errorf("power = %d, want %d", got, tc.power)
			}
			if got := !g.Battlefield.Contains(victim); got != tc.dead {
				t.Errorf("victim dead = %v, want %v (it bites for the pumped power)", got, tc.dead)
			}
		})
	}
}

func TestS82AfrB09PriceOfLoyaltyPumpsOnlyWithTreasureMana(t *testing.T) {
	for _, tc := range []struct {
		name, treasure, plain string
		power                 int
	}{
		{"ordinary mana", "", "RRR", 3},
		{"a Treasure paid", "R", "RR", 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == victim {
						g.Battlefield.Cards[i].Tapped = true
					}
				}
			})
			afrb09Cast(t, g, me, "Price of Loyalty", "Sorcery", "{2}{R}", afrb09PriceOracle, tc.treasure, tc.plain, victim)
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.LayeredCard(t, g, victim)
			if c.Controller != me.ID || c.Tapped || !effectstest.HwHasKeyword(c, "haste") || c.CurrentPower() != tc.power {
				t.Errorf("controller mine=%v tapped=%v haste=%v power=%d, want stolen, untapped, hasty, power %d",
					c.Controller == me.ID, c.Tapped, effectstest.HwHasKeyword(c, "haste"), c.CurrentPower(), tc.power)
			}
		})
	}
}

func TestS82AfrB09JadedSellSwordNeedsTreasureMana(t *testing.T) {
	for _, tc := range []struct {
		name, treasure, plain string
		ok                    bool
	}{
		{"Treasure mana", "R", "RRR", true},
		{"ordinary mana", "", "RRRR", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := afrb09Cast(t, g, me, "Jaded Sell-Sword", "Creature — Dragon Warrior", "{3}{R}", afrb09SellSwordOracle, tc.treasure, tc.plain)
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.LayeredCard(t, g, id)
			if got := effectstest.HwHasKeyword(c, "first strike") && effectstest.HwHasKeyword(c, "haste"); got != tc.ok {
				t.Errorf("first strike and haste = %v, want %v", got, tc.ok)
			}
		})
	}
}

func TestS82AfrB09ForswornPaladinMakesTreasureAndGrantsDeathtouchOnTreasureMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pal := afrb09Creature(g, me.ID, "Forsworn Paladin", afrb09PaladinOracle, "Creature — Human Knight", 1, 1)
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	life := me.Life
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, pal, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the Treasure ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HwCount(g, me.ID, "Treasure") != 1 || me.Life != life-1 {
		t.Errorf("treasures %d life %d, want one Treasure and 1 life paid", effectstest.HwCount(g, me.ID, "Treasure"), life-me.Life)
	}
	for _, tc := range []struct {
		name  string
		kinds game.ManaSourceKinds
		dt    bool
	}{
		{"plain mana", 0, false}, {"Treasure mana", afrb09Treasure, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := effectstest.PushVanillaCreature(g, me.ID, "Bear "+tc.name, 2, 2)
			effectstest.GmrFloatFrom(g, me, "BBB", tc.kinds)
			if err := g.ActivateCatalogAbility(me.ID, pal, 1, game.ActivateAbilityParams{
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
				t.Fatalf("activate the pump: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.LayeredCard(t, g, target)
			if c.CurrentPower() != 4 || effectstest.HwHasKeyword(c, "deathtouch") != tc.dt {
				t.Errorf("power %d deathtouch %v, want 4 and %v", c.CurrentPower(), effectstest.HwHasKeyword(c, "deathtouch"), tc.dt)
			}
		})
	}
}

// --- Mimic and Minimus Containment ----------------------------------------

func TestS82AfrB09MimicAnimatesAndItsManaIsTreasureMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mimic := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mimic", OracleID: afrb09MimicOracle,
		TypeLine: "Artifact — Treasure", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.GmrFloatFrom(g, me, "CC", 0)
	if err := g.ActivateCatalogAbility(me.ID, mimic, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("animate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, mimic)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Shapeshifter") || !c.HasSubtype("Treasure") ||
		c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("animated Mimic: creature=%v artifact=%v types=%v %d/%d, want a 3/3 Shapeshifter artifact creature still a Treasure",
			c.IsCreature(), c.IsArtifact(), c.Effective().Subtypes, c.CurrentPower(), c.CurrentToughness())
	}
	// Its mana is Treasure mana although it is not a token: it pays for Jaded Sell-Sword.
	if err := g.ActivateManaAbility(me.ID, mimic, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack Mimic: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("no colour pick queued by Mimic")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(mimic) {
		t.Error("Mimic is sacrificed for its mana")
	}
	effectstest.FloatForTest(g, me, "RRR")
	id := effectstest.CastFromHandForTest(t, g, me, "Jaded Sell-Sword", "Creature — Dragon Warrior", "{3}{R}", afrb09SellSwordOracle, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HwHasKeyword(effectstest.LayeredCard(t, g, id), "first strike") {
		t.Error("mana from a nontoken Treasure counts as mana from a Treasure")
	}
}

func TestS82AfrB09MinimusContainmentTurnsACreatureIntoATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Flyer", TypeLine: "Creature — Dragon",
		ManaCost: "{4}{R}", Colors: []string{"R"}, Keywords: []string{"flying"}, Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	afrb09Cast(t, g, me, "Minimus Containment", "Enchantment — Aura", "{2}{W}", afrb09ContainOracle, "", "WWW", host)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, host)
	if !ok {
		t.Fatal("the host left")
	}
	e := c.Effective()
	if len(e.Types) != 1 || e.Types[0] != "Artifact" || c.IsCreature() || !c.HasSubtype("Treasure") || c.HasSubtype("Dragon") {
		t.Errorf("types %v subtypes %v, want only a Treasure artifact", e.Types, e.Subtypes)
	}
	if game.HasKeyword(&c, "flying") {
		t.Error("it lost its other abilities")
	}
	if err := g.ActivateManaAbility(opp.ID, host, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the Treasure's mana ability: %v", err)
	}
	if g.Battlefield.Contains(host) {
		t.Error("the Treasure is sacrificed for its mana")
	}
}

func TestS82AfrB09MinimusContainmentCannotEnchantALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.GmrFloatFrom(g, me, "WWW", 0)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Minimus Containment", TypeLine: "Enchantment — Aura", ManaCost: "{2}{W}",
			OracleID: afrb09ContainOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err == nil {
		t.Error("enchant nonland permanent refused a land")
	}
}
