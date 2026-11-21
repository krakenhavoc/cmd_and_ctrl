package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// terror_of_the_peaks_test.go — #3860. "Spells your opponents cast that
// target this creature cost an additional 3 life to cast" is a cost the
// Dragon adds to somebody else's spell once its targets are announced
// (Spec.TargetTaxes, game/target_tax.go).

// castBoltAtFrom casts a Lightning Bolt for `caster` (the active seat,
// since the sandbox cast helpers advance to a main phase on its turn)
// and returns the error.
func CastBoltAtFrom(t *testing.T, g *game.Game, caster *game.Player, target game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	id := PushCatalogHandCard(caster, "Lightning Bolt", "Instant", LightningBoltOracle)
	return id, g.CastSpell(caster.ID, id, game.CastSpellParams{Targets: []game.TargetRef{target}})
}
