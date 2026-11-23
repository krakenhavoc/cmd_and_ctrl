package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cant_become_untapped_cards_test.go — #4574: "can't become untapped"
// (CR 614.17), proven on Frozen in Ice and Spider-Woman, Secret Agent. Act
// of Heroism ("Untap target creature") is the untap effect that must do
// nothing to a locked creature.

const (
	OracleFrozenInIce = "82fa14d8-5019-4e1b-a86c-923d9292b87a"
	OracleSpiderWoman = "04e53f36-9f16-4678-9caf-9afda85f76b4"
)

// untapSpell casts Act of Heroism at id and lets it resolve.
func UntapSpell(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	CastCatalogSpell(t, g, "Act of Heroism", "Instant", ActOfHeroismOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
	PassPriorityAroundTable(t, g)
}
