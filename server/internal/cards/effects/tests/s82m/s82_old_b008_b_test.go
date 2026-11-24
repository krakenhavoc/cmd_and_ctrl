package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 8, group B (#5462): Soulherder, Drana, Tempered Sliver and
// Treasonous Ogre, driven through the real engine.

const (
	s82OldB008Soulherder = "92019547-f6db-4ea6-8356-d0a90ace5662"
	s82OldB008Drana      = "3ae3b901-4d36-40e2-b861-26209fa1e823"
	s82OldB008Sliver     = "536c36d9-5ef4-441c-a6c8-5806644d805c"
	s82OldB008Ogre       = "826d4279-1576-4898-a1c2-26fd547fb0d0"
)

func TestS82OldB008BCardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82OldB008Soulherder: "Soulherder", s82OldB008Drana: "Drana, the Last Bloodchief",
		s82OldB008Sliver: "Tempered Sliver", s82OldB008Ogre: "Treasonous Ogre",
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

// --- Treasonous Ogre ---------------------------------------------------------

func TestS82OldB008BOgreDethronesAtTheLifeLeader(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.PushCatalogPermanent(g, me.ID, "Treasonous Ogre", "Creature — Ogre Shaman", s82OldB008Ogre, false)
	effectstest.DeclareAttack(t, g, opp.ID, ogre) // everyone is tied for the most life
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, ogre); n != 1 {
		t.Errorf("dethrone put %d counters, want 1", n)
	}
}

func TestS82OldB008BOgreDoesNotDethroneATrailingPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.PushCatalogPermanent(g, me.ID, "Treasonous Ogre", "Creature — Ogre Shaman", s82OldB008Ogre, false)
	opp.Life = 5
	effectstest.DeclareAttack(t, g, opp.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, ogre); n != 0 {
		t.Errorf("dethrone put %d counters against a trailing player, want 0", n)
	}
}

func TestS82OldB008BOgrePaysThreeLifeForRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ogre := effectstest.PushCatalogPermanent(g, me.ID, "Treasonous Ogre", "Creature — Ogre Shaman", s82OldB008Ogre, true)
	life := me.Life
	effectstest.ActivateManaFor(t, g, me.ID, ogre, 0, game.ManaAbilityParams{})
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("pool holds %d mana, want 1", got)
	}
	if c, _ := effectstest.BattlefieldCard(g, ogre); c.Tapped {
		t.Error("the ability has no tap cost, but the Ogre is tapped")
	}
}

func TestS82OldB008BOgreCannotPayLifeItDoesNotHave(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ogre := effectstest.PushCatalogPermanent(g, me.ID, "Treasonous Ogre", "Creature — Ogre Shaman", s82OldB008Ogre, true)
	me.Life = 2
	if err := g.ActivateManaAbility(me.ID, ogre, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("paid 3 life with 2 life")
	}
	if me.Life != 2 || me.ManaPool.Total() != 0 {
		t.Errorf("a refused activation changed life to %d or the pool to %d", me.Life, me.ManaPool.Total())
	}
}

// --- Tempered Sliver ---------------------------------------------------------

func TestS82OldB008BSliverGrantsGrowthToEverySliverYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tempered := effectstest.PushCatalogPermanent(g, me.ID, "Tempered Sliver", "Creature — Sliver", s82OldB008Sliver, false)
	mine := effectstest.B12Creature(g, me.ID, "Other Sliver", "Creature — Sliver", 1, 1)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Sliver", "Creature — Sliver", 1, 1)
	g.RunStateChecksForTest()
	for _, id := range []uuid.UUID{tempered, mine, bear, theirs} {
		effectstest.DealCombatDamageToPlayer(g, id, opp.ID, 1)
		effectstest.PassPriorityAroundTable(t, g)
	}
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want int
	}{"Tempered Sliver itself": {tempered, 1}, "my other Sliver": {mine, 1}, "my non-Sliver": {bear, 0}, "their Sliver": {theirs, 0}} {
		if got := effectstest.PlusOnes(g, tc.id); got != tc.want {
			t.Errorf("%s has %d counters, want %d", name, got, tc.want)
		}
	}
}

func TestS82OldB008BSliverGrantsNothingToACreatureThatDealsNoDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tempered := effectstest.PushCatalogPermanent(g, me.ID, "Tempered Sliver", "Creature — Sliver", s82OldB008Sliver, false)
	other := effectstest.B12Creature(g, me.ID, "Other Sliver", "Creature — Sliver", 1, 1)
	g.RunStateChecksForTest()
	// Damage to a creature is not damage to a player.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: tempered, Target: other, Amount: 1, Combat: true})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, tempered); n != 0 {
		t.Errorf("a counter came from damage to a creature: %d", n)
	}
}

// --- Soulherder --------------------------------------------------------------

func s82OldB008Exile(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
}

func TestS82OldB008BSoulherderGrowsWhenAnyCreatureIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herder := effectstest.PushCatalogPermanent(g, me.ID, "Soulherder", "Creature — Spirit", s82OldB008Soulherder, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	s82OldB008Exile(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	s82OldB008Exile(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	s82OldB008Exile(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, herder); n != 2 {
		t.Errorf("Soulherder has %d counters, want 2 (two creatures exiled, an artifact is not one)", n)
	}
}

func TestS82OldB008BSoulherderDoesNotGrowWhenACreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herder := effectstest.PushCatalogPermanent(g, me.ID, "Soulherder", "Creature — Spirit", s82OldB008Soulherder, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, herder); n != 0 {
		t.Errorf("a creature that died gave %d counters, want 0", n)
	}
}

func TestS82OldB008BSoulherderBlinksAnotherCreatureAtYourEndStepAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herder := effectstest.PushCatalogPermanent(g, me.ID, "Soulherder", "Creature — Spirit", s82OldB008Soulherder, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no target prompt at your end step: %+v", g.PendingChoices)
	}
	if effectstest.HasID(pick.PickTargetCards, herder) {
		t.Error("Soulherder offered itself: the clause says another creature")
	}
	if !effectstest.HasID(pick.PickTargetCards, bear) {
		t.Fatal("my other creature is not offered")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	yn := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(yn.ID, me.ID, true); err != nil {
		t.Fatalf("answering yes: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.FindBattlefieldByName(g, "Bear")
	if back == uuid.Nil || back == bear {
		t.Fatalf("Bear back=%v (old %v): want a new object on the battlefield", back, bear)
	}
	if n := effectstest.PlusOnes(g, herder); n != 1 {
		t.Errorf("Soulherder has %d counters after its own blink, want 1", n)
	}
}

func TestS82OldB008BSoulherderMayDeclineTheBlink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herder := effectstest.PushCatalogPermanent(g, me.ID, "Soulherder", "Creature — Spirit", s82OldB008Soulherder, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PickCard(t, g, me.ID, bear)
	yn := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(yn.ID, me.ID, false); err != nil {
		t.Fatalf("declining: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, bear) {
		t.Error("declining still moved the creature")
	}
	if n := effectstest.PlusOnes(g, herder); n != 0 {
		t.Errorf("Soulherder grew (%d) with nothing exiled", n)
	}
}

func TestS82OldB008BSoulherderReturnsAStolenCreatureToItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Soulherder", "Creature — Spirit", s82OldB008Soulherder, false)
	stolen := effectstest.PushVanillaCreature(g, opp.ID, "Stolen Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == stolen {
				g.Battlefield.Cards[i].Controller = me.ID
			}
		}
	})
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PickCard(t, g, me.ID, stolen)
	yn := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(yn.ID, me.ID, true); err != nil {
		t.Fatalf("answering yes: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.FindBattlefieldByName(g, "Stolen Bear")
	if back == uuid.Nil {
		t.Fatal("the creature did not return")
	}
	if c, _ := effectstest.BattlefieldCard(g, back); c.Controller != opp.ID {
		t.Errorf("returned under %v's control, want its owner's (%v)", c.Controller, opp.ID)
	}
}

// --- Drana, the Last Bloodchief ---------------------------------------------

func TestS82OldB008BDranaLetsTheDefenderChooseTheCreatureYouReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drana := effectstest.PushCatalogPermanent(g, me.ID, "Drana, the Last Bloodchief", "Legendary Creature — Vampire Cleric", s82OldB008Drana, false)
	small := effectstest.B38Graveyard(g, me.ID, "Small Zombie", "Creature — Zombie", "{1}{B}", 2, 2)
	big := effectstest.B38Graveyard(g, me.ID, "Big Zombie", "Creature — Zombie", "{4}{B}", 5, 5)
	legend := effectstest.B38Graveyard(g, me.ID, "Legendary Lich", "Legendary Creature — Zombie", "{3}{B}", 4, 4)
	spell := effectstest.B38Graveyard(g, me.ID, "Dark Ritual", "Instant", "{B}", 0, 0)
	theirs := effectstest.B38Graveyard(g, opp.ID, "Their Zombie", "Creature — Zombie", "{1}{B}", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, drana)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the attacker was asked: the defending player chooses")
	}
	pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the defending player was not asked: %+v", g.PendingChoices)
	}
	for id, name := range map[uuid.UUID]string{legend: "a legendary creature card", spell: "an instant card", theirs: "the defender's own creature card"} {
		if effectstest.HasID(pick.ChooseCards, id) {
			t.Errorf("%s was offered", name)
		}
	}
	if !effectstest.HasID(pick.ChooseCards, small) || !effectstest.HasID(pick.ChooseCards, big) {
		t.Fatalf("a nonlegendary creature card of mine is not offered: %v", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.FindBattlefieldByName(g, "Small Zombie")
	if back == uuid.Nil {
		t.Fatal("the chosen card did not return")
	}
	c, _ := effectstest.BattlefieldCard(g, back)
	if c.Controller != me.ID {
		t.Errorf("returned under %v's control, want mine", c.Controller)
	}
	if n := effectstest.PlusOnes(g, back); n != 1 {
		t.Errorf("returned with %d +1/+1 counters, want 1", n)
	}
	if !c.HasSubtype("Vampire") || !c.HasSubtype("Zombie") {
		t.Errorf("types %q: want Vampire in addition to Zombie", c.TypeLine)
	}
	if !me.Graveyard.Contains(big) {
		t.Error("the unchosen card left the graveyard")
	}
}

func TestS82OldB008BDranaAsksNobodyWithNoCreatureCardToReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drana := effectstest.PushCatalogPermanent(g, me.ID, "Drana, the Last Bloodchief", "Legendary Creature — Vampire Cleric", s82OldB008Drana, false)
	effectstest.B38Graveyard(g, me.ID, "Legendary Lich", "Legendary Creature — Zombie", "{3}{B}", 4, 4)
	effectstest.DeclareAttack(t, g, opp.ID, drana)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) != nil {
		t.Error("the defender was asked with only a legendary creature card in the graveyard")
	}
}
