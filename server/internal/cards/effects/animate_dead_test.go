package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// animate_dead_test.go — ADR 0154 (#1780): Animate Dead and Dance of the
// Dead end to end, through the cast, the Aura resolving onto a card in a
// graveyard, the enters trigger, the delayed sacrifice, and the rulings
// the ADR lists.

const (
	animateDeadOracle    = "c0d8fef4-65f4-4769-982d-b397d2b7e977"
	danceOfTheDeadOracle = "15d5b198-edbe-40f3-87e2-d72f7f0c58ab"
)

// adGraveyardCreature puts a 3/3 creature card into p's graveyard.
// `keywords` lets a test give it protection.
func adGraveyardCreature(p *game.Player, name string, keywords ...string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Zombie",
		ManaCost: "{2}{B}", Colors: []string{"B"}, Power: 3, Toughness: 3,
		Owner: p.ID, Controller: p.ID, Keywords: keywords,
	})
	return id
}

// adCast puts a black {1}{B} Aura into the active seat's hand, pays for
// it from the pool and casts it at `target`. The colour matters: a
// creature with protection from black must see a black Aura.
func adCast(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Enchantment — Aura", OracleID: oracle,
		ManaCost: "{1}{B}", Colors: []string{"B"}, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	fillPoolColored(me, "B", 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// adResolveSpell passes priority until the Aura spell has left the stack:
// it is on the battlefield with its enters trigger waiting, or it was
// countered.
func adResolveSpell(t *testing.T, g *game.Game, spell uuid.UUID) {
	t.Helper()
	for i := 0; i < 16; i++ {
		if !g.Stack.Contains(spell) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the Aura spell never left the stack")
}

// adAnimate casts Animate Dead (or Dance of the Dead) at `target` and
// settles everything.
func adAnimate(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID) uuid.UUID {
	t.Helper()
	aura := adCast(t, g, name, oracle, target)
	passPriorityAroundTable(t, g)
	return aura
}

func adCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.WithWriteLock(func() { c, ok = g.LookupCardForEffect(id) })
	if !ok {
		t.Fatalf("card %s is in no zone", id)
	}
	return c
}

func adZone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var k game.ZoneKind
	g.WithWriteLock(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			k = z.Kind
		}
	})
	return k
}

func adSacrificedBy(g *game.Game, id uuid.UUID) (uuid.UUID, bool) {
	for _, ev := range g.Events {
		if ev.Kind == game.EventSacrifice && ev.CardID == id {
			return ev.Actor, true
		}
	}
	return uuid.Nil, false
}

// CR 303.4, CR 608.3c and the first ruling: Animate Dead resolves
// attached to the creature card in the opponent's graveyard, its trigger
// waits on the stack with the card still there, and when it resolves the
// card comes back under the caster's control with the Aura attached, at
// -1/-0.
func TestAnimateDeadReturnsAnOpponentsCreatureUnderYourControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	aura := adCast(t, g, "Animate Dead", animateDeadOracle, zombie)
	adResolveSpell(t, g, aura)

	a := adCard(t, g, aura)
	if adZone(g, aura) != game.ZoneBattlefield || !a.IsAttachedTo(zombie) {
		t.Fatalf("the Aura should be on the battlefield attached to the graveyard card: zone=%v attached=%+v", adZone(g, aura), a.AttachedTo)
	}
	if triggerOnStack(g, aura) == nil {
		t.Fatal("Animate Dead's enters trigger is not waiting on the stack")
	}
	if adZone(g, zombie) != game.ZoneGraveyard {
		t.Fatal("the creature card left the graveyard before the trigger resolved")
	}

	passPriorityAroundTable(t, g)

	z := adCard(t, g, zombie)
	if adZone(g, zombie) != game.ZoneBattlefield || z.Controller != me.ID {
		t.Fatalf("the creature should be on the battlefield under the caster's control: zone=%v controller=%v", adZone(g, zombie), z.Controller)
	}
	a = adCard(t, g, aura)
	if !a.IsAttachedTo(zombie) || !a.EnchantSwapped || a.EnchantReturned.ID != zombie || a.EnchantReturned.Epoch != z.ObjectEpoch {
		t.Fatalf("Animate Dead should enchant the returned creature with its swapped clause: %+v swapped=%v returned=%+v", a.AttachedTo, a.EnchantSwapped, a.EnchantReturned)
	}
	if p, tough := z.CurrentPower(), z.CurrentToughness(); p != 2 || tough != 3 {
		t.Errorf("enchanted creature is %d/%d, want 2/3 (-1/-0)", p, tough)
	}
	if z.Tapped {
		t.Error("Animate Dead returns the creature untapped")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Errorf("%d delayed triggers waiting, want the one sacrifice", len(g.DelayedTriggers))
	}
}

// CR 608.2b: the target exiled in response is illegal, so the Aura spell
// is countered and goes to its owner's graveyard.
func TestAnimateDeadWithItsTargetGoneIsCountered(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	aura := adCast(t, g, "Animate Dead", animateDeadOracle, zombie)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(zombie); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(aura) {
		t.Errorf("Animate Dead should be countered into its owner's graveyard, it is in %v", adZone(g, aura))
	}
	if adZone(g, zombie) != game.ZoneExile {
		t.Error("the exiled card moved")
	}
}

// The second ruling: Animate Dead removed in response to its trigger
// returns nothing and schedules nothing.
func TestAnimateDeadRemovedInResponseReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	aura := adCast(t, g, "Animate Dead", animateDeadOracle, zombie)
	adResolveSpell(t, g, aura)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(aura); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if adZone(g, zombie) != game.ZoneGraveyard {
		t.Fatalf("the creature came back although Animate Dead was gone (zone %v)", adZone(g, zombie))
	}
	if !me.Graveyard.Contains(aura) {
		t.Error("Animate Dead should be in its owner's graveyard")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers scheduled by a trigger whose source was gone", len(g.DelayedTriggers))
	}
}

// The creature card exiled while the trigger waits: the Aura enchants an
// object that left the graveyard, the state-based action buries it, and
// the trigger resolves with its source gone (CR 603.4, Spellweaver
// Volute's ruling).
func TestAnimateDeadsCardExiledWhileTheTriggerWaits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	aura := adCast(t, g, "Animate Dead", animateDeadOracle, zombie)
	adResolveSpell(t, g, aura)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(zombie); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(aura) {
		t.Errorf("Animate Dead should have fallen off into the graveyard, it is in %v", adZone(g, aura))
	}
	if adZone(g, zombie) != game.ZoneExile {
		t.Errorf("the exiled card is in %v", adZone(g, zombie))
	}
}

// Decision 6: the Aura leaving the battlefield makes the creature's
// CURRENT controller sacrifice it — after a control change too.
func TestAnimateDeadLeavingMakesTheCurrentControllerSacrifice(t *testing.T) {
	for _, stolen := range []bool{false, true} {
		name := "same controller"
		if stolen {
			name = "after a control change"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
			zombie := adGraveyardCreature(opp, "Gravecrawler")
			aura := adAnimate(t, g, "Animate Dead", animateDeadOracle, zombie)
			want := me.ID
			if stolen {
				g.WithWriteLock(func() {
					g.GainControlForEffect(uuid.Nil, zombie, third.ID, game.IndefiniteDuration(), "test: steal")
				})
				want = third.ID
			}
			g.WithWriteLock(func() {
				if err := g.DestroyPermanentForEffect(aura); err != nil {
					t.Fatalf("destroy: %v", err)
				}
			})
			passPriorityAroundTable(t, g)
			if !opp.Graveyard.Contains(zombie) {
				t.Fatalf("the creature should be sacrificed into its owner's graveyard, it is in %v", adZone(g, zombie))
			}
			if who, ok := adSacrificedBy(g, zombie); !ok || who != want {
				t.Errorf("sacrificed by %v (ok=%v), want %v", who, ok, want)
			}
			if len(g.DelayedTriggers) != 0 {
				t.Errorf("the delayed trigger is still waiting after it fired (%d)", len(g.DelayedTriggers))
			}
		})
	}
}

// The creature leaving first: the swapped clause names nothing on the
// battlefield, the Aura is buried (CR 704.5m), its delayed trigger fires
// and finds no creature to sacrifice (CR 603.7c).
func TestAnimateDeadsCreatureLeavingFirstLeavesTheTriggerNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	aura := adAnimate(t, g, "Animate Dead", animateDeadOracle, zombie)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(zombie); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(aura) {
		t.Errorf("Animate Dead should be in the graveyard, it is in %v", adZone(g, aura))
	}
	if adZone(g, zombie) != game.ZoneExile {
		t.Errorf("the creature is in %v, want exile", adZone(g, zombie))
	}
	if _, ok := adSacrificedBy(g, zombie); ok {
		t.Error("something was sacrificed")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers still waiting", len(g.DelayedTriggers))
	}
}

// The fifth ruling: a creature with protection from black can be targeted
// in the graveyard and comes back, but Animate Dead can't stay attached
// to it. The Aura goes to the graveyard, its delayed trigger triggers,
// and the creature is sacrificed.
func TestAnimateDeadOnAProtectionFromBlackCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := adGraveyardCreature(opp, "White Knight", "protection from black")
	aura := adAnimate(t, g, "Animate Dead", animateDeadOracle, knight)
	if !me.Graveyard.Contains(aura) {
		t.Errorf("Animate Dead should be in the graveyard, it is in %v", adZone(g, aura))
	}
	if !opp.Graveyard.Contains(knight) {
		t.Errorf("the creature should have been sacrificed, it is in %v", adZone(g, knight))
	}
	if who, ok := adSacrificedBy(g, knight); !ok || who != me.ID {
		t.Errorf("sacrificed by %v (ok=%v), want the caster who controlled it", who, ok)
	}
}

// The third ruling, CR 701.3b: "Attempting to move Animate Dead to
// another creature won't work." It stays on the creature it returned.
func TestAnimateDeadCantBeMovedToAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	bear := seedBear(g, me.ID)
	aura := adAnimate(t, g, "Animate Dead", animateDeadOracle, zombie)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	a := adCard(t, g, aura)
	if adZone(g, aura) != game.ZoneBattlefield || !a.IsAttachedTo(zombie) {
		t.Fatalf("Animate Dead moved: zone=%v attached=%+v", adZone(g, aura), a.AttachedTo)
	}
	if adZone(g, zombie) != game.ZoneBattlefield {
		t.Error("the returned creature was sacrificed")
	}
}

// Decision 6: the delayed trigger has no expiry. It survives the
// cleanup step and an undo snapshot, and still fires in the restored
// game.
func TestAnimateDeadsSacrificeSurvivesCleanupAndASnapshot(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	aura := adAnimate(t, g, "Animate Dead", animateDeadOracle, zombie)
	advanceToUpkeepOf(t, g, 1)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("the sacrifice trigger did not survive the cleanup step (%d waiting)", len(g.DelayedTriggers))
	}
	if d := g.DelayedTriggers[0].Duration; d == nil || d.Kind != game.Indefinite {
		t.Errorf("the sacrifice trigger's duration is %+v, want Indefinite", d)
	}
	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if len(restored.DelayedTriggers) != 1 {
		t.Fatalf("the restore point lost the sacrifice trigger (%d)", len(restored.DelayedTriggers))
	}
	restored.WithWriteLock(func() {
		if err := restored.DestroyPermanentForEffect(aura); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, restored)
	var inGY bool
	restored.ReadSnapshot(func() { inGY = restored.Seats[1].Graveyard.Contains(zombie) })
	if !inGY {
		t.Error("the restored game's trigger did not sacrifice the creature")
	}
}

// A restore point taken while the enters trigger waits (the Aura on the
// battlefield attached to the graveyard card) restores with the Aura
// legal and the trigger intact, and the trigger still returns the card.
func TestAnimateDeadRestoredMidTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	aura := adCast(t, g, "Animate Dead", animateDeadOracle, zombie)
	adResolveSpell(t, g, aura)
	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	passPriorityAroundTable(t, restored)
	var zone game.ZoneKind
	var controller uuid.UUID
	var attached bool
	restored.WithWriteLock(func() {
		if z := restored.FindCardZoneForEffect(zombie); z != nil {
			zone = z.Kind
		}
		if c, ok := restored.LookupCardForEffect(zombie); ok {
			controller = c.Controller
		}
		if a, ok := restored.LookupCardForEffect(aura); ok {
			attached = a.IsAttachedTo(zombie)
		}
	})
	if zone != game.ZoneBattlefield || controller != me.ID || !attached {
		t.Errorf("after a restore mid-trigger: zone=%v controller=%v attached=%v", zone, controller, attached)
	}
}

// The return can pause on a question (a pay-life entry here, a Clone's
// copy choice in play). The Aura must wait with it, still on the card in
// the graveyard, and attach once the creature has entered.
func TestAnimateDeadWaitsForAPausedReturn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(game.ReplacementEffect{
			Watches:       []game.EventKind{game.EventZoneMove},
			EntryLifeCost: 2,
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.CardID == zombie &&
					ev.OldZone == game.ZoneGraveyard && ev.NewZone == game.ZoneBattlefield
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(*game.ReplacementEvent, *game.Game, *game.Card) uuid.UUID { return me.ID },
			Label:      "test: pay 2 life or it enters tapped",
		})
	})
	aura := adAnimate(t, g, "Animate Dead", animateDeadOracle, zombie)
	if entryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the returning creature's entry question was not asked")
	}
	if adZone(g, aura) != game.ZoneBattlefield {
		t.Fatalf("Animate Dead left the battlefield while the return was paused (zone %v)", adZone(g, aura))
	}
	answerEntryPayLife(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	a := adCard(t, g, aura)
	if adZone(g, zombie) != game.ZoneBattlefield || adZone(g, aura) != game.ZoneBattlefield || !a.IsAttachedTo(zombie) {
		t.Fatalf("after the answer: creature zone=%v, aura zone=%v attached=%+v", adZone(g, zombie), adZone(g, aura), a.AttachedTo)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Errorf("%d delayed triggers, want the sacrifice", len(g.DelayedTriggers))
	}
}

// Dance of the Dead: the creature comes back tapped at +1/+1, does not
// untap in its controller's untap step, and that player may pay {1}{B}
// in their upkeep to untap it. The trigger watches the enchanted
// creature's controller's upkeep, not the Aura controller's alone.
func TestDanceOfTheDeadReturnsTappedAndUntapsForAPayment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	aura := adAnimate(t, g, "Dance of the Dead", danceOfTheDeadOracle, zombie)

	z := adCard(t, g, zombie)
	if adZone(g, zombie) != game.ZoneBattlefield || z.Controller != me.ID || !z.Tapped {
		t.Fatalf("Dance of the Dead should return the creature tapped under your control: zone=%v controller=%v tapped=%v", adZone(g, zombie), z.Controller, z.Tapped)
	}
	if p, tough := z.CurrentPower(), z.CurrentToughness(); p != 4 || tough != 4 {
		t.Errorf("enchanted creature is %d/%d, want 4/4", p, tough)
	}

	// The opponent's upkeep is not the enchanted creature's controller's.
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, aura) != nil {
		t.Error("Dance of the Dead triggered in another player's upkeep")
	}

	advanceToUpkeepOf(t, g, 0)
	if !adCard(t, g, zombie).Tapped {
		t.Fatal("the enchanted creature untapped in its controller's untap step")
	}
	if triggerOnStack(g, aura) == nil {
		t.Fatal("Dance of the Dead's upkeep trigger is not on the stack")
	}
	passPriorityAroundTable(t, g)
	choice := untapStaticPendingPay(t, g, me.ID)
	fillPoolColored(me, "B", 2)
	if err := g.ResolvePayUnless(choice.ID, me.ID, true); err != nil {
		t.Fatalf("answer the upkeep payment: %v", err)
	}
	if adCard(t, g, zombie).Tapped {
		t.Error("paying {1}{B} did not untap the creature")
	}
}

// Dance of the Dead's upkeep payment is asked of the creature's
// controller: after a control change, the new controller's upkeep, and
// that player pays.
func TestDanceOfTheDeadAsksTheEnchantedCreaturesController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	zombie := adGraveyardCreature(opp, "Gravecrawler")
	aura := adAnimate(t, g, "Dance of the Dead", danceOfTheDeadOracle, zombie)
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.Nil, zombie, third.ID, game.IndefiniteDuration(), "test: steal")
	})
	advanceToUpkeepOf(t, g, 2)
	if triggerOnStack(g, aura) == nil {
		t.Fatal("Dance of the Dead did not trigger in the new controller's upkeep")
	}
	passPriorityAroundTable(t, g)
	choice := untapStaticPendingPay(t, g, third.ID)
	if choice.Chooser == me.ID {
		t.Fatal("the Aura's controller was asked")
	}
	if err := g.ResolvePayUnless(choice.ID, third.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if !adCard(t, g, zombie).Tapped {
		t.Error("declining untapped the creature")
	}
}
