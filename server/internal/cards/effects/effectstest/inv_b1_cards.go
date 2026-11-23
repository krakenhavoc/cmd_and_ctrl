package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// invB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func InvB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// invB1Activate activates the source's ability `index` for `who` with
// the given targets, without resolving it.
func InvB1Activate(g *game.Game, who, source uuid.UUID, index int, targets ...game.TargetRef) error {
	return g.ActivateCatalogAbility(who, source, index, game.ActivateAbilityParams{Strict: true, Targets: targets})
}
