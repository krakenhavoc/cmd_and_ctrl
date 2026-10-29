package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Selesnya Eulogist — Creature — Centaur Druid {2}{G}, 3/3:
//
//	"{2}{G}: Exile target creature card from a graveyard, then
//	 populate. (Create a token that's a copy of a creature token you
//	 control.)"
//
// A mana-only activated ability (no tap symbol) with a graveyard
// target clause over every graveyard. The populate follows the exile
// as its continuation. If the target card has left the graveyard by
// resolution the ability does nothing at all (CR 608.2b), populate
// included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0335224-9523-47d5-8944-486f2d4a6b8f",
		Name:         "Selesnya Eulogist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}: Exile target creature card from a graveyard, then populate.",
			Cost:    ManaCost("{2}{G}"),
			Targets: TargetCardInGraveyard("target creature card in a graveyard", Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return ExileTarget{
					Target:   item.Targets[0].ID,
					ThenTail: NewContext(g, item).ItemTail(selesnyaEulogistExiled),
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// selesnyaEulogistExiled is the clause after the exile (#4253 slice 9).
var selesnyaEulogistExiled = CardTail("selesnya-eulogist/after-exile", func(ctx *Context, o game.TailOutcome) error {
	c := ctx

	return Populate{}.Apply(c)
})
