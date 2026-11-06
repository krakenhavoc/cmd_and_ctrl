package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boldwyr Aggressor — Creature — Giant Warrior {3}{R}{R}, 2/5 (Lorwyn
// Eclipsed, #4458):
//
//	"Double strike
//	 Other Giants you control have double strike."
//
// She-Hulk, Wallbreaker's lord grant with double strike: a layer 6
// keyword grant over the effective view (TribeFilter), so a changeling or
// a creature made a Giant by another effect gets it, and an opponent's
// Giant does not. Its own double strike is printed.
//
// No simplification.
func init() {
	giants := TribeFilter{Tribes: []string{"Giant"}, Others: true, YoursOnly: true}
	grant := KeywordGrant(giants.Matches, "double strike")
	grant.Label = "Other Giants you control have double strike."

	Register(Spec{
		OracleID:        "ab32c671-7686-435b-a987-14a08f6790eb",
		Name:            "Boldwyr Aggressor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static:          []game.StaticAbility{grant},
	})
}
