package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 BNS batch 10 — Ravnica: Clue Edition (#5929).

const (
	bnsB10Arrester    = "3f2a27a1-f9cf-4dac-a248-5c6af4de3722"
	bnsB10Justiciar   = "cba0eae3-73ea-4ea9-87bf-1015e490da90"
	bnsB10Screecher   = "752ebcba-97fb-43a3-8811-d244164b3e16"
	bnsB10Krushok     = "6a003988-f126-4309-9166-50f3fbf0f30c"
	bnsB10Juggler     = "34400cf5-b8c4-4cf2-9186-51f12037cc51"
	bnsB10Interpreter = "4bc79fb3-81bb-42ed-9bbb-93f8cede3094"
	bnsB10Cyclops     = "a0324f33-3f25-4374-affa-c7b5b54bc8d7"
	bnsB10Seer        = "09bd4e1c-9861-481f-80dc-4de955c8d3af"
	bnsB10Evangelist  = "13c81fc5-dd35-481d-8e32-4d404fb0e4ac"
	bnsB10Caster      = "f1352a81-6fd5-4c4d-ada4-b81d3a9f74c0"
	bnsB10Imp         = "689c5a47-24ec-4cbf-bc93-ebb492732cad"
)

func TestBnsB10CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB10Arrester: "Azorius Arrester", bnsB10Justiciar: "Azorius Justiciar",
		bnsB10Screecher: "Basilica Screecher", bnsB10Krushok: "Battlefront Krushok",
		bnsB10Juggler: "Blade Juggler", bnsB10Interpreter: "Carnage Interpreter",
		bnsB10Cyclops: "Chainwhip Cyclops", bnsB10Seer: "Cloudkin Seer",
		bnsB10Evangelist: "Conclave Evangelist", bnsB10Caster: "Dagger Caster",
		bnsB10Imp: "Daggerclaw Imp",
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

func TestBnsB10CloudkinSeerFliesAndDrawsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	id := effectstest.CastCreatureCard(t, g, "Cloudkin Seer", "Creature — Elemental Wizard", bnsB10Seer, 2, 1)
	effectstest.AssertKeywords(t, g, id, "flying")
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestBnsB10DaggerclawImpFliesAndCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	imp := effectstest.B12Push(g, g.Seats[1].ID, "Daggerclaw Imp", "Creature — Imp", bnsB10Imp, 3, 1)
	effectstest.AssertKeywords(t, g, imp, "flying")
	effectstest.AssertRestrictions(t, g, imp, game.CantBlock)
}

func TestBnsB10ChainwhipCyclopsStopsOneCreatureBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	cyc := effectstest.B12Push(g, me.ID, "Chainwhip Cyclops", "Creature — Cyclops Warrior", bnsB10Cyclops, 4, 4)
	bear := effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, them.ID, "Their Other", "Creature — Bear", 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "RRRR")
	if err := g.ActivateCatalogAbility(me.ID, cyc, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
	if effectstest.BlockRefusal(t, g, bear, attacker) == "" {
		t.Error("the target blocked")
	}
	if got := effectstest.BlockRefusal(t, g, other, attacker); got != "" {
		t.Errorf("another creature was refused: %q", got)
	}
}

func TestBnsB10ChainwhipCyclopsNeedsItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	cyc := effectstest.B12Push(g, me.ID, "Chainwhip Cyclops", "Creature — Cyclops Warrior", bnsB10Cyclops, 4, 4)
	bear := effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, cyc, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("activated without paying {3}{R}")
	}
}

func TestBnsB10DaggerCasterPingsEachOpponentAndTheirCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	theirs2 := effectstest.B12Creature(g, third.ID, "Third Bear", "Creature — Bear", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	l0, l1, l2 := me.Life, opp.Life, third.Life
	effectstest.CastCreatureCard(t, g, "Dagger Caster", "Creature — Lizard Rogue", bnsB10Caster, 2, 3)
	if opp.Life != l1-1 || third.Life != l2-1 || me.Life != l0 {
		t.Errorf("life me %d->%d, opp %d->%d, third %d->%d", l0, me.Life, l1, opp.Life, l2, third.Life)
	}
	for id, want := range map[uuid.UUID]int{theirs: 1, theirs2: 1, mine: 0} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatal("creature gone")
		}
		if c.DamageMarked != want {
			t.Errorf("%s has %d damage, want %d", c.Name, c.DamageMarked, want)
		}
	}
}

func TestBnsB10BasilicaScreecherExtortsOnACast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scr := effectstest.B12Push(g, me.ID, "Basilica Screecher", "Creature — Bat", bnsB10Screecher, 1, 2)
	effectstest.AssertKeywords(t, g, scr, "flying")
	effectstest.E2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("casting a spell offers the extort payment")
	}
	l0, l1 := me.Life, opp.Life
	effectstest.FloatForTest(g, me, "B")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if me.Life != l0+3 || opp.Life != l1-1 {
		t.Errorf("life me %d->%d, opp %d->%d; want a 1-point drain on each of three opponents", l0, me.Life, l1, opp.Life)
	}
}

func TestBnsB10BladeJugglerHurtsYouAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.AdvanceToMain(t, g)
	hand, life := me.Hand.Size(), me.Life
	effectstest.CastCreatureCard(t, g, "Blade Juggler", "Creature — Human Rogue", bnsB10Juggler, 3, 2)
	if me.Life != life-1 || me.Hand.Size()-hand != 1 {
		t.Errorf("life %d->%d, hand %d->%d; want -1 life and +1 card", life, me.Life, hand, me.Hand.Size())
	}
}

func TestBnsB10BladeJugglerSpectacleNeedsAnOpponentToHaveLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	card := effectstest.HandCardFull(me, "Blade Juggler", "Creature — Human Rogue", "{4}{B}", bnsB10Juggler, []string{"B"})
	params := game.CastSpellParams{Strict: true, AlternativeCost: "spectacle"}
	effectstest.FloatForTest(g, me, "BBB")
	if err := g.CastSpell(me.ID, card, params); err == nil {
		t.Fatal("spectacle was accepted with no life lost this turn")
	}
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, victim.ID, -1); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.CastSpell(me.ID, card, params); err != nil {
		t.Fatalf("spectacle after an opponent lost life: %v", err)
	}
}

func TestBnsB10AzoriusArresterDetainsOneOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Their Other", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Azorius Arrester", "Creature — Human Soldier", bnsB10Arrester, 2, 1)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target asked")
	}
	if effectstest.HasID(pick.PickTargetCards, mine) {
		t.Error("a creature I control was offered")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, theirs, effectstest.BnsB10Detained)
	effectstest.AssertRestrictions(t, g, other, 0)
	effectstest.AssertRestrictions(t, g, mine, 0)
}

func TestBnsB10AzoriusArresterDetainEndsAsItsControllersNextTurnBegins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Azorius Arrester", "Creature — Human Soldier", bnsB10Arrester, 2, 1)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(theirs, me.ID); err == nil {
		t.Fatal("a detained creature attacked")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.AssertRestrictions(t, g, theirs, 0)
}

func TestBnsB10AzoriusJusticiarDetainsUpToTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, third.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Azorius Justiciar", "Creature — Human Wizard", bnsB10Justiciar, 2, 2)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target asked")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c},
	}); err == nil {
		t.Fatal("three targets were accepted")
	}
	p = effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the prompt went away after a refused answer")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, a, effectstest.BnsB10Detained)
	effectstest.AssertRestrictions(t, g, b, effectstest.BnsB10Detained)
	effectstest.AssertRestrictions(t, g, c, 0)
}

func TestBnsB10BattlefrontKrushokLimitsItselfAndCounteredCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	krushok := effectstest.B12Push(g, me.ID, "Battlefront Krushok", "Creature — Beast", bnsB10Krushok, 3, 4)
	grown := effectstest.B12Creature(g, me.ID, "Grown Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == grown {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
			}
		}
	})
	b1 := effectstest.B12Creature(g, them.ID, "Blocker 1", "Creature — Bear", 2, 2)
	b2 := effectstest.B12Creature(g, them.ID, "Blocker 2", "Creature — Bear", 2, 2)
	effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{krushok, grown, plain} {
		if err := g.DeclareAttacker(a, them.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareBlockers)
	for _, tc := range []struct {
		attacker uuid.UUID
		name     string
		double   bool
	}{{krushok, "Krushok", false}, {grown, "creature with a +1/+1 counter", false}, {plain, "plain creature", true}} {
		g2 := g.Clone()
		if err := g2.DeclareBlocker(b1, tc.attacker); err != nil {
			t.Fatalf("%s: first block refused: %v", tc.name, err)
		}
		err := g2.DeclareBlocker(b2, tc.attacker)
		if tc.double && err != nil {
			t.Errorf("%s: a second blocker was refused: %v", tc.name, err)
		}
		if !tc.double {
			if err == nil {
				// The count is judged on the whole declaration: finish it.
				err = g2.FinishBlocks(them.ID)
			}
			if err == nil {
				t.Errorf("%s was double-blocked", tc.name)
			}
		}
	}
}

func TestBnsB10CarnageInterpreterDiscardsHandAndInvestigatesFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	id := effectstest.CastCreatureCard(t, g, "Carnage Interpreter", "Creature — Devil Detective", bnsB10Interpreter, 3, 3)
	if me.Hand.Size() != 0 {
		t.Errorf("hand has %d cards, want 0", me.Hand.Size())
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Clue"); n != 4 {
		t.Errorf("%d Clues, want 4", n)
	}
	if p := effectstest.EffectivePower(t, g, id); p != 5 {
		t.Errorf("power %d, want 5 with an empty hand", p)
	}
	effectstest.AssertKeywords(t, g, id, "menace")
}

func TestBnsB10CarnageInterpreterIsSmallWithTwoOrMoreCardsInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Carnage Interpreter", "Creature — Devil Detective", bnsB10Interpreter, 3, 3)
	if me.Hand.Size() < 2 {
		t.Fatalf("hand has %d cards, the test needs at least 2", me.Hand.Size())
	}
	if p := effectstest.EffectivePower(t, g, id); p != 3 {
		t.Errorf("power %d with %d cards in hand, want 3", p, me.Hand.Size())
	}
	if got := effectstest.EffectiveAbilities(t, g, id); effectstest.ContainsString(got, "menace") {
		t.Errorf("menace with a full hand: %v", got)
	}
	// Down to one card: the bonus and menace come on, as a zone move
	// drops the layer cache (CR 611.3a).
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(me.ID, me.Hand.Size()-1); err != nil {
			t.Fatal(err)
		}
	})
	if me.Hand.Size() != 1 {
		t.Fatalf("hand has %d cards, want 1", me.Hand.Size())
	}
	if p := effectstest.EffectivePower(t, g, id); p != 5 {
		t.Errorf("power %d with one card in hand, want 5", p)
	}
	effectstest.AssertKeywords(t, g, id, "menace")
}

func TestBnsB10ConclaveEvangelistMyriadAndCopyOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	ev := effectstest.B12Push(g, me.ID, "Conclave Evangelist", "Creature — Elephant Cleric", bnsB10Evangelist, 4, 4)
	effectstest.DeclareAttack(t, g, opp1.ID, ev)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Conclave Evangelist" && c.InstanceID != ev && effects.IsToken(c) {
			copies++
		}
	}
	if copies != 2 {
		t.Fatalf("%d myriad copies, want 2", copies)
	}
	l1, l2, l3 := opp1.Life, opp2.Life, opp3.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.AnswerOpenOrderingIfAny(t, g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.AnswerOpenOrderingIfAny(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp1.Life != l1-4 || opp2.Life != l2-4 || opp3.Life != l3-4 {
		t.Errorf("each opponent takes 4: %d %d %d", l1-opp1.Life, l2-opp2.Life, l3-opp3.Life)
	}
	// Three hits make three more tokens; the two myriad tokens are exiled
	// at end of combat, and the three new copies stay.
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Conclave Evangelist" && effects.IsToken(c) && c.Controller == me.ID {
			tokens++
		}
	}
	if tokens < 3 {
		t.Errorf("%d token copies on the battlefield, want at least the 3 hit copies", tokens)
	}
}
