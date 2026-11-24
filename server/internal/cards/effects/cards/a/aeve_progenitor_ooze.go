package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aeve, Progenitor Ooze — Legendary Creature — Ooze {2}{G}{G}{G}, 2/2
// (Modern Horizons 2, #5501):
//
//	"Storm (When you cast this spell, copy it for each spell cast before it
//	 this turn. Copies become tokens.)
//	 Aeve isn't legendary if it's a token.
//	 Aeve enters with a +1/+1 counter on it for each other Ooze you
//	 control."
//
// Storm is the keyword (CR 702.40a); its copies are tokens, and a token
// Aeve loses the legendary supertype in layer 4 (so the copies do not trip
// the legend rule). The entry counters count the other Oozes you control as
// the permanent enters (CR 614.1c); a token copy counts them too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44600a74-9584-488b-a342-2fccd8fa7599",
		Name:         "Aeve, Progenitor Ooze",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Storm()},
		Static: []game.StaticAbility{{
			Label:     "Aeve isn't legendary if it's a token.",
			Layer:     game.Layer4Type,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, card *game.Card, _ *game.Game, _ *game.Card) {
				if card == nil || !card.IsToken() {
					return
				}
				kept := c.Supertypes[:0:0]
				for _, s := range c.Supertypes {
					if s != "Legendary" {
						kept = append(kept, s)
					}
				}
				c.Supertypes = kept
			},
		}},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, s82OldB38OtherOozesYouControl,
				"Aeve enters with a +1/+1 counter on it for each other Ooze you control"),
		},
	})
}

func s82OldB38OtherOozesYouControl(g *game.Game, src *game.Card) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == src.Controller && c.InstanceID != src.InstanceID && c.HasSubtype(src.TextSubtype("Ooze")) {
			n++
		}
	}
	return n
}
