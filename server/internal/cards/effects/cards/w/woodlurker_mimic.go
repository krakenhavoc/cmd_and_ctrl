package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Woodlurker Mimic — Creature — Shapeshifter {1}{B/G}, 2/1 (Eventide, S82 OLD
// batch 520, #5590):
//
//	"Whenever you cast a spell that's both black and green, this creature has
//	 base power and toughness 4/5 until end of turn and gains wither until end of
//	 turn. (It deals damage to creatures in the form of -1/-1 counters.)"
//
// A cast trigger for a spell of both colours (a hybrid or multicoloured spell
// is, a mono-coloured one is not), Risen Riptide's layer 7b base-P/T set (so
// counters and pumps still apply on top) and a keyword grant, both until end of
// turn and both for the Mimic itself; it does nothing for a Mimic that has left.
// A second such spell resets nothing worse than the same effect again.
//
// No simplification.
func init() {
	const label = "Woodlurker Mimic — base power and toughness 4/5 and wither until end of turn"
	basePT := effects.S82OldB097ThisHasBasePTUntilEOT(label, 4, 5)
	effects.Register(effects.Spec{
		OracleID:     "5f4ddd52-4940-4557-8531-40c66857d4cd",
		Name:         "Woodlurker Mimic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.And(effects.OfColor("B"), effects.OfColor("G")), label,
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					if err := basePT(g, item); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"wither"}, Label: label}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
