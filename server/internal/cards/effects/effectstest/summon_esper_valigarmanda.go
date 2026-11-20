package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// summon_esper_valigarmanda_test.go — Summon: Esper Valigarmanda
// (#3773, the Terra deck request #3655).

const ValigarmandaOracle = "cecd2f49-8ccb-472d-9cf1-fd0f4d93546d"

// valigarmandaGraveCard seeds a card of `typeLine` into p's graveyard.
func ValigarmandaGraveCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func ValigarmandaCast(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return TdCastFromHand(t, g, game.Card{Name: "Summon: Esper Valigarmanda", TypeLine: "Enchantment Creature — Saga Drake",
		OracleID: ValigarmandaOracle, ManaCost: "{3}{R}", Power: 3, Toughness: 3}, game.CastSpellParams{})
}
