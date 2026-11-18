package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a14c692e-dce6-4121-84ec-7fa28ffe4d6b",
		Name:         "Rohgahh, Kher Keep Overlord",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Kobold"}, Others: true, YoursOnly: true}, 2, 2),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(HasSubtype("Kobold"),
				"Rohgahh, Kher Keep Overlord — you may pay {2} to create a 4/4 Dragon with flying",
				rohgahhOffer),
			WheneverYouCast(HasSubtype("Dragon"),
				"Rohgahh, Kher Keep Overlord — create a 0/1 Kobolds of Kher Keep",
				Do(CreateToken{Template: TokenCard("0/1 red Kobolds of Kher Keep"), N: 1})),
		},
	})
}

func rohgahhOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{2}",
		Question: "Rohgahh, Kher Keep Overlord — pay {2} to create a 4/4 Dragon with flying?",
		Then:     rohgahhPaid,
	}.Apply(NewContext(g, item))
}

var rohgahhPaid = PayUnlessBranches("pay-unless/rohgahh-kher-keep-overlord-dragon",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CreateToken{Template: TokenCard("4/4 red Dragon with flying"), N: 1}.Apply(ctx)
	}, nil)
