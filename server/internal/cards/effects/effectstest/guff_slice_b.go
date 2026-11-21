package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// guff_slice_b_test.go — the Guffster's Party deck request (#4184),
// slice b: Ichormoon Gauntlet, Jaya's Phoenix, Leori, Sparktouched
// Hunter, Repeated Reverberation, Rowan Kenrith, Oath of Teferi, Narset
// of the Ancient Way and Saheeli, Sublime Artificer. The Chain Veil
// followed with #4191 (the_chain_veil_test.go).

const (
	GuffbIchormoonOracle = "9fc7aba7-eda2-4134-9c10-31ee35201b72"
	GuffbJayaOracle      = "c6a67820-5066-462f-a621-1ecab8bf4f97"
	GuffbLeoriOracle     = "ff94e2d5-8ba1-4f6c-916d-58235a60743e"
	GuffbReverbOracle    = "9a664735-6e38-407f-9d2f-a6d279cd6309"
	GuffbRowanOracle     = "e191976c-6ab1-4a23-aeb8-8a88bd4bd203"
	GuffbOathOracle      = "ee00a42d-6ee4-4e46-92d6-62e8e5910880"
	GuffbNarsetOracle    = "b7ac6e89-941f-4ad9-a437-38ec38ec5263"
	GuffbSaheeliOracle   = "ab655810-f67e-4576-9174-35b16f5e3235"
	GuffbChainVeilOracle = "fb88fb3d-6a85-4bb1-b316-7f759b45a0fa"

	// Chandra, Torch of Defiance: its second row is "+1: Add {R}{R}.",
	// a loyalty ability whose every copy is two more mana in the pool.
	GuffbChandraOracle = "12cc800c-c7be-4811-8207-fd01a99cc892"
	GuffbChandraAddRR  = 1
)

// guffbWalker puts a catalog planeswalker with `loyalty` under `owner`.
func GuffbWalker(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, loyalty int) uuid.UUID {
	id := PushCatalogPermanent(g, owner, name, typeLine, oracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, loyalty) })
	return id
}

func GuffbChandra(g *game.Game, owner uuid.UUID) uuid.UUID {
	return GuffbWalker(g, owner, "Chandra, Torch of Defiance", "Legendary Planeswalker — Chandra", GuffbChandraOracle, 4)
}
