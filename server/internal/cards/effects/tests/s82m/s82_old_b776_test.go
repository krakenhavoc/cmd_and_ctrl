package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b776_test.go — S82 OLD batch 776 (Weatherlight: artifacts, #5629).

const (
	oldb776NullRod     = "2f83ca86-e23d-40f7-8085-6928d8cfef9b"
	oldb776Matrix      = "50261dca-bfeb-40b3-95b6-9057491ef3ad"
	oldb776StrawGolem  = "6e48981b-8971-4b7c-ae05-29c7e4091a05"
	oldb776SteelGolem  = "b5cd9c55-34ad-4e1f-92b2-527c9babc4d4"
	oldb776Automaton   = "b6f67539-9f76-42b2-96a8-575f27ba928e"
	oldb776ManaWeb     = "82909ff4-f6f7-44f1-bbd0-6e54afa5ee6d"
	oldb776BraidwoodCp = "d52b72ff-a82d-430e-94c7-675c83b43e50"
)

func TestOldB776EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldb776NullRod: "Null Rod", oldb776Matrix: "Bubble Matrix", oldb776StrawGolem: "Straw Golem",
		oldb776SteelGolem: "Steel Golem", oldb776Automaton: "Jangling Automaton", oldb776ManaWeb: "Mana Web",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

func TestOldB776NullRodStopsArtifactAbilitiesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cup := effectstest.PushCatalogPermanent(g, me.ID, "Braidwood Cup", "Artifact", oldb776BraidwoodCp, false)
	if err := g.ActivateCatalogAbility(me.ID, cup, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the Cup before the Rod: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, cup).Tapped = false })
	rod := effectstest.PushCatalogPermanent(g, me.ID, "Null Rod", "Artifact", oldb776NullRod, false)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, cup, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("an artifact's ability was activated under Null Rod")
	}
	if me.Life != life {
		t.Error("life changed")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rod) })
	if err := g.ActivateCatalogAbility(me.ID, cup, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the Cup after the Rod left: %v", err)
	}
}

func TestOldB776BubbleMatrixPreventsDamageToCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Bubble Matrix", "Artifact", oldb776Matrix, false)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	life := opp.Life
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, bear, 5)
		_ = g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 3)
	})
	var dmg int
	g.ReadSnapshot(func() { dmg = effectstest.FindBattlefieldCardForTest(g, bear).DamageMarked })
	if dmg != 0 || !g.Battlefield.Contains(bear) {
		t.Errorf("the Bear took %d damage", dmg)
	}
	if opp.Life != life-3 {
		t.Errorf("player life %d, want %d", opp.Life, life-3)
	}
}

func TestOldB776StrawGolemIsSacrificedWhenAnOpponentCastsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	golem := effectstest.B12Push(g, opp.ID, "Straw Golem", "Artifact Creature — Golem", oldb776StrawGolem, 2, 3)
	own := effectstest.B12Push(g, me.ID, "Straw Golem", "Artifact Creature — Golem", oldb776StrawGolem, 2, 3)
	sorcery := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: sorcery, Name: "Rite", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(golem) {
		t.Fatal("a noncreature spell sacrificed the Golem")
	}
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) {
		t.Error("the opponent's creature spell did not sacrifice the Golem")
	}
	if !g.Battlefield.Contains(own) {
		t.Error("the caster's own Golem was sacrificed")
	}
}

func TestOldB776SteelGolemStopsItsControllersCreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Steel Golem", "Artifact Creature — Golem", oldb776SteelGolem, 3, 4)
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast a creature spell under Steel Golem")
	}
	sorcery := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: sorcery, Name: "Rite", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a sorcery was refused: %v", err)
	}
	_ = opp
}

func TestOldB776JanglingAutomatonUntapsTheDefendersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	auto := effectstest.B12Push(g, me.ID, "Jangling Automaton", "Artifact Creature — Construct", oldb776Automaton, 3, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, theirs).Tapped = true
		effectstest.FindBattlefieldCardForTest(g, mine).Tapped = true
	})
	effectstest.DeclareAttack(t, g, opp.ID, auto)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, theirs) {
		t.Error("the defending player's creature stayed tapped")
	}
	if !effectstest.IsTapped(g, mine) {
		t.Error("the attacker's own creature untapped")
	}
}

func TestOldB776ManaWebTapsTheOpponentsMatchingLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Mana Web", "Artifact", oldb776ManaWeb, false)
	f1 := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	f2 := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	isle := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	mineF := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	mineF2 := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, mineF, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, mineF2) {
		t.Fatal("Mana Web reacted to its own controller's land")
	}
	if err := g.ActivateManaAbility(opp.ID, f1, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, f2) {
		t.Error("the opponent's other Forest stayed untapped")
	}
	if effectstest.IsTapped(g, isle) {
		t.Error("an Island was tapped by a Forest's mana")
	}
	if effectstest.IsTapped(g, mineF2) {
		t.Error("the Web's controller's land was tapped")
	}
}
