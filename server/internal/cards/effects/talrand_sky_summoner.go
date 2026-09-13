package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Talrand, Sky Summoner — Legendary Creature — Merfolk Wizard {2}{U}{U},
// 2/2 (EDHREC rank 785):
//
//	"Whenever you cast an instant or sorcery spell, create a 2/2 blue
//	 Drake creature token with flying."
//
// The spellslinger commander: every cantrip is a flier. Storm-Kiln
// Artist's cast trigger with a different payout — the spell's type is
// read off the stack when the cast event fires, and the Drake arrives
// when the trigger resolves, before the spell that made it (LIFO), as
// in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ea1eb902-a23c-44ff-9169-19baf71de238",
		Name:         "Talrand, Sky Summoner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (spell.IsInstant() || spell.IsSorcery())
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Talrand, Sky Summoner — create a 2/2 Drake with flying",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   b06BlueDrakeToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
