package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zombie Boa — Creature — Zombie Snake {4}{B}, 3/3 (Apocalypse, S82 OLD,
// #5614):
//
//	"{1}{B}: Choose a color. Whenever this creature becomes blocked by a
//	 creature of that color this turn, destroy that creature. Activate only as
//	 a sorcery."
//
// The colour is chosen as the ability resolves (CR 608.2d), and it creates a
// delayed triggered ability (CR 603.7a) that lasts the turn and fires once for
// each blocker of that colour (CR 509.3d; the colour is read as the creature is
// declared as a blocker). The destroy is the shared payload body: it hits that
// blocker only while it is still the same object on the battlefield (CR 603.7c).
// A Boa that left and came back is a new object and never fires it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4949d62-c2f9-4425-a77e-444824955ecc",
		Name:         "Zombie Boa",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{B}: Choose a color. Whenever this creature becomes blocked by a creature of that color this turn, destroy that creature. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{1}{B}"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				effects.ChooseColorTail(game.ColorForBenefit, g, ctx.Controller(), ctx.Source(),
					"Zombie Boa — choose a color", ctx.ItemTail(zombieBoaColorChosen))
				return nil
			},
		}},
	})
}

// zombieBoaBlockedByColor is the delayed trigger's each-condition: a block
// declaration by a creature of the chosen colour (CondParams.Name) against the
// Boa named by CondParams.Object, while it is still that object. EventBlock
// carries the blocker in CardID and the attacker in Target. The blocker is the
// instance's payload, its "that creature".
var zombieBoaBlockedByColor = game.DelayedEachCondition("block/s82-old-b632-zombie-boa-blocked-by-color",
	func(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
		if ev.Kind != game.EventBlock || p.Object.ID == uuid.Nil || ev.Target != p.Object.ID || p.Name == "" {
			return nil
		}
		boa, ok := effects.BattlefieldCardForEffect(g, p.Object.ID)
		if !ok || boa.ObjectEpoch != p.Object.Epoch {
			return nil
		}
		blocker, ok := effects.BattlefieldCardForEffect(g, ev.CardID)
		if !ok || !blocker.IsCreature() || !blocker.HasColor(p.Name) {
			return nil
		}
		return []uuid.UUID{ev.CardID}
	})

// zombieBoaColorChosen schedules the turn's delayed trigger for the chosen
// colour. An on-disk key.
var zombieBoaColorChosen = effects.CardTail("then/s82-old-b632-zombie-boa-color", func(ctx *effects.Context, o game.TailOutcome) error {
	color := o.Tail.Word
	boa, ok := effects.BattlefieldCardForEffect(ctx.Game, ctx.Source())
	if color == "" || !ok {
		return nil // a Boa that has left can no longer become blocked
	}
	return effects.DelayedOnEvent{
		Label:      "Zombie Boa — destroy a " + game.ColorName(color) + " creature that blocks it this turn",
		On:         []game.EventKind{game.EventBlock},
		Condition:  zombieBoaBlockedByColor,
		CondParams: game.EffectParams{Object: game.ObjectRef{ID: boa.InstanceID, Epoch: boa.ObjectEpoch}, Name: color},
		Body:       effects.DestroyThePayloadPermanentBody,
		Repeats:    true,
	}.Apply(ctx)
})
