package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Batterskull — Artifact — Equipment for {5}:
//
//	"Living weapon (When this Equipment enters, create a 0/0 black
//	 Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +4/+4 and has vigilance and lifelink.
//	 {3}: Return this Equipment to its owner's hand.
//	 Equip {5}"
//
// The original living weapon, and the card that makes the keyword
// worth building shared machinery for: five mana produces a 4/4
// vigilance lifelink body out of nothing, and the {3} bounce means a
// removal spell aimed at the Germ costs the opponent a card while
// Batterskull walks back to hand and does it again.
//
// The bounce ability is an ordinary Spec.Activated entry with a mana
// cost and no target: BounceToHand on the ability's own source. It
// is deliberately NOT sorcery-speed — equip is gated by CR 702.6a
// and this is not equip, so returning Batterskull in response to a
// Disenchant is a real play the card is famous for.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d12e5ce0-5705-4c35-9a93-b883db52c80c",
		Name:         "Batterskull",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LivingWeapon("Batterskull"),
		},
		Static: []game.StaticAbility{
			PumpAttached(4, 4),
			GrantToAttached("vigilance", "lifelink"),
		},
		Activated: []ActivatedAbility{
			{
				Label: "{3}: Return this Equipment to its owner's hand",
				Cost:  ManaCost("{3}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			},
			EquipAbility("{5}"),
		},
	})
}
