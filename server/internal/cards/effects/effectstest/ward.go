package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ward_test.go — CR 702.21. The cases that matter are the ones that
// separate ward from hexproof, because a ward implemented at the
// targeting gate would pass none of them:
//
//	announce succeeds        a warded permanent is a LEGAL target
//	the payer is the caster   not the ward permanent's controller
//	paying resolves the spell a refused target could never do this
//	declining counters it     and the removal never resolves
//
// Plus the one hexproof-shaped case ward shares: your own spell on
// your own warded creature does not trigger it.

const (
	RimeshieldOracle   = "ec47cf67-2580-464f-8118-7eabea5be11c"
	HulkingRaptorOracl = "9f2e7533-cfc1-4dd8-a9b8-09cdc712ef2f"
)

// castAtWardedCreature seeds a Doom Blade in `caster`'s hand and
// casts it at `victim`, returning the spell's ID.
func CastAtWardedCreature(t *testing.T, g *game.Game, caster *game.Player, victim uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Doom Blade", TypeLine: "Instant",
		OracleID: DoomBladeOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("a warded permanent is a LEGAL target — announce must succeed: %v", err)
	}
	return id
}
