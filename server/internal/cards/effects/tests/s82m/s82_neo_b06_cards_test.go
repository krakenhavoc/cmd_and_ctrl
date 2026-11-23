package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 NEO batch 6 (#5493): the ninja creatures. Every test puts the ninja
// onto the battlefield through its real ninjutsu ability and plays the
// combat damage step out.

const (
	neoMukotaiAmbusher     = "3a3f0bf4-7a02-4a57-8a18-ff22f012e0ef"
	neoDokuchiShadowWalker = "1e79b7bf-8a2d-4585-81d0-755dce78a4ea"
	neoBladeBlizzardKit    = "84fb1c2f-cf1b-4166-834f-a887aa182d09"
	neoMoonsnareSpecialist = "ba858b68-ae93-4e1c-8a3a-3e4ac97282e8"
	neoSpringLeafAvenger   = "19797dd6-6f48-47bc-a9ac-84d4acad3fa5"
	neoCovertTechnician    = "b713ee3a-b6cd-4cd7-b403-f0bbc683ea99"
	neoCoilingStalker      = "54dc1c02-2d71-465b-a285-d6d11113991c"
	neoMoonCircuitHacker   = "7cd8b017-8d59-4b55-bb9a-8d8e492ef1ac"
	neoNezumiProwler       = "a93b5956-c11e-434e-bed5-94af9de55382"
	neoDokuchiSilencer     = "d377b28f-c887-4937-9f14-74484b612551"
	neoBitingPalmNinja     = "5c22447b-7a7f-4a27-8d3b-1bb664ea94c3"
	neoKappaTechWrecker    = "2bc92c16-cd1e-4c32-9102-4593d77a9a9f"
)

// neoConfirm answers the open yes/no prompt (a "you may" resolved inside an
// ability), passing priority until one opens.
func neoConfirm(t *testing.T, g *game.Game, who uuid.UUID, yes bool) {
	t.Helper()
	for i := 0; i < 8; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == who {
				if err := g.ResolveConfirm(c.ID, who, yes); err != nil {
					t.Fatalf("ResolveConfirm: %v", err)
				}
				return
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	t.Fatal("no confirm prompt opened")
}

func TestNeoB06EveryCardIsFullAndHasItsNinjutsuCost(t *testing.T) {
	for _, tc := range []struct{ name, oracle, mana string }{
		{"Mukotai Ambusher", neoMukotaiAmbusher, "{1}{B}"},
		{"Dokuchi Shadow-Walker", neoDokuchiShadowWalker, "{3}{B}"},
		{"Blade-Blizzard Kitsune", neoBladeBlizzardKit, "{3}{W}"},
		{"Moonsnare Specialist", neoMoonsnareSpecialist, "{2}{U}"},
		{"Spring-Leaf Avenger", neoSpringLeafAvenger, "{3}{G}"},
		{"Covert Technician", neoCovertTechnician, "{1}{U}"},
		{"Coiling Stalker", neoCoilingStalker, "{1}{G}"},
		{"Moon-Circuit Hacker", neoMoonCircuitHacker, "{U}"},
		{"Nezumi Prowler", neoNezumiProwler, "{1}{B}"},
		{"Dokuchi Silencer", neoDokuchiSilencer, "{1}{B}"},
		{"Biting-Palm Ninja", neoBitingPalmNinja, "{2}{B}"},
		{"Kappa Tech-Wrecker", neoKappaTechWrecker, "{1}{G}"},
	} {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: not registered Full with no caveats", tc.name)
		}
		abs := game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle})
		if len(abs) != 1 || abs[0].Cost.Mana != tc.mana || abs[0].Cost.ReturnToHand.Empty() {
			t.Errorf("%s: want one ninjutsu ability costing %s with a return cost", tc.name, tc.mana)
		}
	}
}

func TestNeoB06MukotaiAmbusherLifelinkGainsLife(t *testing.T) {
	g, _, _, me, opp := effectstest.NeoNinjutsu(t, "Mukotai Ambusher", neoMukotaiAmbusher, "Artifact Creature — Rat Ninja", 3, 2, "{1}{B}")
	life, oppLife := me.Life, opp.Life
	effectstest.NeoToDamage(t, g)
	if got := oppLife - opp.Life; got != 3 {
		t.Errorf("defender lost %d, want 3", got)
	}
	if got := me.Life - life; got != 3 {
		t.Errorf("gained %d life, want 3 (lifelink)", got)
	}
}

func TestNeoB06BladeBlizzardKitsuneDealsDamageTwice(t *testing.T) {
	g, _, _, _, opp := effectstest.NeoNinjutsu(t, "Blade-Blizzard Kitsune", neoBladeBlizzardKit, "Creature — Fox Ninja", 2, 2, "{3}{W}")
	life := opp.Life
	effectstest.NeoToDamage(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if got := life - opp.Life; got != 4 {
		t.Errorf("defender lost %d, want 4 (double strike, 2 power)", got)
	}
}

func TestNeoB06DokuchiShadowWalkerArrivesTappedAndAttacking(t *testing.T) {
	g, ninja, attacker, me, opp := effectstest.NeoNinjutsu(t, "Dokuchi Shadow-Walker", neoDokuchiShadowWalker, "Creature — Ogre Ninja", 5, 5, "{3}{B}")
	c, _ := g.LookupCardForEffect(ninja)
	if !c.Tapped || c.AttackingTarget != opp.ID || !me.Hand.Contains(attacker) {
		t.Errorf("tapped=%v attacking=%v returned=%v", c.Tapped, c.AttackingTarget == opp.ID, me.Hand.Contains(attacker))
	}
	life := opp.Life
	effectstest.NeoToDamage(t, g)
	if got := life - opp.Life; got != 5 {
		t.Errorf("defender lost %d, want 5", got)
	}
}

func TestNeoB06MoonsnareSpecialistBouncesATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ninja := effectstest.PushNinjaToHand(me, "Moonsnare Specialist", "Creature — Human Ninja", neoMoonsnareSpecialist, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, effectstest.NeoPool("{2}{U}")); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{attacker}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the targeted creature was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(ninja) {
		t.Error("the ninja is gone")
	}
}

func TestNeoB06NezumiProwlerGivesDeathtouchAndLifelinkUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	ninja := effectstest.PushNinjaToHand(me, "Nezumi Prowler", "Artifact Creature — Rat Ninja", neoNezumiProwler, 3, 1)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, effectstest.NeoPool("{1}{B}")); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{attacker}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, ally, "deathtouch") || !effectstest.HasEffectiveKeyword(t, g, ally, "lifelink") {
		t.Error("the target did not gain deathtouch and lifelink")
	}
	if effectstest.HasEffectiveKeyword(t, g, ninja, "deathtouch") {
		t.Error("the Prowler gained deathtouch though it was not the target")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if effectstest.HasEffectiveKeyword(t, g, ally, "deathtouch") {
		t.Error("deathtouch lasted past end of turn")
	}
}

func TestNeoB06SpringLeafAvengerReturnsAPermanentCard(t *testing.T) {
	g, _, _, me, _ := effectstest.NeoNinjutsu(t, "Spring-Leaf Avenger", neoSpringLeafAvenger, "Creature — Insect Ninja", 6, 5, "{3}{G}")
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Shock", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	effectstest.NeoToDamage(t, g)
	effectstest.PickIfAsked(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) || me.Graveyard.Contains(land) {
		t.Error("the permanent card was not returned to hand")
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("a nonpermanent card was taken")
	}
}

func TestNeoB06CovertTechnicianPutsAnArtifactWithManaValueAtMostTheDamage(t *testing.T) {
	g, _, _, me, _ := effectstest.NeoNinjutsu(t, "Covert Technician", neoCovertTechnician, "Artifact Creature — Human Ninja", 2, 4, "{1}{U}")
	small := effectstest.PushHandCardWithManaCost(me, "Small Rock", "Artifact", "{2}")
	big := effectstest.PushHandCardWithManaCost(me, "Big Rock", "Artifact", "{3}")
	creature := effectstest.PushHandCardWithManaCost(me, "Bear", "Creature — Bear", "{1}")
	effectstest.NeoToDamage(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no hand-put prompt")
	}
	if !effectstest.HasID(pick.ChooseCards, small) || effectstest.HasID(pick.ChooseCards, big) || effectstest.HasID(pick.ChooseCards, creature) {
		t.Fatalf("options %v: only the artifact with mana value <= 2 may be offered", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, small)
	if !g.Battlefield.Contains(small) {
		t.Error("the artifact did not enter")
	}
}

func TestNeoB06CoilingStalkerPutsACounterOnACreatureWithoutOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	ninja := effectstest.PushNinjaToHand(me, "Coiling Stalker", "Creature — Snake Ninja", neoCoilingStalker, 2, 1)
	grown := effectstest.B12Creature(g, me.ID, "Grown", "Creature — Bear", 2, 2)
	bare := effectstest.B12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(grown, game.CounterPlusOne, 1) })
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, effectstest.NeoPool("{1}{G}")); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{attacker}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.NeoToDamage(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, grown) {
		t.Error("a creature that already has a +1/+1 counter was offered")
	}
	effectstest.PickIfAsked(t, g, me.ID, bare)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bare); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("bare has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
	}
	if c, _ := effectstest.BattlefieldCard(g, grown); c.Counters[game.CounterPlusOne] != 1 {
		t.Error("the creature that already had a counter got another")
	}
}

func TestNeoB06MoonCircuitHackerEnteredThisTurnDrawsWithoutDiscarding(t *testing.T) {
	g, _, _, me, _ := effectstest.NeoNinjutsu(t, "Moon-Circuit Hacker", neoMoonCircuitHacker, "Enchantment Creature — Human Ninja", 2, 1, "{U}")
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	effectstest.NeoToDamage(t, g)
	neoConfirm(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("asked to discard though the Hacker entered this turn")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1", got)
	}
}

func TestNeoB06MoonCircuitHackerFromAnEarlierTurnMustDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hacker := effectstest.B12Push(g, me.ID, "Moon-Circuit Hacker", "Enchantment Creature — Human Ninja", neoMoonCircuitHacker, 2, 1)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	keep := effectstest.PushHandCardOfType(me, "Keeper", "Sorcery")
	effectstest.DeclareAttack(t, g, opp.ID, hacker)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	neoConfirm(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt though the Hacker did not enter this turn")
	}
	effectstest.AnswerDiscard(t, g, me.ID, keep)
	if !me.Graveyard.Contains(keep) {
		t.Error("the chosen card was not discarded")
	}
}

func TestNeoB06MoonCircuitHackerDecliningTheDrawSkipsTheDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hacker := effectstest.B12Push(g, me.ID, "Moon-Circuit Hacker", "Enchantment Creature — Human Ninja", neoMoonCircuitHacker, 2, 1)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, hacker)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	neoConfirm(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("declining the draw still drew or discarded")
	}
}

func TestNeoB06DokuchiSilencerDiscardsACreatureToDestroyOne(t *testing.T) {
	g, _, _, me, opp := effectstest.NeoNinjutsu(t, "Dokuchi Silencer", neoDokuchiSilencer, "Creature — Human Ninja", 2, 1, "{1}{B}")
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	fodder := effectstest.PushHandCardOfType(me, "Fodder", "Creature — Goblin")
	effectstest.NeoToDamage(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(fodder) {
		t.Error("the creature card was not discarded")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the target survived")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was destroyed")
	}
}

func TestNeoB06DokuchiSilencerDiscardingNothingDestroysNothing(t *testing.T) {
	g, _, _, me, opp := effectstest.NeoNinjutsu(t, "Dokuchi Silencer", neoDokuchiSilencer, "Creature — Human Ninja", 2, 1, "{1}{B}")
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.PushHandCardOfType(me, "Fodder", "Creature — Goblin")
	effectstest.NeoToDamage(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("destroyed a creature without a discard")
	}
}

func TestNeoB06DokuchiSilencerRefusesANonCreatureDiscard(t *testing.T) {
	g, _, _, me, _ := effectstest.NeoNinjutsu(t, "Dokuchi Silencer", neoDokuchiSilencer, "Creature — Human Ninja", 2, 1, "{1}{B}")
	spell := effectstest.PushHandCardOfType(me, "Shock", "Instant")
	effectstest.NeoToDamage(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{spell}); err == nil {
		t.Error("a noncreature card was accepted for 'discard a creature card'")
	}
}

func TestNeoB06BitingPalmNinjaEntersWithAMenaceCounterAndTakesACard(t *testing.T) {
	g, ninja, _, me, opp := effectstest.NeoNinjutsu(t, "Biting-Palm Ninja", neoBitingPalmNinja, "Creature — Human Ninja", 3, 3, "{2}{B}")
	if c, _ := g.LookupCardForEffect(ninja); c.Counters[game.CounterMenace] != 1 {
		t.Fatalf("menace counters %d, want 1", c.Counters[game.CounterMenace])
	}
	ids := effectstest.RevealHand(opp, effectstest.RhBolt(), game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.NeoToDamage(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(ninja); c.Counters[game.CounterMenace] != 0 {
		t.Error("the menace counter was not removed")
	}
	pick := effectstest.OpenVariant(t, g)
	if effectstest.HasID(pick.DiscardOptions, ids[1]) || !effectstest.HasID(pick.DiscardOptions, ids[0]) {
		t.Fatalf("options %v: only the nonland card may be chosen", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || opp.Graveyard.Contains(ids[0]) {
		t.Error("the chosen card was not exiled")
	}
}

func TestNeoB06BitingPalmNinjaDecliningKeepsTheCounter(t *testing.T) {
	g, ninja, _, me, opp := effectstest.NeoNinjutsu(t, "Biting-Palm Ninja", neoBitingPalmNinja, "Creature — Human Ninja", 3, 3, "{2}{B}")
	ids := effectstest.RevealHand(opp, effectstest.RhBolt())
	effectstest.NeoToDamage(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(ninja); c.Counters[game.CounterMenace] != 1 {
		t.Error("declining removed the counter")
	}
	if !opp.Hand.Contains(ids[0]) {
		t.Error("their card was taken")
	}
}

func TestNeoB06KappaTechWreckerExilesAnArtifactOrEnchantment(t *testing.T) {
	g, ninja, _, me, opp := effectstest.NeoNinjutsu(t, "Kappa Tech-Wrecker", neoKappaTechWrecker, "Creature — Turtle Ninja", 1, 3, "{1}{G}")
	if c, _ := g.LookupCardForEffect(ninja); c.Counters[game.CounterDeathtouch] != 1 || !effectstest.HasEffectiveKeyword(t, g, ninja, "deathtouch") {
		t.Fatal("the Wrecker did not enter with a deathtouch counter")
	}
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.NeoToDamage(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !g.Exile.Contains(rock) {
		t.Error("the artifact was not exiled")
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(mine) {
		t.Error("something other than the target left")
	}
	if c, _ := g.LookupCardForEffect(ninja); c.Counters[game.CounterDeathtouch] != 0 {
		t.Error("the deathtouch counter was not removed")
	}
}

func TestNeoB06KappaTechWreckerHasNoTargetOfTheirsToTake(t *testing.T) {
	g, ninja, _, me, _ := effectstest.NeoNinjutsu(t, "Kappa Tech-Wrecker", neoKappaTechWrecker, "Creature — Turtle Ninja", 1, 3, "{1}{G}")
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.NeoToDamage(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("exiled my own artifact")
	}
	_ = ninja
}
