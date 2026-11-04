package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain America, First Avenger — Legendary Creature — Human Soldier Hero
// {R}{W}{U}, 4/4 (Marvel, #4799):
//
//	"Throw ... — {3}, Unattach an Equipment from Captain America: He deals
//	 damage equal to that Equipment's mana value divided as you choose
//	 among one, two, or three targets.
//	 ... Catch — At the beginning of combat on your turn, attach up to one
//	 target Equipment you control to Captain America."
//
// Throw is the unattach cost (UnattachAnEquipment, game.AbilityCost.
// UnattachEquipment): the activator names one Equipment attached to him,
// whoever controls it, and it stays on the battlefield unattached
// (CR 701.3d). With none attached the ability cannot be activated
// (CR 118.3). The damage is divided as the ability is activated
// (CR 601.2d via 602.2b), so its amount is read from the Equipment the same
// announcement names, while it is still attached: its mana value
// (DivideUnattachedManaValue). Each target takes at least 1, so an
// Equipment with mana value 0 pays for no target and the ability cannot be
// activated with it; one with mana value 1 hits exactly one target. A
// target that leaves in response takes nothing and its share is lost
// (CR 608.2b). He deals the damage, as he last existed if he has left.
//
// Catch is Cloud, Ex-SOLDIER's attach (AttachTargetsToSource) at the
// beginning of combat on your turn, with an "up to one" clause: no
// Equipment is a legal choice, and an Equipment that left or changed
// control in response is not attached.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d038055-035d-4698-b6b6-e2859bf2f1d1",
		Name:         "Captain America, First Avenger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, Unattach an Equipment from Captain America: He deals damage equal to that Equipment's mana value divided as you choose among one, two, or three targets.",
			Cost:    Plus(ManaCost("{3}"), UnattachAnEquipment()),
			Targets: TargetAny().WithCount(1, 3).Dividing(DivideBy(DivideUnattachedManaValue)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DealDividedDamage(NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Captain America, First Avenger — attach up to one target Equipment you control to Captain America",
					AttachTargetsToSource),
				TargetPermanent("up to one target Equipment you control", HasSubtype("Equipment"), YouControl()).WithCount(0, 1)),
		},
	})
}
