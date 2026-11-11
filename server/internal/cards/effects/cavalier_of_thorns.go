package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cavalier of Thorns — Creature — Elemental Knight {2}{G}{G}{G}, 5/6 (Lorwyn
// Eclipsed Commander, #5368):
//
//	"Reach
//	 When this creature enters, reveal the top five cards of your library. Put
//	 a land card from among them onto the battlefield and the rest into your
//	 graveyard.
//	 When this creature dies, you may exile it. If you do, put another target
//	 card from your graveyard on top of your library."
//
// The enters trigger is Genesis Wave's put with a land filter: one land card
// (not optional), the rest put into the graveyard. The dies trigger asks "you
// may" as its own yes/no, targets another card in the graveyard, and exiles the
// Cavalier first: a Cavalier that has left the graveyard in response cannot be
// exiled, so "if you do" fails and nothing is put on top (Greenwarden of
// Murasa's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "96b909ab-55f1-4cde-a675-2504de3da772",
		Name:            "Cavalier of Thorns",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Purpose:         game.Purpose{Lands: 1, LandsUntapped: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cavalier of Thorns — reveal the top five cards, put a land card from among them onto the battlefield and the rest into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					revealed := g.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 5,
						"Cavalier of Thorns — revealed from the top of the library")
					return PutFromLibraryOntoBattlefield{
						Player:   ctx.Controller(),
						Cards:    revealed,
						Match:    Land(),
						Max:      1,
						Label:    "Cavalier of Thorns — put a land card onto the battlefield",
						ThenTail: PutRestIntoGraveyardTail,
					}.Apply(ctx)
				}),
			func() game.TriggeredAbility {
				t := Targeting(
					WhenThisDies("Cavalier of Thorns — exile it, then put another target card from your graveyard on top of your library", cmdr2B15CavalierOfThornsExile),
					Another(TargetCardInGraveyard("another target card from your graveyard", YouOwn())))
				t.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Cavalier of Thorns — exile it to put another card from your graveyard on top of your library?"}
				return t
			}(),
		},
	})
}

// cmdr2B15CavalierOfThornsExile exiles the Cavalier from the graveyard and, if
// that happened, tucks the chosen card.
func cmdr2B15CavalierOfThornsExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ExileTarget{
		Target:   item.SourceCardID,
		ThenTail: ctx.ItemTail(cmdr2B15CavalierExiled),
	}.Apply(ctx)
}

// cmdr2B15CavalierExiled is the clause after the exile: only when it was
// exiled is the target put on top. An on-disk key.
var cmdr2B15CavalierExiled = CardTail("cmdr2-b15/cavalier-of-thorns-exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	return PutChosenTargetOnTopOfLibrary(ctx.Game, ctx.Item)
})
