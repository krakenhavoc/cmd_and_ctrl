package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Traxos, Scourge of Kroog —
//
// "Trample Traxos enters tapped and doesn't untap during your untap step.
// Whenever you cast a historic spell, untap Traxos. (Artifacts, legendaries,
// and Sagas are historic.)"
func init() {
	effects.Register(effects.Spec{
		OracleID:              "c1c78144-b335-4d22-a668-9173ab6a0d04",
		Name:                  "Traxos, Scourge of Kroog",
		Completeness:          effects.CompletenessFull,
		PrintedKeywords:       []string{"trample"},
		Replacements:          []game.ReplacementEffect{effects.SelfEntersTapped()},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return effects.B17HistoricSpellCastByYou(ev, source, g)
		}, "Traxos — untap", func(g *game.Game, item *game.StackItem) error {
			return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
		})},
	})
}
