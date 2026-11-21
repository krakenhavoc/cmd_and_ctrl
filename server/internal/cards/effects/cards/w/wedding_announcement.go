package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wedding Announcement // Wedding Festivity — {2}{W} Enchantment //
// Enchantment (Innistrad: Crimson Vow, #5495):
//
//	Front: "At the beginning of your end step, put an invitation counter on
//	        this enchantment. If you attacked with two or more creatures
//	        this turn, draw a card. Otherwise, create a 1/1 white Human
//	        creature token. Then if this enchantment has three or more
//	        invitation counters on it, transform it."
//	Back:  "Creatures you control get +1/+1."
//
// One trigger, one resolution in the printed order: the counter, then the
// attack-count branch (Minas Tirith's reader of this turn's attack
// declarations), then the counter count read as it is now. A counter put
// on by anything else counts too. TransformThis is a no-op if the
// enchantment has already turned over since the trigger went on the stack
// (CR 701.27f).
//
// No simplification.
func init() {
	const oracle = "259ac30c-cc05-4c04-9b23-71283f84b808"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Wedding Announcement",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Wedding Announcement — put an invitation counter; draw a card if you attacked with two or more creatures, otherwise create a 1/1 Human; then transform at three counters",
				weddingAnnouncementEndStep),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Wedding Festivity",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller
		}, 1, 1)},
	})
}

func weddingAnnouncementEndStep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.AddCounter{Target: ctx.Source(), Kind: "invitation", N: 1}).Apply(ctx); err != nil {
		return err
	}
	if effects.MinasTirithAttackedWithTwo(g, item.Controller, uuid.Nil) {
		if err := (effects.DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
			return err
		}
	} else if err := (effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Human"), N: 1}).Apply(ctx); err != nil {
		return err
	}
	if c, ok := g.LookupCardForEffect(ctx.Source()); ok && c.Counters["invitation"] >= 3 {
		return effects.TransformThis{}.Apply(ctx)
	}
	return nil
}
