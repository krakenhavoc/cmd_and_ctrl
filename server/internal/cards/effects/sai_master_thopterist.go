package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sai, Master Thopterist — Legendary Creature — Human Artificer
// {2}{U}, 1/4 (EDHREC rank 820):
//
//	"Whenever you cast an artifact spell, create a 1/1 colorless
//	 Thopter artifact creature token with flying.
//	 {1}{U}, Sacrifice two artifacts: Draw a card."
//
// The artifact deck's token engine: every rock, every Treasure spell,
// every equipment is a flier. The cast trigger is Beast Whisperer's
// shape; the Thopter is itself an artifact, so it feeds the
// artifact-ETB payoffs (Reckless Fireweaver) as printed.
//
// Sandbox simplification, declared: the draw ability is NOT
// registered. "Sacrifice two artifacts" is a sacrifice cost with a
// count of two, and validateSacrificeCostLocked accepts exactly one
// permanent for a SacrificeOther clause — there is no two-card
// sacrifice cost in the engine yet. Omitting the ability is the
// weaker direction; inventing a one-artifact cost would be stronger
// than printed. The Thopter engine, which is the card, is whole.
func init() {
	Register(Spec{
		OracleID:     "52241b9c-7a69-4176-9234-8bdab09d8e64",
		Name:         "Sai, Master Thopterist",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The draw ability isn't implemented — only the Thopter-making trigger works."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.IsArtifact()
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Sai, Master Thopterist — create a Thopter",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   ThopterToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
