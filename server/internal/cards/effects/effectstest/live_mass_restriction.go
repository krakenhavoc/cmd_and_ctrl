package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// live_mass_restriction_test.go — #1650. A mass "can't block / can't
// be blocked this turn" changes the rules, not a characteristic, so
// CR 611.2c does not lock its set: RestrictUntilEOT's Scope form reads
// it live until cleanup. The single-Target form still follows exactly
// one object.

const (
	FalterOracle           = "d4b50749-a016-4aff-8d70-a1707cabf57b"
	MagmaticChasmOracle    = "6cfd2cd8-a86e-48fe-a85a-9a3444b5030c"
	SeismicStompOracle     = "a1d02a70-2543-45c6-a9a1-c1941f2f68f3"
	CosmotronicWaveOracle  = "079ac521-9414-46c1-acd9-2ff2ff047d22"
	HazardousBlastOracle   = "ffd666c7-a8ca-4467-8733-878efb127268"
	GlaringSpotlightOracle = "21570481-190e-4cfa-a5fd-a0642e2af569"
	ArtfulDodgeOracle      = "c174dcbb-03a0-439c-b3d8-ed61bd46dc67"
)

// pushSizedCreature is a creature of the given size and printed
// keywords under `owner`.
func PushSizedCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int, keywords ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test",
		ManaCost: "{1}{G}", Colors: []string{"G"},
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
		Keywords: keywords,
	})
}
