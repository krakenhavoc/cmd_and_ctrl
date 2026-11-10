package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voracious Bibliophile — Creature — Dragon {3}{U}, 3/3 (Tarkir: Dragonstorm
// Commander):
//
//	"Flying, vigilance
//	 Whenever you cast a spell with one or more targets, draw that many
//	 cards."
//
// The number of targets is read when the spell is cast and kept on the item
// (Params.Amount, a fill-in Build), because the spell may be countered or
// its targets may leave before the trigger resolves, and "that many" is the
// count the spell had. The cast path announces targets before the cast event
// (CR 601.2c), so the stack item already carries them; a spell with no target
// does not trigger. An Aura spell's enchant target counts (the ruling), and
// a spell that targets one object twice has two targets.
//
// No simplification.
func init() {
	const label = "Voracious Bibliophile — draw a card for each target of the spell"
	t := On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		if !YouCast(nil)(ev, source, lki, g) {
			return false
		}
		item := g.StackItemForEffect(ev.CardID)
		return item != nil && len(item.Targets) > 0
	}, label, func(g *game.Game, item *game.StackItem) error {
		if item.Params.Amount <= 0 {
			return nil
		}
		return DrawCards{N: item.Params.Amount}.Apply(NewContext(g, item))
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		if spell := g.StackItemForEffect(ev.CardID); spell != nil {
			item.Params.Amount = len(spell.Targets)
		}
		return item
	}
	Register(Spec{
		OracleID:        "dcc323d6-ddae-4fd9-9c9e-466b07895d73",
		Name:            "Voracious Bibliophile",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
