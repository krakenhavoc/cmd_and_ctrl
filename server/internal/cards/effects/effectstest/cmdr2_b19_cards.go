package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2B19Attach puts an Aura of the catalog onto the battlefield attached
// to host, controlled by owner.
func Cmdr2B19Attach(g *game.Game, owner, host uuid.UUID, name, oracle string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: AuraTypeLine,
		OracleID: oracle, Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}})
}
