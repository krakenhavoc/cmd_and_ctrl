package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "58e9f5db-c369-42c6-a26d-fc31e25d5f83",
		Name:         "Emergency Regeneration",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GrantKeywordUntilEOT{
				Match:    And(Artifact(), YouControl()),
				Keywords: []string{"hexproof", "indestructible"},
				Label:    "Emergency Regeneration — hexproof and indestructible",
			}.Apply(ctx)
		},
	})
}
