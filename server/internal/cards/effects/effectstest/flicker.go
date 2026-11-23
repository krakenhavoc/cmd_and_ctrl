package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flicker_test.go — S22 exile-and-return. Covers the return-from-
// exile primitive (new object, re-triggered ETB, replacement
// pipeline still consulted), the immediate blink (Y'shtola Rhul),
// and the delayed blink (Waterbender's Restoration, Cosmic
// Intervention). The delayed-trigger queue's own timing and undo
// contract live in server/internal/game/delayed_test.go.

const (
	YshtolaRhulOracle             = "a6a7bf77-0560-4572-a826-3bc9df1f78d1"
	waterbendersRestorationOracle = "285046f6-b3c4-4eb7-8712-9dffebabc762"
	CosmicInterventionOracle      = "cddccc2a-a76e-48b3-b4dd-dfeab89e1619"
	MulldrifterOracle             = "24d0f5e7-0d9e-4b76-900e-a7274e80312d"
	FlickerProbeOracle            = "test-flicker-etb-probe"
)

// flickerProbeETBs counts direct AsEnters-hook fires on the probe
// permanent. Mulldrifter covers the other ETB path (a declared
// `Triggered` watching EventETB, harvested onto the stack); the two
// are separate mechanisms and a return has to re-fire both.
var FlickerProbeETBs int

func init() {
	effects.Register(effects.Spec{
		OracleID: FlickerProbeOracle,
		Name:     "Flicker ETB Probe",
		AsEnters: func(_ *game.Card, _ *effects.Context) error {
			FlickerProbeETBs++
			return nil
		},
	})
}

// advanceToEndStepOf walks the turn engine forward to the named
// seat's end step — where EventBeginEndStep fires and the delayed
// return triggers drain.
func AdvanceToEndStepOf(t *testing.T, g *game.Game, seat int) {
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
func BattlefieldIDsNamed(g *game.Game, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func ExileHas(g *game.Game, id uuid.UUID) bool {
	return g.Exile.Contains(id)
}

// pushFlickerCreature puts a catalog creature straight onto the
// battlefield. Direct pushes emit no events, so no ETB fires from
// the setup itself — every count in these tests comes from a return.
func PushFlickerCreature(g *game.Game, owner uuid.UUID, name, oracleID string) uuid.UUID {
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
