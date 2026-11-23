package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d12e5ce0-5705-4c35-9a93-b883db52c80c",
		Name:         "Batterskull",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LivingWeapon("Batterskull"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(4, 4),
			effects.GrantToAttached("vigilance", "lifelink"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}: Return this Equipment to its owner's hand",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{3}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BounceToHand{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				},
			},
			effects.EquipAbility("{5}"),
		},
	})
}
