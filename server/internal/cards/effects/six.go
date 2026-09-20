package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Six — Legendary Creature — Treefolk {2}{G}, 2/4:
//
//	"Reach
//	 Whenever Six attacks, mill three cards. You may put a land card
//	 from among them into your hand.
//	 During your turn, nonland permanent cards in your graveyard have
//	 retrace."
//
// The first two lines are the card that gets attacked with: a
// three-mana 2/4 reach body that turns every attack into a land and
// three cards of graveyard fuel. Reach rides PrintedKeywords; the
// attack trigger is the shared mill-then-take shape
// (mill_then_take.go), filtered to lands.
//
// The mill is NOT optional — "mill three cards" is flat — and only
// the pick is a "may", which is why the trigger has no
// OptionalPrompt and the choose-cards prompt floors at zero.
//
// DECLARED SIMPLIFICATION, weaker than printed: the third line is
// dropped. Retrace is unimplemented (seam #652, the same one that
// keeps Wrenn and Six's ultimate off the board) and it needs two
// things that do not exist — the alternative cost itself, "cast this
// from your graveyard by discarding a land card in addition to
// paying its other costs", and a granted cast permission one player
// holds over a SET of cards nobody printed it on. Six's clause adds
// a third that the emblem does not: the grant is live only during
// its controller's turn, which is a duration the permission model
// has no shape for either. Registering it would put a line on the
// card the engine never honours; leaving it out costs the card its
// recursion and nothing else. The day #652 lands, this is one static
// and a caveat deletion.
func init() {
	Register(Spec{
		OracleID:     "dbcbdf37-c40f-4068-b4a7-a849cab1056c",
		Name:         "Six",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Nonland permanent cards in your graveyard don't have retrace — retrace isn't implemented yet.",
		},
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Six — mill three cards", sixAttackMill),
		},
	})
}

// sixAttackMill is the attack trigger: mill three, then offer a land
// from among them.
func sixAttackMill(g *game.Game, item *game.StackItem) error {
	return MillToZone{
		N: 3,
		Then: func(ctx *Context, milled []uuid.UUID) error {
			return mayTakeOneFromAmongThem(ctx, milled, Land(),
				"Six — you may put a land card from among them into your hand")
		},
	}.Apply(NewContext(g, item))
}
