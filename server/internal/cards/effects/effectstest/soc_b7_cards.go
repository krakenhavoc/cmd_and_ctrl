package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b7_cards_test.go — Secrets of Strixhaven Commander, batch 7
// (#4834). Each test plays the real catalog entry and asserts what its
// text says.

const (
	SocB7HofriOracle        = "663d451b-f848-41cc-b541-0c6fe37d3d54"
	SocB7ProdigyOracle      = "6f498ff7-38c9-44ea-9cb5-994a003825ee"
	SocB7SkypainterOracle   = "d751580b-bb72-4d1d-be27-d0410f883b2a"
	SocB7ChirographyOracle  = "6c948f71-5c86-450c-8828-df4d8deaf1dd"
	SocB7JadarOracle        = "fb3defc0-c93a-4423-9ff9-c399f8a8ef1f"
	SocB7KamiOracle         = "e937e0f5-ca35-4dce-9f05-8f2075731628"
	SocB7DuelistOracle      = "b8602f19-6fc2-43c4-b0c4-2cd11d5dbda2"
	SocB7MentorOracle       = "9c54bd34-667d-415c-ad26-82e0eb0ff409"
	SocB7InkDuelistOracle   = "e233cb5b-7237-4271-8157-9179cc897b6c"
	SocB7KineticOozeOracle  = "ab74c6b1-2099-4429-94b7-a5656ac04db3"
	SocB7DoomBladeOracle    = "59e7f2ae-4535-4191-98be-3e65b6b2befa"
	SocB7MindRotOracle      = "ad44cf74-b717-48fb-9fa2-77512024d76a"
	SocB7SkypainterPrepared = "Maestro's Gift"
	SocB7GrizzlyBearsOracle = "14c8f55d-d177-4c25-a931-ebeb9e6062a0"
)

// socB7Zone is the zone the card is in now, or "" when it is nowhere.
func SocB7Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	if z := g.FindCardZoneForEffect(id); z != nil {
		return z.Kind
	}
	return ""
}
