package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slaughter-Priest of Mogis — Creature — Minotaur Shaman {B}{R}, 2/2 (Theros
// Beyond Death, S82 OLD, #5514):
//
//	"Whenever you sacrifice a permanent, this creature gets +2/+0 until end
//	 of turn.
//	 {2}, Sacrifice another creature or an enchantment: This creature gains
//	 first strike until end of turn."
//
// The trigger is Rose, Cutthroat Raider's sacrifice watcher, one trigger for
// each permanent its controller sacrifices (the cost of the second ability
// included, and a token too), pumping the Priest. The activated ability's cost
// is a mana payment and a sacrifice of another creature or an enchantment
// (Broadside Bombardiers' "another" shape), paid before the ability goes on
// the stack, so a response cannot undo it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a8c1002-e237-42d8-b152-aefecab283d0",
		Name:         "Slaughter-Priest of Mogis",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			}, "Slaughter-Priest of Mogis — this creature gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Slaughter-Priest of Mogis — +2/+0 until end of turn")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice another creature or an enchantment: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost: effects.Plus(effects.ManaCost("{2}"),
				effects.SacrificeAnotherN(1, "another creature or an enchantment", effects.Or(effects.Creature(), effects.Enchantment()))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"first strike"},
					Label: "Slaughter-Priest of Mogis — first strike"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
