package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Origin of the Hidden Ones — Enchantment — Saga {3}{R}:
//
//	"I — This Saga deals 4 damage to any target.
//	 II — Create two 1/1 black Assassin creature tokens with menace.
//	 III — Whenever an Assassin you control attacks this turn, create a
//	 1/1 black Assassin creature token with menace that's tapped and
//	 attacking."
//
// Chapter III sets up a repeating delayed trigger (CR 603.7b) lasting the
// turn, Indulge's shape: it fires once for each Assassin its controller
// attacks with, and each new Assassin enters tapped and attacking a player
// or planeswalker its controller chooses (CR 508.4). The tokens are not
// declared as attackers, so they trigger nothing, this trigger included.
func init() {
	const name = "Origin of the Hidden Ones"
	effects.Register(effects.Spec{
		OracleID:     "9c8f1abb-bd08-41e0-91ed-56ff25d7f9f2",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "4 damage to any target"),
				effects.TargetAny(),
				effects.SourceDealsDamageToEachLegalTarget(4)),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "create two 1/1 black Assassin creature tokens with menace"),
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black Assassin with menace"), N: 2})),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "whenever an Assassin you control attacks this turn, create a tapped and attacking Assassin"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.DelayedOnEvent{
						Label:      "Origin of the Hidden Ones — create a 1/1 black Assassin creature token with menace that's tapped and attacking",
						On:         []game.EventKind{game.EventAttack},
						Condition:  effects.GapB41AssassinYouControlAttacks,
						CondParams: game.EffectParams{Name: ctx.TextSubtype("Assassin")},
						Body:       effects.GapB41AssassinAttackingBody,
						Repeats:    true,
					}.Apply(ctx)
				}),
		},
	})
}
