package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Savvy Hunter — Creature — Human Warrior {1}{B}{G}, 3/3 (EDHREC
// rank 3235):
//
//	"Whenever this creature attacks or blocks, create a Food token.
//	 (It's an artifact with "{2}, {T}, Sacrifice this token: You gain
//	 3 life.")
//	 Sacrifice two Foods: Draw a card."
//
// The Food deck's three-drop. One printed ability with two trigger
// conditions is one TriggeredAbility watching two kinds (the Sun
// Titan shape): the source's own attack declaration, or the source's
// own block declaration — EventBlock names the blocker in CardID and
// is emitted once per attacker it is declared against. The Food is
// the shared template.
//
// "Sacrifice two Foods: Draw a card" is a sacrifice cost with a count
// of two (#747, SacrificeN): the activator names exactly two Foods
// they control, and both leave as one simultaneous exit. Until #747
// the ability was omitted, because a cost could sacrifice only one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "602132c2-8ee8-41f8-bfac-cb17d32203f5",
		Name:         "Savvy Hunter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source) || b28SelfBlocked(ev, source)
			}, "Savvy Hunter — create a Food", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:  "Sacrifice two Foods: Draw a card.",
			Cost:   SacrificeN(2, "two Foods", HasSubtype("Food")),
			Effect: Do(DrawCards{N: 1}),
		}},
	})
}
