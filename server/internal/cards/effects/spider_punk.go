package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Punk — Legendary Creature — Spider Human Hero {1}{R}, 2/1:
//
//	"Riot (This creature enters with your choice of a +1/+1 counter or
//	 haste.)
//	 Other Spiders you control have riot.
//	 Spells and abilities can't be countered.
//	 Damage can't be prevented."
//
// Riot is the engine's keyword (ADR 0109 §10), and the second line is
// a layer-6 grant of it: a Spider entering under Spider-Punk has riot
// as it would exist on the battlefield, so the CR 614.12 look-ahead
// asks it as it enters. "Spiders" names a creature type, which a
// kindred permanent shares (CR 205.3m), so the grant is not narrowed
// to creatures: Firdoch Core, a changeling kindred artifact, is a
// Spider under it. The type is read through the source's text, as a
// lord's is (CR 612.2).
//
// The third line is ADR 0106 §4's battlefield static in its "any
// player" form, with the ability half #2014 added: every player's
// spells, and every player's activated and triggered abilities, can't
// be countered (CR 701.6a, CR 101.2). A Stifle aimed at one still
// resolves and does nothing.
//
// The fourth is ADR 0107 §5's battlefield static over every damage
// event (CR 615.12).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "03bf7a90-78c1-4e83-8553-e22d85e49876",
		Name:            "Spider-Punk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRiot},
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
					target.HasSubtype(source.TextSubtype("Spider"))
			}, game.KeywordRiot), "Other Spiders you control have riot."),
		},
		SpellsCantBeCountered: []game.CounterShieldStatic{
			SpellsAndAbilitiesCantBeCountered("Spells and abilities can't be countered."),
		},
		DamageCantBePrevented: DamageCantBePreventedStatic(),
	})
}
