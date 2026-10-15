package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// necromancy_test.go — #3156: Necromancy end to end. The cast at sorcery
// and at instant speed, the cleanup sacrifice and the ruling that only
// its own flash permission earns it, the enters trigger that makes it an
// Aura, the target leaving before resolution, and the Aura losing its
// creature.

const necromancyOracle = "699b192c-f3f2-4ac5-9ca9-e35bcf08ad8a"

// nmAdvanceTo moves the turn forward until it reaches `step`, with the
// stack empty.
func nmAdvanceTo(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for i := 0; i < 32 && g.Turn.Step != step; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Step != step {
		t.Fatalf("never reached %s (at %s)", step, g.Turn.Step)
	}
}

// nmCast puts Necromancy into the active seat's hand and casts it, in the
// current step, for {2}{B}. It is not targeted.
func nmCast(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Necromancy", TypeLine: "Enchantment", OracleID: necromancyOracle,
		ManaCost: "{2}{B}", Colors: []string{"B"}, Owner: me.ID, Controller: me.ID,
	})
	fillPoolColored(me, "B", 3)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Necromancy: %v", err)
	}
	return id
}

// nmResolveAndTarget resolves the spell and answers its enters trigger's
// target with `target`, leaving the trigger on the stack.
func nmResolveAndTarget(t *testing.T, g *game.Game, spell, target uuid.UUID) {
	t.Helper()
	adResolveSpell(t, g, spell)
	if adZone(g, spell) != game.ZoneBattlefield {
		t.Fatalf("Necromancy is in %v, want the battlefield", adZone(g, spell))
	}
	pickTriggerTarget(t, g, g.Seats[g.Turn.ActiveSeat].ID, target)
	if triggerOnStack(g, spell) == nil {
		t.Fatal("Necromancy's enters trigger is not on the stack")
	}
}

// nmCleanupTriggers counts the delayed triggers waiting for a cleanup step.
func nmCleanupTriggers(g *game.Game) int {
	n := 0
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.At == game.StepCleanup {
			n++
		}
	}
	return n
}

// Cast in a main phase with the stack empty: it enters as an enchantment
// that is not yet an Aura, its trigger makes it one, returns the
// opponent's creature under the caster's control and attaches to it. No
// cleanup sacrifice is owed.
func TestNecromancyAtSorcerySpeedReanimatesAndBecomesAnAura(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	nmAdvanceTo(t, g, game.StepPrecombatMain)
	n := nmCast(t, g)
	if item := g.StackMeta[n]; item == nil || item.OwnFlashCast {
		t.Fatalf("a sorcery-speed cast is marked as cast with its own flash: %+v", item)
	}
	nmResolveAndTarget(t, g, n, zombie)
	if adCard(t, g, n).IsAura() {
		t.Fatal("Necromancy is an Aura before its trigger resolved; it enters as an enchantment")
	}

	passPriorityAroundTable(t, g)

	z := adCard(t, g, zombie)
	if adZone(g, zombie) != game.ZoneBattlefield || z.Controller != me.ID {
		t.Fatalf("the creature should be on the battlefield under the caster's control: zone=%v controller=%v", adZone(g, zombie), z.Controller)
	}
	c := adCard(t, g, n)
	if !c.IsAura() {
		t.Error("Necromancy did not become an Aura")
	}
	if !c.IsAttachedTo(zombie) || !c.EnchantSwapped || c.EnchantReturned.ID != zombie {
		t.Fatalf("Necromancy should enchant the creature it returned: attached=%+v swapped=%v returned=%+v", c.AttachedTo, c.EnchantSwapped, c.EnchantReturned)
	}
	if nmCleanupTriggers(g) != 0 {
		t.Error("a sorcery-speed cast owes a cleanup sacrifice")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Errorf("%d delayed triggers waiting, want the one leaves-the-battlefield sacrifice", len(g.DelayedTriggers))
	}

	nmAdvanceTo(t, g, game.StepEnd)
	advanceToUpkeepOf(t, g, 1)
	if adZone(g, n) != game.ZoneBattlefield || adZone(g, zombie) != game.ZoneBattlefield {
		t.Errorf("a sorcery-speed Necromancy was sacrificed at cleanup: aura in %v, creature in %v", adZone(g, n), adZone(g, zombie))
	}
}

// Cast in the end step on its own flash: the permanent it becomes is
// sacrificed at the beginning of the next cleanup step, and the creature
// it returned goes with it through the leaves-the-battlefield trigger.
func TestNecromancyAtInstantSpeedIsSacrificedAtCleanup(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	nmAdvanceTo(t, g, game.StepEnd)
	n := nmCast(t, g)
	if item := g.StackMeta[n]; item == nil || !item.OwnFlashCast {
		t.Fatal("an end-step cast is not marked as cast with its own flash")
	}
	nmResolveAndTarget(t, g, n, zombie)
	passPriorityAroundTable(t, g)
	if adZone(g, zombie) != game.ZoneBattlefield || !adCard(t, g, n).IsAttachedTo(zombie) {
		t.Fatalf("the creature should be back with Necromancy on it: zone=%v", adZone(g, zombie))
	}
	if nmCleanupTriggers(g) != 1 {
		t.Fatalf("%d cleanup sacrifices scheduled, want 1", nmCleanupTriggers(g))
	}

	// The cleanup step: the trigger gives the active player priority
	// (CR 514.3a), and resolving it sacrifices Necromancy, whose own
	// delayed trigger then sacrifices the creature.
	// An empty hand, so the CR 514.1 discard does not hold the step.
	me.Hand.Cards = nil
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if g.Turn.Step != game.StepCleanup {
		t.Fatalf("at %s, want the cleanup step", g.Turn.Step)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(n) {
		t.Errorf("Necromancy should have been sacrificed at cleanup, it is in %v", adZone(g, n))
	}
	if who, ok := adSacrificedBy(g, n); !ok || who != me.ID {
		t.Errorf("Necromancy sacrificed by %v (ok=%v), want its controller %v", who, ok, me.ID)
	}
	if !opp.Graveyard.Contains(zombie) {
		t.Errorf("the returned creature should have been sacrificed with it, it is in %v", adZone(g, zombie))
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers still waiting", len(g.DelayedTriggers))
	}
}

// Ruling 2022-12-08: the sacrifice applies only when Necromancy was cast
// using its own permission. With another flash permission (a Vedalken
// Orrery's), an end-step cast owes nothing.
func TestNecromancyCastWithAnotherFlashPermissionIsKept(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	g.WithWriteLock(func() {
		g.GrantCastTimingForEffect(me.ID, game.CastTimingRule{Timing: game.TimingFlash},
			"test: you may cast spells as though they had flash", uuid.Nil, game.Duration{})
	})
	nmAdvanceTo(t, g, game.StepEnd)
	n := nmCast(t, g)
	if item := g.StackMeta[n]; item == nil || item.OwnFlashCast {
		t.Fatal("a cast another flash permission allowed is marked as cast with Necromancy's own")
	}
	nmResolveAndTarget(t, g, n, zombie)
	passPriorityAroundTable(t, g)
	if nmCleanupTriggers(g) != 0 {
		t.Errorf("%d cleanup sacrifices scheduled, want none", nmCleanupTriggers(g))
	}
}

// Without its rider Necromancy would be sorcery-speed: a Teferi-style
// restriction closes the instant-speed window it opens (CR 101.2).
func TestNecromancysFlashIsClosedBySorceryOnlyRestriction(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	nmAdvanceTo(t, g, game.StepEnd)
	id := uuid.New()
	card := game.Card{InstanceID: id, Name: "Necromancy", TypeLine: "Enchantment", OracleID: necromancyOracle,
		ManaCost: "{2}{B}", Owner: me.ID, Controller: me.ID}
	var open bool
	g.ReadSnapshot(func() { open = g.CastTimingOpenLocked(me.ID, card, game.ZoneHand, nil) })
	if !open {
		t.Fatal("Necromancy's own flash should open the end step")
	}
	g.WithWriteLock(func() {
		g.GrantCastTimingForEffect(me.ID, game.CastTimingRule{Timing: game.TimingSorcery},
			"test: cast spells only any time you could cast a sorcery", uuid.Nil, game.Duration{})
	})
	g.ReadSnapshot(func() { open = g.CastTimingOpenLocked(me.ID, card, game.ZoneHand, nil) })
	if open {
		t.Error("a sorcery-only restriction should close Necromancy's flash")
	}
}

// The target leaves before the trigger resolves: the trigger's only
// target is illegal, so it does nothing (CR 608.2b). Necromancy stays on
// the battlefield as an enchantment, not an Aura, and nothing is owed.
func TestNecromancysTargetLeavingLeavesAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	nmAdvanceTo(t, g, game.StepPrecombatMain)
	n := nmCast(t, g)
	nmResolveAndTarget(t, g, n, zombie)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(zombie); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if adZone(g, n) != game.ZoneBattlefield {
		t.Fatalf("Necromancy should stay on the battlefield, it is in %v", adZone(g, n))
	}
	c := adCard(t, g, n)
	if c.IsAura() || c.EnchantSwapped || c.IsAttached() {
		t.Errorf("Necromancy changed although its trigger had no legal target: aura=%v swapped=%v attached=%+v", c.IsAura(), c.EnchantSwapped, c.AttachedTo)
	}
	if adZone(g, zombie) != game.ZoneExile {
		t.Errorf("the exiled card moved to %v", adZone(g, zombie))
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers scheduled by a trigger that did nothing", len(g.DelayedTriggers))
	}
}

// The Aura loses its creature: the creature dies, so "enchant creature
// put onto the battlefield with Necromancy" names nothing, Necromancy
// goes to the graveyard (CR 704.5m), and its delayed trigger finds no
// creature to sacrifice (CR 603.7c).
func TestNecromancyLosingItsCreatureGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	nmAdvanceTo(t, g, game.StepPrecombatMain)
	n := nmCast(t, g)
	nmResolveAndTarget(t, g, n, zombie)
	passPriorityAroundTable(t, g)
	if !adCard(t, g, n).IsAttachedTo(zombie) {
		t.Fatal("Necromancy is not on the returned creature")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(zombie); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(n) {
		t.Errorf("Necromancy should be in its owner's graveyard, it is in %v", adZone(g, n))
	}
	if !opp.Graveyard.Contains(zombie) {
		t.Errorf("the creature is in %v, want its owner's graveyard", adZone(g, zombie))
	}
	if _, ok := adSacrificedBy(g, zombie); ok {
		t.Error("the destroyed creature was also sacrificed")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers still waiting", len(g.DelayedTriggers))
	}
	// A card in a graveyard is a new object: it is not an Aura any more.
	if adCard(t, g, n).IsAura() {
		t.Error("Necromancy in the graveyard kept the Aura subtype its trigger gave it")
	}
}

// Necromancy destroyed: the creature's controller sacrifices it, as for
// Animate Dead.
func TestNecromancyLeavingSacrificesTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	nmAdvanceTo(t, g, game.StepPrecombatMain)
	n := nmCast(t, g)
	nmResolveAndTarget(t, g, n, zombie)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(n); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if who, ok := adSacrificedBy(g, zombie); !ok || who != me.ID {
		t.Errorf("sacrificed by %v (ok=%v), want %v", who, ok, me.ID)
	}
	if !opp.Graveyard.Contains(zombie) {
		t.Errorf("the creature is in %v, want its owner's graveyard", adZone(g, zombie))
	}
}

// The cast memory and the cleanup trigger survive a restore point.
func TestNecromancysCleanupSacrificeSurvivesASnapshot(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	nmAdvanceTo(t, g, game.StepEnd)
	n := nmCast(t, g)
	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if item := restored.StackMeta[n]; item == nil || !item.OwnFlashCast {
		t.Fatal("the restore point lost the cast memory")
	}
	nmResolveAndTarget(t, restored, n, zombie)
	passPriorityAroundTable(t, restored)
	again, err := restored.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if nmCleanupTriggers(again) != 1 {
		t.Errorf("the restore point lost the cleanup sacrifice (%d)", nmCleanupTriggers(again))
	}
}
