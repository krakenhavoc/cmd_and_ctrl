package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mirB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func MirB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, keywords ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner, Keywords: keywords,
	})
}

// mirB1Activate activates the source's ability `index` for `who` with
// the given targets, without resolving it.
func MirB1Activate(g *game.Game, who, source uuid.UUID, index int, targets ...game.TargetRef) error {
	return g.ActivateCatalogAbility(who, source, index, game.ActivateAbilityParams{Strict: true, Targets: targets})
}
