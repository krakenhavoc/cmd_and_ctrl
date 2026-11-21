package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illuminator Virtuoso — Creature — Human Rogue {1}{W}, 1/1 (Streets of
// New Capenna, #5491):
//
//	"Double strike
//	 Whenever this creature becomes the target of a spell you control, it
//	 connives. (Draw a card, then discard a card. If you discarded a
//	 nonland card, put a +1/+1 counter on this creature.)"
//
// EventBecomesTarget carries no discriminator between spells and abilities,
// so the spell half is read off the stack (b40TargetedByASpell); its Actor is
// the targeting spell's controller, which must be the Virtuoso's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ed2af15e-955b-43a1-80af-2484be373db3",
		Name:            "Illuminator Virtuoso",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventBecomesTarget && ev.CardID == source.InstanceID &&
					ev.Actor == source.Controller && effects.B40TargetedByASpell(ev, g)
			}, "Illuminator Virtuoso — it connives", effects.Do(effects.Instruction(effects.SourceConnives))),
		},
	})
}
