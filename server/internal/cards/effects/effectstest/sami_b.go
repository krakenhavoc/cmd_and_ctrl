package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sami_b_test.go — Sami Whammy slice B (cantrip rocks, Food, relics),
// relates to #2190.

const (
	SamiChromaticSphere = "2e03e44a-9fff-4490-859f-b42e89e8563a"
	SamiPropheticPrism  = "134a9877-5cfb-4a2a-a0f0-930dce45f58b"
	SamiGoldenEgg       = "25977c02-e2d4-4afd-b12d-37ce4d58f453"
	SamiManaGeode       = "0844f4e6-2b98-4d09-a5c5-92f0a3b6a517"
	SamiOrazcaRelic     = "48b84b58-1a06-4bb4-be1f-ad3ca69e66dc"
	SamiPhial           = "08f17ebc-c0fd-493f-84b3-e9250694543e"
	SamiSkyclaveRelic   = "eb44b65d-6b56-4f20-a4fb-c5dd147a54c4"
	SamiInstantRamen    = "2283e409-c6c7-4de9-899b-2b3caea5f35e"
	SamiLembas          = "f71fcdc3-5e96-416e-a49d-37019806e2e2"
	SamiEnergyRefractor = "fd9dde4c-bacd-48d8-b4c9-a6a04f0b3338"
	SamiSemblanceAnvil  = "bbd0c406-c995-46e0-898d-375fdbede203"
	SamiSkyscanner      = "974f788a-039f-4310-a2fe-16b14a1e2d35"
	SamiWeddingInvite   = "d9e752f2-d552-4d72-babd-29eca3508820"
)

func SamiGiveColorless(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
}

func SamiInGraveyard(g *game.Game, p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}
