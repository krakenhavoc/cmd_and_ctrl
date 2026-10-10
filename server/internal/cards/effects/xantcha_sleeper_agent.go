package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Xantcha, Sleeper Agent — Legendary Creature — Phyrexian Minion
// {1}{B}{R}, 5/5:
//
//	"Xantcha enters under the control of an opponent of your choice.
//	 Xantcha attacks each combat if able and can't attack its owner or
//	 planeswalkers its owner controls.
//	 {3}: Xantcha's controller loses 2 life and you draw a card. Any
//	 player may activate this ability."
//
// The card ADR 0106 §1 and §2 were written for (#1793, #1794). Each line
// is an engine seam the card names in one call:
//
//   - The entry is ADR 0102's replacement (CR 614.1d, 614.12): asked as
//     Xantcha would enter, so a reanimated or blinked Xantcha asks too.
//     Harm, for the bot: the recipient is the one who loses life.
//   - "Attacks each combat if able" is a CR 508.1d requirement and "can't
//     attack its owner or planeswalkers its owner controls" a CR 508.1c
//     restriction, and the requirement is counted only over targets the
//     restriction allows. Both are Xantcha's own abilities, so a Xantcha
//     that loses its abilities is free of both (CR 613.1f). The owner is
//     read live (CR 108.3).
//   - "Any player may activate this ability" (CR 602.2, 602.1b) is the
//     row's AnyPlayer bit. "You" is whoever activated it (CR 109.5,
//     2018-07-13 ruling: the player who activated it draws), and the cost
//     is theirs to pay (CR 602.1a). Xantcha's controller may activate it
//     too. "Xantcha's controller" is read at resolution off the
//     permanent, as it last existed if it has left (CR 608.2h).
//
// The life loss comes before the draw, in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f0f3712-8d13-41a5-b332-2ab34e48d79d",
		Name:         "Xantcha, Sleeper Agent",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersUnderTheControlOfAnOpponentOfYourChoice("Xantcha, Sleeper Agent", game.ControlForHarm),
		},
		Static: []game.StaticAbility{
			AttacksEachCombat(),
			CantAttackItsOwnerOrItsOwnersPlaneswalkers(),
		},
		Activated: []ActivatedAbility{{
			Label:     "{3}: Xantcha's controller loses 2 life and you draw a card. Any player may activate this ability.",
			Cost:      game.AbilityCost{Mana: "{3}"},
			AnyPlayer: true,
			Purpose:   game.Purpose{Answers: game.AnswerValue, Draws: 1, ControllerLosesLife: 2},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if info, ok := ctx.SourcePermanent(); ok {
					if err := g.ChangePlayerLifeForEffect(ctx.Source(), info.Controller, -2); err != nil {
						return err
					}
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			},
		}},
	})
}
