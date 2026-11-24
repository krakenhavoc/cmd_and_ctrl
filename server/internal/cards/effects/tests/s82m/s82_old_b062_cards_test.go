package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b062_cards_test.go — Kaldheim, batch 062: each test plays the real
// catalog entry through the engine.

const (
	oldB062Narfi    = "18beb8eb-6f67-482f-9a7d-282f1e7ecbc1"
	oldB062Glimpse  = "27065d34-b22a-47af-aea2-980f47bafcea"
	oldB062Skald    = "34bccfc0-453d-4d1a-8046-70f9c3065e88"
	oldB062Blood    = "45f386ba-ad62-4c0e-b6ad-ae02ca8f2bec"
	oldB062WeighDn  = "51f1ceb9-2305-4a4e-b422-7dd7f8deef98"
	oldB062Reaper   = "70c268af-3fb4-42b8-a87a-d06d33c1c641"
	oldB062Tergrid  = "8485cfaa-1dbf-432b-b5d0-92a6aa6a329b"
	oldB062Draugr   = "8727284f-e9a6-49ea-9fb8-1076883ceecf"
	oldB062Egon     = "8fa7bd8f-f712-4dd1-9ffb-e30dfe744389"
	oldB062Jorn     = "bd43b485-e42e-4c20-adcb-799b0ce18a59"
	oldB062SnowType = "Basic Snow Land — Island"
)

func TestOldB062EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB062Narfi:          "Narfi, Betrayer King",
		oldB062Glimpse:        "Glimpse the Cosmos",
		oldB062Skald:          "Master Skald",
		oldB062Blood:          "Blood on the Snow",
		oldB062WeighDn:        "Weigh Down",
		oldB062Reaper:         "Cleaving Reaper",
		oldB062Tergrid:        "Tergrid, God of Fright",
		oldB062Tergrid + "#1": "Tergrid's Lantern",
		oldB062Draugr:         "Raise the Draugr",
		oldB062Egon:           "Egon, God of Death",
		oldB062Egon + "#1":    "Throne of Death",
		oldB062Jorn:           "Jorn, God of Winter",
		oldB062Jorn + "#1":    "Kaldring, the Rimestaff",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v, want Full with no caveats", name, spec, ok)
		}
	}
}

func oldB062Lands(g *game.Game, owner uuid.UUID, n int, snow bool, basic string) {
	for i := 0; i < n; i++ {
		if snow {
			effectstest.BlbB1Push(g, owner, "Snow-Covered "+basic, "Basic Snow Land — "+basic, "", 0, 0)
		} else {
			effectstest.BlbB1Push(g, owner, basic, "Basic Land — "+basic, "", 0, 0)
		}
	}
}

func TestOldB062NarfiPumpsOtherSnowAndZombieCreaturesAndReturnsForSnowMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	narfi := effectstest.B12Push(g, me.ID, "Narfi, Betrayer King", "Legendary Snow Creature — Zombie Wizard", oldB062Narfi, 4, 3)
	zombie := effectstest.B12Creature(g, me.ID, "Zed", "Creature — Zombie", 2, 2)
	snow := effectstest.B12Creature(g, me.ID, "Yeti", "Snow Creature — Yeti", 2, 2)
	both := effectstest.B12Creature(g, me.ID, "Snow Zombie", "Snow Creature — Zombie", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"Narfi", narfi, 4}, {"Zombie", zombie, 3}, {"snow", snow, 3}, {"snow Zombie", both, 3}, {"Bear", bear, 2}, {"their Zombie", theirs, 2}} {
		if p, _ := effectstest.PtOf(t, g, tc.id); p != tc.want {
			t.Errorf("%s has power %d, want %d", tc.name, p, tc.want)
		}
	}

	// From the graveyard: three snow mana return it tapped; plain mana cannot pay.
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, narfi, 9) })
	g.RunStateChecksForTest()
	if !me.Graveyard.Contains(narfi) {
		t.Fatal("Narfi did not die")
	}
	oldB062Lands(g, me.ID, 3, false, "Island")
	if err := g.ActivateCatalogAbility(me.ID, narfi, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("plain mana paid for {S}{S}{S}")
	}
	oldB062Lands(g, me.ID, 3, true, "Island")
	if err := g.ActivateCatalogAbility(me.ID, narfi, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, narfi)
	if !ok || !c.Tapped {
		t.Errorf("Narfi back on the battlefield = %v, tapped = %v, want true/true", ok, c.Tapped)
	}
}

func TestOldB062GlimpseTheCosmosTakesOneOfThreeAndRecastsFromGraveyardWithAGiant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	effectstest.ToMainForCost(t, g)
	oldB062Lands(g, me.ID, 2, false, "Island")
	id := effectstest.HandCardFull(me, "Glimpse the Cosmos", "Sorcery", "{1}{U}", oldB062Glimpse, []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("want a mandatory pick of one of three, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Fatal("Two did not reach the hand")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no bottom-order prompt for the rest")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[0], ids[2]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Fatal("the Glimpse is not in the graveyard")
	}

	// Without a Giant the graveyard cast is refused.
	oldB062Lands(g, me.ID, 1, false, "Island")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true, FromZone: "graveyard", AlternativeCost: "glimpse-the-cosmos"}); err == nil {
		t.Fatal("cast from the graveyard with no Giant")
	}
	// With one it costs {U}, and the spell is exiled afterwards.
	effectstest.B12Creature(g, me.ID, "Giant", "Creature — Giant", 3, 3)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true, FromZone: "graveyard", AlternativeCost: "glimpse-the-cosmos"}); err != nil {
		t.Fatalf("cast from the graveyard with a Giant: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(id) {
		t.Error("the recast Glimpse went back to the graveyard instead of exile")
	}
}

func TestOldB062MasterSkaldExilesACreatureCardToReturnAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	creature := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	relic := effectstest.PushGraveyardCardTyped(me, "Old Relic", "Artifact")
	effectstest.CastRingCreature(t, g, "Master Skald", "Creature — Dwarf Warrior", oldB062Skald, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, relic)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("pick %+v, want up to one creature card", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{creature}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) {
		t.Error("the artifact card did not return to hand")
	}
	if me.Graveyard.Contains(creature) {
		t.Error("the creature card was not exiled")
	}
}

func TestOldB062MasterSkaldDeclinedReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	relic := effectstest.PushGraveyardCardTyped(me, "Old Relic", "Artifact")
	effectstest.CastRingCreature(t, g, "Master Skald", "Creature — Dwarf Warrior", oldB062Skald, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, relic)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(relic) || !me.Graveyard.Contains(relic) {
		t.Error("the artifact card returned although no creature card was exiled")
	}
}

func TestOldB062WeighDownNeedsACreatureCardToExileAndShrinksTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.ToMainForCost(t, g)
	effectstest.B06AddMana(me, "B")
	id := effectstest.HandCardFull(me, "Weigh Down", "Sorcery", "{B}", oldB062WeighDn, []string{"B"})
	target := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("cast with no creature card in the graveyard")
	}
	fuel := effectstest.PushGraveyardCardTyped(me, "Fuel", "Creature — Zombie")
	land := effectstest.PushGraveyardCardTyped(me, "Mountain", "Basic Land — Mountain")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: target, ExileIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land card paid the creature-card cost")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: target, ExileIDs: []uuid.UUID{fuel}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(fuel) {
		t.Error("the creature card is still in the graveyard")
	}
	if p, tough := effectstest.PtOf(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("the Ogre is %d/%d, want 1/1", p, tough)
	}
}

func TestOldB062RaiseTheDraugrReturnsOneOrTwoSharingCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elf := effectstest.PushGraveyardCardTyped(me, "Elf", "Creature — Elf Druid")
	warrior := effectstest.PushGraveyardCardTyped(me, "Elf Warrior", "Creature — Elf Warrior")
	bear := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	effectstest.ToMainForCost(t, g)
	effectstest.B06AddMana(me, "B", "B", "B", "B")
	id := effectstest.HandCardFull(me, "Raise the Draugr", "Instant", "{1}{B}", oldB062Draugr, []string{"B"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{1}, Targets: effectstest.CardRefs(elf, bear)}); err == nil {
		t.Fatal("two creature cards that share no type were accepted")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{1}, Targets: effectstest.CardRefs(elf, warrior)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(elf) || !me.Hand.Contains(warrior) || me.Hand.Contains(bear) {
		t.Error("the two Elves, and only they, should be back in hand")
	}
}

func TestOldB062CleavingReaperReturnsOnlyAfterAnAngelOrBerserkerEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	reaper := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: reaper, Name: "Cleaving Reaper", TypeLine: "Creature — Angel Berserker", OracleID: oldB062Reaper,
		Power: 5, Toughness: 3, Owner: me.ID, Controller: me.ID})
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, reaper, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no Angel or Berserker having entered")
	}
	effectstest.CastRingCreature(t, g, "Berserker Bob", "Creature — Human Berserker", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, reaper, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(reaper) {
		t.Error("the Reaper did not return to hand")
	}
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
}

func TestOldB062BloodOnTheSnowDestroysThenReturnsAGraveyardCardWithinTheSnowManaSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	pw := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 4}, Owner: opp.ID, Controller: opp.ID})
	cheap := game.Card{InstanceID: uuid.New(), Name: "Cheap Elf", TypeLine: "Creature — Elf", ManaCost: "{1}{G}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	dear := game.Card{InstanceID: uuid.New(), Name: "Dear Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}{R}", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(cheap)
	me.Graveyard.PushTop(dear)
	effectstest.ToMainForCost(t, g)
	oldB062Lands(g, me.ID, 3, true, "Swamp")
	oldB062Lands(g, me.ID, 3, false, "Swamp")
	id := effectstest.HandCardFull(me, "Blood on the Snow", "Snow Sorcery", "{4}{B}{B}", oldB062Blood, []string{"B"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true, Modes: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("the creatures were not all destroyed")
	}
	if !g.Battlefield.Contains(pw) {
		t.Error("the planeswalker was destroyed by the creature mode")
	}
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no return prompt")
	}
	if effectstest.HasID(pick.ChooseCards, dear.InstanceID) {
		t.Error("a mana value 6 card is offered with only 3 snow mana spent")
	}
	if !effectstest.HasID(pick.ChooseCards, cheap.InstanceID) || !effectstest.HasID(pick.ChooseCards, mine) {
		t.Errorf("the mana value 3 card and the creature the wipe just killed should be offered: %v", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{cheap.InstanceID}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(cheap.InstanceID) {
		t.Error("the chosen creature card did not return to the battlefield")
	}
}

func TestOldB062BloodOnTheSnowPlaneswalkerModeSparesCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	pw := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 4}, Owner: opp.ID, Controller: opp.ID})
	effectstest.ToMainForCost(t, g)
	oldB062Lands(g, me.ID, 6, false, "Swamp")
	id := effectstest.HandCardFull(me, "Blood on the Snow", "Snow Sorcery", "{4}{B}{B}", oldB062Blood, []string{"B"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true, Modes: []int{1}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("the planeswalker mode destroyed a creature")
	}
	if g.Battlefield.Contains(pw) {
		t.Error("the planeswalker survived")
	}
}

func TestOldB062TergridTakesWhatAnOpponentSacrificesOrDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Tergrid, God of Fright", "Legendary Creature — God", oldB062Tergrid, false)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.SacrificeNow(t, g, opp.ID, victim)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if c, ok := effectstest.BattlefieldCard(g, victim); !ok || c.Controller != me.ID {
		t.Error("the sacrificed card is not on the battlefield under my control")
	}

	// A discarded permanent card comes too; a nonpermanent one does not.
	bolt := uuid.New()
	opp.Hand.Cards = nil
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if g.Battlefield.Contains(bolt) {
		t.Error("a discarded instant came to the battlefield")
	}
	rock := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: rock, Name: "Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if c, ok := effectstest.BattlefieldCard(g, rock); !ok || c.Controller != me.ID {
		t.Error("the discarded artifact card did not come to my battlefield")
	}

	// My own sacrifice does nothing, and "no" declines.
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.SacrificeNow(t, g, me.ID, mine)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if g.Battlefield.Contains(mine) {
		t.Error("my own sacrifice triggered Tergrid")
	}
	other := effectstest.B12Creature(g, opp.ID, "Another Bear", "Creature — Bear", 2, 2)
	effectstest.SacrificeNow(t, g, opp.ID, other)
	effectstest.AssistCardSettle(t, g, me.ID, nil, false)
	if g.Battlefield.Contains(other) {
		t.Error("declining the may still returned the card")
	}
}

func TestOldB062TergridsLanternMakesThemChooseAndUntapsForFourMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lantern := effectstest.PushCatalogPermanent(g, me.ID, "Tergrid's Lantern", "Legendary Artifact", oldB062Tergrid+"#1", false)
	victimLand := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	target := &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	activate := func() {
		t.Helper()
		if err := g.ActivateCatalogAbility(me.ID, lantern, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{*target}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	// Lose 3 life.
	life := opp.Life
	activate()
	c := effectstest.LatestOptionPickFor(g, opp.ID)
	if c == nil || len(c.PickOptions) != 3 {
		t.Fatalf("want three options (life, sacrifice, discard), got %+v", c)
	}
	if err := g.ResolveOptionPick(c.ID, opp.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
	// Untap it for {3}{B} and sacrifice instead.
	effectstest.B06AddMana(me, "B", "B", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, lantern, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("untap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if card, _ := effectstest.BattlefieldCard(g, lantern); card.Tapped {
		t.Fatal("the Lantern is still tapped")
	}
	activate()
	c = effectstest.LatestOptionPickFor(g, opp.ID)
	if c == nil {
		t.Fatal("no choice")
	}
	if err := g.ResolveOptionPick(c.ID, opp.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, victimLand)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victimLand) {
		t.Error("the chosen nonland permanent was not sacrificed")
	}
	if opp.Life != life-3 {
		t.Errorf("life %d after sacrificing, want unchanged %d", opp.Life, life-3)
	}
}

func TestOldB062EgonExilesTwoGraveyardCardsAtUpkeepOrIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	egon := effectstest.PushCatalogPermanent(g, me.ID, "Egon, God of Death", "Legendary Creature — God", oldB062Egon, false)
	a := effectstest.PushGraveyardCardTyped(me, "A", "Creature — Bear")
	b := effectstest.PushGraveyardCardTyped(me, "B", "Creature — Bear")
	c := effectstest.PushGraveyardCardTyped(me, "C", "Creature — Bear")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) || me.Graveyard.Contains(a) || !me.Graveyard.Contains(c) {
		t.Error("the two chosen cards, and only they, should be exiled")
	}
	if !g.Battlefield.Contains(egon) {
		t.Error("Egon was sacrificed with enough cards in the graveyard")
	}

	// Next upkeep one card is left: Egon is sacrificed and a card is drawn.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(egon) {
		t.Error("Egon survived an upkeep it couldn't pay")
	}
	if got := me.Hand.Size(); got != hand+1 { // only Egon's draw: the draw step has not come yet
		t.Errorf("hand %d -> %d, want the one card Egon draws", hand, got)
	}
}

func TestOldB062ThroneOfDeathMillsAtUpkeepAndDrawsForACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	throne := effectstest.PushCatalogPermanent(g, me.ID, "Throne of Death", "Legendary Artifact", oldB062Egon+"#1", false)
	before := me.Graveyard.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != before+1 {
		t.Errorf("graveyard %d -> %d, want one card milled", before, got)
	}
	fuel := effectstest.PushGraveyardCardTyped(me, "Fuel", "Creature — Zombie")
	land := effectstest.PushGraveyardCardTyped(me, "Mountain", "Basic Land — Mountain")
	effectstest.B06AddMana(me, "B", "B", "B")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land card paid the creature-card cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, throne, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{fuel}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
	if me.Graveyard.Contains(fuel) {
		t.Error("the creature card was not exiled")
	}
}

func TestOldB062JornUntapsEverySnowPermanentWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jorn := effectstest.PushCatalogPermanent(g, me.ID, "Jorn, God of Winter", "Legendary Snow Creature — God", oldB062Jorn, false)
	snow := effectstest.BlbB1Push(g, me.ID, "Snow-Covered Island", oldB062SnowType, "", 0, 0)
	plain := effectstest.BlbB1Push(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{snow, plain} {
			_ = g.TapTargetForEffect(id)
		}
	})
	effectstest.DeclareAttack(t, g, opp.ID, jorn)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, snow); c.Tapped {
		t.Error("the snow land is still tapped")
	}
	if c, _ := effectstest.BattlefieldCard(g, plain); !c.Tapped {
		t.Error("the plain land was untapped")
	}
}

func TestOldB062KaldringLetsYouPlayASnowPermanentCardFromYourGraveyardTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	staff := effectstest.PushCatalogPermanent(g, me.ID, "Kaldring, the Rimestaff", "Legendary Snow Artifact", oldB062Jorn+"#1", false)
	snowLand := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: snowLand, Name: "Snow-Covered Island", TypeLine: oldB062SnowType, Owner: me.ID, Controller: me.ID})
	plainLand := effectstest.PushGraveyardCardTyped(me, "Island", "Basic Land — Island")
	effectstest.ToMainForCost(t, g)
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plainLand}}}); err == nil {
		t.Fatal("a nonsnow card was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: snowLand}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, snowLand, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("play the land from the graveyard: %v", err)
	}
	c, ok := effectstest.BattlefieldCard(g, snowLand)
	if !ok || !c.Tapped {
		t.Errorf("the land is on the battlefield = %v, tapped = %v, want true/true", ok, c.Tapped)
	}
}
