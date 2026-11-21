package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Black Sun's Twilight — Instant {X}{B} (Phyrexia: All Will Be One):
//
//	"Up to one target creature gets -X/-X until end of turn. If X is 5 or
//	 more, return a creature card with mana value X or less from your
//	 graveyard to the battlefield tapped."
//
// The shrink is Blue Sun's Zenith-style X read from the announced cost
// (ctx.X()), applied to the one optional target if it is still legal
// (CR 608.2b). The return is a second sentence and does not depend on the
// target: with X of 5 or more it runs even when no target was chosen or the
// target is gone. Nothing is targeted by it: the controller picks on
// resolution among the creature cards of their graveyard with mana value X or
// less (a card with {X} in its cost has X = 0 in the graveyard), which are
// read after the -X/-X is applied (no creature has died yet; deaths are a
// state-based action after the spell finishes). It enters tapped under the
// caster's control. The choice is "a creature card", so it is mandatory
// when one exists.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af105cda-1f7c-4868-9407-a97c46bd762d",
		Name:         "Black Sun's Twilight",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("up to one target creature").WithCount(0, 1),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				x := ctx.X()
				return effects.BoostUntilEOT{
					Target:    t.ID,
					Power:     -x,
					Toughness: -x,
					Label:     "Black Sun's Twilight — -X/-X until end of turn",
				}.Apply(ctx)
			}),
			effects.Instruction(blackSunsTwilightReturn),
		),
	})
}

// blackSunsTwilightReturn is "If X is 5 or more, return a creature card with
// mana value X or less from your graveyard to the battlefield tapped": the
// candidates are read now and the pick is asked on resolution.
func blackSunsTwilightReturn(ctx *effects.Context) error {
	x := ctx.X()
	if x < 5 {
		return nil
	}
	you := ctx.Controller()
	candidates := effects.GraveyardCardIDs(ctx, you, func(c game.Card) bool {
		return c.IsCreature() && !c.IsToken() && c.ManaValue() <= x
	})
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    you,
		FromPlayer: you,
		Source:     ctx.Source(),
		Question:   "Black Sun's Twilight — return a creature card with mana value X or less from your graveyard to the battlefield tapped",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		ThenKey:    blackSunsTwilightReturnPick,
		LiveZone:   game.ZoneGraveyard,
	})
	return nil
}

// blackSunsTwilightReturnPick puts the chosen card onto the battlefield
// tapped under its chooser's control. An on-disk key: never renamed.
var blackSunsTwilightReturnPick = game.RegisterChooseCardsThen("choose-cards/black-suns-twilight-return",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return nil
		}
		return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(r.Picked[0], r.Chooser, true, game.KeyedTail{})
	})
