package testss82m

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b496_cards_test.go — Zendikar landfall, pump and token makers, S82 OLD
// batch 496 (#5583), each driven through the engine.

const (
	zenSoulStair   = "24b11ede-26c8-40ed-b874-c0f6b21572c3"
	zenZektar      = "45bd5e97-f34e-47bc-ad21-da42d3444431"
	zenVessel      = "c8fa1319-6623-4999-acb5-a5e15e6d868c"
	zenSunspring   = "f5aeed95-24f8-4f28-9e7c-1e3502d013cd"
	zenPrimalBell  = "5cd064b9-1277-4edc-bf2c-82eb5e62f9d7"
	zenLethargy    = "7356319d-1485-426b-80a7-9700d6684ac9"
	zenLarva       = "e53532db-5b14-429e-8437-a369c5b8632b"
	zenArmament    = "f1dca9ac-8886-4688-9245-763ef486701c"
	zenKalitas     = "003bf21f-6815-4098-a186-aa15f5947541"
	zenTurntimber  = "391fdf1e-e489-410e-9f33-b674fd98d0fc"
	zenLullmage    = "752566b4-d3f5-4f10-b6f5-7185809f7a9d"
	zenCountersTag = "quest"
)

func TestS82OldB496EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		zenSoulStair: "Soul Stair Expedition", zenZektar: "Zektar Shrine Expedition", zenVessel: "Eternity Vessel",
		zenSunspring: "Sunspring Expedition", zenPrimalBell: "Primal Bellow", zenLethargy: "Lethargy Trap",
		zenLarva: "Timbermaw Larva", zenArmament: "Armament Master", zenKalitas: "Kalitas, Bloodchief of Ghet",
		zenTurntimber: "Turntimber Ranger", zenLullmage: "Lullmage Mentor",
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

// zenExpeditionCounts drives the landfall counter of an Expedition: a yes counts, a no does not.
func zenExpeditionCounts(t *testing.T, name, oracle string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	q := effectstest.B12Push(g, me.ID, name, "Enchantment", oracle, 0, 0)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.DrcSettle(t, g, me.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, q, zenCountersTag); got != 1 {
		t.Fatalf("%s: %d quest counters after one land and a yes, want 1", name, got)
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.DrcSettle(t, g, me.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, q, zenCountersTag); got != 1 {
		t.Fatalf("%s: %d quest counters after declining, want 1", name, got)
	}
	return g, me, q
}

func zenExpeditionPay(t *testing.T, g *game.Game, me *game.Player, q uuid.UUID, params game.ActivateAbilityParams) {
	t.Helper()
	effectstest.SetCountersOn(g, q, map[string]int{"quest": 2})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, q, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with two quest counters")
	}
	if _, ok := effectstest.LookupBattlefield(g, q); !ok {
		t.Fatal("a refused activation sacrificed the enchantment")
	}
	effectstest.SetCountersOn(g, q, map[string]int{"quest": 3})
	params.Strict = true
	if err := g.ActivateCatalogAbility(me.ID, q, 0, params); err != nil {
		t.Fatalf("activate with three counters: %v", err)
	}
	if _, ok := effectstest.LookupBattlefield(g, q); ok {
		t.Fatal("the enchantment is still on the battlefield after paying its sacrifice")
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestS82OldB496SoulStairExpeditionReturnsUpToTwoCreatureCards(t *testing.T) {
	g, me, q := zenExpeditionCounts(t, "Soul Stair Expedition", zenSoulStair)
	var dead []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: fmt.Sprintf("Dead Bear %d", i), TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		dead = append(dead, id)
	}
	rite := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: rite, Name: "Dead Rite", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	zenExpeditionPay(t, g, me, q, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead[0], dead[2])})
	if !me.Hand.Contains(dead[0]) || !me.Hand.Contains(dead[2]) {
		t.Error("a chosen creature card did not return")
	}
	if me.Hand.Contains(dead[1]) || me.Hand.Contains(rite) {
		t.Error("a card that was not chosen returned")
	}
}

func TestS82OldB496SunspringExpeditionGainsEight(t *testing.T) {
	g, me, q := zenExpeditionCounts(t, "Sunspring Expedition", zenSunspring)
	before := me.Life
	zenExpeditionPay(t, g, me, q, game.ActivateAbilityParams{})
	if me.Life != before+8 {
		t.Errorf("life %d -> %d, want +8", before, me.Life)
	}
}

func TestS82OldB496ZektarShrineExpeditionMakesAnElementalThatIsExiledAtEndStep(t *testing.T) {
	g, me, q := zenExpeditionCounts(t, "Zektar Shrine Expedition", zenZektar)
	zenExpeditionPay(t, g, me, q, game.ActivateAbilityParams{})
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Elemental" {
			token = c.InstanceID
			if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 7 || tough != 1 {
				t.Errorf("the token is %d/%d, want 7/1", p, tough)
			}
			if !effectstest.P1g2HasKeyword(g, c.InstanceID, "trample") || !effectstest.P1g2HasKeyword(g, c.InstanceID, "haste") {
				t.Error("the Elemental lacks trample or haste")
			}
		}
	}
	if token == uuid.Nil {
		t.Fatal("no Elemental token")
	}
	effectstest.EndTheTurn(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the Elemental was not exiled at the beginning of the end step")
	}
}

func TestS82OldB496EternityVesselEntersWithYourLifeInCountersAndResetsItOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 17-me.Life) })
	vessel := effectstest.TdCastFromHand(t, g, game.Card{Name: "Eternity Vessel", TypeLine: "Artifact", OracleID: zenVessel}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, vessel, "charge"); got != 17 {
		t.Fatalf("%d charge counters, want 17", got)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 8) })
	if me.Life != 25 {
		t.Fatalf("life %d, test setup wants 25", me.Life)
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.DrcSettle(t, g, me.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 25 {
		t.Fatalf("declined, life is %d, want 25", me.Life)
	}
	// A second land (it needs another land drop).
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.DrcSettle(t, g, me.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 17 {
		t.Errorf("life %d, want 17 (the charge counters)", me.Life)
	}
	if got := effectstest.CountersOn(g, vessel, "charge"); got != 17 {
		t.Errorf("the counters changed to %d", got)
	}
}

func TestS82OldB496PrimalBellowPumpsPerForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.PushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	}
	effectstest.PushPermanentForTest(g, me.ID, "Plains", "", "Basic Land — Plains")
	effectstest.PushPermanentForTest(g, opp.ID, "Forest", "", "Basic Land — Forest")
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Primal Bellow", "Instant", zenPrimalBell, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5 for three Forests", p, tough)
	}
}

func TestS82OldB496LethargyTrapNeedsThreeAttackersAndShrinksOnlyThem(t *testing.T) {
	setup := func(t *testing.T, attackers int) (g *game.Game, def *game.Player, id uuid.UUID, raiders []uuid.UUID, bystander uuid.UUID) {
		g = effectstest.NewCatalogGame(t)
		atk, d := effectstest.SeatsForCombat(g)
		def = d
		for i := 0; i < 3; i++ {
			raiders = append(raiders, effectstest.B12Creature(g, atk.ID, "Raider", "Creature — Bear", 4, 4))
		}
		bystander = effectstest.B12Creature(g, atk.ID, "Bystander", "Creature — Bear", 4, 4)
		effectstest.AdvanceToMain(t, g)
		id = uuid.New()
		def.Hand.PushTop(game.Card{InstanceID: id, Name: "Lethargy Trap", TypeLine: "Instant — Trap", ManaCost: "{3}{U}", OracleID: zenLethargy, Owner: def.ID, Controller: def.ID})
		if attackers > 0 {
			effectstest.DeclareAttack(t, g, def.ID, raiders[:attackers]...)
		}
		effectstest.FloatForTest(g, def, "U")
		return g, def, id, raiders, bystander
	}
	for _, n := range []int{0, 2} {
		g, def, id, _, _ := setup(t, n)
		if err := g.CastSpell(def.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "trap"}); err == nil {
			t.Errorf("the trap cost was offered with %d attackers", n)
		}
	}
	g, def, id, raiders, bystander := setup(t, 3)
	if err := g.CastSpell(def.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "trap"}); err != nil {
		t.Fatalf("the trap cost was refused with three attackers: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, r := range raiders {
		if p, tough := effectstest.PtOf(t, g, r); p != 1 || tough != 4 {
			t.Errorf("an attacker is %d/%d, want 1/4", p, tough)
		}
	}
	if p, _ := effectstest.PtOf(t, g, bystander); p != 4 {
		t.Errorf("a non-attacking creature got -3/-0: power %d", p)
	}
}

func TestS82OldB496TimbermawLarvaGrowsPerForestWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	atk, def := effectstest.SeatsForCombat(g)
	for i := 0; i < 2; i++ {
		effectstest.PushPermanentForTest(g, atk.ID, "Forest", "", "Basic Land — Forest")
	}
	effectstest.PushPermanentForTest(g, def.ID, "Forest", "", "Basic Land — Forest")
	larva := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Timbermaw Larva", TypeLine: "Creature — Beast",
		OracleID: zenLarva, Power: 2, Toughness: 2, Owner: atk.ID, Controller: atk.ID})
	effectstest.DeclareAttack(t, g, def.ID, larva)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, larva); p != 4 || tough != 4 {
		t.Errorf("the attacking Larva is %d/%d, want 4/4 for two Forests", p, tough)
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	effectstest.Settle(t, g)
	if p, tough := effectstest.PtOf(t, g, larva); p != 2 || tough != 2 {
		t.Errorf("after the turn the Larva is %d/%d, want 2/2", p, tough)
	}
}

func TestS82OldB496ArmamentMasterGivesOtherKorCreaturesTwoPerEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	master := effectstest.PushCatalogPermanent(g, me.ID, "Armament Master", "Creature — Kor Soldier", zenArmament, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == master {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 2
			}
		}
	})
	kor := effectstest.B12Creature(g, me.ID, "Kor Scout", "Creature — Kor Scout", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tough := effectstest.PtOf(t, g, kor); p != 1 || tough != 1 {
		t.Fatalf("with no Equipment the Kor is %d/%d, want 1/1", p, tough)
	}
	swords := []uuid.UUID{effectstest.SeedEquipment(g, me.ID, "Sword A", ""), effectstest.SeedEquipment(g, me.ID, "Sword B", "")}
	attach := func(eq, host uuid.UUID) {
		g.WithWriteLock(func() {
			if err := g.AttachForEffect(eq, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
				t.Fatalf("attach: %v", err)
			}
		})
	}
	attach(swords[0], master)
	if p, tough := effectstest.PtOf(t, g, kor); p != 3 || tough != 3 {
		t.Errorf("with one Equipment the Kor is %d/%d, want 3/3", p, tough)
	}
	attach(swords[1], master)
	if p, tough := effectstest.PtOf(t, g, kor); p != 5 || tough != 5 {
		t.Errorf("with two Equipment the Kor is %d/%d, want 5/5", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, master); p != 2 || tough != 2 {
		t.Errorf("the Master itself is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("a non-Kor is %d/%d, want 2/2", p, tough)
	}
	// Equipment on the Kor is not on the Master.
	attach(swords[1], kor)
	if p, tough := effectstest.PtOf(t, g, kor); p != 3 || tough != 3 {
		t.Errorf("with one Equipment on the Master the Kor is %d/%d, want 3/3", p, tough)
	}
}

func TestS82OldB496TurntimberRangerMakesAWolfAndGrowsOnAYes(t *testing.T) {
	for _, yes := range []bool{true, false} {
		t.Run(fmt.Sprintf("yes=%v", yes), func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			ranger := effectstest.TdCastFromHand(t, g, game.Card{Name: "Turntimber Ranger", TypeLine: "Creature — Elf Scout Ranger Ally", OracleID: zenTurntimber, Power: 2, Toughness: 2}, game.CastSpellParams{})
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
			effectstest.PassPriorityAroundTable(t, g)
			wolves := 0
			for _, c := range g.Battlefield.Cards {
				if c.Controller == me.ID && c.IsToken() && c.Name == "Wolf" {
					wolves++
				}
			}
			counters := effectstest.CountersOn(g, ranger, game.CounterPlusOne)
			if yes && (wolves != 1 || counters != 1) {
				t.Errorf("a yes made %d Wolves and %d counters, want 1 and 1", wolves, counters)
			}
			if !yes && (wolves != 0 || counters != 0) {
				t.Errorf("a no made %d Wolves and %d counters, want 0 and 0", wolves, counters)
			}
		})
	}
}

func TestS82OldB496TurntimberRangerTriggersForAnotherAllyToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ranger := effectstest.B12Push(g, me.ID, "Turntimber Ranger", "Creature — Elf Scout Ranger Ally", zenTurntimber, 2, 2)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "New Ally", TypeLine: "Creature — Kor Ally", Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ranger, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after another Ally entered, want 1", got)
	}
}

func TestS82OldB496LullmageMentorCountersASpellByTappingSevenMerfolkAndMakesAMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mentor := effectstest.PushCatalogPermanent(g, me.ID, "Lullmage Mentor", "Creature — Merfolk Wizard", zenLullmage, false)
	var merfolk []uuid.UUID
	for i := 0; i < 5; i++ {
		merfolk = append(merfolk, effectstest.B12Creature(g, me.ID, "Merfolk Token", "Creature — Merfolk", 1, 1))
	}
	effectstest.FillPool(me, 1)
	spell := effectstest.CastFromHandForTest(t, g, me, "Spark", "Instant", "{1}", "", game.CastSpellParams{Strict: true})
	// Six untapped Merfolk are not enough.
	six := append([]uuid.UUID{mentor}, merfolk...)
	if err := g.ActivateCatalogAbility(me.ID, mentor, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardRefs(spell), TapIDs: six}); err == nil {
		t.Fatal("activated with six untapped Merfolk")
	}
	merfolk = append(merfolk, effectstest.B12Creature(g, me.ID, "Merfolk Token", "Creature — Merfolk", 1, 1))
	seven := append([]uuid.UUID{mentor}, merfolk...)
	if err := g.ActivateCatalogAbility(me.ID, mentor, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardRefs(spell), TapIDs: seven}); err != nil {
		t.Fatalf("activate with seven Merfolk: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(spell) {
		t.Error("the spell was not countered")
	}
	for _, id := range append(merfolk, mentor) {
		if tapped, ok := effectstest.BattlefieldCardTapped(g, id); ok && !tapped {
			t.Error("a Merfolk was left untapped after paying the cost")
		}
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	made := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Merfolk" && c.Power == 1 {
			made++
		}
	}
	if made < 1 {
		t.Error("no 1/1 blue Merfolk token was created for countering a spell")
	}
}

func TestS82OldB496KalitasDestroysACreatureAndMakesAVampireWithItsSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kalitas := effectstest.PushCatalogPermanent(g, me.ID, "Kalitas, Bloodchief of Ghet", "Legendary Creature — Vampire Warrior", zenKalitas, false)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 4)
	effectstest.ApaMana(me, "B", "B", "B")
	effectstest.ApaActivate(t, g, me, kalitas, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not destroyed")
	}
	var found bool
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Vampire" {
			found = true
			if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 3 || tough != 4 {
				t.Errorf("the Vampire is %d/%d, want 3/4", p, tough)
			}
			if len(c.Colors) != 1 || c.Colors[0] != "B" {
				t.Errorf("the Vampire's colours are %v, want black", c.Colors)
			}
		}
	}
	if !found {
		t.Error("no Vampire token was created")
	}
}

func TestS82OldB496KalitasMakesNoVampireWhenTheCreatureDoesNotDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kalitas := effectstest.PushCatalogPermanent(g, me.ID, "Kalitas, Bloodchief of Ghet", "Legendary Creature — Vampire Warrior", zenKalitas, false)
	tough := effectstest.P1g2Indestructible(g, opp.ID, 5, 5)
	effectstest.ApaMana(me, "B", "B", "B")
	effectstest.ApaActivate(t, g, me, kalitas, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(tough)})
	if !g.Battlefield.Contains(tough) {
		t.Fatal("the indestructible creature was destroyed")
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Vampire" {
			t.Error("a Vampire token was created though the creature did not die")
		}
	}
}
