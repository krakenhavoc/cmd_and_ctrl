package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b556_test.go — Lorwyn, counters, other creatures and champion, batch
// 556 (#5593). Every test plays the real catalog entry through the engine.

const (
	oldB556Handservant = "16fb84e1-0bc7-469c-bd73-40747be56d9e"
	oldB556Hunter      = "34d5f249-df46-49d1-a692-da2348fbc51e"
	oldB556Witch       = "4d165c0f-3e40-45ad-92a1-53b13c07b34d"
	oldB556Veteran     = "70bcec56-1485-4e4e-8fff-940d7b133c5d"
	oldB556BlindSpot   = "37165859-87bc-4d44-95a2-39051ab20a14"
	oldB556Caterwaul   = "ab3f4455-7329-4145-b9c4-93eef38bf327"
	oldB556Axegrinder  = "c5355c5e-d07d-48ca-90c5-98edcab1a3c7"
	oldB556Dodger      = "e120f341-4445-4b96-9777-2290471832bf"
	oldB556Trio        = "15fca8d6-8a86-4c24-9178-f0b6496518ab"
	oldB556NovaChaser  = "997f5847-04c2-498a-a294-d3355fbb7424"
	oldB556Mob         = "f5e2da6c-35fa-4f46-bfae-8ac8907c8300"
	oldB556Bear        = "Creature — Bear"
)

func TestOldB556EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB556Handservant: "Elvish Handservant", oldB556Hunter: "Hunter of Eyeblights", oldB556Witch: "Knucklebone Witch",
		oldB556Veteran: "Veteran of the Depths", oldB556BlindSpot: "Blind-Spot Giant", oldB556Caterwaul: "Caterwauling Boggart",
		oldB556Axegrinder: "Axegrinder Giant", oldB556Dodger: "Goldmeadow Dodger", oldB556Trio: "Thoughtweft Trio",
		oldB556NovaChaser: "Nova Chaser", oldB556Mob: "Boggart Mob",
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

func oldB556AnswerMay(t *testing.T, g *game.Game, who *game.Player, yes bool) {
	t.Helper()
	p := effectstest.LatestTriggerPromptFor(g, who.ID)
	if p == nil {
		t.Fatal("no optional trigger prompt is open")
	}
	if err := g.ResolveTriggerPrompt(p.ID, who.ID, yes); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
}

func TestOldB556ElvishHandservantGrowsForAnyPlayersGiantSpellIfYouSayYes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.B27Push(g, me.ID, "Elvish Handservant", "Creature — Elf Warrior", oldB556Handservant, "{G}", 1, 1, "G")
	effectstest.AdvanceToMain(t, g)
	cast := func(caster *game.Player, typeLine string) {
		id := uuid.New()
		caster.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: typeLine, Owner: caster.ID, Controller: caster.ID})
		if err := g.CastSpell(caster.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	cast(me, "Instant")
	cast(opp, "Instant — Bear")
	if n := effectstest.CounterCount(g, elf, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters after spells that are no Giants, want 0", n)
	}
	cast(me, "Instant — Giant Warrior")
	oldB556AnswerMay(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	cast(opp, "Instant — Giant")
	oldB556AnswerMay(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	cast(opp, "Instant — Giant")
	oldB556AnswerMay(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, elf, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters, want 2 (two Giants said yes to, one said no to)", n)
	}
}

func TestOldB556HunterOfEyeblightsCountersAnOpposingCreatureAndDestroysOneWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB556Bear, 2, 2)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", oldB556Bear, 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other Foe", oldB556Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	hunter := effectstest.CastHoldCreature(t, g, "Hunter of Eyeblights", "Creature — Elf Assassin", oldB556Hunter, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardTargets(mine)); err == nil {
		t.Fatal("targeted my own creature")
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, foe, game.CounterPlusOne); n != 1 {
		t.Fatalf("the target has %d +1/+1 counters, want 1", n)
	}
	// The cast Hunter is summoning sick; a second one that has been here a while taps for the ability.
	hunter = effectstest.PushCatalogPermanent(g, me.ID, "Hunter of Eyeblights", "Creature — Elf Assassin", oldB556Hunter, false)
	effectstest.FloatForTest(g, me, "BBB")
	if err := g.ActivateCatalogAbility(me.ID, hunter, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(other)}); err == nil {
		t.Fatal("destroyed a creature with no counter")
	}
	effectstest.FloatForTest(g, me, "BBB")
	if err := g.ActivateCatalogAbility(me.ID, hunter, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(foe)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, foe) {
		t.Error("the creature with a counter survived")
	}
	if !effectstest.BattlefieldHas(g, other) {
		t.Error("the creature without one was destroyed")
	}
}

func TestOldB556KnucklebonWitchGrowsWhenAGoblinOfYoursGoesToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	witch := effectstest.B27Push(g, me.ID, "Knucklebone Witch", "Creature — Goblin Shaman", oldB556Witch, "{B}", 1, 1, "B")
	mine := effectstest.B12Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	token := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 black Goblin Rogue"))
	theirs := effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB556Bear, 1, 1)
	effectstest.AdvanceToMain(t, g)
	kill := func(id uuid.UUID, controller uuid.UUID, answer bool, expectPrompt bool) {
		effectstest.KillCreature(t, g, controller, id)
		effectstest.PassPriorityAroundTable(t, g)
		if expectPrompt {
			oldB556AnswerMay(t, g, me, answer)
			effectstest.PassPriorityAroundTable(t, g)
		}
	}
	kill(theirs, opp.ID, false, false)
	kill(bear, me.ID, false, false)
	if n := effectstest.CounterCount(g, witch, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters for creatures that don't count, want 0", n)
	}
	kill(mine, me.ID, true, true)
	kill(token, me.ID, true, true)
	if n := effectstest.CounterCount(g, witch, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters, want 2 (a Goblin and a Goblin token of mine)", n)
	}
}

func TestOldB556VeteranOfTheDepthsMayGrowWhenItBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vet := effectstest.B27Push(g, me.ID, "Veteran of the Depths", "Creature — Merfolk Soldier", oldB556Veteran, "{3}{W}", 2, 2, "W")
	other := effectstest.B12Creature(g, me.ID, "Bear", oldB556Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.TapForTest(t, g, other)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPromptFor(g, me.ID) != nil || effectstest.CounterCount(g, vet, game.CounterPlusOne) != 0 {
		t.Fatal("another creature's tap triggered the Veteran")
	}
	effectstest.TapForTest(t, g, vet)
	effectstest.PassPriorityAroundTable(t, g)
	oldB556AnswerMay(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, vet, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}
}

func TestOldB556BlindSpotGiantNeedsAnotherGiantToAttackAndBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.B12Push(g, me.ID, "Blind-Spot Giant", "Creature — Giant Warrior", oldB556BlindSpot, 4, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(giant, opp.ID); err == nil {
		t.Fatal("the Giant attacked with no other Giant")
	}
	effectstest.B12Creature(g, me.ID, "Other Giant", "Creature — Giant", 3, 3)
	if err := g.DeclareAttacker(giant, opp.ID); err != nil {
		t.Errorf("the Giant could not attack with another Giant: %v", err)
	}
}

func TestOldB556BlindSpotGiantCantBlockWithoutAnotherGiant(t *testing.T) {
	for _, withGiant := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		giant := effectstest.B12Push(g, opp.ID, "Blind-Spot Giant", "Creature — Giant Warrior", oldB556BlindSpot, 4, 3)
		if withGiant {
			effectstest.B12Creature(g, opp.ID, "Other Giant", "Creature — Giant", 3, 3)
		}
		atk := effectstest.B12Creature(g, me.ID, "Attacker", oldB556Bear, 1, 1)
		effectstest.B64ReadyToAttack(g, atk)
		effectstest.AdvanceToMain(t, g)
		effectstest.LciB1Attack(t, g, atk, opp.ID)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(giant, atk)
		if withGiant && err != nil {
			t.Errorf("the Giant could not block with another Giant: %v", err)
		}
		if !withGiant && err == nil {
			t.Error("the Giant blocked with no other Giant")
		}
	}
}

func TestOldB556CaterwaulingBoggartGivesMenaceToYourGoblinsAndElementalsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boggart := effectstest.B12Push(g, me.ID, "Caterwauling Boggart", "Creature — Goblin Shaman", oldB556Caterwaul, 2, 2)
	goblin := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	elemental := effectstest.B12Creature(g, me.ID, "Elemental", "Creature — Elemental", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB556Bear, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	for _, id := range []uuid.UUID{boggart, goblin, elemental} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "menace") {
			t.Errorf("%s has no menace", id)
		}
	}
	for _, id := range []uuid.UUID{bear, theirs} {
		if effectstest.HasEffectiveKeyword(t, g, id, "menace") {
			t.Errorf("%s has menace", id)
		}
	}
}

func TestOldB556GoldmeadowDodgerCantBeBlockedByPowerFourOrGreater(t *testing.T) {
	for _, tc := range []struct {
		power   int
		blocked bool
	}{{3, true}, {4, false}, {6, false}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dodger := effectstest.B12Push(g, me.ID, "Goldmeadow Dodger", "Creature — Kithkin Rogue", oldB556Dodger, 1, 1)
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB556Bear, tc.power, 4)
		effectstest.B64ReadyToAttack(g, dodger)
		effectstest.AdvanceToMain(t, g)
		effectstest.LciB1Attack(t, g, dodger, opp.ID)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(blocker, dodger)
		if tc.blocked && err != nil {
			t.Errorf("a creature with power %d could not block the Dodger: %v", tc.power, err)
		}
		if !tc.blocked && err == nil {
			t.Errorf("a creature with power %d blocked the Dodger", tc.power)
		}
	}
}

func TestOldB556ThoughtweftTrioChampionsAKithkinAndReturnsItWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	kithkin := effectstest.B12Creature(g, me.ID, "Goldmeadow Harrier", "Creature — Kithkin Soldier", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB556Bear, 2, 2)
	trio := effectstest.CastHoldCreature(t, g, "Thoughtweft Trio", "Creature — Kithkin Soldier", oldB556Trio, 5, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, kithkin)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(trio) || g.Battlefield.Contains(kithkin) || !g.Battlefield.Contains(bear) {
		t.Fatalf("trio on battlefield %v, kithkin %v, bear %v", g.Battlefield.Contains(trio), g.Battlefield.Contains(kithkin), g.Battlefield.Contains(bear))
	}
	for _, kw := range []string{"first strike", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, trio, kw) {
			t.Errorf("the Trio has no %s", kw)
		}
	}
	effectstest.KillCreature(t, g, me.ID, trio)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(oldB556Named(g, "Goldmeadow Harrier", me.ID)); got != 1 {
		t.Errorf("the championed Kithkin is on the battlefield %d times after the Trio left, want 1", got)
	}
}

func TestOldB556ThoughtweftTrioBlocksSeveralAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trio := effectstest.B12Push(g, opp.ID, "Thoughtweft Trio", "Creature — Kithkin Soldier", oldB556Trio, 5, 5)
	a := effectstest.B12Creature(g, me.ID, "Attacker A", oldB556Bear, 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Attacker B", oldB556Bear, 1, 1)
	effectstest.B64ReadyToAttack(g, a)
	effectstest.B64ReadyToAttack(g, b)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(trio, a); err != nil {
		t.Fatalf("first block: %v", err)
	}
	if err := g.DeclareBlocker(trio, b); err != nil {
		t.Errorf("the Trio could not block a second attacker: %v", err)
	}
}

func TestOldB556NovaChaserChampionsAnElementalOrIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	chaser := effectstest.CastHoldCreature(t, g, "Nova Chaser", "Creature — Elemental Warrior", oldB556NovaChaser, 10, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(chaser) {
		t.Fatal("the Chaser stayed with no Elemental to champion")
	}
	elemental := effectstest.B12Creature(g, me.ID, "Elemental", "Creature — Elemental", 1, 1)
	chaser2 := effectstest.CastHoldCreature(t, g, "Nova Chaser", "Creature — Elemental Warrior", oldB556NovaChaser, 10, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, elemental)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(chaser2) || g.Battlefield.Contains(elemental) {
		t.Error("the Chaser did not champion the Elemental")
	}
	if !effectstest.HasEffectiveKeyword(t, g, chaser2, "trample") {
		t.Error("the Chaser has no trample")
	}
}

func TestOldB556BoggartMobChampionsAGoblinAndMakesARogueForEachGoblinThatHitsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	spare := effectstest.B12Creature(g, me.ID, "Spare Goblin", "Creature — Goblin", 1, 1)
	mob := effectstest.CastHoldCreature(t, g, "Boggart Mob", "Creature — Goblin Warrior", oldB556Mob, 5, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, spare)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mob) || g.Battlefield.Contains(spare) {
		t.Fatal("the Mob did not champion the Goblin")
	}
	goblin := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB556Bear, 2, 2)
	for _, id := range []uuid.UUID{goblin, bear} {
		effectstest.B64ReadyToAttack(g, id)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.DeclareAttack(t, g, opp.ID, goblin, bear)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 8 && effectstest.LatestTriggerPromptFor(g, me.ID) == nil; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if _, err := g.AdvanceStep(); err != nil {
			break
		}
	}
	if effectstest.LatestTriggerPromptFor(g, me.ID) == nil {
		t.Fatal("no Rogue offer after a Goblin hit a player")
	}
	oldB556AnswerMay(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPromptFor(g, me.ID) != nil {
		t.Fatal("the Bear's damage offered a Rogue")
	}
	rogues := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Goblin Rogue" {
			rogues++
		}
	}
	if rogues != 1 {
		t.Errorf("%d Goblin Rogue tokens, want 1", rogues)
	}
}

func oldB556Named(g *game.Game, name string, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}
