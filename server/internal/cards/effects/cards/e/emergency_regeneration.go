package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emergency Regeneration — Instant {1}{W} (Star Trek Commander, #4974):
//
//	"Artifacts you control gain hexproof and indestructible until end
//	 of turn."
//
// Heroic Intervention narrowed to artifacts. The set is read as the
// spell resolves (CR 611.2c): an artifact that enters later this turn
// is unprotected. It names artifacts, not artifact creatures, so mana
// rocks and Equipment are covered as well as Borg.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58e9f5db-c369-42c6-a26d-fc31e25d5f83",
		Name:         "Emergency Regeneration",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Artifact(), effects.YouControl()),
				Keywords: []string{"hexproof", "indestructible"},
				Label:    "Emergency Regeneration — hexproof and indestructible",
			}.Apply(ctx)
		},
	})
}
