package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// master_of_cruelties_test.go — #3531, ADR 0187 amendment 2026-10-16:
// "This creature can only attack alone" (CR 506.5, 508.1c-d) and "that
// player's life total becomes 1. This creature assigns no combat damage
// this combat."

const mocOracle = "d135c457-9be5-4a6e-a6fb-5b35cfb70cb0"

func mocMaster(g *game.Game) uuid.UUID {
	me := g.Seats[0]
	return apaPush(g, me.ID, me.ID, apaCreature("Master of Cruelties", mocOracle, 1, 4))
}

// Unblocked at a player: their life total becomes 1, and Master deals no
// combat damage, so it stays 1. The mark ends with the combat.
func TestMasterOfCrueltiesSetsLifeToOne(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	master := mocMaster(g)
	ncdPlay(t, g, master, true)
	advanceTo(t, g, game.StepEndCombat)
	if opp.Life != 1 {
		t.Fatalf("defender at %d life, want 1", opp.Life)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if ncdMarked(g, master) {
		t.Error("the mark outlived the combat (this combat, not this turn)")
	}
}

// Blocked, the trigger does nothing and Master fights as usual.
func TestMasterOfCrueltiesBlockedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	master := mocMaster(g)
	wall := apaPush(g, opp.ID, opp.ID, apaCreature("Wall", "", 0, 9))
	life := opp.Life
	ncdPlayBlocked(t, g, master, wall, true)
	if opp.Life != life || ncdMarked(g, master) {
		t.Errorf("defender %d → %d, marked %v; want no change", life, opp.Life, ncdMarked(g, master))
	}
}

// Master can't join an attack, and nothing can join Master's.
func TestMasterOfCrueltiesCanOnlyAttackAlone(t *testing.T) {
	for _, masterFirst := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		master := mocMaster(g)
		bear := apaPush(g, me.ID, me.ID, apaCreature("Bear", "", 2, 2))
		advanceTo(t, g, game.StepDeclareAttackers)
		first, second := bear, master
		if masterFirst {
			first, second = master, bear
		}
		if err := g.DeclareAttacker(first, opp.ID); err != nil {
			t.Fatalf("masterFirst=%v: the first attacker was refused: %v", masterFirst, err)
		}
		err := g.DeclareAttacker(second, opp.ID)
		if !errors.Is(err, game.ErrAttackLimit) {
			t.Fatalf("masterFirst=%v: the second attacker got %v, want ErrAttackLimit", masterFirst, err)
		}
		var le *game.AttackLimitError
		if errors.As(err, &le) && le.Sentence(me.ID) != "Master of Cruelties can only attack alone." {
			t.Errorf("sentence %q", le.Sentence(me.ID))
		}
	}
}

// CR 508.1d with "can only attack alone": the declaration obeys as many
// requirements as it could. A goaded creature beside a free Master must
// attack, so Master may not; a goaded Master beside a free creature must
// attack, so the other may not; with both goaded either is legal and the
// declaration ends with one attacker.
func TestMasterOfCrueltiesAloneAgainstRequirements(t *testing.T) {
	cases := []struct {
		name                     string
		goadMaster, goadBear     bool
		masterRefused, bearRefus bool
	}{
		{"goaded bear", false, true, true, false},
		{"goaded master", true, false, false, true},
		{"both goaded", true, true, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, declareMaster := range []bool{true, false} {
				g := newCatalogGame(t)
				me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
				master := mocMaster(g)
				bear := apaPush(g, me.ID, me.ID, apaCreature("Bear", "", 2, 2))
				g.WithWriteLock(func() {
					if tc.goadMaster {
						g.GoadForEffect(master, opp.ID)
					}
					if tc.goadBear {
						g.GoadForEffect(bear, opp.ID)
					}
				})
				advanceTo(t, g, game.StepDeclareAttackers)
				id, refused := bear, tc.bearRefus
				if declareMaster {
					id, refused = master, tc.masterRefused
				}
				err := g.DeclareAttacker(id, third.ID)
				if refused != errors.Is(err, game.ErrAttackRequirement) || (!refused && err != nil) {
					t.Fatalf("declaring master=%v: err %v, want refused=%v", declareMaster, err, refused)
				}
				if refused {
					continue
				}
				// The declaration is complete: the checkpoint passes.
				if err := g.PassPriority(); err != nil {
					t.Errorf("declaring master=%v: the active player's pass was refused: %v", declareMaster, err)
				}
			}
		})
	}
}
