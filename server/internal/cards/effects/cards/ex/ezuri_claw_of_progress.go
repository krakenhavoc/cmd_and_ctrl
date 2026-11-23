package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ezuri, Claw of Progress — Legendary Creature — Phyrexian Elf Warrior
// {2}{G}{U}, 3/3:
//
//	"Whenever a creature you control with power 2 or less enters, you get
//	 an experience counter.
//	 At the beginning of combat on your turn, put X +1/+1 counters on
//	 another target creature you control, where X is the number of
//	 experience counters you have."
//
// X is read as the combat trigger resolves. With no experience counters
// the trigger still targets (and puts nothing).
//
// No simplification.
func init() {
	combat := effects.AtBeginningOfYourCombat("Ezuri, Claw of Progress — put X +1/+1 counters on another target creature you control, where X is your experience counters",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			x := effects.ExperienceCounters(g, item.Controller)
			if x <= 0 {
				return nil
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
				}
			}
			return nil
		})
	combat.Targets = effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))
	effects.Register(effects.Spec{
		OracleID:     "40eacdae-fae5-46b9-a9e5-9211469f74fc",
		Name:         "Ezuri, Claw of Progress",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.CreatureWithPower2OrLessEnteredUnderYourControl,
				"Ezuri, Claw of Progress — you get an experience counter", effects.YouGetAnExperienceCounter),
			combat,
		},
	})
}
