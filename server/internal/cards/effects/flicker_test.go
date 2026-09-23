package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flicker_test.go — S22 exile-and-return. Covers the return-from-
// exile primitive (new object, re-triggered ETB, replacement
// pipeline still consulted), the immediate blink (Y'shtola Rhul),
// and the delayed blink (Waterbender's Restoration, Cosmic
// Intervention). The delayed-trigger queue's own timing and undo
// contract live in server/internal/game/delayed_test.go.

const (
	yshtolaRhulOracle             = "a6a7bf77-0560-4572-a826-3bc9df1f78d1"
	waterbendersRestorationOracle = "285046f6-b3c4-4eb7-8712-9dffebabc762"
	cosmicInterventionOracle      = "cddccc2a-a76e-48b3-b4dd-dfeab89e1619"
	mulldrifterOracle             = "24d0f5e7-0d9e-4b76-900e-a7274e80312d"
	flickerProbeOracle            = "test-flicker-etb-probe"
)

// flickerProbeETBs counts direct AsEnters-hook fires on the probe
// permanent. Mulldrifter covers the other ETB path (a declared
// `Triggered` watching EventETB, harvested onto the stack); the two
// are separate mechanisms and a return has to re-fire both.
var flickerProbeETBs int

func init() {
	Register(Spec{
		OracleID: flickerProbeOracle,
		Name:     "Flicker ETB Probe",
		AsEnters: func(_ *game.Card, _ *Context) error {
			flickerProbeETBs++
			return nil
		},
	})
}

// advanceToEndStepOf walks the turn engine forward to the named
// seat's end step — where EventBeginEndStep fires and the delayed
// return triggers drain.
func advanceToEndStepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepEnd && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward end step of seat %d: %v", seat, err)
		}
	}
	t.Fatalf("never reached the end step of seat %d", seat)
}

// battlefieldIDsNamed returns the instance IDs of every battlefield
// card with the given name.
func battlefieldIDsNamed(g *game.Game, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func exileHas(g *game.Game, id uuid.UUID) bool {
	return g.Exile.Contains(id)
}

// pushFlickerCreature puts a catalog creature straight onto the
// battlefield. Direct pushes emit no events, so no ETB fires from
// the setup itself — every count in these tests comes from a return.
func pushFlickerCreature(g *game.Game, owner uuid.UUID, name, oracleID string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Test",
		OracleID:   oracleID,
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// --- the primitive -----------------------------------------------

// TestReturnFromExileMintsANewObject: the returned permanent is a
// new object (CR 400.7) — fresh InstanceID, counters and damage
// gone, untapped — and its ETB fires again.
func TestReturnFromExileMintsANewObject(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0]
	oldID := pushFlickerCreature(g, owner.ID, "Probe", flickerProbeOracle)
	flickerProbeETBs = 0

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
	if exileHas(g, oldID) {
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
	if flickerProbeETBs != 1 {
		t.Errorf("AsEnters fired %d times on the return, want 1", flickerProbeETBs)
	}
}

// TestReturnFromExileUnderYourControl: passing a controller returns
// the permanent under that player rather than its owner ("under your
// control").
func TestReturnFromExileUnderYourControl(t *testing.T) {
	g := newCatalogGame(t)
	owner, thief := g.Seats[1], g.Seats[0]
	id := pushFlickerCreature(g, owner.ID, "Probe", flickerProbeOracle)

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = seedReplacementPermanent(g, authorityOfTheConsulsOracle, "Authority of the Consuls", me.ID)
	id := pushFlickerCreature(g, opp.ID, "Their Creature", "")

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushFlickerCreature(g, me.ID, "Y'shtola Rhul", yshtolaRhulOracle)
	drifter := pushFlickerCreature(g, me.ID, "Mulldrifter", mulldrifterOracle)

	advanceToEndStepOf(t, g, 0)
	handBefore := me.Hand.Size()
	pickCard(t, g, me.ID, drifter)
	passPriorityAroundTable(t, g)

	if exileHas(g, drifter) {
		t.Error("Mulldrifter stayed in exile — the return half never ran")
	}
	ids := battlefieldIDsNamed(g, "Mulldrifter")
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	// Walk into the opponent's turn BEFORE Y'shtola exists, so the
	// only end step she could have triggered on is theirs.
	aangAdvanceToMain(t, g, 1)
	pushFlickerCreature(g, me.ID, "Y'shtola Rhul", yshtolaRhulOracle)
	drifter := pushFlickerCreature(g, me.ID, "Mulldrifter", mulldrifterOracle)

	advanceToEndStepOf(t, g, 1)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Error("Y'shtola triggered on an opponent's end step")
	}
	ids := battlefieldIDsNamed(g, "Mulldrifter")
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	drifter := pushFlickerCreature(g, me.ID, "Mulldrifter", mulldrifterOracle)
	probe := pushFlickerCreature(g, me.ID, "Probe", flickerProbeOracle)
	flickerProbeETBs = 0

	// S22: waterbend {X} is charged now (issue #259), so the blink is
	// "exile X target creatures" with X announced at cast time — two
	// targets means X=2, and the {2} is paid by tapping two of the
	// caster's own permanents rather than arriving free.
	helpers := pushTapCostSoldiers(g, me.ID, 2)
	if err := castRestoration(t, g, game.CastSpellParams{
		XValue: 2,
		TapIDs: helpers,
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: drifter},
			{Kind: game.TargetCard, ID: probe},
		},
	}); err != nil {
		t.Fatalf("CastSpell Waterbender's Restoration: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !exileHas(g, drifter) || !exileHas(g, probe) {
		t.Fatal("targets were not exiled")
	}
	if len(battlefieldIDsNamed(g, "Mulldrifter")) != 0 {
		t.Error("a target is still on the battlefield after the exile half")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed-return queue holds %d triggers, want 1", len(g.DelayedTriggers))
	}
	if flickerProbeETBs != 0 {
		t.Fatalf("an ETB fired on the exile half: %d", flickerProbeETBs)
	}

	handBefore := me.Hand.Size()
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	drifters := battlefieldIDsNamed(g, "Mulldrifter")
	probes := battlefieldIDsNamed(g, "Probe")
	if len(drifters) != 1 || len(probes) != 1 {
		t.Fatalf("returned %d Mulldrifters and %d Probes, want 1 each", len(drifters), len(probes))
	}
	if drifters[0] == drifter || probes[0] == probe {
		t.Error("a returned creature kept its pre-exile InstanceID")
	}
	if exileHas(g, drifter) || exileHas(g, probe) {
		t.Error("a target is still in exile after the delayed return")
	}
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d cards off the re-triggered ETB, want 2", got)
	}
	if flickerProbeETBs != 1 {
		t.Errorf("AsEnters fired %d times on the delayed return, want 1", flickerProbeETBs)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("delayed-return queue not drained: %d", len(g.DelayedTriggers))
	}
}

// --- Cosmic Intervention (replacement + delayed return) ----------

func TestCosmicInterventionExilesInsteadThenReturns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drifter := pushFlickerCreature(g, me.ID, "Mulldrifter", mulldrifterOracle)

	castCatalogSpell(t, g, "Cosmic Intervention", "Instant", cosmicInterventionOracle, nil)
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(drifter); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})

	if me.Graveyard.Contains(drifter) {
		t.Error("permanent hit the graveyard; Cosmic Intervention should have exiled it instead")
	}
	if !exileHas(g, drifter) {
		t.Fatal("permanent was not exiled")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed-return queue holds %d triggers, want 1", len(g.DelayedTriggers))
	}

	handBefore := me.Hand.Size()
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	ids := battlefieldIDsNamed(g, "Mulldrifter")
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
// So both outcomes must be reachable, whichever order the controller
// puts the two replacements in: "no" to the command zone leaves the
// commander in exile and it walks back at the end step; "yes" sends it
// home and nothing comes back.
func TestCosmicInterventionSavesACommanderWhoseOwnerDeclines(t *testing.T) {
	for _, tc := range []struct {
		name            string
		reverseOrder    bool
		takeCommandZone bool
	}{
		{"declines, commander rule first", false, false},
		{"declines, Cosmic Intervention first", true, false},
		{"takes the command zone", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			commander := uuid.New()
			g.Battlefield.PushTop(game.Card{
				InstanceID: commander, Name: "My Commander", TypeLine: "Legendary Creature — Human Avatar",
				Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true,
			})

			castCatalogSpell(t, g, "Cosmic Intervention", "Instant", cosmicInterventionOracle, nil)
			passPriorityAroundTable(t, g)
			g.WithWriteLock(func() {
				if err := g.DestroyPermanentForEffect(commander); err != nil {
					t.Fatalf("DestroyPermanentForEffect: %v", err)
				}
			})

			asked := false
			for i := 0; i < 4 && len(g.PendingChoices) > 0; i++ {
				c := g.PendingChoices[len(g.PendingChoices)-1]
				if c.Chooser != me.ID {
					t.Fatalf("prompt %s addressed to someone other than the commander's owner", c.Kind)
				}
				switch c.Kind {
				case game.PendingChoiceReplacementOrder:
					ids := append([]game.ReplacementEffectID(nil), c.ReplacementEffectIDs...)
					if tc.reverseOrder {
						for l, r := 0, len(ids)-1; l < r; l, r = l+1, r-1 {
							ids[l], ids[r] = ids[r], ids[l]
						}
					}
					if err := g.ResolveReplacementOrder(c.ID, me.ID, ids); err != nil {
						t.Fatalf("ResolveReplacementOrder: %v", err)
					}
				case game.PendingChoiceOptionalReplacement:
					asked = true
					if err := g.ResolveOptionalReplacement(c.ID, me.ID, tc.takeCommandZone); err != nil {
						t.Fatalf("ResolveOptionalReplacement: %v", err)
					}
				default:
					t.Fatalf("unexpected prompt %s", c.Kind)
				}
			}
			if !asked {
				t.Fatal("the commander's owner was never offered the command zone")
			}
			if me.Graveyard.Contains(commander) {
				t.Fatal("the commander hit the graveyard under Cosmic Intervention")
			}

			if tc.takeCommandZone {
				if !me.Command.Contains(commander) {
					t.Fatal("the owner took the command zone but the commander is not there")
				}
				advanceToEndStepOf(t, g, 0)
				passPriorityAroundTable(t, g)
				if got := battlefieldIDsNamed(g, "My Commander"); len(got) != 0 {
					t.Errorf("a commander in the command zone came back at the end step: %v", got)
				}
				return
			}
			if !exileHas(g, commander) {
				t.Fatal("declined commander is not in exile")
			}
			advanceToEndStepOf(t, g, 0)
			passPriorityAroundTable(t, g)
			if got := battlefieldIDsNamed(g, "My Commander"); len(got) != 1 {
				t.Fatalf("%d copies of the commander returned at the end step, want 1", len(got))
			}
		})
	}
}

// An opponent's permanent dying is untouched — "a permanent YOU
// control".
func TestCosmicInterventionIgnoresOpponentsPermanents(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushFlickerCreature(g, opp.ID, "Their Creature", "")

	castCatalogSpell(t, g, "Cosmic Intervention", "Instant", cosmicInterventionOracle, nil)
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(theirs); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})

	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's permanent was diverted from their graveyard")
	}
	if left := battlefieldIDsNamed(g, "Their Creature"); len(left) != 0 {
		t.Errorf("opponent's creature is still on the battlefield: %v", left)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("scheduled %d returns for a permanent the caster does not control", len(g.DelayedTriggers))
	}
}
