package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stonehoof Chieftain — Creature — Centaur Warrior {7}{G}, 8/8
// (Commander 2016, #5537):
//
//	"Trample
//	 Indestructible
//	 Whenever another creature you control attacks, it gains trample and
//	 indestructible until end of turn."
//
// Glory Bearers' trigger (one per attacker you declare, the Chieftain's
// own attack excluded) with Dreadhorde Invasion's grant to the attacking
// creature, which keeps both until end of turn even if it is removed
// from combat.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "241d00bf-3301-4385-b559-d50565603dcc",
		Name:            "Stonehoof Chieftain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "indestructible"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && ev.CardID != source.InstanceID
			}, "Stonehoof Chieftain — the attacking creature gains trample and indestructible until end of turn",
				stonehoofChieftainGrant),
		},
	})
}

func stonehoofChieftainGrant(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
		return nil
	}
	attacker := item.Trigger.Event.CardID
	return GrantKeywordUntilEOT{
		Target:   attacker,
		Keywords: []string{"trample", "indestructible"},
		Label:    "Stonehoof Chieftain — trample and indestructible",
	}.Apply(NewContext(g, item))
}
