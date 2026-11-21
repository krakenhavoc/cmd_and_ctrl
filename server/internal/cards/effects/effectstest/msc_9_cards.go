package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_9_cards_test.go — Marvel Super Heroes Commander, ninth batch
// (#4455). Each test plays the real catalog entry and asserts what its
// text says.

const (
	Msc9AbsorbingManOracle  = "a1afb796-1b0b-4f41-9c8f-5320c9af5989"
	Msc9AncestralOracle     = "a45bb088-16fd-4d91-8fd6-3c52f1c24748"
	Msc9ArmedAssailantOracl = "53d7722a-b60b-4520-abce-f31b1e6bba00"
	Msc9AstonishingOracle   = "24059283-edbd-4817-9f35-4be02f07b8f8"
	Msc9QuinjetOracle       = "905e4bd1-f5f2-4691-b13f-ce2628d26a38"
	Msc9BastOracle          = "e3233686-7703-460c-80ee-8cd37b3014ff"
	Msc9BionicBlowOracle    = "27683e43-7c3f-45b3-8ca4-968428000315"
	Msc9PantherOracle       = "7ba593fd-93db-452f-ad80-7f0f37afb09d"
	Msc9BlurOracle          = "b4a9b328-f976-4390-be9a-711ef357a50d"
	Msc9BrawnOracle         = "19910a90-d334-4d99-89f2-cd81cea3c824"
	Msc9TeamLeaderOracle    = "916cf34a-fd2c-47fe-bebb-f7b52b3dabed"
)

// --- Ancestral Communion -------------------------------------------------

func Msc9Commander(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := B12Creature(g, owner, "My Commander", "Legendary Creature — Human", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	return id
}
