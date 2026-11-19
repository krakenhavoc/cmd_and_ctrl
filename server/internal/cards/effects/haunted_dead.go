package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haunted Dead — Creature — Zombie {3}{B}, 2/2 (Innistrad Remastered,
// S82 GAP, #5948):
//
//	"When this creature enters, create a 1/1 white Spirit creature token
//	 with flying.
//	 {1}{B}, Discard two cards: Return this card from your graveyard to
//	 the battlefield tapped."
//
// The second ability works only from the graveyard (CR 113.6),
// Fleshless Gladiator's shape with a two-card discard in the cost. If
// the card has left the graveyard by resolution nothing happens
// (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03259490-a789-4f96-8c5b-b410425bdd02",
		Name:         "Haunted Dead",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Haunted Dead — create a 1/1 white Spirit creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Discard two cards: Return this card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{B}"), DiscardN(2, "two cards")),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
					return nil
				}
				return ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneBattlefield, Tapped: true}.Apply(NewContext(g, item))
			},
		}},
	})
}
