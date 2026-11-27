package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b401_cards_test.go — Magic 2014, S82 OLD batch 401 (#5563): twelve
// cards, each driven through the engine.

const (
	m14MoltenBirth      = "aeaae9cc-f30f-4d7e-86f3-3132ef47e987"
	m14GrimReturn       = "c1d54d0c-53fc-474c-ba89-2aa1be3bf27d"
	m14DevoutInvocation = "fab491b0-b820-4027-996f-390c97df1e96"
	m14Shrivel          = "1be478ee-87f3-4265-94d7-08521703c888"
	m14Rootwalla        = "4dda5f21-11d6-40ce-be24-7c9529272c25"
	m14IntoTheWilds     = "70d6cd9c-6311-4d8e-b74a-8edc11d1c05e"
	m14DismissIntoDream = "b76566a9-f998-4a20-b073-4b60862843c9"
	m14SpellBlast       = "04477339-7ed5-4770-9e5c-6e481ffcc858"
	m14GuardianAges     = "1bf88f66-b673-4652-981b-3ea5d28f5c29"
	m14ChandrasPhoenix  = "65ddd04c-7207-4f6c-ae1b-c5e02a520a51"
	m14GoblinDiplomats  = "84b55e6b-5ce3-49af-95fe-46c029e0260c"
	m14SuntailHawk      = "a8d31e2f-7b2e-4135-8074-9e6ef778bd80"
	m14LavaAxe          = "387b6b07-a283-412d-94c3-f7f1dc76e858"
	m14Shock            = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"
)

func TestS82OldB401EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		m14MoltenBirth: "Molten Birth", m14GrimReturn: "Grim Return", m14DevoutInvocation: "Devout Invocation",
		m14Shrivel: "Shrivel", m14Rootwalla: "Rootwalla", m14IntoTheWilds: "Into the Wilds",
		m14DismissIntoDream: "Dismiss into Dream", m14SpellBlast: "Spell Blast", m14GuardianAges: "Guardian of the Ages",
		m14ChandrasPhoenix: "Chandra's Phoenix", m14GoblinDiplomats: "Goblin Diplomats", m14SuntailHawk: "Suntail Hawk",
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

func m14Toughness(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.FindBattlefieldCardForTest(g, id)
	if c == nil {
		t.Fatal("not on the battlefield")
	}
	return c.CurrentToughness()
}

func m14Power(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.FindBattlefieldCardForTest(g, id)
	if c == nil {
		t.Fatal("not on the battlefield")
	}
	return c.CurrentPower()
}

func TestS82OldB401SuntailHawkFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hawk := effectstest.B21Push(g, me.ID, "Suntail Hawk", "Creature — Bird", m14SuntailHawk, 1, 1, "W")
	if !effectstest.HasEffectiveKeyword(t, g, hawk, "flying") {
		t.Error("no flying")
	}
}

func TestS82OldB401ShrivelShrinksEveryCreatureUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Mouse", "Creature — Mouse", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Shrivel", "Sorcery", m14Shrivel, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := m14Power(t, g, mine), m14Toughness(t, g, mine); p != 1 || th != 1 {
		t.Errorf("my bear %d/%d, want 1/1", p, th)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the 1/1 survived")
	}
	// A creature arriving later is not affected (CR 611.2c).
	late := effectstest.B12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if p := m14Power(t, g, late); p != 2 {
		t.Errorf("a creature that arrived later has power %d, want 2", p)
	}
}

func TestS82OldB401SpellBlastCountersAnXSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	effectstest.CastXSpell(t, g, "Spell Blast", "Instant", m14SpellBlast, "{X}{U}", 0, effectstest.CardTarget(spell))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(spell) || effectstest.FindBattlefieldByName(g, "A Bear") != uuid.Nil {
		t.Error("the spell was not countered")
	}
}

func TestS82OldB401SpellBlastRefusesASpellOfAnotherManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.HandCardFull(me, "Pricey Bear", "Creature — Bear", "{3}", "", nil)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	blast := effectstest.HandCardFull(me, "Spell Blast", "Instant", "{X}{U}", m14SpellBlast, nil)
	if err := g.CastSpell(me.ID, blast, game.CastSpellParams{Strict: true, XValue: 1, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("X=1 targeted a spell with mana value 3")
	}
}

func TestS82OldB401RootwallaPumpsOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	walla := effectstest.B21Push(g, me.ID, "Rootwalla", "Creature — Lizard", m14Rootwalla, 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, walla, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	effectstest.B06AddMana(me, "C", "G", "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, walla, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := m14Power(t, g, walla), m14Toughness(t, g, walla); p != 4 || th != 4 {
		t.Errorf("after one activation %d/%d, want 4/4", p, th)
	}
	if err := g.ActivateCatalogAbility(me.ID, walla, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a second time in one turn")
	}
}

func TestS82OldB401GoblinDiplomatsForcesTheAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	diplomats := effectstest.B21Push(g, me.ID, "Goblin Diplomats", "Creature — Goblin", m14GoblinDiplomats, 2, 1, "R")
	bear := effectstest.B21Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, diplomats, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != bear {
		t.Fatalf("pass with the Bear home = %v, want an attack-requirement refusal naming it", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
}

func TestS82OldB401GoblinDiplomatsAloneForcesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Goblin Diplomats", "Creature — Goblin", m14GoblinDiplomats, 2, 1, "R")
	effectstest.B21Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); errors.As(err, &re) {
		t.Fatalf("a requirement exists without the ability: %v", err)
	}
}

func TestS82OldB401IntoTheWildsPutsATopLandOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Into the Wilds", "Enchantment", m14IntoTheWilds, false)
	land := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("the land did not enter the battlefield")
	}
}

func TestS82OldB401IntoTheWildsLeavesANonlandOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Into the Wilds", "Enchantment", m14IntoTheWilds, false)
	top := effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a nonland card was offered")
	}
	if g.Battlefield.Contains(top) || !me.Library.Contains(top) {
		t.Error("the nonland card left the library")
	}
}

func TestS82OldB401MoltenBirthMakesTwoElementalsAndMayComeBack(t *testing.T) {
	won, lost := 0, 0
	for table := 0; table < 40 && (won == 0 || lost == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		birth := effectstest.HandCardFull(me, "Molten Birth", "Sorcery", "", m14MoltenBirth, []string{"R"})
		if err := g.CastSpell(me.ID, birth, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.CountOnBattlefield(g, "Elemental", me.ID); n != 2 {
			t.Fatalf("%d Elementals, want 2", n)
		}
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		flips := len(effectstest.RandomEvents(g, game.EventFlipCoin))
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		if len(evs) != flips+1 {
			t.Fatalf("%d coins flipped, want 1", len(evs)-flips)
		}
		inHand := me.Hand.Contains(birth)
		if evs[len(evs)-1].Won {
			won++
			if !inHand || me.Graveyard.Contains(birth) {
				t.Errorf("a won flip did not return Molten Birth to its owner's hand (hand=%v gy=%v stack=%v)", inHand, me.Graveyard.Contains(birth), g.Stack.Contains(birth))
			}
		} else {
			lost++
			if inHand || !me.Graveyard.Contains(birth) {
				t.Error("a lost flip did not leave Molten Birth in the graveyard")
			}
		}
	}
	if won == 0 || lost == 0 {
		t.Fatalf("saw %d won and %d lost flips; want both", won, lost)
	}
}

func TestS82OldB401GrimReturnReanimatesACreatureThatDiedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dead := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dead) })
	milled := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: milled, Name: "Milled Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	if !opp.Graveyard.Contains(dead) {
		t.Fatal("the Ogre did not reach the graveyard")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Grim Return", "Instant", m14GrimReturn, effectstest.CardTarget(milled)); err == nil {
		t.Fatal("a card that was not put there from the battlefield was targeted")
	}
	effectstest.CastCatalogSpell(t, g, "Grim Return", "Instant", m14GrimReturn, effectstest.CardTarget(dead))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardForTest(g, dead)
	if c == nil || c.Controller != me.ID {
		t.Fatal("the Ogre did not return under my control")
	}
}

func TestS82OldB401DevoutInvocationMakesAnAngelForEachCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Devout Invocation", "Sorcery", m14DevoutInvocation, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedOnBattlefield(t, g, a) || !effectstest.TappedOnBattlefield(t, g, b) || effectstest.TappedOnBattlefield(t, g, c) {
		t.Error("the wrong creatures were tapped")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Angel"); n != 2 {
		t.Errorf("%d Angels, want 2", n)
	}
}

func TestS82OldB401DevoutInvocationMayTapNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Devout Invocation", "Sorcery", m14DevoutInvocation, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Angel"); n != 0 {
		t.Errorf("%d Angels, want 0", n)
	}
}

func TestS82OldB401GuardianOfTheAgesLosesDefenderWhenAttackedAndGainsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := effectstest.B21Push(g, opp.ID, "Guardian of the Ages", "Artifact Creature — Golem", m14GuardianAges, 7, 7)
	attacker := effectstest.B21Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	if !effectstest.HasEffectiveKeyword(t, g, guard, "defender") {
		t.Fatal("no defender to start with")
	}
	effectstest.AttackIntoBlocks(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, guard, "defender") {
		t.Error("still has defender")
	}
	if !effectstest.HasEffectiveKeyword(t, g, guard, "trample") {
		t.Error("no trample")
	}
}

func TestS82OldB401GuardianOfTheAgesIgnoresAttacksOnOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	guard := effectstest.B21Push(g, opp.ID, "Guardian of the Ages", "Artifact Creature — Golem", m14GuardianAges, 7, 7)
	attacker := effectstest.B21Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	effectstest.AttackIntoBlocks(t, g, third.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, guard, "defender") || effectstest.HasEffectiveKeyword(t, g, guard, "trample") {
		t.Error("the Guardian changed when another player was attacked")
	}
}

func TestS82OldB401ChandrasPhoenixReturnsOnRedSpellDamageToAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	phoenix := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: phoenix, Name: "Chandra's Phoenix", TypeLine: "Creature — Phoenix", OracleID: m14ChandrasPhoenix,
		Colors: []string{"R"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	// A colourless sorcery that damages the opponent does not bring it back.
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(phoenix) {
		t.Fatal("the Phoenix returned on damage from no red instant, sorcery or planeswalker")
	}
	// A red sorcery that deals damage to the opponent (Lava Axe).
	axe := effectstest.HandCardFull(me, "Lava Axe", "Sorcery", "", m14LavaAxe, []string{"R"})
	if err := g.CastSpell(me.ID, axe, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast Lava Axe: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(phoenix) {
		t.Error("the Phoenix did not return to hand")
	}
}

func TestS82OldB401DismissIntoDreamMakesOpponentsCreaturesIllusionsThatDieWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Dismiss into Dream", "Enchantment", m14DismissIntoDream, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hasIllusion := func(id uuid.UUID) bool {
		for _, s := range effectstest.LayeredCard(t, g, id).Effective().Subtypes {
			if s == "Illusion" {
				return true
			}
		}
		return false
	}
	if !hasIllusion(theirs) {
		t.Error("the opponent's creature is not an Illusion")
	}
	if hasIllusion(mine) {
		t.Error("my creature is an Illusion")
	}
	if !hasIllusion(theirs) || !containsString(effectstest.LayeredCard(t, g, theirs).Effective().Subtypes, "Bear") {
		t.Error("its other types were not kept")
	}
	// Targeting my own creature does nothing to it.
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", m14Shock, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, mine) == nil && !me.Graveyard.Contains(mine) {
		t.Fatal("my creature vanished")
	}
	// Targeting theirs makes them sacrifice it.
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", m14Shock, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the targeted Illusion was not sacrificed")
	}
}

func containsString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
