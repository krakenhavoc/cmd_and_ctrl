package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_batch_d_cards_test.go — ADR 0129 PR 1's energy cards, batch D:
// each pays energy as printed (CR 107.14), is refused when short
// (CR 118.3) with nothing paid, and does its effect.

const (
	EdHeliosOne        = "cfb1a656-0bf1-484d-b099-33087914250b"
	EdMultiformWonder  = "9849c743-9492-4a1b-83bf-1be54614b85f"
	EdPeemaAetherSeer  = "4a00f56e-e558-4239-bc75-fb0af7639042"
	EdPeemaTrailblazer = "88600bd4-4dcc-4788-bba4-78b6cf5ad8f8"
	EdIronworks        = "b3c87cb9-a37a-428e-95d1-e6dd90b31097"
	EdRoilCartographer = "3cd7ef87-4cb4-42ff-842f-e9a465dab165"
	EdScurryOfGremlins = "ff24d18d-9c5d-4ffe-846f-a8d568b32724"
	EdSphinxRevelation = "9b8f6c6a-49a1-4061-a2c4-5be612e138fa"
	EdStoneIdol        = "411c3636-a5b3-42fa-8020-ceacf581a4f3"
	EdSynthEradicator  = "c68f09d8-a1eb-449b-8bf5-4251d7a14337"
)

func EdCard(t *testing.T, g *game.Game, owner *game.Player, name, oracle, typeLine string, power, toughness int) uuid.UUID {
	t.Helper()
	return ApaPush(g, owner.ID, owner.ID, game.Card{Name: name, OracleID: oracle, TypeLine: typeLine, Power: power, Toughness: toughness})
}

func EdCardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}
