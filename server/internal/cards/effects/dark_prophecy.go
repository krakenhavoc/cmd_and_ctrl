package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Prophecy — Enchantment {B}{B}{B} (EDHREC rank 4332):
//
//	"Whenever a creature you control dies, you draw a card and you
//	 lose 1 life."
//
// Three black pips for an engine that turns a sacrifice deck's fodder
// into cards. The life loss is not a drawback the deck minds — it is
// the same deck that is draining the table — but it IS mandatory, and
// a Dark Prophecy on a board wipe has killed its own controller more
// than once. That is the card.
//
// "A creature you control", not "another": the Prophecy is an
// enchantment and cannot die as a creature itself, so the distinction
// never arises. Tokens count — WheneverACreatureYouControlDies is the
// plain dies-trigger, with no nontoken clause to add (contrast
// Midnight Reaper, whose whole text is that clause).
//
// The engine emits one EventLTB per card even for a simultaneous
// batch, so a board wipe fires the Prophecy once per creature, which
// is what the printed "whenever a creature dies" asks for — this is
// not one of the "one or more" triggers that needs the per-batch
// dedup.
//
// Printed order is draw THEN lose, which matters when the draw is off
// an empty library or when the life loss would kill you: the card is
// in hand before the loss, and the loss is the last thing that
// happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f47659b5-d834-4153-b9fc-6a0702c503d9",
		Name:         "Dark Prophecy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverACreatureYouControlDies("Dark Prophecy — draw a card, lose 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
				}),
		},
	})
}
