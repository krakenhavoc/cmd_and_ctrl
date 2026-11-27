package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b555_test.go — Lorwyn, evoke and graveyard recursion, batch 555
// (#5593). Every test plays the real catalog entry through the engine.

const (
	oldB555Mournwhelk = "10277cc5-f53b-4b33-bd68-edb0534a2040"
	oldB555Acolyte    = "699812af-4dba-49dc-987d-67e472a68971"
	oldB555Dawnfluke  = "9ca9168f-54cb-4091-b80c-7c264773c1a2"
	oldB555Glare      = "9d3136da-3efa-4103-9d35-8da87d318544"
	oldB555Wispmare   = "a3bb8ce0-e61e-438d-8875-32473bb5b61c"
	oldB555School     = "00909d51-3da8-41eb-a535-0c325fedc59f"
	oldB555Rite       = "030e9e3e-ce78-4c3b-bd83-969c07c70459"
	oldB555Vinebreed  = "759e454e-ee8c-4b0e-9503-26e4c0a2a8f9"
	oldB555Pilferers  = "9f90dcd1-e372-40c5-9de2-81392b0add06"
	oldB555Bear       = "Creature — Bear"
)

func TestOldB555EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB555Mournwhelk: "Mournwhelk", oldB555Acolyte: "Inner-Flame Acolyte", oldB555Dawnfluke: "Dawnfluke",
		oldB555Glare: "Glarewielder", oldB555Wispmare: "Wispmare", oldB555School: "Summon the School",
		oldB555Rite: "Boggart Birth Rite", oldB555Vinebreed: "Scarred Vinebreeder", oldB555Pilferers: "Warren Pilferers",
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

func TestOldB555EvokeCostsAreThePrintedOnes(t *testing.T) {
	for oracle, cost := range map[string]string{
		oldB555Mournwhelk: "{3}{B}", oldB555Acolyte: "{R}", oldB555Dawnfluke: "{W}", oldB555Glare: "{1}{R}", oldB555Wispmare: "{W}",
	} {
		if ac := game.AlternativeCostByKey(oracle, "evoke"); ac == nil || ac.ManaCost != cost {
			t.Errorf("%s: evoke is %+v, want %s", oracle, ac, cost)
		}
	}
}

func oldB555TokensOf(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// oldB555Evoke casts the card for its evoke cost, answers the trigger's targets
// and lets everything resolve.
func oldB555Evoke(t *testing.T, g *game.Game, name, oracle string, targets ...uuid.UUID) uuid.UUID {
	t.Helper()
	return oldB555EvokeWith(t, g, name, oracle, func(me *game.Player) {
		if len(targets) > 0 {
			effectstest.AnswerTriggerTargets(t, g, me.ID, targets...)
		}
	})
}

func oldB555EvokeWith(t *testing.T, g *game.Game, name, oracle string, answer func(me *game.Player)) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastWithAltCost(t, g, name, "Creature — Elemental", oracle, "evoke")
	effectstest.PassPriorityAroundTable(t, g)
	answer(me)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB555MournwhelkMakesTheTargetDiscardTwoWhetherCastOrEvoked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	var spare []uuid.UUID
	for i := 0; i < 4; i++ {
		spare = append(spare, effectstest.B13HandCard(opp, "Spare", "Sorcery", ""))
	}
	before := opp.Hand.Size()
	id := effectstest.CastCatalogSpell(t, g, "Mournwhelk", "Creature — Elemental", oldB555Mournwhelk, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets2(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, spare[0], spare[1])
	if opp.Hand.Size() != before-2 {
		t.Errorf("the opponent has %d cards in hand, want %d", opp.Hand.Size(), before-2)
	}
	if !g.Battlefield.Contains(id) {
		t.Error("a hard-cast Mournwhelk was sacrificed")
	}
	// Evoked: the discard still happens, then the sacrifice.
	g2 := effectstest.NewCatalogGame(t)
	opp2 := g2.Seats[1]
	var spare2 []uuid.UUID
	for i := 0; i < 3; i++ {
		spare2 = append(spare2, effectstest.B13HandCard(opp2, "Spare", "Sorcery", ""))
	}
	before2 := opp2.Hand.Size()
	id2 := oldB555EvokeWith(t, g2, "Mournwhelk", oldB555Mournwhelk, func(me *game.Player) {
		effectstest.AnswerTriggerTargets2(t, g2, me.ID, opp2.ID)
	})
	effectstest.AnswerDiscard(t, g2, opp2.ID, spare2[0], spare2[1])
	effectstest.PassPriorityAroundTable(t, g2)
	if opp2.Hand.Size() != before2-2 {
		t.Errorf("evoked: the opponent has %d cards in hand, want %d", opp2.Hand.Size(), before2-2)
	}
	if g2.Battlefield.Contains(id2) {
		t.Error("an evoked Mournwhelk was not sacrificed")
	}
}

func TestOldB555InnerFlameAcolytePumpsAndGivesHasteEvenEvoked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB555Bear, 2, 2)
	acolyte := oldB555Evoke(t, g, "Inner-Flame Acolyte", oldB555Acolyte, bear)
	if p, tt := effectstest.PtOf(t, g, bear); p != 4 || tt != 2 {
		t.Errorf("the Bear is %d/%d, want 4/2", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the Bear has no haste")
	}
	if g.Battlefield.Contains(acolyte) {
		t.Error("the evoked Acolyte was not sacrificed")
	}
}

func TestOldB555DawnflukePreventsTheNextThreeDamageAndHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB555Bear, 5, 5)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", oldB555Bear, 5, 5)
	effectstest.CastCatalogSpell(t, g, "Dawnfluke", "Creature — Elemental", oldB555Dawnfluke, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DealDamageTo(t, g, foe, bear, 5)
	if got := effectstest.DamageMarkedOn(g, bear); got != 2 {
		t.Errorf("the shielded Bear has %d damage, want 2 (5 minus 3 prevented)", got)
	}
	effectstest.DealDamageTo(t, g, foe, bear, 1)
	if got := effectstest.DamageMarkedOn(g, bear); got != 3 {
		t.Errorf("the Bear has %d damage after the shield was used up, want 3", got)
	}
	spec, _ := effects.Lookup(oldB555Dawnfluke)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
}

func TestOldB555GlarewielderStopsTwoCreaturesBlockingAndEvokeStillDoesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", oldB555Bear, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", oldB555Bear, 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", oldB555Bear, 2, 2)
	glare := oldB555Evoke(t, g, "Glarewielder", oldB555Glare, a, b)
	if !effectstest.SbCantBlock(t, g, a) || !effectstest.SbCantBlock(t, g, b) {
		t.Error("a targeted creature can still block")
	}
	if effectstest.SbCantBlock(t, g, c) {
		t.Error("an untargeted creature can't block")
	}
	if g.Battlefield.Contains(glare) {
		t.Error("the evoked Glarewielder was not sacrificed")
	}
}

func TestOldB555WispmareDestroysATargetEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ench := effectstest.B12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Wispmare", "Creature — Elemental", oldB555Wispmare, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, ench)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, ench) {
		t.Error("the enchantment survived")
	}
	if !effectstest.BattlefieldHas(g, rock) {
		t.Error("the artifact was destroyed")
	}
}

func TestOldB555SummonTheSchoolMakesTwoMerfolkAndReturnsFromTheGraveyardForFourTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	card := effectstest.CastCatalogSpell(t, g, "Summon the School", "Kindred Sorcery — Merfolk", oldB555School, nil)
	effectstest.PassPriorityAroundTable(t, g)
	wizards := oldB555TokensOf(g, me.ID, "Merfolk Wizard")
	if len(wizards) != 2 {
		t.Fatalf("%d Merfolk Wizard tokens, want 2", len(wizards))
	}
	if !me.Graveyard.Contains(card) {
		t.Fatal("the spell is not in the graveyard")
	}
	two := []uuid.UUID{wizards[0].InstanceID, wizards[1].InstanceID}
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true, TapIDs: two}); err == nil {
		t.Fatal("returned the card for two Merfolk")
	}
	m1 := effectstest.B12Creature(g, me.ID, "Merman One", "Creature — Merfolk", 1, 1)
	m2 := effectstest.B12Creature(g, me.ID, "Merman Two", "Creature — Merfolk", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB555Bear, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true, TapIDs: append(append([]uuid.UUID{}, two...), m1, bear)}); err == nil {
		t.Fatal("tapped a non-Merfolk for the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true, TapIDs: append(append([]uuid.UUID{}, two...), m1, m2)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, card) {
		t.Error("the card did not return to hand")
	}
	for _, id := range append(two, m1, m2) {
		if !effectstest.TappedNow(g, id) {
			t.Error("a Merfolk was not tapped for the cost")
		}
	}
}

func TestOldB555BoggartBirthRiteReturnsAGoblinCardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	goblin := effectstest.GraveyardCardOf(g, me, "Goblin Guide", "Creature — Goblin Scout", "{R}", "")
	kindred := effectstest.GraveyardCardOf(g, me, "Goblin Kindred", "Kindred Sorcery — Goblin", "{B}", "")
	elf := effectstest.GraveyardCardOf(g, me, "Elf", "Creature — Elf", "{G}", "")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Boggart Birth Rite", "Kindred Sorcery — Goblin", oldB555Rite, effectstest.CardTargets(elf)); err == nil {
		t.Fatal("targeted an Elf card")
	}
	effectstest.CastCatalogSpell(t, g, "Boggart Birth Rite", "Kindred Sorcery — Goblin", oldB555Rite, effectstest.CardTargets(kindred))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, kindred) {
		t.Error("the Kindred Goblin card did not return")
	}
	if !me.Graveyard.Contains(goblin) || !me.Graveyard.Contains(elf) {
		t.Error("another card left the graveyard")
	}
}

func TestOldB555ScarredVinebreederExilesAnElfCardForPlusThreePlusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	vine := effectstest.PushCatalogPermanent(g, me.ID, "Scarred Vinebreeder", "Creature — Elf Shaman", oldB555Vinebreed, false)
	goblin := effectstest.GraveyardCardOf(g, me, "Goblin", "Creature — Goblin", "{R}", "")
	effectstest.FloatForTest(g, me, "BBB")
	if err := g.ActivateCatalogAbility(me.ID, vine, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{goblin}}); err == nil {
		t.Fatal("exiled a Goblin card card in the graveyard")
	}
	elf := effectstest.GraveyardCardOf(g, me, "Elf", "Creature — Elf", "{G}", "")
	if err := g.ActivateCatalogAbility(me.ID, vine, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{elf}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, vine); p != 4 || tt != 4 {
		t.Errorf("the Vinebreeder is %d/%d, want 4/4", p, tt)
	}
	if me.Graveyard.Contains(elf) || !me.Graveyard.Contains(goblin) {
		t.Error("the wrong card was exiled")
	}
}

func TestOldB555WarrenPilferersReturnsACreatureAndHastesOnlyForAGoblin(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		haste          bool
	}{{"Goblin Guide", "Creature — Goblin Scout", true}, {"Grizzly Bears", "Creature — Bear", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			target := effectstest.GraveyardCardOf(g, me, tc.name, tc.typeLine, "{1}", "")
			land := effectstest.GraveyardCardOf(g, me, "Island", "Basic Land — Island", "", "")
			pil := effectstest.CastCatalogSpell(t, g, "Warren Pilferers", "Creature — Goblin Rogue", oldB555Pilferers, nil)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerTriggerTargets(t, g, me.ID, target)
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.InHandOf(me, target) {
				t.Fatal("the creature card did not return")
			}
			if me.Graveyard.Contains(land) == false {
				t.Error("the land card left the graveyard")
			}
			if got := effectstest.HasEffectiveKeyword(t, g, pil, "haste"); got != tc.haste {
				t.Errorf("haste = %v, want %v", got, tc.haste)
			}
		})
	}
}
