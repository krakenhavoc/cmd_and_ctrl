package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB454Crescendo = "3c6a8ad8-9b40-4136-9cfd-2b4cf9a9d900"
	oldB454Grunt     = "43fbfeec-bcaf-48b8-befe-b7346fec5a3a"
	oldB454Captain   = "7182e6f4-6caf-474a-8759-13c56f391bef"
	oldB454VishKal   = "ba3bfe0c-2bd6-469c-802c-440b7caf7fb5"
	oldB454Szadek    = "c40c823c-5954-457a-81a6-02683da57f7f"
)

func TestOldB454CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB454Crescendo: "Crescendo of War", oldB454Grunt: "Jötun Grunt", oldB454Captain: "Sigil Captain",
		oldB454VishKal: "Vish Kal, Blood Arbiter", oldB454Szadek: "Szadek, Lord of Secrets",
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

func oldB454Power(g *game.Game, id uuid.UUID) int {
	return effectstest.ApaLive(g, id).PowerForComparison()
}

func TestOldB454CrescendoGrowsAtEveryUpkeepAndBoostsAttackersAndYourBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cres := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Crescendo of War", TypeLine: "Enchantment", OracleID: oldB454Crescendo,
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	// The turn order passed every seat's upkeep on the way to mine: one counter
	// for each, not only for the controller's.
	if got := effectstest.CountersOn(g, cres, "strife"); got != len(g.Seats) {
		t.Fatalf("%d strife counters, want %d (one at each player's upkeep)", got, len(g.Seats))
	}

	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 9)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 9)
	idleTheirs := effectstest.B12Creature(g, opp.ID, "IdleTheirs", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{attacker, idle, blocker, idleTheirs} {
		if p := oldB454Power(g, id); p != 2 {
			t.Fatalf("before combat a creature has power %d, want 2", p)
		}
	}
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	effectstest.LockInBlocks(t, g)
	if p := oldB454Power(g, attacker); p != 2+effectstest.CountersOn(g, cres, "strife") {
		t.Errorf("the attacker has power %d, want +1 per strife counter", p)
	}
	if p := oldB454Power(g, blocker); p != 2+effectstest.CountersOn(g, cres, "strife") {
		t.Errorf("your blocker has power %d, want +1 per strife counter", p)
	}
	for name, id := range map[string]uuid.UUID{"idle": idle, "idleTheirs": idleTheirs} {
		if p := oldB454Power(g, id); p != 2 {
			t.Errorf("a creature that is not in combat (%s) has power %d, want 2", name, p)
		}
	}
}

func TestOldB454CrescendoBoostsOnlyItsControllersBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Crescendo of War", TypeLine: "Enchantment", OracleID: oldB454Crescendo,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{"strife": 3},
	})
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 9)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 9)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	effectstest.LockInBlocks(t, g)
	if p := oldB454Power(g, attacker); p != 5 {
		t.Errorf("the attacker has power %d, want 5 (attackers of any controller)", p)
	}
	if p := oldB454Power(g, blocker); p != 2 {
		t.Errorf("an opponent's blocker has power %d, want 2 (only blockers you control)", p)
	}
}

func TestOldB454GruntPaysTwoCardsFromOneGraveyardOrIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grunt := effectstest.B12Push(g, me.ID, "Jötun Grunt", "Creature — Giant Soldier", oldB454Grunt, 4, 4)
	mk := func(z *game.Zone, owner uuid.UUID, name string) uuid.UUID {
		id := uuid.New()
		z.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", Owner: owner, Controller: owner})
		return id
	}
	var mine, theirs []uuid.UUID
	g.WithWriteLock(func() {
		mine = append(mine, mk(me.Graveyard, me.ID, "M1"), mk(me.Graveyard, me.ID, "M2"))
		theirs = append(theirs, mk(opp.Graveyard, opp.ID, "O1"))
	})
	libMe := me.Library.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	a := prompt.PayAction()
	if a == nil || a.Kind != game.PayActionGraveyardToLibraryBottom || a.Count != 2 || a.GroupSize != 2 {
		t.Fatalf("the payment is %+v, want two cards from a single graveyard", a)
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{mine[0], theirs[0]}); err == nil {
		t.Fatal("a split pair paid the upkeep")
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, mine); err != nil {
		t.Fatalf("pay with my two: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, grunt) != game.ZoneBattlefield {
		t.Fatalf("the Grunt left after its upkeep was paid: zone %v", effectstest.ZoneOfCard(g, grunt))
	}
	if got := effectstest.CountersOn(g, grunt, game.CounterAge); got != 1 {
		t.Errorf("%d age counters, want 1", got)
	}
	if got := me.Library.Size(); got != libMe+2 {
		t.Errorf("library %d, want %d (two cards on the bottom)", got, libMe+2)
	}

	// The next upkeep wants four cards; with two in the graveyards it cannot be paid.
	effectstest.AdvancePastCleanupForTest(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.RemoraUpkeepPrompt(t, g, me.ID); p != nil {
		if err := g.ResolvePayUnlessWithCards(p.ID, me.ID, false, nil); err != nil {
			t.Fatalf("decline: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.ZoneOfCard(g, grunt) == game.ZoneBattlefield {
		t.Error("the Grunt survived an unpaid cumulative upkeep")
	}
}

func TestOldB454SigilCaptainPutsTwoCountersOnAOneOneOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sigil Captain", "Creature — Rhino Soldier", oldB454Captain, 3, 3)
	effectstest.AdvanceToMain(t, g)
	small := effectstest.HandCardFull(me, "Squire", "Creature — Human", "", "", nil)
	big := effectstest.HandCardFull(me, "Knight", "Creature — Human", "", "", nil)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			switch me.Hand.Cards[i].InstanceID {
			case small:
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 1, 1
			case big:
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 2, 2
			}
		}
	})
	for _, id := range []uuid.UUID{big, small} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CountersOn(g, small, game.CounterPlusOne); got != 2 {
		t.Errorf("the 1/1 has %d +1/+1 counters, want 2", got)
	}
	if got := effectstest.CountersOn(g, big, game.CounterPlusOne); got != 0 {
		t.Errorf("the 2/2 has %d +1/+1 counters, want 0", got)
	}
	if p := oldB454Power(g, small); p != 3 {
		t.Errorf("the 1/1 is power %d, want 3", p)
	}
}

func TestOldB454VishKalGrowsByThePowerSacrificedAndSpendsItAsMinusCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vk := effectstest.B12Push(g, me.ID, "Vish Kal, Blood Arbiter", "Legendary Creature — Vampire", oldB454VishKal, 5, 5)
	fodder := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, vk, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without naming a creature to sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, vk, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, vk, game.CounterPlusOne); got != 3 {
		t.Fatalf("Vish Kal has %d counters, want 3 (the Ogre's power)", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, vk, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err != nil {
		t.Fatalf("remove counters: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, vk, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters left, want all removed", got)
	}
	if effectstest.ZoneOfCard(g, victim) != game.ZoneGraveyard {
		t.Errorf("the 2/2 target got -3/-3 and is in zone %v, want the graveyard", effectstest.ZoneOfCard(g, victim))
	}
	if !effectstest.HasEffectiveKeyword(t, g, vk, "flying") || !effectstest.HasEffectiveKeyword(t, g, vk, "lifelink") {
		t.Error("Vish Kal lacks flying or lifelink")
	}
}

func TestOldB454SzadekTurnsCombatDamageIntoCountersAndMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sz := effectstest.B12Push(g, me.ID, "Szadek, Lord of Secrets", "Legendary Creature — Vampire", oldB454Szadek, 5, 5)
	for i := 0; i < 8; i++ {
		effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Lib", TypeLine: "Instant"})
	}
	life, lib := opp.Life, opp.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, sz)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("the opponent's life went %d -> %d, want unchanged", life, opp.Life)
	}
	if got := opp.Library.Size(); got != lib-5 {
		t.Errorf("the opponent's library is %d, want %d (milled five)", got, lib-5)
	}
	if got := effectstest.CountersOn(g, sz, game.CounterPlusOne); got != 5 {
		t.Errorf("Szadek has %d +1/+1 counters, want 5", got)
	}
}

func TestOldB454SzadekStillDamagesCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sz := effectstest.B12Push(g, me.ID, "Szadek, Lord of Secrets", "Legendary Creature — Vampire", oldB454Szadek, 5, 5)
	blocker := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Spider", TypeLine: "Creature — Spider", Power: 0, Toughness: 4, Keywords: []string{"reach"}})
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, sz, blocker)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, blocker) != game.ZoneGraveyard {
		t.Error("a blocking creature took no damage from Szadek")
	}
	if got := effectstest.CountersOn(g, sz, game.CounterPlusOne); got != 0 {
		t.Errorf("Szadek got %d counters for damage dealt to a creature, want 0", got)
	}
}
