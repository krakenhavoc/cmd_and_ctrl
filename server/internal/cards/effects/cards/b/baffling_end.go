package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baffling End — Enchantment {1}{W}:
//
//	"When this enchantment enters, exile target creature an opponent controls
//	 with mana value 3 or less.
//	 When this enchantment leaves the battlefield, target opponent creates a
//	 3/3 green Dinosaur creature token with trample."
//
// The exile is permanent: the creature never comes back. The leaves-the-
// battlefield trigger is Hunted Dragon's shape (RavB4TargetOpponentCreates for a
// different trigger): the opponent chosen as it goes on the stack creates the
// token and owns it, so a destroyed or bounced Baffling End still gives an
// opponent a Dinosaur.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03c35670-84ac-4825-b395-cefb76ceb89a",
		Name:         "Baffling End",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Baffling End — exile target creature an opponent controls with mana value 3 or less", effects.ExileFirstLegalCardTarget),
				effects.TargetCreature("target creature an opponent controls with mana value 3 or less", effects.OpponentControls(), effects.ManaValueLE(3))),
			effects.Targeting(effects.WhenThisLeaves("Baffling End — target opponent creates a 3/3 green Dinosaur creature token with trample",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return effects.CreateToken{Controller: t.ID, Template: effects.TokenCard("3/3 green Dinosaur with trample"), N: 1}.Apply(ctx)
						}
					}
					return nil
				}), effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
