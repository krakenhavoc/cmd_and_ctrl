package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b319_test.go — S82 older sets, batch 319: the Magic Origins Auras.

const (
	oldB319ConsecratedByBlood = "270508aa-b2e7-4066-be2f-4308cffe2b94"
	oldB319CallOfTheFullMoon  = "3e34a3de-ae30-4d7b-88f1-03e2b79c4451"
	oldB319GraspOfHieromancer = "51310af7-796d-4cf0-871e-5055765ffcdd"
	oldB319MantleOfWebs       = "7549206e-22a9-4d56-b3bc-f6f2f5e87e8a"
	oldB319InfectiousBloodlst = "9cd8ca01-e8a9-47ea-8899-b016a2dcaea8"
	oldB319StratusWalk        = "eda37caa-d676-460a-804b-3ba9d8a8d044"
)

func TestOldB319EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB319ConsecratedByBlood: "Consecrated by Blood", oldB319CallOfTheFullMoon: "Call of the Full Moon",
		oldB319GraspOfHieromancer: "Grasp of the Hieromancer", oldB319MantleOfWebs: "Mantle of Webs",
		oldB319InfectiousBloodlst: "Infectious Bloodlust", oldB319StratusWalk: "Stratus Walk",
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

func oldB319Attach(g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	aura := effectstest.PushCatalogPermanent(g, owner, name, effectstest.AuraTypeLine, oracle, false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	})
	return aura
}

func TestOldB319ConsecratedByBloodPumpsFliesAndRegeneratesForTwoSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	fodder1 := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Ogre", 1, 1)
	fodder2 := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Ogre", 1, 1)
	oldB319Attach(g, me.ID, "Consecrated by Blood", oldB319ConsecratedByBlood, host)
	effectstest.AdvanceToMain(t, g)
	if p, tough := effectstest.PtOf(t, g, host); p != 4 || tough != 4 {
		t.Errorf("the host is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, host, "flying") {
		t.Error("the host has no flying")
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the enchanted creature has no granted ability")
	}
	// One other creature is not two.
	if err := g.ActivateCatalogAbility(me.ID, host, idx, game.ActivateAbilityParams{Ref: ref, SacrificeIDs: []uuid.UUID{fodder1}}); err == nil {
		t.Error("sacrificing one creature paid for the ability")
	}
	// The creature itself cannot be one of the two.
	if err := g.ActivateCatalogAbility(me.ID, host, idx, game.ActivateAbilityParams{Ref: ref, SacrificeIDs: []uuid.UUID{host, fodder1}}); err == nil {
		t.Error("the enchanted creature itself was sacrificed to its own ability")
	}
	effectstest.B16Activate(t, g, me.ID, host, idx, game.ActivateAbilityParams{Ref: ref, SacrificeIDs: []uuid.UUID{fodder1, fodder2}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder1) || g.Battlefield.Contains(fodder2) {
		t.Error("the two creatures were not sacrificed")
	}
	effectstest.P1Destroy(t, g, host)
	if !g.Battlefield.Contains(host) {
		t.Error("the regenerated creature should survive being destroyed")
	}
}

func TestOldB319CallOfTheFullMoonPumpsAndIsSacrificedAfterTwoSpellsByOnePlayer(t *testing.T) {
	setup := func(t *testing.T, spells func(g *game.Game)) (*game.Game, uuid.UUID, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
		aura := oldB319Attach(g, me.ID, "Call of the Full Moon", oldB319CallOfTheFullMoon, host)
		if p, tough := effectstest.PtOf(t, g, host); p != 5 || tough != 4 {
			t.Fatalf("the host is %d/%d, want 5/4", p, tough)
		}
		if !effectstest.HasKeywordOnBattlefield(t, g, host, "trample") {
			t.Fatal("the host has no trample")
		}
		g.WithWriteLock(func() {
			if g.SpellsCastThisTurn == nil {
				g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{}
			}
		})
		spells(g)
		next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
		effectstest.AdvanceToUpkeepOfSeat(t, g, next)
		effectstest.PassPriorityAroundTable(t, g)
		return g, aura, host
	}
	t.Run("one player cast two", func(t *testing.T) {
		g, aura, host := setup(t, func(g *game.Game) {
			g.WithWriteLock(func() { g.SpellsCastThisTurn[g.Seats[g.Turn.ActiveSeat].ID] = game.CastTally{Total: 2} })
		})
		if g.Battlefield.Contains(aura) {
			t.Error("the Aura was not sacrificed")
		}
		if !g.Battlefield.Contains(host) {
			t.Error("the creature went with it")
		}
	})
	t.Run("two players cast one each", func(t *testing.T) {
		g, aura, _ := setup(t, func(g *game.Game) {
			g.WithWriteLock(func() {
				g.SpellsCastThisTurn[g.Seats[0].ID] = game.CastTally{Total: 1}
				g.SpellsCastThisTurn[g.Seats[1].ID] = game.CastTally{Total: 1}
			})
		})
		if !g.Battlefield.Contains(aura) {
			t.Error("the Aura was sacrificed though no player cast two spells")
		}
	})
	t.Run("one spell", func(t *testing.T) {
		g, aura, _ := setup(t, func(g *game.Game) {
			g.WithWriteLock(func() { g.SpellsCastThisTurn[g.Seats[g.Turn.ActiveSeat].ID] = game.CastTally{Total: 1} })
		})
		if !g.Battlefield.Contains(aura) {
			t.Error("the Aura was sacrificed after a single spell")
		}
	})
}

func TestOldB319GraspOfTheHieromancerTapsADefendingCreatureOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	oldB319Attach(g, me.ID, "Grasp of the Hieromancer", oldB319GraspOfHieromancer, host)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Ogre", 2, 2)
	if p, tough := effectstest.PtOf(t, g, host); p != 3 || tough != 3 {
		t.Fatalf("the host is %d/%d, want 3/3", p, tough)
	}
	effectstest.DeclareAttack(t, g, opp.ID, host)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.PickCard(t, g, me.ID, blocker)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, blocker) {
		t.Error("the defending player's creature was not tapped")
	}
}

func TestOldB319MantleOfWebsGivesPlusOnePlusThreeAndReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	aura := oldB319Attach(g, me.ID, "Mantle of Webs", oldB319MantleOfWebs, host)
	if p, tough := effectstest.PtOf(t, g, host); p != 3 || tough != 5 {
		t.Errorf("the host is %d/%d, want 3/5", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, host, "reach") {
		t.Error("the host has no reach")
	}
	effectstest.B27Kill(g, aura)
	if effectstest.HasKeywordOnBattlefield(t, g, host, "reach") {
		t.Error("the host kept reach after the Aura left")
	}
}

func TestOldB319InfectiousBloodlustForcesAttackAndFindsAnotherWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	oldB319Attach(g, me.ID, "Infectious Bloodlust", oldB319InfectiousBloodlst, host)
	other := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Infectious Bloodlust", TypeLine: effectstest.AuraTypeLine, Owner: me.ID, Controller: me.ID})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bystander", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if p, tough := effectstest.PtOf(t, g, host); p != 4 || tough != 3 {
		t.Errorf("the host is %d/%d, want 4/3", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, host, "haste") {
		t.Error("the host has no haste")
	}
	if len(effectstest.GaLayered(t, g, host).Effective().AttackRequirements) == 0 {
		t.Error("the host does not have to attack each combat")
	}
	hand := me.Hand.Size()
	effectstest.B27Kill(g, host)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchPromptFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after the enchanted creature died")
	}
	if err := g.BeginSearchLibrary(c.ID, me.ID); err != nil {
		t.Fatalf("BeginSearchLibrary: %v", err)
	}
	c = effectstest.SearchPromptFor(g, me.ID)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{other}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !effectstest.HandHasID(me, other) || me.Hand.Size() != hand+1 {
		t.Error("the named card did not reach the hand")
	}
}

func TestOldB319InfectiousBloodlustDoesNothingWhenTheCreatureIsNotEnchantedByIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Infectious Bloodlust", TypeLine: effectstest.AuraTypeLine, Owner: me.ID, Controller: me.ID})
	effectstest.B27Kill(g, host)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchPromptFor(g, me.ID) != nil {
		t.Error("a search was offered with no Aura on the creature")
	}
}

func TestOldB319StratusWalkDrawsFliesAndBlocksOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	effectstest.AuraSeedLibrary(me, 3)
	bear := effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Stratus Walk", effectstest.AuraTypeLine, oldB319StratusWalk, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d, want %d (the Aura drew a card)", got, handBefore+1)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Fatal("the enchanted creature does not have flying")
	}
	ground := effectstest.B12Creature(g, me.ID, "Ground attacker", "Creature — Bear", 2, 2)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flying attacker", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AttackIntoBlocks(t, g, them.ID, ground, flyer)
	if effectstest.BlockRefusal(t, g, bear, ground) == "" {
		t.Error("the enchanted creature blocked a creature without flying")
	}
	if got := effectstest.BlockRefusal(t, g, bear, flyer); got != "" {
		t.Errorf("the enchanted creature was refused a flyer: %q", got)
	}
}
