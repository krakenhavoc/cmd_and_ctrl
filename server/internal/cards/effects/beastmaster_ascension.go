package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beastmaster Ascension — Enchantment {2}{G} (EDHREC rank 561):
//
//	"Whenever a creature you control attacks, you may put a quest
//	 counter on this enchantment.
//	 As long as this enchantment has seven or more quest counters on
//	 it, creatures you control get +5/+5."
//
// Two turns of attacking with a wide board, then Overrun forever.
// The counter trigger fires once per attacking creature (the text is
// "a creature", not "one or more") and is optional, so each one is a
// yes/no prompt — tedious across a seven-creature attack, and
// exactly the printed card. The anthem is an ordinary Layer 7c static
// whose AppliesTo reads the enchantment's own quest counters, and the
// layer engine recomputes on every counter change, so the seventh
// counter switches it on mid-combat.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11b5308d-5bc0-4782-875f-a28be36e665d",
		Name:         "Beastmaster Ascension",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Beastmaster Ascension — put a quest counter on it?",
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Beastmaster Ascension — put a quest counter",
					func(g *game.Game, item *game.StackItem) error {
						return AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					source.Counters["quest"] >= 7
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 5
				c.Toughness += 5
			},
		}},
	})
}
