package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const (
	legolasUntapLabel = "Legolas, Counter of Kills — you may untap Legolas"
	legolasUntapDone  = "Legolas, Counter of Kills — untapped this turn"
)

// Legolas, Counter of Kills — Legendary Creature — Elf Archer {2}{G}{U},
// 2/3 (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Reach
//	 Whenever you scry, if Legolas is tapped, you may untap it. Do this
//	 only once each turn.
//	 Whenever a creature an opponent controls dies, put a +1/+1 counter on
//	 Legolas."
//
// Reach is a printed keyword. The scry trigger is Ancient Cornucopia's
// once-each-turn "may": its "if" is an intervening one (Legolas tapped as
// the scry finishes and again as the ability resolves, CR 603.4), and a
// "yes" is what uses up the turn (Calix's per-object done mark), so a
// "no" leaves the next scry free to ask again. The counter trigger is
// Body Count's died-under-an-opponent's-control watch.
//
// No simplification.
func init() {
	untap := effects.InterveningIf(effects.AvatarB5CondThisIsTapped{}, effects.Optional(
		effects.On(game.EventScry, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return ev.Actor == source.Controller && source.Tapped && !effects.B28DoneThisTurn(g, source.InstanceID, legolasUntapDone)
		}, legolasUntapLabel, func(g *game.Game, item *game.StackItem) error {
			if effects.B28DoneThisTurn(g, item.SourceCardID, legolasUntapDone) {
				return nil
			}
			effects.B28NoteDoneThisTurn(g, item.SourceCardID, legolasUntapDone)
			return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
		}),
		"Legolas, Counter of Kills — untap Legolas?"))
	effects.Register(effects.Spec{
		OracleID:        "ddc12487-f577-4def-acdb-491f7ba6dc51",
		Name:            "Legolas, Counter of Kills",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			untap,
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B18OpponentsCreatureDied(ev, source, g)
			}, "Legolas, Counter of Kills — put a +1/+1 counter on Legolas", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
