package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jamie McCrimmon — Legendary Creature — Human Warrior {2}{G}, 2/2 (Doctor Who):
//
//	"Trample
//	 Whenever you cast a historic spell, Jamie McCrimmon gets +X/+X until end
//	 of turn, where X is that spell's mana value. (Artifacts, legendaries, and
//	 Sagas are historic.)
//	 Doctor's companion"
//
// The trigger goes on the stack above the spell and resolves first, so X is
// the spell's mana value with its announced X while it is on the stack. Only
// your own casts count. Doctor's companion is a deck-construction rule read
// off the oracle text.
//
// No simplification.
func init() {
	const label = "Jamie McCrimmon — gets +X/+X until end of turn, where X is that spell's mana value"
	effects.Register(effects.Spec{
		OracleID:        "c9f074cb-6dcb-4c25-8332-9b15e4881d33",
		Name:            "Jamie McCrimmon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.B14Historic(), label, func(g *game.Game, item *game.StackItem) error {
				x := effects.TriggeringSpellManaValue(g, item)
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Toughness: x, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
