package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loran's Escape — Instant {W} (EDHREC rank 2233):
//
//	"Target artifact or creature gains hexproof and indestructible
//	 until end of turn. Scry 1."
//
// Heroic Intervention's grant narrowed to one permanent instead of a
// mass "permanents you control": both keywords are layer-6 ability
// additions on the same clause and the same timestamp, so
// GrantKeywordUntilEOT takes them together. The scry has no ordering
// dependency on the grant and runs after it, matching the printed
// sentence break.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1a9048d6-f3fa-4f3b-9baa-35318177bc6e",
		Name:         "Loran's Escape",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"hexproof", "indestructible"},
					Label:    "Loran's Escape — hexproof and indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
				break
			}
			return Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
