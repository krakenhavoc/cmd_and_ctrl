package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lightstall Inquisitor — Creature — Angel Wizard {W}, 2/1 (Edge of
// Eternities, #4461):
//
//	"Vigilance
//	 When this creature enters, each opponent exiles a card from their
//	 hand and may play that card for as long as it remains exiled. Each
//	 spell cast this way costs {1} more to cast. Each land played this way
//	 enters tapped."
//
// Ruthless Negotiation's pick, once for each opponent: the opponent chooses
// the card from their own hand (a floor of one, "exiles a card" is not a
// "may"; an empty hand has nothing to exile, CR 609.3). The exile carries
// a play permission for that card's OWNER that lasts while it stays exiled
// (Zero, Incorporeal Sensate's CostMore 1: the spell's total cost is raised
// by {1} with the other increases, CR 601.2f). The permission is the
// card's, not the Inquisitor's, so it outlives it. "Each land played this
// way enters tapped" is the permission's EntersTapped on a land play
// (CastPermission.EntersTapped, seeded on the land's entry event), so it
// never touches a spell or a land played any other way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "57012049-06d7-4354-8125-9ea9e742f14d",
		Name:            "Lightstall Inquisitor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lightstall Inquisitor — each opponent exiles a card from their hand and may play it",
				func(g *game.Game, item *game.StackItem) error {
					eachOpponent(g, item.Controller, func(opp uuid.UUID) bool {
						p := g.PlayerByIDForEffect(opp)
						if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
							return false
						}
						cards := make([]uuid.UUID, 0, p.Hand.Size())
						for _, c := range p.Hand.Cards {
							cards = append(cards, c.InstanceID)
						}
						g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
							Chooser:    opp,
							FromPlayer: opp,
							Source:     item.SourceCardID,
							Question:   "Lightstall Inquisitor — exile a card from your hand; you may play it for as long as it remains exiled",
							Cards:      cards,
							Min:        1,
							Max:        1,
							ThenKey:    lightstallInquisitorExiled,
							LiveZone:   game.ZoneHand,
						})
						return false
					})
					return nil
				}),
		},
	})
}

// lightstallInquisitorExiled is the keyed answer: exile the picked card,
// then grant its owner the play permission, over the card once it has
// really landed in exile. An on-disk key: never renamed, never reused.
var lightstallInquisitorExiled = game.RegisterChooseCardsThen("choose-cards/lightstall-inquisitor-exile",
	func(g *game.Game, r game.CardsChosen) error {
		perm := ExileWithPermission{WhileExiled: true, CostMore: 1}.permission()
		perm.EntersTapped = true
		return exileAllWithPermission(g, r.Picked, perm)
	})
