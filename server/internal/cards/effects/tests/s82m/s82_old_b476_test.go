package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB476Unholy     = "090d88a9-7f2d-4bd1-a30a-7c48d05068be"
	oldB476Cocoon     = "79f575b3-ebd7-4d6a-acb9-27c16bcc7a39"
	oldB476Ascension  = "916db499-7f0a-45f7-8abe-94818a53e494"
	oldB476Holy       = "9357de36-f8be-4f49-b2c8-9fe9eaf82b07"
	oldB476Plague     = "9d637892-26d7-46f1-bd7d-89bda523f649"
	oldB476Cloud      = "0b751cc7-3a19-4207-9e64-83e18d279588"
	oldB476Excomm     = "6c2cf22e-5dbb-4fa9-8ece-17069e6f1d7f"
	oldB476Moth       = "8dcfa697-5350-43b9-896f-c5eefcc3893a"
	oldB476Phoenix    = "9c6edb99-2625-4d80-a2e8-1b6039c0a960"
	oldB476Bloodcraze = "a10e5b31-c9c1-45ca-85cc-4a50ed8c4946"
)

func TestOldB476CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB476Unholy: "Unholy Strength", oldB476Cocoon: "Primal Cocoon", oldB476Ascension: "Armored Ascension",
		oldB476Holy: "Holy Strength", oldB476Plague: "Necrotic Plague", oldB476Cloud: "Cloud Elemental",
		oldB476Excomm: "Excommunicate", oldB476Moth: "Goldenglow Moth", oldB476Phoenix: "Magma Phoenix",
		oldB476Bloodcraze: "Bloodcrazed Goblin",
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

// oldB476Aura pushes an Aura of mine and attaches it to host.
func oldB476Aura(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := effectstest.B12Push(g, owner, name, "Enchantment — Aura", oracle, 0, 0)
	effectstest.NcdAttach(t, g, aura, host)
	return aura
}

func TestOldB476StrengthAurasPumpTheEnchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	unholy := oldB476Aura(t, g, me.ID, "Unholy Strength", oldB476Unholy, a)
	oldB476Aura(t, g, me.ID, "Holy Strength", oldB476Holy, b)
	if p, tough := effectstest.PtOf(t, g, a); p != 4 || tough != 3 {
		t.Errorf("Unholy Strength: %d/%d, want 4/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, b); p != 3 || tough != 4 {
		t.Errorf("Holy Strength: %d/%d, want 3/4", p, tough)
	}
	effectstest.B27Kill(g, unholy)
	if p, tough := effectstest.PtOf(t, g, a); p != 2 || tough != 2 {
		t.Errorf("after the Aura left: %d/%d, want 2/2", p, tough)
	}
}

func TestOldB476ArmoredAscensionCountsYourPlainsAndGrantsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB476Aura(t, g, me.ID, "Armored Ascension", oldB476Ascension, bear)
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("with no Plains: power %d, want 2", p)
	}
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	}
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("with three Plains: %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the enchanted creature has no flying")
	}
}

func TestOldB476PrimalCocoonGrowsEachUpkeepAndGoesWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := oldB476Aura(t, g, me.ID, "Primal Cocoon", oldB476Cocoon, bear)
	start := g.Turn.Seq
	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepUpkeep || g.Turn.Seq <= start {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, bear); n != 1 {
		t.Fatalf("+1/+1 counters after my upkeep = %d, want 1", n)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, aura); ok {
		t.Error("Primal Cocoon stayed on the battlefield after the creature attacked")
	}
	if n := effectstest.PlusOneCounters(g, bear); n != 1 {
		t.Errorf("the counter is gone: %d", n)
	}
}

func TestOldB476CloudElementalBlocksOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	walker := effectstest.B12Creature(g, me.ID, "Walker", "Creature — Bear", 2, 2)
	flyer := effectstest.B16Creature(g, me.ID, "Flyer", "Creature — Bird", 2, 2, "W")
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, flyer)
		c.Keywords = append(c.Keywords, "flying")
	})
	cloud := effectstest.B12Push(g, opp.ID, "Cloud Elemental", "Creature — Elemental", oldB476Cloud, 2, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{walker, flyer} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if got := effectstest.BlockRefusal(t, g, cloud, walker); got == "" {
		t.Error("Cloud Elemental blocked a creature without flying")
	}
	if got := effectstest.BlockRefusal(t, g, cloud, flyer); got != "" {
		t.Errorf("Cloud Elemental could not block a flyer: %v", got)
	}
}

func TestOldB476ExcommunicatePutsTheCreatureOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// Steal it first: it goes to its OWNER's library, not mine.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, victim).Controller = me.ID })
	effectstest.CastCatalogSpell(t, g, "Excommunicate", "Sorcery", oldB476Excomm, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Fatal("the creature stayed on the battlefield")
	}
	if top := effectstest.LibraryTopIDs(opp, 1); len(top) != 1 || top[0] != victim {
		t.Errorf("top of the owner's library = %v, want the creature", top)
	}
}

func TestOldB476GoldenglowMothMayGainFourWhenItBlocks(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
		moth := effectstest.B12Push(g, opp.ID, "Goldenglow Moth", "Creature — Insect", oldB476Moth, 0, 1)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		before := opp.Life
		if err := g.DeclareBlocker(moth, attacker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
		effectstest.LockInBlocks(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, opp.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := before
		if yes {
			want += 4
		}
		if opp.Life != want {
			t.Errorf("answered %v: life %d, want %d", yes, opp.Life, want)
		}
	}
}

func TestOldB476MagmaPhoenixBlastsOnDeathAndComesBackFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phoenix := effectstest.B12Push(g, me.ID, "Magma Phoenix", "Creature — Phoenix", oldB476Phoenix, 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)
	small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	lifeMe, lifeOpp := me.Life, opp.Life
	effectstest.B27Kill(g, phoenix)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != lifeMe-3 || opp.Life != lifeOpp-3 {
		t.Errorf("life me %d→%d, opponent %d→%d, want 3 each", lifeMe, me.Life, lifeOpp, opp.Life)
	}
	if _, ok := effectstest.BattlefieldCard(g, small); ok {
		t.Error("the 2/2 survived")
	}
	if got := effectstest.DamageMarkedOn(g, big); got != 3 {
		t.Errorf("the 5/5 has %d damage, want 3", got)
	}
	if !effectstest.InGraveyard(me, phoenix) {
		t.Fatal("the Phoenix is not in the graveyard")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, phoenix, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(phoenix) {
		t.Error("the Phoenix was not returned to hand")
	}
}

func TestOldB476BloodcrazedGoblinNeedsAnOpponentDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	goblin := effectstest.B12Push(g, me.ID, "Bloodcrazed Goblin", "Creature — Goblin Berserker", oldB476Bloodcraze, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(goblin, opp.ID); err == nil {
		t.Fatal("the Goblin attacked before any opponent was dealt damage")
	}
	// Damage to its own controller does not count; damage to an opponent does.
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(other, me.ID, 2) })
	if err := g.DeclareAttacker(goblin, opp.ID); err == nil {
		t.Fatal("the Goblin attacked after only its controller was dealt damage")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(other, opp.ID, 1) })
	if err := g.DeclareAttacker(goblin, opp.ID); err != nil {
		t.Fatalf("the Goblin could not attack after an opponent was dealt damage: %v", err)
	}
}

func TestOldB476NecroticPlagueSpreadsToAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// My Aura on THEIR creature: the opponent chooses, among creatures they
	// don't control, the one it comes back on.
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirOther := effectstest.B12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)
	aura := oldB476Aura(t, g, me.ID, "Necrotic Plague", oldB476Plague, victim)
	effectstest.B27Kill(g, victim)
	ch := effectstest.LatestPickTarget(g, opp.ID)
	if ch == nil {
		effectstest.PassPriorityAroundTable(t, g)
		ch = effectstest.LatestPickTarget(g, opp.ID)
	}
	if ch == nil {
		t.Fatal("the dead creature's controller was not asked for a target")
	}
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirOther}); err == nil {
		t.Fatal("the chooser's own creature was accepted")
	}
	effectstest.PickCard(t, g, opp.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, aura); !ok {
		t.Fatal("Necrotic Plague did not return to the battlefield")
	}
	if !effectstest.AttachedToNow(t, g, aura, mine) {
		t.Error("it returned attached to the wrong permanent")
	}
}

func TestOldB476NecroticPlagueMakesItsCreatureSacrificeItselfAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	aura := oldB476Aura(t, g, me.ID, "Necrotic Plague", oldB476Plague, victim)
	start := g.Turn.Seq
	for g.Turn.ActiveSeat != 1 || g.Turn.Step != game.StepUpkeep || g.Turn.Seq <= start {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Fatal("the enchanted creature was not sacrificed at its controller's upkeep")
	}
	effectstest.PickCard(t, g, opp.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedToNow(t, g, aura, mine) {
		t.Error("the Aura did not return attached to my creature")
	}
}
