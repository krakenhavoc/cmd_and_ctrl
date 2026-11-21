package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardssy "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/sy"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_z2_restore_test.go — #4253's engine closure doors: the card callers
// that still reached an engine door with a Go closure (ExileCardThenForEffect,
// SacrificeAllThenForEffect, TuckToLibraryThenForEffect,
// ExileTopWithPermissionThenForEffect, AddCounterByThenForEffect,
// ScryThenForEffect) went through the keyed siblings instead, and the doors
// that no card needed any more are gone. Each test pauses the move on a CR
// 616 ordering prompt from two catalog probe replacements over a card named
// Needle, takes a STRICT restore (the boot path), answers on the restored
// table, and checks the rest of the card ran off the disk.

// --- Banishing Stroke: PutIntoLibrary with no "then" -------------------

// Banishing Stroke over an opponent's Needle: the tuck asks for the order,
// and the restored answer puts it on the bottom of its owner's library.
func TestSweepZ2BanishingStrokeTuckIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	needle := effectstest.PushVanillaCreature(g, opp.ID, "Needle", 2, 2)
	effectstest.SeedOrderProbes(g, opp.ID, effectstest.LibraryHandOrderProbeOracles, "Library Hand Order Probe")
	effectstest.CastCatalogSpell(t, g, "Banishing Stroke", "Instant", effectstest.BanishingStrokeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: needle}})
	effectstest.PassPriorityAroundTable(t, g)

	g = strictRestoreAtOrderPrompt(t, g)
	opp = g.Seats[1]
	if !g.Battlefield.Contains(needle) {
		t.Fatal("the Needle left the battlefield before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if g.Battlefield.Contains(needle) {
		t.Fatal("the Needle did not leave the battlefield after the restored answer")
	}
	if len(opp.Library.Cards) == 0 || opp.Library.Cards[0].InstanceID != needle {
		t.Error("the Needle is not on the bottom of its owner's library")
	}
}

// --- Momentum Breaker: the edict's sacrifice and the discard after it ---

// Momentum Breaker over an opponent whose only creature is a Needle: the
// Needle's sacrifice asks for the order (the sacrifice is one simultaneous
// exit, its continuation the edict's registered run), and the restored
// answer finishes the sacrifice and then asks the opponents who could not
// sacrifice to discard (a keyed discard run).
func TestSweepZ2MomentumBreakerSacrificeIsARestorePointAndStillDiscards(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	a := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	needle := effectstest.PushCatalogPermanent(g, a.ID, "Needle", "Creature — Bear", "", false)
	effectstest.SeedOrderProbes(g, a.ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	handA := a.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Momentum Breaker", "Enchantment", effectstest.MomentumBreakerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	g = strictRestoreAtOrderPrompt(t, g)
	a, b = g.PlayerByIDForEffect(a.ID), g.PlayerByIDForEffect(b.ID)
	if !g.Battlefield.Contains(needle) {
		t.Fatal("the Needle left the battlefield before the restored answer")
	}
	handB := b.Hand.Size()
	effectstest.AnswerExileOrder(t, g)
	if g.Battlefield.Contains(needle) {
		t.Fatal("the Needle was not sacrificed after the restored answer")
	}
	// The opponents who had no creature discard; the one who sacrificed does not.
	effectstest.DiscardFromHand(t, g, b.ID)
	if got := b.Hand.Size(); got != handB-1 {
		t.Errorf("an opponent who could not sacrifice: hand %d, want %d", got, handB-1)
	}
	if got := a.Hand.Size(); got != handA {
		t.Errorf("the opponent who sacrificed discarded: hand %d, want %d", got, handA)
	}
	_ = me
}

// --- Tersa Lightshatter: an exile that grants the permission -----------

// Tersa Lightshatter's attack exiles a card from the graveyard at random
// and lets its controller play it this turn. Every card in the graveyard is
// a Needle, so whichever it picks stops on the order prompt; the restored
// answer exiles it and grants the permission (the grant rides the move as a
// registered tail).
func TestSweepZ2ExileWithPermissionIsARestorePointAndStillGrants(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		me.Graveyard.Cards = nil
		for i := 0; i < 7; i++ {
			me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Needle", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		}
	})
	effectstest.SeedExileOrderProbes(g, me.ID)
	source := effectstest.PushVanillaCreature(g, me.ID, "Tersa Stand-in", 3, 3)
	g.WithWriteLock(func() {
		if err := cardssy.TersaExileRandom(g, &game.StackItem{Controller: me.ID, SourceCardID: source}); err != nil {
			t.Fatalf("tersaExileRandom: %v", err)
		}
	})

	g = strictRestoreAtOrderPrompt(t, g)
	me = g.Seats[0]
	if me.Graveyard.Size() != 7 {
		t.Fatalf("the graveyard has %d cards before the restored answer, want 7", me.Graveyard.Size())
	}
	effectstest.AnswerExileOrder(t, g)
	if me.Graveyard.Size() != 6 {
		t.Fatalf("the graveyard has %d cards after the restored answer, want 6", me.Graveyard.Size())
	}
	exiled := g.Exile.Cards
	if len(exiled) != 1 {
		t.Fatalf("%d cards in exile, want the one", len(exiled))
	}
	if effectstest.PermissionFor(g, me.ID, exiled[0].InstanceID) == nil {
		t.Error("the exiled card was not granted its permission: the grant did not ride the move")
	}
}

// --- Annihilator: the defender's sacrifice -----------------------------

// Annihilator 1 over a defender whose only permanent is a Needle: the
// defender's pick is made, the sacrifice asks for the order, and the
// restored answer sacrifices it. The sacrifice is the engine's own, with no
// card tail.
func TestSweepZ2AnnihilatorSacrificeIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eldrazi := effectstest.PushAnnihilatorCreature(g, me.ID, "annihilator 1")
	needle := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Needle", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.SeedOrderProbes(g, opp.ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")

	effectstest.DeclareAttack(t, g, opp.ID, eldrazi)
	for i := 0; i < 64 && !effectstest.OrderPromptOpen(g); i++ {
		answered := false
		for _, c := range g.PendingChoices {
			switch {
			case c != nil && c.Kind == game.PendingChoiceTriggerOrder:
				if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				answered = true
			case c != nil && c.Kind == game.PendingChoiceOwnPermanents:
				if err := g.ResolveOwnPermanents(c.ID, c.Chooser, []uuid.UUID{needle}); err != nil {
					t.Fatalf("ResolveOwnPermanents: %v", err)
				}
				answered = true
			}
			if answered {
				break
			}
		}
		if !answered && !effectstest.OrderPromptOpen(g) {
			effectstest.PassPriorityAroundTable(t, g)
		}
	}
	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = strictRestoreAtOrderPrompt(t, g)
	if !effectstest.AnnOnBattlefield(g, needle) {
		t.Fatal("the Needle left the battlefield before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if effectstest.AnnOnBattlefield(g, needle) {
		t.Fatal("the defender kept the Needle after the restored answer")
	}
}

// --- Sephiroth: exile and return transformed ---------------------------

// A permanent exiled and returned transformed (Fable of the Mirror-Breaker's
// chapter III, here through the primitive) whose exile asks for the order:
// the restored answer exiles it, sets its back face and returns it, the
// return a registered tail carrying the face and who controls it.
func TestSweepZ2ExileAndReturnTransformedIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	seph := effectstest.HooksSephirothOnTheBattlefield(t, g, me)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == seph {
				g.Battlefield.Cards[i].Name = "Needle"
			}
		}
	})
	effectstest.SeedExileOrderProbes(g, me.ID)
	var err error
	g.WithWriteLock(func() { err = g.ExileAndReturnTransformedForEffect(seph, uuid.Nil) })
	if err != nil {
		t.Fatalf("ExileAndReturnTransformedForEffect: %v", err)
	}

	g = strictRestoreAtOrderPrompt(t, g)
	if !g.Battlefield.Contains(seph) {
		t.Fatal("the permanent left the battlefield before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if _, ok := effectstest.E2BattlefieldNamed(g, effectstest.SephirothBackName); !ok {
		t.Fatal("the permanent did not return on its back face after the restored answer")
	}
}

// --- an explicit proliferate behind a counter order prompt -------------

// Proliferate with its permanents named (no prompt) over two creatures with
// Hardened Scales beside Doubling Season: each creature's counter asks for
// the order, each is a restore point, and the walk goes on to the next
// creature and then the rest of the sentence off the disk.
func TestSweepZ2ExplicitProliferateWalksAsDataAcrossRestores(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hs, ds := effectstest.CounterOrderBoard(g, me.ID)
	a := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	// The setup placements ask their own order prompts, one at a time.
	for _, id := range []uuid.UUID{a, b} {
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, "+1/+1", 1) })
		for effectstest.OrderPromptOpen(g) {
			effectstest.AnswerOrderBySource(t, g, ds, hs)
		}
	}
	before := map[uuid.UUID]int{a: effectstest.CountersOn(g, a, "+1/+1"), b: effectstest.CountersOn(g, b, "+1/+1")}
	var err error
	g.WithWriteLock(func() {
		err = effects.Proliferate{Cards: []uuid.UUID{a, b}, ThenTail: effectstest.HookTail(func(*game.Game, game.TailOutcome) error { return nil })}.
			Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	if err != nil {
		t.Fatalf("Proliferate: %v", err)
	}
	for placed := 1; placed <= 2; placed++ {
		if !effectstest.OrderPromptOpen(g) {
			t.Fatalf("placement %d: no order prompt is open: %+v", placed, g.PendingChoices)
		}
		g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
		if placed == 2 && effectstest.CountersOn(g, b, "+1/+1") != before[b] {
			t.Fatalf("the second creature has new counters before its own order was answered")
		}
		effectstest.AnswerOrderBySource(t, g, ds, hs)
	}
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.CountersOn(g, id, "+1/+1"); got <= before[id] {
			t.Errorf("a creature has %d counters, want more than the %d it began with", got, before[id])
		}
	}
}
