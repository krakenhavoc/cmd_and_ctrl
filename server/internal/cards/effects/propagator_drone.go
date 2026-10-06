package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Propagator Drone — Creature — Eldrazi Drone {1}{G}, 2/2:
//
//	"Devoid (This card has no color.)
//	 Creature tokens you control have evolve. (They have "Whenever a
//	 creature you control enters, if it has greater power or toughness
//	 than this token, put a +1/+1 counter on this token." They see this
//	 creature enter.)
//	 {3}{G}: Create a 0/1 colorless Eldrazi Spawn creature token with
//	 "Sacrifice this token: Add {C}.""
//
// The second evolve GRANT (#1805, ADR 0106 §3): a layer-6 keyword grant
// to creature tokens, read by the engine's evolve trigger
// (game/evolve.go) like a printed evolve.
//
// "They see this creature enter" is CR 603.6a working as written: the
// entry harvest brings the layer cache up to date before it asks who
// triggers, so the grant is already on the tokens when the Drone's own
// EventETB is judged, and a 0/1 Spawn evolves off the 2/2 Drone.
//
// Devoid is declared in PrintedKeywords, and the engine reads it as CR
// 702.114a's colour-defining ability (#2152), so the card is colourless
// in every zone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9485db26-7e03-4a66-beeb-627a3bc6f367",
		Name:            "Propagator Drone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.IsToken() && target.Controller == source.Controller
			}, game.KeywordEvolve),
		},
		Activated: []ActivatedAbility{{
			Label: `{3}{G}: Create a 0/1 colorless Eldrazi Spawn creature token with "Sacrifice this token: Add {C}."`,
			Cost:  ManaCost("{3}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: EldraziSpawnToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
