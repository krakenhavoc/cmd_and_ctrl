package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spend_only_cards_test.go — #1600's proof cards for mana only some
// spells may use and costs only some mana may pay (ADR 0040's
// 2026-10-03 amendment): Throne of Eldraine, Crypt Rats, Crimson
// Hellkite, Pillar of the Paruns and Obsidian Obelisk. The engine half
// is game/spend_only_test.go. The agreement tests at the bottom ask one
// board three questions — the view's legal_actions, the bot's
// enumerator and the real payment — and require one answer.

const (
	ThroneOfEldraineOracle  = "5b9a2b81-a645-43be-8001-03817ac210ca"
	CryptRatsOracle         = "104095ed-55e3-408e-bf70-4fe06bb16d2f"
	CrimsonHellkiteOracle   = "37a1f340-10d5-407e-bf2b-cbbbb6cab985"
	PillarOfTheParunsOracle = "677b8ce7-f922-4ee3-b311-f199da9b352b"
	ObsidianObeliskOracle   = "92a5e50d-8477-4654-9c1b-219e4fe3c5ab"
)

// basics puts n untapped basic lands of `land` (Plains, Island, …)
// under p's control and returns them.
func Basics(g *game.Game, p *game.Player, land string, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, PushCatalogPermanent(g, p.ID, land, "Basic Land — "+land, "", false))
	}
	return out
}
