package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reflector Mage — Creature — Human Wizard {1}{W}{U}, 2/3:
//
//	"When this creature enters, return target creature an opponent
//	 controls to its owner's hand. That creature's owner can't cast
//	 spells with the same name as that creature until your next turn."
//
// The ban is a granted cast restriction (cast_ban.go) on the creature's
// OWNER, not its controller, with a name filter
// (PermissionFilter.Named, ADR 0066's 2026-10-19 amendment) and a
// duration of Reflector Mage's controller's next turn. It is read by
// the one cast gate every cast already passes through, so the card is
// greyed in hand and refused from the command zone alike -- a bounced
// commander cannot be recast for its name either.
//
// The name is read from the creature BEFORE it moves, so a Clone that
// was copying a Bear bans "Grizzly Bears" and not "Clone", and a token
// bans its name too. The ban is not conditioned on the bounce
// succeeding: it is a second instruction of the same ability about "that
// creature", so a creature that was replaced on its way (a commander
// the owner sends to the command zone) still bans its name.
//
// The target is re-checked at resolution (CR 608.2b); an illegal one
// does nothing, ban included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5398362d-c45e-4d56-a622-8be1124730b9",
		Name:         "Reflector Mage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Reflector Mage — return target creature an opponent controls to its owner's hand; its owner can't cast spells with that name until your next turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					target, found := g.LookupCardForEffect(id)
					if !found {
						return nil
					}
					if err := (RestrictCasting{
						Player: target.Owner,
						Rule: game.CastBanRule{
							Kind:   game.CastBanNamed,
							Filter: game.PermissionFilter{Named: target.Name},
						},
						Label:    "Reflector Mage — can't cast spells named " + target.Name,
						Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
					}).Apply(ctx); err != nil {
						return err
					}
					return BounceToHand{Target: id}.Apply(ctx)
				}), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
