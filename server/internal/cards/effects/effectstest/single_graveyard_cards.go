package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// single_graveyard_cards_test.go — #1807 (ADR 0106 §5), the cards that
// target "from a single graveyard". The engine half (the sameness rule
// and its four readers) is pinned in game/target_sameness_test.go;
// these are the cards on top of it.

const (
	OracleDigsiteConservator = "ced2ae56-988a-4c60-93d0-85a2261be72c"
	OracleCarrionBeetles     = "d64a51be-2dd1-485f-98fa-c93d4c296e68"
	OracleRagDealer          = "493acba2-92b0-43a4-9f4b-d49a91ca77f5"
	OracleDecompose          = "12c2b6bb-05f8-4291-b9a5-c4ddb7a69b9e"
	OracleFamishedGhoul      = "22c74e4b-b1fe-4a35-acd8-fa4840af3c06"
	OracleGriffnautTracker   = "6eff5e17-946b-4433-9f48-88f103844c42"
	OracleScarabFeast        = "8b9043d3-a1c6-4f49-9c49-ef78bbfbd4ac"
	OracleRapidDecay         = "2c769dfc-79be-4fcb-840c-b72cafdd0a23"
	OracleShredMemory        = "16cae855-d93a-445d-9cac-dbc5a7c18cb1"
	OracleArashinSunshield   = "1175d482-d8a2-467f-bc50-2f4b241966bb"
	OracleGravegouger        = "2677c0d9-1251-478e-b157-072859c593a9"
	OracleSoulShackledZombie = "f81ca3a5-356d-4b5d-afc2-a33aa774816b"
	OracleEbonyCharm         = "2f3237fe-9e07-4aad-8bc8-7ef4b510dd36"
	OracleRatsFeast          = "327c5ca2-4eeb-49cb-84fb-f6a4b06b24bc"
	OracleSereneRemembrance  = "f980c8b4-4cd7-42d2-82d5-96e3275b2647"
	OracleUnlicensedHearse   = "c640654c-487e-4a2c-aced-126ed835b78f"
	OracleWasteManagement    = "8d4e0866-d8f5-4eb6-a0fd-3fa9d4b9cf4a"
	OracleKayaOrzhovUsurper  = "7dd4a1a1-d5f4-4ac7-a9f6-34af411f070b"
	OracleCeaseDesist        = "ba4d644f-1931-4fc4-aed5-681a476a5a58"
	OraclePushPull           = "7e050495-bed7-43b9-abce-866c61beb1da"
)

func SgCards(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func SgInExile(g *game.Game, ids ...uuid.UUID) bool {
	for _, id := range ids {
		if !g.Exile.Contains(id) {
			return false
		}
	}
	return true
}

// --- entry triggers -----------------------------------------------

// sgAnswerTargets answers the open pick_target prompt with refs and
// returns the error.
func SgAnswerTargets(t *testing.T, g *game.Game, ids ...uuid.UUID) error {
	t.Helper()
	ch := PickTargetPrompt(t, g)
	return g.ResolvePickTargets(ch.ID, ch.Chooser, SgCards(ids...))
}
