package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Essence of Orthodoxy — Creature — Phyrexian {3}{W}{W}, 3/3 (March of
// the Machine):
//
//	"Flying
//	 Whenever this creature or another Phyrexian you control enters,
//	 incubate 2."
//
// One ability with two conditions: this creature entering, or another
// Phyrexian entering under its controller's control. The Phyrexian Token
// face an Incubator turns into does not trigger it, because a transform
// is not an entry (CR 712.18) and the Incubator token itself is not a
// Phyrexian.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "51329e73-aebe-4f42-a4db-e092b1984a4b",
		Name:            "Essence of Orthodoxy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				// "Phyrexian" is a word of the source's text (CR 612.2, #3984).
				return ok && (ev.CardID == source.InstanceID || c.HasSubtype(source.TextSubtype("Phyrexian")))
			}, "Essence of Orthodoxy — incubate 2", effects.Do(effects.Incubate{N: 2})),
		},
	})
}
