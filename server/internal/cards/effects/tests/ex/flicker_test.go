package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the primitive -----------------------------------------------

// TestReturnFromExileMintsANewObject: the returned permanent is a
// new object (CR 400.7) — fresh InstanceID, counters and damage
// gone, untapped — and its ETB fires again.
func TestReturnFromExileMintsANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0]
	oldID := effectstest.PushFlickerCreature(g, owner.ID, "Probe", effectstest.FlickerProbeOracle)
	effectstest.FlickerProbeETBs = 0

	var newID uuid.UUID
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == oldID {
				g.Battlefield.Cards[i].Tapped = true
				g.Battlefield.Cards[i].DamageMarked = 1
				g.Battlefield.Cards[i].Counters = map[string]int{"+1/+1": 2}
			}
		}
		if err := g.ExileCardForEffect(oldID); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
		var err error
		newID, err = g.ReturnFromExileToBattlefieldForEffect(oldID, uuid.Nil, false)
		if err != nil {
			t.Fatalf("ReturnFromExileToBattlefieldForEffect: %v", err)
		}
	})

	if newID == uuid.Nil {
		t.Fatal("return produced no new instance ID")
	}
	if newID == oldID {
		t.Error("returned permanent kept its old InstanceID — it must be a new object")
	}
	if effectstest.ExileHas(g, oldID) {
		t.Error("old instance still sitting in exile")
	}
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == newID {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatal("returned permanent is not on the battlefield")
	}
	if back.Tapped {
		t.Error("returned permanent came back tapped")
	}
	if back.DamageMarked != 0 {
		t.Errorf("damage carried over: %d", back.DamageMarked)
	}
	if back.Counters["+1/+1"] != 0 {
		t.Errorf("+1/+1 counters carried over: %d", back.Counters["+1/+1"])
	}
	if back.Controller != owner.ID {
		t.Errorf("returned under %v, want the owner %v", back.Controller, owner.ID)
	}
	if effectstest.FlickerProbeETBs != 1 {
		t.Errorf("AsEnters fired %d times on the return, want 1", effectstest.FlickerProbeETBs)
	}
}

// TestReturnFromExileUnderYourControl: passing a controller returns
// the permanent under that player rather than its owner ("under your
// control").
func TestReturnFromExileUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, thief := g.Seats[1], g.Seats[0]
	id := effectstest.PushFlickerCreature(g, owner.ID, "Probe", effectstest.FlickerProbeOracle)

	var newID uuid.UUID
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
		var err error
		newID, err = g.ReturnFromExileToBattlefieldForEffect(id, thief.ID, true)
		if err != nil {
			t.Fatalf("return: %v", err)
		}
	})

	for _, c := range g.Battlefield.Cards {
		if c.InstanceID != newID {
			continue
		}
		if c.Controller != thief.ID {
			t.Errorf("controller %v, want %v", c.Controller, thief.ID)
		}
		if c.Owner != owner.ID {
			t.Errorf("owner changed to %v; ownership never changes", c.Owner)
		}
		if !c.Tapped {
			t.Error("asked for a tapped return, came back untapped")
		}
		return
	}
	t.Fatal("returned permanent not on the battlefield")
}

// TestReturnFromExileRunsTheReplacementPipeline: a blinked creature
// entering under Authority of the Consuls enters tapped. This is the
// reason the return consults CR 614 rather than pushing straight
// onto the battlefield.
func TestReturnFromExileRunsTheReplacementPipeline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = effectstest.SeedReplacementPermanent(g, effectstest.AuthorityOfTheConsulsOracle, "Authority of the Consuls", me.ID)
	id := effectstest.PushFlickerCreature(g, opp.ID, "Their Creature", "")

	var newID uuid.UUID
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
		var err error
		newID, err = g.ReturnFromExileToBattlefieldForEffect(id, uuid.Nil, false)
		if err != nil {
			t.Fatalf("return: %v", err)
		}
	})

	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == newID {
			if !c.Tapped {
				t.Error("Authority of the Consuls did not tap the returning creature")
			}
			return
		}
	}
	t.Fatal("returned permanent not on the battlefield")
}

// --- Y'shtola Rhul (immediate blink on an end-step trigger) -------

func TestYshtolaRhulBlinksAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushFlickerCreature(g, me.ID, "Y'shtola Rhul", effectstest.YshtolaRhulOracle)
	drifter := effectstest.PushFlickerCreature(g, me.ID, "Mulldrifter", effectstest.MulldrifterOracle)

	effectstest.AdvanceToEndStepOf(t, g, 0)
	handBefore := me.Hand.Size()
	effectstest.PickCard(t, g, me.ID, drifter)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.ExileHas(g, drifter) {
		t.Error("Mulldrifter stayed in exile — the return half never ran")
	}
	ids := effectstest.BattlefieldIDsNamed(g, "Mulldrifter")
	if len(ids) != 1 {
		t.Fatalf("found %d Mulldrifters on the battlefield, want 1", len(ids))
	}
	if ids[0] == drifter {
		t.Error("blinked creature kept its InstanceID — it must return as a new object")
	}
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d cards off the re-triggered ETB, want 2", got)
	}
}

// The trigger is "your end step", not every end step.
func TestYshtolaRhulDoesNotTriggerOnAnOpponentsEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// Walk into the opponent's turn BEFORE Y'shtola exists, so the
	// only end step she could have triggered on is theirs.
	effectstest.AangAdvanceToMain(t, g, 1)
	effectstest.PushFlickerCreature(g, me.ID, "Y'shtola Rhul", effectstest.YshtolaRhulOracle)
	drifter := effectstest.PushFlickerCreature(g, me.ID, "Mulldrifter", effectstest.MulldrifterOracle)

	effectstest.AdvanceToEndStepOf(t, g, 1)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Error("Y'shtola triggered on an opponent's end step")
	}
	ids := effectstest.BattlefieldIDsNamed(g, "Mulldrifter")
	if len(ids) != 1 || ids[0] != drifter {
		t.Errorf("Mulldrifter was blinked on the wrong end step: %v", ids)
	}
}

// --- Waterbender's Restoration (delayed blink) -------------------

// Two creatures leave together and come back together a step
// boundary later. One is a Mulldrifter (its ETB is a declared
// `Triggered` that goes on the stack), the other is the probe (a
// direct AsEnters hook) — the delayed return has to re-fire both
// mechanisms, and they are wired independently.
func TestWaterbendersRestorationReturnsAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drifter := effectstest.PushFlickerCreature(g, me.ID, "Mulldrifter", effectstest.MulldrifterOracle)
	probe := effectstest.PushFlickerCreature(g, me.ID, "Probe", effectstest.FlickerProbeOracle)
	effectstest.FlickerProbeETBs = 0

	// S22: waterbend {X} is charged now (issue #259), so the blink is
	// "exile X target creatures" with X announced at cast time — two
	// targets means X=2, and the {2} is paid by tapping two of the
	// caster's own permanents rather than arriving free.
	helpers := effectstest.PushTapCostSoldiers(g, me.ID, 2)
	if err := effectstest.CastRestoration(t, g, game.CastSpellParams{
		XValue: 2,
		TapIDs: helpers,
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: drifter},
			{Kind: game.TargetCard, ID: probe},
		},
	}); err != nil {
		t.Fatalf("CastSpell Waterbender's Restoration: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.ExileHas(g, drifter) || !effectstest.ExileHas(g, probe) {
		t.Fatal("targets were not exiled")
	}
	if len(effectstest.BattlefieldIDsNamed(g, "Mulldrifter")) != 0 {
		t.Error("a target is still on the battlefield after the exile half")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed-return queue holds %d triggers, want 1", len(g.DelayedTriggers))
	}
	if effectstest.FlickerProbeETBs != 0 {
		t.Fatalf("an ETB fired on the exile half: %d", effectstest.FlickerProbeETBs)
	}

	handBefore := me.Hand.Size()
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)

	drifters := effectstest.BattlefieldIDsNamed(g, "Mulldrifter")
	probes := effectstest.BattlefieldIDsNamed(g, "Probe")
	if len(drifters) != 1 || len(probes) != 1 {
		t.Fatalf("returned %d Mulldrifters and %d Probes, want 1 each", len(drifters), len(probes))
	}
	if drifters[0] == drifter || probes[0] == probe {
		t.Error("a returned creature kept its pre-exile InstanceID")
	}
	if effectstest.ExileHas(g, drifter) || effectstest.ExileHas(g, probe) {
		t.Error("a target is still in exile after the delayed return")
	}
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d cards off the re-triggered ETB, want 2", got)
	}
	if effectstest.FlickerProbeETBs != 1 {
		t.Errorf("AsEnters fired %d times on the delayed return, want 1", effectstest.FlickerProbeETBs)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("delayed-return queue not drained: %d", len(g.DelayedTriggers))
	}
}

// --- Cosmic Intervention (replacement + delayed return) ----------

func TestCosmicInterventionExilesInsteadThenReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drifter := effectstest.PushFlickerCreature(g, me.ID, "Mulldrifter", effectstest.MulldrifterOracle)

	effectstest.CastCatalogSpell(t, g, "Cosmic Intervention", "Instant", effectstest.CosmicInterventionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(drifter); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})

	if me.Graveyard.Contains(drifter) {
		t.Error("permanent hit the graveyard; Cosmic Intervention should have exiled it instead")
	}
	if !effectstest.ExileHas(g, drifter) {
		t.Fatal("permanent was not exiled")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed-return queue holds %d triggers, want 1", len(g.DelayedTriggers))
	}

	handBefore := me.Hand.Size()
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)

	ids := effectstest.BattlefieldIDsNamed(g, "Mulldrifter")
	if len(ids) != 1 {
		t.Fatalf("%d Mulldrifters came back, want 1", len(ids))
	}
	if ids[0] == drifter {
		t.Error("returned permanent kept its pre-exile InstanceID")
	}
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d cards off the re-triggered ETB, want 2", got)
	}
}

// A dying commander. CR 903.9a lets its owner send it to the command
// zone once it has been exiled; Cosmic Intervention does the exiling.
// Since ADR 0115 that is the engine's order too: Cosmic Intervention's
// replacement is the only one that applies (no CR 616 ordering prompt),
// and the state-based action asks afterwards. "No" leaves the commander
// in exile and it walks back at the end step; "yes" sends it home and
// nothing comes back.
func TestCosmicInterventionExilesACommanderThenAsksItsOwner(t *testing.T) {
	for _, tc := range []struct {
		name            string
		takeCommandZone bool
	}{
		{"declines", false},
		{"takes the command zone", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			commander := uuid.New()
			g.Battlefield.PushTop(game.Card{
				InstanceID: commander, Name: "My Commander", TypeLine: "Legendary Creature — Human Avatar",
				Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true,
			})

			effectstest.CastCatalogSpell(t, g, "Cosmic Intervention", "Instant", effectstest.CosmicInterventionOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			g.WithWriteLock(func() {
				if err := g.DestroyPermanentForEffect(commander); err != nil {
					t.Fatalf("DestroyPermanentForEffect: %v", err)
				}
			})

			for _, c := range g.PendingChoices {
				if c != nil && (c.Kind == game.PendingChoiceReplacementOrder || c.Kind == game.PendingChoiceOptionalReplacement) {
					t.Fatalf("unexpected %s prompt: only Cosmic Intervention's replacement applies", c.Kind)
				}
			}
			if me.Graveyard.Contains(commander) {
				t.Fatal("the commander hit the graveyard under Cosmic Intervention")
			}
			if !effectstest.ExileHas(g, commander) {
				t.Fatal("Cosmic Intervention did not exile the dying commander")
			}
			if effectstest.CommanderReturnPromptFor(g, me.ID) == nil {
				t.Fatal("the commander's owner was never offered the command zone")
			}
			effectstest.AnswerCommanderReturn(t, g, me.ID, tc.takeCommandZone)

			if tc.takeCommandZone {
				if !me.Command.Contains(commander) {
					t.Fatal("the owner took the command zone but the commander is not there")
				}
				effectstest.AdvanceToEndStepOf(t, g, 0)
				effectstest.PassPriorityAroundTable(t, g)
				if got := effectstest.BattlefieldIDsNamed(g, "My Commander"); len(got) != 0 {
					t.Errorf("a commander in the command zone came back at the end step: %v", got)
				}
				return
			}
			if !effectstest.ExileHas(g, commander) {
				t.Fatal("declined commander is not in exile")
			}
			effectstest.AdvanceToEndStepOf(t, g, 0)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.BattlefieldIDsNamed(g, "My Commander"); len(got) != 1 {
				t.Fatalf("%d copies of the commander returned at the end step, want 1", len(got))
			}
		})
	}
}

// An opponent's permanent dying is untouched — "a permanent YOU
// control".
func TestCosmicInterventionIgnoresOpponentsPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.PushFlickerCreature(g, opp.ID, "Their Creature", "")

	effectstest.CastCatalogSpell(t, g, "Cosmic Intervention", "Instant", effectstest.CosmicInterventionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(theirs); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})

	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's permanent was diverted from their graveyard")
	}
	if left := effectstest.BattlefieldIDsNamed(g, "Their Creature"); len(left) != 0 {
		t.Errorf("opponent's creature is still on the battlefield: %v", left)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("scheduled %d returns for a permanent the caster does not control", len(g.DelayedTriggers))
	}
}
