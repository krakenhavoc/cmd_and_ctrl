package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Comforting Counsel — Enchantment {1}{G} (Secrets of Strixhaven, #4456):
//
//	"Whenever you gain life, put a growth counter on this enchantment.
//	 As long as there are five or more growth counters on this
//	 enchantment, creatures you control get +3/+3."
//
// The trigger is one per life-gain event, and each source that gains you
// life is its own event, so lifelink damage from two creatures in one
// combat damage step is two counters.
// The anthem is a layer 7c static that reads the counters on THIS
// enchantment every time the layers are applied, so it switches on the
// moment the fifth counter lands and off if counters are removed.
//
// No simplification.
func init() {
	anthem := effects.B16Anthem(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.B16CreaturesYouControl(target, g, source) && source.Counters["growth"] >= 5
	}, 3, 3)
	anthem.Label = "As long as there are five or more growth counters on this enchantment, creatures you control get +3/+3."
	effects.Register(effects.Spec{
		OracleID:     "5fd193c0-ed9f-4afb-b6cd-c1a20008fc0b",
		Name:         "Comforting Counsel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Comforting Counsel — put a growth counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.AddCounter{Target: item.SourceCardID, Kind: "growth", N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
		Static: []game.StaticAbility{anthem},
	})
}
