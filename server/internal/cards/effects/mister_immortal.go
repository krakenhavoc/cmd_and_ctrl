package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mister Immortal — Legendary Creature — Mutant Hero {1}{G}, 2/1 (Marvel
// Super Heroes Commander, #4455):
//
//	"{2}{G}: Return this card from your graveyard or from exile to the
//	 battlefield tapped."
//
// One ability that functions from two zones (CR 113.6, Zones): the
// Chronosavant / Reassembling Skeleton return when the card is in a
// graveyard, and the exile return (ReturnFromExile, the card's new object
// enters tapped under its owner's control) when it is in exile. The zone is
// read as the ability resolves, so a card that left both zones in response
// is not returned. The battlefield is not one of its zones: on the
// battlefield the ability does not exist.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f016ad56-4f73-4091-a1c2-937782642ecb",
		Name:         "Mister Immortal",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}: Return this card from your graveyard or from exile to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{2}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard, game.ZoneExile},
			Effect:  misterImmortalReturn,
		}},
	})
}

func misterImmortalReturn(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil {
		return nil
	}
	switch z.Kind {
	case game.ZoneGraveyard:
		return returnThisFromGraveyardTapped(g, item)
	case game.ZoneExile:
		return ReturnFromExile{Target: item.SourceCardID, Controller: item.Controller, Tapped: true}.Apply(NewContext(g, item))
	}
	return nil
}
