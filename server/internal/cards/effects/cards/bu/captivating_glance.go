package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Captivating Glance — Enchantment — Aura {2}{U} (Lorwyn, #5593):
//
//	"Enchant creature
//	 At the beginning of your end step, clash with an opponent. If you win,
//	 gain control of enchanted creature. Otherwise, that player gains
//	 control of enchanted creature. (Each clashing player reveals the top
//	 card of their library, then puts that card on their choice of the top
//	 or bottom. A player wins if their card had a greater mana value.)"
//
// The enchanted creature is read as the clash finishes: the Aura must still
// be attached. "That player" is the opponent the clash was with, which the
// clash's tail is told. Winning gives control to the Aura's controller,
// anything else (a loss, a tie) to that opponent, and with no opponent
// chosen nothing changes. Neither has a duration (CR 611.2a), so control
// stays after the Aura goes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11a372c7-3c7c-431c-9532-f2bae93f1c33",
		Name:         "Captivating Glance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Captivating Glance — clash with an opponent; the winner of the clash gains control of enchanted creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ClashWithAnOpponent{Then: ctx.ItemTail(captivatingGlanceClashed)}.Apply(ctx)
				}),
		},
	})
}

// captivatingGlanceClashed is "If you win, gain control of enchanted
// creature. Otherwise, that player gains control of it". An on-disk key.
var captivatingGlanceClashed = effects.CardTail("then/captivating-glance-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	host := effects.EnchantedCreatureOf(ctx)
	if host == uuid.Nil {
		return nil
	}
	taker := ctx.Controller()
	if !effects.WonTheClash(o) {
		taker = effects.ClashOpponent(o)
	}
	if taker == uuid.Nil {
		return nil
	}
	return effects.GainControl{Target: host, Controller: taker, Duration: game.IndefiniteDuration(), Label: "Captivating Glance — gains control"}.Apply(ctx)
})
