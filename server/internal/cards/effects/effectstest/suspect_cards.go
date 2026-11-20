package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// suspect_cards_test.go — #2698, suspect (CR 701.60) through the real
// catalog: the primitives, the predicate, and the nineteen cards that
// ship with the seam.

const (
	SbBarbedServitor      = "e3d82066-d8b9-47bf-8821-1370c506970b"
	SbPersonOfInterest    = "3d18cd9b-6013-48d7-9269-4787e7585a51"
	SbRuneBrandJuggler    = "8c105948-4b8f-4615-b66a-2bbdb3f3d6dd"
	SbJonahJameson        = "de861715-fd0b-493e-9a7c-c470a23044c0"
	SbRubblebeltBraggart  = "4ae7f959-9f44-4769-8bd3-49ca18928aba"
	SbRepeatOffender      = "4d1885d1-212e-47ae-865f-e6dc4148ddf2"
	SbClandestineMeddler  = "f11ec1c9-347d-4d79-aaa9-5919db5f9044"
	SbAbsolvingLammasu    = "dbedd45e-f76c-4be7-ad89-ff57bed6626f"
	SbAgrusKos            = "11b7ee78-fca6-4b34-93e4-e60492583d50"
	SbEliminateImpossible = "c78d4cec-6764-4cce-96d2-a2d85a57b218"
	SbCaughtRedHanded     = "484b3586-c0ed-4793-a3c0-e99f265e092d"
	SbItDoesntAddUp       = "dcfc8c59-d032-414d-9da6-125ebcb1366c"
	SbPresumedDead        = "a2d3e3f0-7399-4eb9-a0cf-831710c3d431"
	SbConvenientTarget    = "2f214dcf-9601-4ccb-bdac-2f9b9f6dff99"
	SbStashedSkeleton     = "64219e17-1302-4a43-85d2-8a6214448890"
	SbReasonableDoubt     = "020503f5-0915-4cd5-a617-1c7bde045e10"
	SbIncriminatingImp    = "75a7cad2-782c-4a5b-9193-5ba789bf5326"
	SbAgencyCoroner       = "3686fd97-49ff-4a92-9cb2-7d8e9438d945"
	SbDeadlyComplication  = "c53c75a0-c248-4e76-8d28-f7e53b2dcc49"
)

// sbSuspected reads the designation off the battlefield.
func SbSuspected(g *game.Game, id uuid.UUID) bool {
	var out bool
	g.ReadSnapshot(func() { out = g.IsSuspected(id) })
	return out
}

// sbCantBlock is whether the creature's effective restrictions bar it
// from blocking.
func SbCantBlock(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return AuraRestrictions(t, g, id)&game.CantBlock != 0
}

// sbCastPicking casts a creature whose enters trigger asks for a target
// and answers it.
func SbCastPicking(t *testing.T, g *game.Game, name, oracle string, pick uuid.UUID) uuid.UUID {
	t.Helper()
	id := CastCatalogSpell(t, g, name, "Creature — Test", oracle, nil)
	PassPriorityAroundTable(t, g)
	PickCard(t, g, g.Seats[0].ID, pick)
	PassPriorityAroundTable(t, g)
	return id
}
