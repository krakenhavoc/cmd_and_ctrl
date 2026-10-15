package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func slicerRow() cards.Card {
	return transformerRow(slicerOracleID,
		disturbFace{"Slicer, Hired Muscle", "Legendary Artifact Creature — Robot", "{4}{R}", "3", "4"},
		disturbFace{"Slicer, High-Speed Antagonist", "Legendary Artifact — Vehicle", "", "3", "2"},
		"More Than Meets the Eye {2}{R}\nDouble strike, haste\nAt the beginning of each opponent's upkeep, you may have that player gain control of Slicer until end of turn. If you do, untap Slicer, goad it, and it can't be sacrificed this turn. If you don't, convert it.",
		"Living metal\nFirst strike, haste\nWhenever Slicer deals combat damage to a player, convert it at end of combat.",
		"Living metal", "Convert", "Double strike", "First strike", "Haste")
}

// slicerAtOpponentsUpkeep seats Slicer (tapped) for the active player and
// walks to the next seat's upkeep, where its trigger asks its controller.
func slicerAtOpponentsUpkeep(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, slicerRow(), me)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	oppSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	advanceToUpkeepOfSeat(t, g, oppSeat)
	tfSettle(t, g)
	if latestConfirmFor(g, me.ID) == nil {
		t.Fatalf("Slicer's controller was not asked at %s's upkeep", opp.Name)
	}
	return g, me, opp, id
}

// "If you do": the upkeep's player gains control until end of turn, and
// Slicer is untapped, goaded by its owner and can't be sacrificed.
func TestSlicerHandedOverIsUntappedGoadedAndCantBeSacrificed(t *testing.T) {
	g, me, opp, id := slicerAtOpponentsUpkeep(t)
	answerMayChoice(t, g, me.ID, true)
	tfSettle(t, g)

	c := wantFace(t, g, id, 0, "Slicer, Hired Muscle")
	if c.Controller != opp.ID {
		t.Fatalf("controller = %v, want the upkeep's player %v", c.Controller, opp.ID)
	}
	if c.Tapped {
		t.Error("Slicer should be untapped")
	}
	if !c.Goaded() || len(c.Goads) == 0 || c.Goads[0].By != me.ID {
		t.Errorf("goads = %+v, want goaded by Slicer's owner", c.Goads)
	}
	var cant bool
	g.WithWriteLock(func() { cant = g.CantBeSacrificedForEffect(id) })
	if !cant {
		t.Fatal("Slicer can be sacrificed this turn")
	}
	// The borrower's edict passes over it.
	g.WithWriteLock(func() {
		if cands := g.EffectSacrificeCandidatesForEffect(opp.ID, nil); len(cands) != 0 {
			for _, cid := range cands {
				if cid == id {
					t.Error("Slicer is an edict choice for the borrower")
				}
			}
		}
		if err := g.SacrificePermanentForEffect(id); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})
	if findBattlefieldCardForTest(g, id) == nil {
		t.Fatal("a sacrifice instruction took Slicer")
	}

	// Goaded by its owner, it attacks someone else if it can (CR
	// 701.15b). End of turn: it goes home and may be sacrificed again.
	defender := me.ID
	for _, p := range g.Seats {
		if p.ID != me.ID && p.ID != opp.ID {
			defender = p.ID
			break
		}
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, defender); err != nil {
		t.Fatalf("the borrowed Slicer can't attack: %v", err)
	}
	lockInAttacks(t, g)
	tfSettle(t, g)
	advanceToUpkeepOfSeat(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if got := controllerOf(t, g, id); got != me.ID {
		t.Errorf("after the turn, controller = %v, want the owner %v", got, me.ID)
	}
	g.WithWriteLock(func() { cant = g.CantBeSacrificedForEffect(id) })
	if cant {
		t.Error("\"this turn\" outlived the turn")
	}
}

// "If you don't, convert it."
func TestSlicerKeptConverts(t *testing.T) {
	g, me, _, id := slicerAtOpponentsUpkeep(t)
	answerMayChoice(t, g, me.ID, false)
	tfSettle(t, g)
	c := wantFace(t, g, id, 1, "Slicer, High-Speed Antagonist")
	if c.Controller != me.ID {
		t.Errorf("controller = %v, want unchanged", c.Controller)
	}
	var cant bool
	g.WithWriteLock(func() { cant = g.CantBeSacrificedForEffect(id) })
	if cant {
		t.Error("a kept Slicer can't be sacrificed")
	}
}

// The back face converts at end of combat after it hits a player, once
// for two hits (first strike and regular).
func TestSlicerBackFaceConvertsAtEndOfCombat(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerConverted(t, g, slicerRow(), me)
	transformerAttacks(t, g, id, opp.ID)
	emitCombatDamage(g, id, opp.ID, 3)
	emitCombatDamage(g, id, opp.ID, 3)
	tfSettle(t, g)
	wantFace(t, g, id, 1, "Slicer, High-Speed Antagonist")

	advanceTo(t, g, game.StepEndCombat)
	tfSettle(t, g)
	c := wantFace(t, g, id, 0, "Slicer, Hired Muscle")
	if c.TurnOvers != 2 {
		t.Errorf("turn-overs = %d, want 2 (converted once by the two hits)", c.TurnOvers)
	}
}
