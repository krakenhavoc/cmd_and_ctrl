package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// slice293_test.go — card-level coverage for the play-rate slice cut
// from tracker #293's gap lists (rites and rituals, flicker, the
// Phyrexian-mana instants, the pay-unless and counter spells, and three
// counter-themed creatures).

const (
	S293RiteOfFlameOracle      = "8a2e53f9-8100-488f-8504-b59e9bd1cc29"
	S293PyreticRitualOracle    = "86ef6474-613f-41fb-931c-d4279b03ed99"
	S293CullingRitualOracle    = "15e5136e-ed15-49a8-b027-e09436673fb4"
	S293GhostlyFlickerOracle   = "ad0070db-6454-41cb-861f-8f5b8fc2a3b8"
	S293NoxiousRevivalOracle   = "97cabeda-9fe3-490d-99b4-4c8d87c17157"
	S293ArchdruidsCharmOracle  = "3c1ef404-e2c6-486d-a5a2-d5779c71d498"
	S293TirelessTrackerOracle  = "e6555cca-e608-4c87-b835-9cd47fd1f543"
	S293ChasmSkulkerOracle     = "69facbc7-3859-4716-b627-5199571fb3cf"
	S293TyvarsStandOracle      = "4247b667-a0a8-4995-97c7-622d20132f7d"
	S293SpellPierceOracle      = "d64b0848-0193-4025-ba62-63ecd8fb9f50"
	S293DismemberOracle        = "fd74f8eb-0253-42dd-8277-186d4934da38"
	S293MuddleTheMixtureOracle = "0826cf7d-7ccc-459b-a92f-29dc169628f8"
	S293EvolutionWitnessOracle = "0e07f1af-5ff6-4da5-8683-aea4cc390975"
)

func PushGraveyardCardWithTypeLineID(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}
