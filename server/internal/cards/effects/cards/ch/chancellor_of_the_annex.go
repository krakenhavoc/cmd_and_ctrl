package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chancellor of the Annex — Creature — Phyrexian Angel {4}{W}{W}{W}, 5/6:
//
//	"You may reveal this card from your opening hand. If you do, when each
//	 opponent casts their first spell of the game, counter that spell unless
//	 that player pays {1}.
//	 Flying
//	 Whenever an opponent casts a spell, counter it unless that player pays
//	 {1}."
//
// The reveal is the Chancellor seam (RevealFromOpeningHand, CR 103.6): one
// delayed trigger per opponent, firing on that opponent's first cast. The
// permanent's own line is the ordinary cast trigger; both ask the caster the
// CR 118.12 question over the spell still on the stack.
func init() {
	const label = "Chancellor of the Annex — counter it unless that player pays {1}"
	effects.Register(effects.Spec{
		OracleID:        "85c0a376-da6f-4f69-853c-a093708d0952",
		Name:            "Chancellor of the Annex",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OpeningHand:     effects.RevealFromOpeningHand(game.RevealAtEachOpponentsFirstSpell, label, chancellorOfTheAnnexFirstSpell),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(nil), label, func(g *game.Game, item *game.StackItem) error {
				return effects.CounterThatSpellUnlessItsCasterPays(effects.NewContext(g, item), "{1}", label)
			}),
		},
	})
}

var chancellorOfTheAnnexFirstSpell = game.SimpleDelayedBody("chancellor-of-the-annex/first-spell-counter", func(g *game.Game, item *game.StackItem) error {
	return effects.CounterThatSpellUnlessItsCasterPays(effects.NewContext(g, item), "{1}",
		"Chancellor of the Annex — counter that spell unless that player pays {1}")
})
