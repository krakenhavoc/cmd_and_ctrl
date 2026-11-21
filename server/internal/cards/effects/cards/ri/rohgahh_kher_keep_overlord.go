package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rohgahh, Kher Keep Overlord — Legendary Creature — Kobold Warrior
// {3}{B}{R}, 4/4 (Dominaria United Commander, S82):
//
//	"Other Kobolds you control get +2/+2.
//	 Whenever you cast a Kobold spell, you may pay {2}. If you do, create a
//	 4/4 red Dragon creature token with flying.
//	 Whenever you cast a Dragon spell, create a 0/1 red Kobold creature
//	 token named Kobolds of Kher Keep."
//
// The anthem is the shared tribal lord; the two cast triggers are separate
// printed abilities, the first with a may-pay prompt whose paid branch makes
// the Dragon. The Kobold token is named Kobolds of Kher Keep and is a
// Kobold, so it takes the anthem too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a14c692e-dce6-4121-84ec-7fa28ffe4d6b",
		Name:         "Rohgahh, Kher Keep Overlord",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Kobold"}, Others: true, YoursOnly: true}, 2, 2),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.HasSubtype("Kobold"),
				"Rohgahh, Kher Keep Overlord — you may pay {2} to create a 4/4 Dragon with flying",
				rohgahhOffer),
			effects.WheneverYouCast(effects.HasSubtype("Dragon"),
				"Rohgahh, Kher Keep Overlord — create a 0/1 Kobolds of Kher Keep",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 red Kobolds of Kher Keep"), N: 1})),
		},
	})
}

func rohgahhOffer(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{2}",
		Question: "Rohgahh, Kher Keep Overlord — pay {2} to create a 4/4 Dragon with flying?",
		Then:     rohgahhPaid,
	}.Apply(effects.NewContext(g, item))
}

var rohgahhPaid = effects.PayUnlessBranches("pay-unless/rohgahh-kher-keep-overlord-dragon",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Template: effects.TokenCard("4/4 red Dragon with flying"), N: 1}.Apply(ctx)
	}, nil)
