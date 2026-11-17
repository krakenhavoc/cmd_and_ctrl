package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fleshless Gladiator — Creature — Phyrexian Skeleton {1}{B}, 2/2:
//
//	"Corrupted — {2}{B}: Return this card from your graveyard to the
//	 battlefield tapped. You lose 1 life. Activate only if an opponent has
//	 three or more poison counters."
//
// Activated only from the graveyard (CR 113.6), Reassembling Skeleton's
// shape. If the card has left the graveyard by resolution, nothing happens
// (CR 400.7), and the life is lost only if it came back: the card is its
// own payload, and an ability that finds it gone does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e33d6e5-30a2-48b4-826c-09aa75bbba27",
		Name:         "Fleshless Gladiator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "Corrupted — {2}{B}: Return this card from your graveyard to the battlefield tapped. You lose 1 life. Activate only if an opponent has three or more poison counters.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      ManaCost("{2}{B}"),
			Zones:     []game.ZoneKind{game.ZoneGraveyard},
			Condition: corruptedAnd(nil),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
					return nil
				}
				ctx := NewContext(g, item)
				if err := (ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneBattlefield, Tapped: true}).Apply(ctx); err != nil {
					return err
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
			},
		}},
	})
}
