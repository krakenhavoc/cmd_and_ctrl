package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Patient Naturalist — Creature — Human Scout {2}{G}, 2/3 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature enters, mill three cards. Put a land card from
//	 among the milled cards into your hand. If you can't, create a
//	 Treasure token. (To mill three cards, put the top three cards of
//	 your library into your graveyard.)"
//
// Ainok Wayfarer's mill and pick, made mandatory: with a land card among
// the cards that reached your graveyard you must take one of them (one
// prompt, exactly one card), and with none, or with every milled land
// gone from the graveyard by then, you create a Treasure instead.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba7c1062-f777-4ba9-84df-ec340a9ec514",
		Name:         "Patient Naturalist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Patient Naturalist — mill three cards; put a land card from among them into your hand, or create a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 3, ThenTail: ctx.ItemTail(patientNaturalistMilled)}.Apply(ctx)
				}),
		},
	})
}

var patientNaturalistMilled = CardTail("patient-naturalist/after-mill", func(ctx *Context, o game.TailOutcome) error {
	player := ctx.Controller()
	var lands []uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !Land()(ctx.Game, player, c) {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		lands = append(lands, id)
	}
	if len(lands) == 0 {
		return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Patient Naturalist — put a land card from among the milled cards into your hand",
		Cards:    lands,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  patientNaturalistPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

var patientNaturalistPick = askedPick("choose-cards/patient-naturalist-land", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		// The land left the graveyard while the prompt was open.
		return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
	}
	return ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}.Apply(ctx)
})
