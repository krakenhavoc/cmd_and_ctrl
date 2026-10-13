package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quagnoth — Creature — Beast {5}{G}, 4/5:
//
//	"Split second
//	 Shroud
//	 When a spell or ability an opponent controls causes you to discard
//	 this card, return it to your hand."
//
// Split second and shroud are printed keywords. The discard line is the
// family's "this card" trigger (ADR 0153 §4), watching from the GRAVEYARD
// on a discard an opponent's spell or ability caused (the stack item's
// controller, CR 113.8), returning the card at once when the ability
// resolves. It is the same object it was when it triggered, or nothing
// returns (CR 400.7).
//
// Caveat, declared: if madness or Rest in Peace sends the discarded card
// to exile instead of the graveyard, its ability does not trigger from
// exile (owner decision 6).
func init() {
	Register(Spec{
		OracleID:          "9250eb9e-690f-4563-8132-d737f6d22883",
		Name:              "Quagnoth",
		Completeness:      CompletenessCaveats,
		Caveats:           []string{discardTriggerFromExileCaveat},
		PrintedKeywords:   []string{"split second", "shroud"},
		DiscardByOpponent: game.DiscardByOpponentTrigger,
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventDiscardCard, DiscardedByAnOpponent(true),
				"Quagnoth — discarded: return it to your hand",
				func(g *game.Game, item *game.StackItem) error {
					if sourceIsNewObject(g, item) {
						return nil
					}
					if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					return ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
				})),
		},
	})
}
