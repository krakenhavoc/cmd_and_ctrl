package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawnhand Eulogist — {3}{B} Creature — Elf Warlock 3/3 (Lorwyn Eclipsed,
// #4458):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 When this creature enters, mill three cards. Then if there is an Elf
//	 card in your graveyard, each opponent loses 2 life and you gain 2
//	 life. (To mill three cards, put the top three cards of your library
//	 into your graveyard.)"
//
// The mill is Cantankerous Keepers': the clause after it runs once the
// cards have landed (MillToZone.ThenTail, so a milled commander's
// command-zone question is answered first). The "if" is not an
// intervening one (CR 603.4): the trigger always goes on the stack, and
// the graveyard is read after the mill, every Elf card in it counting, not
// only the three just milled. A changeling card is an Elf card. The drain
// is Sneering Shadewriter's, a life change with no damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "57bbd80b-290f-461e-8dc6-26ba4d673fa9",
		Name:            "Dawnhand Eulogist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dawnhand Eulogist — mill three cards, then if there is an Elf card in your graveyard, each opponent loses 2 life and you gain 2 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 3, ThenTail: ctx.ItemTail(dawnhandEulogistMilled)}.Apply(ctx)
				}),
		},
	})
}

// dawnhandEulogistMilled is the clause after the mill: the drain, when an
// Elf card is in the controller's graveyard. An on-disk key.
var dawnhandEulogistMilled = CardTail("dawnhand-eulogist/after-mill", func(ctx *Context, _ game.TailOutcome) error {
	if !dawnhandEulogistElfInGraveyard(ctx) {
		return nil
	}
	if err := (EachOpponentLosesLife{N: 2}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Amount: 2}.Apply(ctx)
})

// dawnhandEulogistElfInGraveyard reports whether the controller's
// graveyard holds an Elf card, the word read through the Eulogist's text
// (CR 612.2).
func dawnhandEulogistElfInGraveyard(ctx *Context) bool {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil {
		return false
	}
	elf := ctx.TextSubtype("Elf")
	for _, c := range p.Graveyard.Cards {
		if c.HasSubtype(elf) {
			return true
		}
	}
	return false
}
