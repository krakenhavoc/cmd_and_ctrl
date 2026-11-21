package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Balrog of Moria — Legendary Creature — Avatar Demon {4}{B}{B}{R}, 8/8
// (Tales of Middle-earth Commander, #5948):
//
//	"Trample, haste
//	 When The Balrog of Moria dies, you may exile it. When you do, for each
//	 opponent, exile up to one target creature that player controls.
//	 Cycling {3}{R} ({3}{R}, Discard this card: Draw a card.)
//	 When you cycle this card, create two Treasure tokens."
//
// The dies trigger is Undead Butler's CR 603.12 reflexive shape
// (exileItFromTheGraveyardThen): the "you may" belongs to the parent, the
// Balrog is exiled out of the graveyard, and only once it really is in exile
// does the reflexive trigger go on the stack, with one "up to one target
// creature that player controls" clause per opponent in seat order, as Havoc
// Eater's are (perOpponentClauses). Each target is re-checked as the reflexive
// trigger resolves (CR 608.2b) and each creature still legal is exiled.
// Cycling is the keyword's own hand ability, and the cycle trigger fires from
// the graveyard, where the card already is (CR 702.29c), Choking Tethers'
// InGraveyard(On(EventCycle, Self, …)).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d73191d8-6f94-4fba-acd2-2d0490e3ac00",
		Name:            "The Balrog of Moria",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Activated:       []ActivatedAbility{Cycling("{3}{R}")},
		Triggered: []game.TriggeredAbility{
			Optional(
				WhenThisDies("The Balrog of Moria — exile it",
					exileItFromTheGraveyardThen("The Balrog of Moria — for each opponent, exile up to one target creature that player controls", balrogOfMoriaExileBody)),
				"The Balrog of Moria — exile it to exile up to one creature each opponent controls?"),
			InGraveyard(On(game.EventCycle, Self,
				"The Balrog of Moria — cycled: create two Treasure tokens",
				Do(CreateToken{Template: TreasureToken(), N: 2}))),
		},
	})
}

// balrogOfMoriaExileBody exiles every creature the reflexive trigger's
// per-opponent clauses chose that is still legal.
var balrogOfMoriaExileBody = game.ReflexiveBody("the-balrog-of-moria/exile",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (ExileTarget{Target: t.ID}).Apply(ctx.asGroupMember()); err != nil {
				return err
			}
		}
		return nil
	}),
	func(_ uuid.UUID, _ game.EffectParams) *game.TargetSpec {
		return perOpponentClauses(game.MaxPlayers-1, "creature", Creature())
	})
