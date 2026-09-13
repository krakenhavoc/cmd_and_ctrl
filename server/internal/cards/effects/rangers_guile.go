package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ranger's Guile — Instant for {G}:
//
//	"Target creature you control gets +1/+1 and gains hexproof until
//	 end of turn."
//
// Blossoming Defense with one less point in each direction — the
// same card printed four years earlier and kept in the catalog
// because a 100-card singleton deck plays both. See
// blossoming_defense.go for the layer argument (7c and 6 cannot
// share a registry entry) and for why "you control" is enforced
// rather than assumed.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "694cb93a-bcc5-44b0-a76c-19ae4a4e5e0f",
		Name:         "Ranger's Guile",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (BoostUntilEOT{
				Target:    target,
				Power:     1,
				Toughness: 1,
				Label:     "Ranger's Guile — +1/+1",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"hexproof"},
				Label:    "Ranger's Guile — hexproof",
			}.Apply(ctx)
		},
	})
}
