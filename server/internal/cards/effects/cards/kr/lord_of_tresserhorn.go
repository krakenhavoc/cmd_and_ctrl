package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lord of Tresserhorn — Legendary Creature — Zombie {1}{U}{B}{R}, 10/4
// (Alliances, #5634):
//
//	"When Lord of Tresserhorn enters, you lose 2 life, you sacrifice two
//	 creatures, and target opponent draws two cards.
//	 {B}: Regenerate Lord of Tresserhorn."
//
// One trigger with a target opponent, its three instructions in printed order
// (CR 608.2c): the controller loses 2 life, then picks two creatures they
// control (the Lord itself among them) to sacrifice, as many as they have if
// fewer, and once those have gone the targeted opponent draws two. The draw
// waits on the sacrifice (a keyed pick continuation), and is skipped if the
// opponent is no longer a legal target.
//
// No simplification.
func init() {
	const name = "Lord of Tresserhorn"
	effects.Register(effects.Spec{
		OracleID:     "283e7a23-ae25-48d1-abb5-cd021f66cc29",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(name+" — you lose 2 life, you sacrifice two creatures, and target opponent draws two cards", oldB814LordEnters),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
		Activated: []effects.ActivatedAbility{effects.RegenerateThisRow("{B}")},
	})
}

func oldB814LordEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := g.ChangePlayerLifeForEffect(item.SourceCardID, ctx.Controller(), -2); err != nil {
		return err
	}
	creatures := effects.PermanentsControlledByMatching(g, ctx.Controller(), effects.Creature())
	if len(creatures) == 0 {
		return oldB814LordDraw(ctx)
	}
	n := min(2, len(creatures))
	return effects.ChoosePermanents{
		Question: "Lord of Tresserhorn — sacrifice two creatures",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			return effects.PermanentsControlledByMatching(g, of, effects.Creature()), n, n
		},
		Keyed: oldB814LordSacrifice,
	}.Apply(ctx)
}

// oldB814LordSacrifice sacrifices the pair, then the targeted opponent draws.
// An on-disk key.
var oldB814LordSacrifice = effects.RegisterPermanentsPick("old-b814/lord-of-tresserhorn-sacrifice-two", effects.PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *effects.Context, _ game.PromptedPicks) error {
		return oldB814LordDraw(ctx)
	},
})

// oldB814LordDraw is "target opponent draws two cards", if they are still a
// legal target.
func oldB814LordDraw(ctx *effects.Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return effects.DrawCards{Player: t.ID, N: 2}.Apply(ctx)
		}
	}
	return nil
}
