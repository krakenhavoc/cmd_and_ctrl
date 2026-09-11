package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pawn of Ulamog — Creature — Vampire Shaman {1}{B}{B}, 2/2:
//
//	"Whenever this creature or another nontoken creature you control
//	 dies, you may create a 0/1 colorless Eldrazi Spawn creature
//	 token. It has 'Sacrifice this token: Add {C}.'"
//
// The aristocrats deck's mana engine: every nontoken death refunds
// itself as a body that is itself sacrifice fuel AND a mana source.
// Pair it with a free sacrifice outlet and each creature dies twice.
//
// Two clauses doing real work, both of which the catalog can now
// express:
//
//   - NONTOKEN. The Spawn it makes is a token, so the Spawns do not
//     loop into more Spawns — the card would be an infinite engine
//     otherwise. Midnight Reaper's split, used for the opposite
//     reason.
//   - "this creature OR ANOTHER" — the Pawn sees its own death, so
//     sacrificing the Pawn itself still yields a Spawn. The trigger
//     is found through the LKI map after it has left the battlefield.
//
// The token's "Sacrifice this token: Add {C}" rides on the template
// (a token has no oracle ID for the catalog to key on), so the Spawn
// taps for mana through the ordinary sacrifice-cost mana-ability
// path with no special-casing here.
func init() {
	Register(Spec{
		OracleID: "9bcaf141-1f1f-491f-aced-13dc093b9e2c",
		Name:     "Pawn of Ulamog",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && dead.Controller == source.Controller && !IsToken(dead)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Pawn of Ulamog — create a 0/1 Eldrazi Spawn",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   EldraziSpawnToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Pawn of Ulamog — create a 0/1 Eldrazi Spawn?",
			},
		}},
	})
}
