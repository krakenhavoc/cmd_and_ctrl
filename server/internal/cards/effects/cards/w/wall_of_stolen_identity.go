package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wall of Stolen Identity — Creature — Shapeshifter Wall {3}{U}, 0/0
// (Commander 2019):
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield, except it's a Wall in addition to its other types and has
//	 defender. When you do, tap the copied creature and it doesn't untap
//	 during its controller's untap step for as long as you control this
//	 creature."
//
// Clone's entry copy with an except clause that adds the Wall subtype and
// defender as copiable values (CR 707.9b), and Superior Spider-Man's "when you
// do": the except clause names the copied creature on the entry event and
// declares a CR 603.12 reflexive trigger that the engine creates once the
// copy has landed. It taps that creature and holds it tapped for as long as
// you control the Wall (Spider-Woman's lock, DurationWhileYouControlSource);
// a copied creature that has since left or come back as a new object is left
// alone (CR 400.7). Declined, the Wall is a 0/0 that dies as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92c58354-e106-402f-8348-6c7cdb67c6df",
		Name:         "Wall of Stolen Identity",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Wall of Stolen Identity", effects.AnyCreatureOnBattlefield, wallOfStolenIdentityExcept),
		},
	})
}

// wallOfStolenIdentityExcept is the except clause and the "when you do".
func wallOfStolenIdentityExcept(ev *game.ReplacementEvent, v *game.PrintedValues, g *game.Game, source *game.Card) {
	v.AddSubtype("Wall")
	v.AddKeyword("defender")
	copied := ev.CopySourceForEffect()
	c, ok := g.LookupCardForEffect(copied)
	if !ok {
		return
	}
	// The declaration is refused only for a malformed one, which this is
	// not; a copy has nothing to do about a refusal anyway.
	_ = g.QueueReflexiveTriggerAfterReplacedMoveForEffect(ev, source, game.ReflexiveTrigger{
		Label:  "Wall of Stolen Identity — tap the copied creature; it doesn't untap while you control this creature",
		Body:   wallOfStolenIdentityTapBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: copied, Epoch: c.ObjectEpoch}},
	})
}

var wallOfStolenIdentityTapBody = game.ReflexiveBody("wall-of-stolen-identity/tap-and-hold", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	ref := item.Params.Object
	cur, ok := g.PermanentRefForEffect(ref.ID)
	if !ok || cur != ref {
		return nil
	}
	return effects.TapAndHoldWhileYouControlThis(effects.NewContext(g, item), []uuid.UUID{ref.ID})
}), nil)
