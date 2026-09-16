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
// One declared simplification, weaker than printed: the draw
// ability is not implemented. "Sacrifice two Foods" is a cost paid
// with two permanents, and an ability cost sacrifices exactly one
// (Samwise Gamgee's gap). Shipping the draw for one Food would be
// stronger than printed (#259), so the ability is left off.
func init() {
	Register(Spec{
		OracleID:     "602132c2-8ee8-41f8-bfac-cb17d32203f5",
		Name:         "Savvy Hunter",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The draw ability isn't available — a cost can't sacrifice two Foods, so only the Food-making attack and block trigger works."},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source) || b28SelfBlocked(ev, source)
			}, "Savvy Hunter — create a Food", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
	})
}
