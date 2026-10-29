package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azog, Moria's Ruin — Legendary Creature — Goblin Soldier {2}{B}, 1/3
// (The Hobbit):
//
//	"When Azog enters, destroy up to one other target creature. Its
//	 controller amasses Goblins X, where X is that creature's power. If
//	 you controlled that creature, draw a card."
//
// One trigger with an optional, other-than-Azog target. The creature's
// controller and power are read BEFORE the destroy (CR 608.2h: last-known
// information would give the same answer, but reading first needs no
// reconstruction), because the destroy may move it off the battlefield.
// The sentences after the destroy do not depend on it succeeding: an
// indestructible creature still gives its controller the Goblins, and the
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
	owner, power := c.Controller, c.CurrentPower()
	if power < 0 {
		power = 0
	}
	if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	if err := (Amass{Controller: owner, Subtype: "Goblin", N: power}).Apply(ctx); err != nil {
		return err
	}
	if owner == item.Controller {
		return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
	}
	return nil
}
