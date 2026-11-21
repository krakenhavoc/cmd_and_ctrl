package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venomous Brutalizer — Creature — Phyrexian Knight {2}{G}{G}, 4/4 (Phyrexia:
// All Will Be One, S82):
//
//	"Toxic 3
//	 When this creature enters, you may pay {1}{G}. If you do, proliferate."
//
// Toxic rides PrintedKeywords. {1}{G} is MayPay's optional payment as the
// trigger resolves; the proliferate is its paid branch, so a decline or a
// pool that cannot pay proliferates nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "766c3528-2c5c-4e92-aa85-b1e0e115d491",
		Name:            "Venomous Brutalizer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"toxic 3"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Venomous Brutalizer — you may pay {1}{G}; if you do, proliferate",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{G}",
						Question: "Venomous Brutalizer — pay {1}{G} to proliferate?",
						Then:     effects.S82OneB06ProliferatePaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
