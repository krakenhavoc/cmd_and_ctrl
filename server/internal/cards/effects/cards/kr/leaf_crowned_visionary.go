package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leaf-Crowned Visionary — Creature — Elf Druid {G}{G}, 1/1 (EDHREC
// rank 1913):
//
//	"Other Elves you control get +1/+1.
//	 Whenever you cast an Elf spell, you may pay {G}. If you do, draw
//	 a card."
//
// The Elf lord that draws. The anthem is TribalAnthem over "other
// Elves you control" — both words printed, so this one buffs only
// your side, unlike Elvish Champion. The draw is a MayPay prompt on
// the cast trigger: the spell is read off the stack for its Elf
// subtype (a changeling Elf counts), the controller is asked for
// {G}, and the draw rides the "yes".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef693b9f-11e3-49bf-8387-b8f480b9007a",
		Name:         "Leaf-Crowned Visionary",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Elf"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B17ElfSpellCastByYou(ev, source, g)
			}, "Leaf-Crowned Visionary — pay {G} to draw a card", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{G}",
					Question: "Leaf-Crowned Visionary — pay {G} to draw a card?",
					Then:     effects.MayPayDrawACard,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
