package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Azog, Moria's Ruin — Legendary Creature — Goblin Soldier {2}{B}, 1/3
// (The Hobbit):
//
//	"When Azog enters, destroy up to one other target creature. Its
//	 controller amasses Goblins X, where X is that creature's power. If
//	 you controlled that creature, draw a card."
//
// One trigger with an optional, other-than-Azog target. The creature's
// controller and power are read BEFORE the destroy (CR 608.2h), because the
// destroy may move it off the battlefield, and carried to a keyed
// continuation that runs once the destruction has settled (a replacement
// prompt in between pauses the ability). The sentences after the destroy
// do not depend on it succeeding: an indestructible creature still gives its controller the Goblins, and the
// controller of the target amasses, not Azog's (Amass.Controller, the one
// printed amass that is not "you"). A negative power amasses 0 (CR 107.1b:
// X is never negative; the Army is still made and chosen). "If you
// controlled that creature" is Azog's controller against the controller
// read before the destroy. A target that is illegal on resolution
// fizzles the whole ability, as it should.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8b018a7-0350-4ee0-9582-8d391018bdee",
		Name:         "Azog, Moria's Ruin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Azog — destroy up to one other target creature; its controller amasses Goblins X; draw if you controlled it",
				azogDestroyAndAmass),
				Another(TargetCreature("up to one other target creature").WithCount(0, 1))),
		},
	})
}

func azogDestroyAndAmass(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, found := g.LookupCardForEffect(id)
	if !found {
		return nil
	}
	power := c.CurrentPower()
	if power < 0 {
		power = 0
	}
	// The rest of the ability runs from the destroy's continuation: a
	// creature that stops the destruction to answer a replacement prompt
	// (a commander) must not have been amassed for yet. The tail does not
	// ask whether anything was destroyed; the sentences do not depend on it.
	return g.DestroyPermanentsKeyedForEffect([]uuid.UUID{id}, withN(withCarry(ctx.Tail(azogAmassAndDraw), c.Controller), power))
}

// azogAmassAndDraw is the ability's second and third sentences: the target's
// controller (carried) amasses Goblins X (N), and Azog's controller draws if
// they controlled the creature.
var azogAmassAndDraw = CardTail("azog/amass-goblins-and-draw", func(ctx *Context, o game.TailOutcome) error {
	owner := tailCarry0(o)
	if err := (Amass{Controller: owner, Subtype: "Goblin", N: o.Tail.N}).Apply(ctx); err != nil {
		return err
	}
	if owner == ctx.Controller() {
		return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}
	return nil
})
