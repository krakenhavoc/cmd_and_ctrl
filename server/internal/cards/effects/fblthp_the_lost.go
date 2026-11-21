package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fblthp, the Lost — Legendary Creature — Homunculus {1}{U}, 1/1 (Ravnica
// Remastered, #5948):
//
//	"When Fblthp enters, draw a card. If it entered from your library or was
//	 cast from your library, draw two cards instead.
//	 When Fblthp becomes the target of a spell, shuffle Fblthp into its owner's
//	 library."
//
// The draw is read as the trigger resolves, off the entering event (the zone
// the permanent came from, and whose it was) and off the spell it was cast as
// (CastProvenance.FromZone, CR 400.7d); "your library" is the Fblthp's
// controller's. The second ability is Dream Strix's trigger (SelfTargetedByASpell,
// a spell only, not an ability) with Cerulean Sphinx's tuck: the card goes into
// its owner's library and that owner shuffles, only if it is still the
// permanent the ability came from (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba5a7ac0-b625-42e9-af05-2680a01a95ed",
		Name:         "Fblthp, the Lost",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fblthp, the Lost — draw a card; two instead if it entered from or was cast from your library", gapB75FblthpDraw),
			On(game.EventBecomesTarget, SelfTargetedByASpell,
				"Fblthp, the Lost — shuffle it into its owner's library", ceruleanSphinxShuffleItself),
		},
	})
}

// gapB75FblthpDraw draws one card, or two when the Fblthp came from its
// controller's library or was cast from it.
func gapB75FblthpDraw(g *game.Game, item *game.StackItem) error {
	n := 1
	if item.Trigger != nil {
		ev := item.Trigger.Event
		fromLibrary := ev.EnteredFrom == game.ZoneLibrary && ev.EnteredFromOwner == item.Controller
		if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && c.Provenance.FromZone == game.ZoneLibrary && c.Controller == item.Controller {
			fromLibrary = true
		}
		if fromLibrary {
			n = 2
		}
	}
	return DrawCards{N: n}.Apply(NewContext(g, item))
}
