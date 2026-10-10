package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eldrazi Ravager — Creature — Eldrazi {5}{C}, 6/6:
//
//	"Annihilator 1
//	 Sacrifice two Eldrazi: Return this card from your graveyard to
//	 your hand.
//	 Cycling {2}"
//
// Annihilator 1 is the canonical keyword token (ADR 0113 §2). The
// return is an ability of the card in the graveyard: its cost is two
// Eldrazi permanents you control, sacrificed together, and a card that
// left the graveyard before it resolves stays where it is. Cycling is
// the shared constructor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "25a73cfc-a40d-47ea-a7b1-61c7b62a9e6c",
		Name:            "Eldrazi Ravager",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"annihilator 1"},
		Activated: []ActivatedAbility{
			{
				Label:   "Sacrifice two Eldrazi: Return this card from your graveyard to your hand.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    SacrificeN(2, "two Eldrazi", HasSubtype("Eldrazi")),
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Effect:  returnThisCardFromYourGraveyardToYourHand,
			},
			Cycling("{2}"),
		},
	})
}

// returnThisCardFromYourGraveyardToYourHand is a graveyard ability's
// "Return this card from your graveyard to your hand." The card may
// have left the graveyard between activation and resolution, and then
// nothing happens.
func returnThisCardFromYourGraveyardToYourHand(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}
