package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Night Nurse, Healer of Heroes — {1}{W} Legendary Creature — Human Doctor
// Hero 2/1 (Marvel Super Heroes, #4455):
//
//	"Flash
//	 Lifelink
//	 When Night Nurse enters, choose target permanent card in your graveyard
//	 that was put there from anywhere this turn. Return it to your hand."
//
// The target is a permanent card in your graveyard with a graveyard arrival
// in this turn's events (the same scan descend uses), so a card that sat
// there since an earlier turn is not offered, and a card that went away and
// came back is. The clause is re-judged on resolution (CR 608.2b), and the
// shared return-to-hand step does the move.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e3559565-0512-4525-addb-bf27903bcd13",
		Name:            "Night Nurse, Healer of Heroes",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Night Nurse — return target permanent card put into your graveyard this turn to your hand", returnTargetCardToHand),
				TargetCardInGraveyard("target permanent card in your graveyard that was put there from anywhere this turn",
					YouOwn(), mshB6PutIntoGraveyardThisTurn)),
		},
	})
}

// mshB6PutIntoGraveyardThisTurn passes a permanent card with a graveyard
// arrival from any other zone in this turn's events.
func mshB6PutIntoGraveyardThisTurn(g *game.Game, _ uuid.UUID, c game.Card) bool {
	if !isPermanentCard(c) {
		return false
	}
	for _, ev := range g.EventsThisTurn() {
		if ev.CardID == c.InstanceID && ev.NewZone == game.ZoneGraveyard && ev.OldZone != game.ZoneGraveyard {
			return true
		}
	}
	return false
}
