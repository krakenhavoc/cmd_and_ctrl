package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b8_cards_test.go — Secrets of Strixhaven Commander, eighth batch
// (#4834). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	SocB8LatticeOracle     = "0030578c-afaf-4072-a62b-2510bdceb64a"
	SocB8ComposerOracle    = "61816f98-9d1a-4b16-9194-cb7249880710"
	SocB8CampusOracle      = "45773715-3f46-4671-b633-bf087e892e26"
	SocB8OpusOracle        = "985f9dd7-e693-48ff-9dd6-a1a3cbaa236b"
	SocB8BloomOracle       = "7c699180-d18e-4599-acd1-27b951d82136"
	SocB8HellkiteOracle    = "f426aa1a-e1dc-4563-b450-2e23d6bd980b"
	SocB8ImpetusOracle     = "c8decfdf-1421-4649-acde-5215e7f08bb9"
	SocB8MerchantOracle    = "3918eaa7-5a34-4d6a-be93-a95e733d041b"
	SocB8MirrorwingOracle  = "9ea14a03-c6a2-49ec-953d-4249a2c27e1d"
	SocB8GiantGrowthOracle = "5748ebf1-24e3-499d-ab7c-c2cebd462a24"
)

// socB8Named lists the permanents `controller` controls with that name.
func SocB8Named(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}
