package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undead Augur — Creature — Zombie Wizard {B}{B}, 2/2 (EDHREC rank
// 1842):
//
//	"Whenever this creature or another Zombie you control dies, you
//	 draw a card and you lose 1 life."
//
// The Zombie deck's card draw. One dies trigger with two conditions:
// the Augur's own death (the card is already in the graveyard when
// the harvester finds it, and cardDied reads it by ID) or a Zombie
// its controller controlled dying. The Zombie is read post-move from
// its printed type line — a changeling counts — so a creature that
// was a Zombie only through a layer effect when it died is missed:
// weaker, never stronger. Draw first, lose life second, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c26887e1-27f4-4550-921b-53460e43c079",
		Name:         "Undead Augur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b17SelfOrZombieYouControlDied(ev, source, g)
			}, "Undead Augur — draw a card and lose 1 life", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
			}),
		},
	})
}
