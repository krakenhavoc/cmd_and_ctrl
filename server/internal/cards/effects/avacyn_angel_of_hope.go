package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avacyn, Angel of Hope — Creature — Angel, {5}{W}{W}{W}, 8/8:
//
//	"Flying, vigilance, indestructible"
//	"Other permanents you control have indestructible."
//
// Eight mana for a board that cannot be destroyed. The card is the
// reason indestructible had to become a real keyword rather than a
// badge: it is the format's canonical "your side of the table stops
// dying" effect, and until S25 (#77) it would have been an 8/8 flier
// with two lines of decoration.
//
// # Both halves, and why only one of them is a Static
//
// Avacyn's OWN indestructible is printed on her, so it rides
// `PrintedKeywords` — the engine synthesises a self-only layer-6
// static from that slot at catalog load, which is the same place a
// granted keyword lands. The second line is a genuine battlefield
// static granting to OTHER permanents, so it is written out.
//
// The self-exclusion in `AppliesTo` is not redundant with the
// printed keyword: leaving it out would grant Avacyn the keyword she
// already has, and while `GrantKeywordUntilEOT`-style dedupe would
// keep `HasKeyword` correct, the doubled string would render as two
// badges on the client's keyword row. Excluding self is cheaper than
// deduping.
//
// # "Permanents", not "creatures"
//
// The predicate is controller-only with no type test, so it covers
// lands, artifacts, enchantments and planeswalkers — Armageddon and
// Vandalblast bounce off a board with Avacyn on it. What still gets
// through: sacrifice effects, exile effects, -X/-X to zero toughness,
// and a planeswalker spending itself to 0 loyalty. None of those are
// destruction and the engine models every one of those gaps
// faithfully; game/indestructible.go enumerates them.
//
// # Avacyn does not protect herself from everything either
//
// Her own indestructible is subject to the same list. A Merciless
// Eviction on "creatures" exiles her, and the whole board loses the
// grant the instant she leaves — the layer engine recomputes from
// the battlefield, so removing the source removes the effect
// (CR 113.6).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "216cb26e-8da9-478b-bfbc-8030f7adee72",
		Name:            "Avacyn, Angel of Hope",
		PrintedKeywords: []string{"flying", "vigilance", "indestructible"},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID &&
					target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = append(c.Abilities, "indestructible")
			},
		}},
	})
}
