package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conditional_cast_zones_test.go — #3105. Spec.CastableZonesIf and the
// three cards that use it: Lightwheel Enhancements (max speed),
// Gravecrawler (a Zombie) and Marang River Prowler (a black or green
// permanent). Each test casts the card out of the graveyard through the
// engine's own CastSpell, so the claim is "the cast is accepted /
// refused", not "a helper returns true".

const (
	LightwheelEnhancementsOracle = "17fae94e-558f-449a-b30c-e3c37a068b79"
	GravecrawlerOracle           = "09ff28b1-b6c9-48e6-b12e-2f0e644f709f"
	MarangRiverProwlerOracle     = "002e767c-0b30-4ad1-a924-88da400a4cff"
)

// condZoneGraveyardCard puts a catalog card into the seat's graveyard at a
// main phase.
func CondZoneGraveyardCard(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	t.Helper()
	AdvanceTo(t, g, game.StepPrecombatMain)
	c := game.NewCard(name, owner.ID)
	c.TypeLine = typeLine
	c.ManaCost = cost
	c.OracleID = oracle
	owner.Graveyard.PushTop(c)
	return c.InstanceID
}

func CondZoneCastFromGraveyard(g *game.Game, owner *game.Player, id uuid.UUID, targets ...game.TargetRef) error {
	return g.CastSpell(owner.ID, id, game.CastSpellParams{FromZone: "graveyard", Targets: targets})
}
