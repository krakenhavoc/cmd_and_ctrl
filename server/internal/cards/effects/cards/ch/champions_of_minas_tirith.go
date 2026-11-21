package cardsch

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Champions of Minas Tirith — Creature — Human Soldier {5}{W}, 4/6 (Tales
// of Middle-earth Commander):
//
//	"When this creature enters, you become the monarch.
//	 At the beginning of combat on each opponent's turn, if you're the
//	 monarch, that opponent may pay {X}, where X is the number of cards in
//	 their hand. If they don't, they can't attack you this combat."
//
// The combat trigger has an intervening "if" (CR 603.4), checked as the
// step begins and again on resolution. X is that opponent's hand size as
// the ability resolves; an empty hand pays {0}, so there is nothing to
// ask and they may attack. The payer is the opponent (PayUnless), and a
// refusal, or mana they cannot find, registers "can't attack you" (you,
// not your planeswalkers, as printed) until the end of this combat phase
// (game.GrantCantAttackPlayerThisCombatForEffect).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "464a930a-79d5-4436-864d-3ce9f29475ae",
		Name:         "Champions of Minas Tirith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Champions of Minas Tirith"),
			effects.InterveningIf(effects.CondMonarch{},
				effects.On(game.EventStepBegan, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Kind == game.EventStepBegan && ev.Step == game.StepBeginCombat &&
						ev.Actor != uuid.Nil && ev.Actor != source.Controller
				}, "Champions of Minas Tirith — that opponent may pay {X}, X their hand size; if they don't, they can't attack you this combat",
					effects.Do(effects.Instruction(s82GapB118ChampionsAsk)))),
		},
	})
}

func s82GapB118ChampionsAsk(ctx *effects.Context) error {
	opp := ctx.Trigger().Event.Actor
	p := ctx.Game.PlayerByIDForEffect(opp)
	if p == nil || p.Eliminated || p.Hand == nil || p.Hand.Size() == 0 {
		return nil
	}
	return effects.PayUnless{
		Chooser:  opp,
		Cost:     fmt.Sprintf("{%d}", p.Hand.Size()),
		Question: "Champions of Minas Tirith — pay {X} (X is the number of cards in your hand), or you can't attack its controller this combat",
		Else:     s82GapB118ChampionsDeclined,
	}.Apply(ctx)
}

var s82GapB118ChampionsDeclined = effects.PayUnlessElse("champions-of-minas-tirith/cant-attack", func(ctx *effects.Context, r game.PayUnlessAnswer) error {
	ctx.Game.GrantCantAttackPlayerThisCombatForEffect(r.Chooser, ctx.Controller(), game.CantAttackScope{PlayerOnly: true},
		"Champions of Minas Tirith — can't attack its controller this combat", r.Source)
	return nil
})
