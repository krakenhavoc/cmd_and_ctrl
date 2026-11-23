package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_cost_replacements_test.go — #1710: Devoted Druid's -1/-1
// cost and a planeswalker's loyalty cost now go through the CR 614
// counter window (game.payCostCounterLocked), marked CounterFromCost,
// instead of writing the counter map directly. Doubling Season still
// does nothing to either — a cost is not the effect of a resolving
// spell or ability (CR 614.16) — but a replacement that names no
// effect (Vizier of Remedies, Winding Constrictor, Vorinclex,
// Monstrous Raider) now applies, which is the long-standing Devoted
// Druid + Vizier of Remedies ruling and the loyalty-cost half ADR 0073's
// 2026-09-28 amendment confirms against the Doubling Season ruling.

const (
	CcrVizierOracle    = "79770e65-740a-44c7-bea2-a24e6a722c22"
	CcrDoublingSeason  = "01546b7d-a233-4176-8843-d732074dc5b6"
	CcrWindingOracle   = "c9404d7d-a026-4082-9fcb-1ab571a136b5"
	CcrVorinclexOracle = "5a3fdf5a-bff8-4896-b288-3f43f9a72d9b"
)

// ccrPushReplacement puts a named replacement-only permanent (Vizier
// of Remedies, Doubling Season, Winding Constrictor, Vorinclex) on the
// battlefield under `owner`.
func CcrPushReplacement(g *game.Game, owner uuid.UUID, name, typeLine, oracleID string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: owner, Controller: owner,
	})
}

// ccrLoyaltyWalker seats a planeswalker under `owner` with one
// activated ability whose only cost is `delta` loyalty — the same
// catalog-agnostic shape internal/game/loyalty_test.go and
// internal/protocol/loyalty_view_test.go use, reproduced here because
// it is package-private in both.
func CcrLoyaltyWalker(g *game.Game, owner uuid.UUID, loyalty, delta int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Planeswalker",
		TypeLine:   "Legendary Planeswalker — Test",
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{game.CounterLoyalty: loyalty},
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "loyalty ability",
			Cost:  game.AbilityCost{Loyalty: &delta},
		}},
	})
	return id
}
