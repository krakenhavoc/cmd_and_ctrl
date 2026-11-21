package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Securitron Squadron — Artifact Creature — Robot {1}{W}, 2/2
// (Fallout, #5469):
//
//	"Squad {3} (As an additional cost to cast this spell, you may pay
//	 {3} any number of times. When this creature enters, create that
//	 many tokens that are copies of it.)
//	 Vigilance
//	 Whenever a creature token you control enters, put a +1/+1 counter
//	 on it."
//
// Imperial Hullbreachers' squad (SquadMana, Squad: CR 702.157a, the
// engine's cap of 20 payments on one announcement, multikicker's
// reason). The counter goes on the token that entered, The Great
// Henge's "put a +1/+1 counter on it", read off the trigger's event;
// a token that has left by then gets nothing. The Squadron's own token
// copies are creature tokens, so each Squadron on the battlefield
// counts them as they enter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "96312b75-3110-4506-9f95-0c6c17d9da11",
		Name:            "Securitron Squadron",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		OptionalCosts:   []game.AdditionalCost{effects.SquadMana("{3}", 20)},
		Triggered: []game.TriggeredAbility{
			effects.Squad("Securitron Squadron"),
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && effects.IsToken(c)
			}, "Securitron Squadron — put a +1/+1 counter on the creature token", effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				entered := ctx.Item.Trigger.Event.CardID
				if !effects.OnBattlefield(ctx.Game, entered) {
					return nil
				}
				return effects.AddCounter{Target: entered, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}))),
		},
	})
}
