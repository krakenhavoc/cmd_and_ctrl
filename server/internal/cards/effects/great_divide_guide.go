package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Great Divide Guide — Creature — Human Scout Ally {1}{G}, 2/3:
//
//	"Each land and Ally you control has "{T}: Add one mana of any
//	 color.""
//
// The Guide is an Ally, so it has the ability too. A land's granted
// mana is an ordinary auto-tap source; an Ally CREATURE's is kept for
// last (ADR 0093 Decision 6).
//
// No simplification.
const greatDivideGuideGrant = "great-divide-guide/any-color"

func init() {
	Register(Spec{
		OracleID:     "79e69a91-d580-47fb-be76-1e32c50d2fa0",
		Name:         "Great Divide Guide",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{AnyColorManaGrant(greatDivideGuideGrant)},
		Static: []game.StaticAbility{GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.Controller == source.Controller && (target.IsLand() || target.HasSubtype("Ally"))
		}, greatDivideGuideGrant)},
	})
}
