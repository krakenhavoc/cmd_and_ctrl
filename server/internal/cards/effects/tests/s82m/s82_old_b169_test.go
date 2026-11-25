package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b169_test.go — War of the Spark, batch 169 (#5520): damage cards.
// Each test plays the real catalog entry and asserts what its text says.

const (
	s82o169Storrev  = "159098f1-172f-408e-a962-86763a29a8e1"
	s82o169Command  = "198d31e4-1ae1-4bf9-a05a-7fa45c0aeddb"
	s82o169Greeting = "1be474c6-5a9f-4c58-adc6-c1508348adea"
	s82o169Outburst = "5e715875-28e0-44b4-af25-f6bb8ddec823"
	s82o169Bond     = "826f9f42-f2ff-4b9e-be79-a9de6b5b7b38"
	s82o169Kaya     = "8b3febb9-0597-46af-b005-982f1a31c83e"
	s82o169Triumph  = "c0cc881c-b63b-44dd-8157-67680c65fc3d"
)

func TestS82Old169EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82o169Storrev:  "Storrev, Devkarin Lich",
		s82o169Command:  "Command the Dreadhorde",
		s82o169Greeting: "Jaya's Greeting",
		s82o169Outburst: "Ral's Outburst",
		s82o169Bond:     "Bond of Passion",
		s82o169Kaya:     "Oath of Kaya",
		s82o169Triumph:  "Chandra's Triumph",
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

// s82o169Grave puts a creature card in a player's graveyard.
func s82o169Grave(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

// s82o169Walker puts a planeswalker with the given name and subtype on the
// battlefield with 5 loyalty (names differ so the legend rule stays quiet).
func s82o169Walker(g *game.Game, owner uuid.UUID, name, subtype string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Planeswalker — " + subtype,
		Counters: map[string]int{game.CounterLoyalty: 5}, Owner: owner, Controller: owner,
	})
}

// s82o169Settle resolves the stack, picking `pick` if a target prompt opens.
func s82o169Settle(t *testing.T, g *game.Game, who uuid.UUID, pick uuid.UUID) {
	t.Helper()
	if effectstest.LatestPickTarget(g, who) != nil && pick != uuid.Nil {
		effectstest.PickCard(t, g, who, pick)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestS82Old169StorrevReturnsACardNotPutThereThisCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	st := effectstest.PushCatalogPermanent(g, me.ID, "Storrev, Devkarin Lich", "Legendary Creature — Zombie Elf Wizard", s82o169Storrev, false)
	old := s82o169Grave(me, "Old Bear", "Creature — Bear", "{1}{G}")
	fresh := s82o169Grave(me, "Fresh Bear", "Creature — Bear", "{1}{G}")
	g.EmitEvent(game.Event{Kind: game.EventStepBegan, Actor: me.ID, Step: game.StepBeginCombat})
	g.EmitEvent(game.Event{Kind: game.EventLTB, CardID: fresh, Actor: me.ID, OldZone: game.ZoneBattlefield, NewZone: game.ZoneGraveyard})
	effectstest.DealCombatDamageToPlayer(g, st, opp.ID, 5)
	s82o169Settle(t, g, me.ID, old)
	if !me.Hand.Contains(old) {
		t.Error("the card that was already in the graveyard did not return")
	}
	if me.Hand.Contains(fresh) || !me.Graveyard.Contains(fresh) {
		t.Error("the card put into the graveyard this combat was returned")
	}
}

func TestS82Old169StorrevHasNothingToReturnButACardPutThereThisCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	st := effectstest.PushCatalogPermanent(g, me.ID, "Storrev, Devkarin Lich", "Legendary Creature — Zombie Elf Wizard", s82o169Storrev, false)
	fresh := s82o169Grave(me, "Fresh Bear", "Creature — Bear", "{1}{G}")
	g.EmitEvent(game.Event{Kind: game.EventStepBegan, Actor: me.ID, Step: game.StepBeginCombat})
	g.EmitEvent(game.Event{Kind: game.EventLTB, CardID: fresh, Actor: me.ID, OldZone: game.ZoneBattlefield, NewZone: game.ZoneGraveyard})
	effectstest.DealCombatDamageToPlayer(g, st, opp.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(fresh) {
		t.Error("Storrev returned a card put into the graveyard this combat")
	}
}

func TestS82Old169CommandTheDreadhordeCostsTheTotalManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := s82o169Grave(me, "Big Bear", "Creature — Bear", "{3}{G}")
	b := s82o169Grave(opp, "Their Elf", "Creature — Elf", "{1}{G}")
	land := s82o169Grave(opp, "Their Land", "Land", "")
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Command the Dreadhorde", "Sorcery", s82o169Command, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-6 {
		t.Errorf("life %d, want %d (4 + 2 damage)", me.Life, life-6)
	}
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("the chosen cards did not enter the battlefield")
	}
	if c, _ := g.LookupCardForEffect(b); c.Controller != me.ID {
		t.Error("their Elf did not enter under the caster's control")
	}
	if g.Battlefield.Contains(land) {
		t.Error("an unchosen card entered")
	}
}

func TestS82Old169CommandTheDreadhordeRefusesALandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := s82o169Grave(opp, "Their Land", "Land", "")
	err := effectstest.CastCatalogSpellErr(t, g, "Command the Dreadhorde", "Sorcery", s82o169Command, effectstest.CardRefs(land))
	if err == nil {
		t.Error("a land card was accepted as a target")
	}
}

func TestS82Old169JayasGreetingDamagesAndScrys(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 3)
	effectstest.CastCatalogSpell(t, g, "Jaya's Greeting", "Instant", s82o169Greeting, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("no scry 1")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(big) {
		t.Error("the 3-toughness creature survived 3 damage")
	}
}

func TestS82Old169JayasGreetingRefusesAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	if err := effectstest.CastCatalogSpellErr(t, g, "Jaya's Greeting", "Instant", s82o169Greeting, effectstest.PlayerRefs(opp.ID)); err == nil {
		t.Error("a player was accepted as a target")
	}
}

func TestS82Old169RalsOutburstDamagesThenSortsTheTopTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three")
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Ral's Outburst", "Instant", s82o169Outburst, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-3)
	}
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 2 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("pick = %+v, want exactly one of the top two", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{pick.ChooseCards[0]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	hand, yard := 0, 0
	for _, id := range ids {
		if me.Hand.Contains(id) {
			hand++
		}
		if me.Graveyard.Contains(id) {
			yard++
		}
	}
	if hand != 1 || yard != 1 {
		t.Errorf("%d in hand, %d in the graveyard, want 1 and 1", hand, yard)
	}
}

func TestS82Old169BondOfPassionStealsAndDamagesAnotherTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Bond of Passion", "Sorcery", s82o169Bond,
		append(effectstest.CardRefs(bear), effectstest.PlayerRefs(opp.ID)...))
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(bear); c.Controller != me.ID {
		t.Error("the creature was not stolen")
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
}

func TestS82Old169BondOfPassionCannotDamageTheStolenCreatureItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	err := effectstest.CastCatalogSpellErr(t, g, "Bond of Passion", "Sorcery", s82o169Bond, effectstest.CardRefs(bear, bear))
	if err == nil {
		t.Error("the same creature was accepted for both clauses")
	}
}

func TestS82Old169OathOfKayaEntersForThreeDamageAndThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Oath of Kaya", "Legendary Enchantment", s82o169Kaya, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-3 || me.Life != life+3 {
		t.Errorf("life: opponent %d (want %d), you %d (want %d)", opp.Life, theirs-3, me.Life, life+3)
	}
}

func TestS82Old169OathOfKayaPunishesAnAttackOnAPlaneswalkerOncePerPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Oath of Kaya", "Legendary Enchantment", s82o169Kaya, false)
	pw1 := s82o169Walker(g, me.ID, "Walker One", "Test")
	pw2 := s82o169Walker(g, me.ID, "Walker Two", "Test")
	a1 := effectstest.B12Creature(g, opp.ID, "Attacker One", "Creature — Bear", 2, 2)
	a2 := effectstest.B12Creature(g, opp.ID, "Attacker Two", "Creature — Bear", 2, 2)
	life, theirs := me.Life, opp.Life
	g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: a1, Actor: opp.ID, Target: pw1})
	g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: a2, Actor: opp.ID, Target: pw1})
	g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: a2, Actor: opp.ID, Target: pw2})
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-4 {
		t.Errorf("attacker life %d, want %d (two planeswalkers attacked, 2 each)", opp.Life, theirs-4)
	}
	if me.Life != life+4 {
		t.Errorf("your life %d, want %d", me.Life, life+4)
	}
}

func TestS82Old169OathOfKayaIgnoresAnAttackOnYouAndYourOwnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Oath of Kaya", "Legendary Enchantment", s82o169Kaya, false)
	pw := s82o169Walker(g, opp.ID, "Their Walker", "Test")
	a1 := effectstest.B12Creature(g, opp.ID, "Attacker One", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	life, theirs := me.Life, opp.Life
	g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: a1, Actor: opp.ID, Target: me.ID})
	g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: mine, Actor: me.ID, Target: pw})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs || me.Life != life {
		t.Errorf("life moved (%d, %d) for an attack that is not on your planeswalker", me.Life-life, opp.Life-theirs)
	}
}

func TestS82Old169ChandrasTriumphDealsFiveWithAChandra(t *testing.T) {
	for _, tc := range []struct {
		name    string
		chandra bool
		dead    bool
	}{{"no Chandra, three damage leaves a 4/5", false, false}, {"a Chandra, five damage kills it", true, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			if tc.chandra {
				s82o169Walker(g, me.ID, "Chandra, Test", "Chandra")
			}
			big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 5)
			effectstest.CastCatalogSpell(t, g, "Chandra's Triumph", "Instant", s82o169Triumph, effectstest.CardRefs(big))
			effectstest.PassPriorityAroundTable(t, g)
			g.RunStateChecksForTest()
			if g.Battlefield.Contains(big) == tc.dead {
				t.Errorf("creature on battlefield = %v, want %v", !tc.dead, !tc.dead)
			}
		})
	}
}

func TestS82Old169ChandrasTriumphRefusesYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Chandra's Triumph", "Instant", s82o169Triumph, effectstest.CardRefs(mine)); err == nil {
		t.Error("your own creature was accepted")
	}
}
