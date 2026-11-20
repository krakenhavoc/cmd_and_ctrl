package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// role_tokens_test.go — #1945: Role tokens (CR 111.10, 303.7) and the
// CR 704.5z state-based action, through the real proof cards.

const (
	MonstrousRageOracle  = "646a2371-54c0-4492-ac2f-20f109d6108c"
	RoyalTreatmentOracle = "fd0f8fdc-4065-41f4-b8fb-ecb8f185774d"
	CursedCourtierOracle = "17916cf9-1e6c-41bd-96ce-2b050d828838"
	SpitefulHexmageOracl = "d07e0d88-8b7f-423d-97b7-cf7b80d8924c"
	FaunsbaneTrollOracle = "b707c131-de13-4d4d-839d-b9f47d62f090"
)

// rolesOn lists the Role tokens on the battlefield attached to host.
func RolesOn(g *game.Game, host uuid.UUID) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.IsRole() && c.IsAttachedTo(host) {
				out = append(out, c)
			}
		}
	})
	return out
}
