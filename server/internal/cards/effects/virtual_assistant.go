package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Virtual Assistant — {1}{U} Artifact Creature — Illusion Advisor 3/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Defender
//	 Whenever you cast a spell using teamwork, create a 1/1 colorless Robot
//	 Hero artifact creature token with flying."
//
// A spell is cast using teamwork when its optional teamwork cost was paid
// (CR 702.194a); the cast record the spell put on the stack says so, read
// when the trigger fires. The token row is Shuri, Vibranium Technologist's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "69db9cb8-9152-4d1b-9ac9-16aac42c0908",
		Name:            "Virtual Assistant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, youCastUsingTeamwork,
				"Virtual Assistant — create a 1/1 Robot Hero artifact creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Robot Hero artifact with flying"), N: 1})),
		},
	})
}

// youCastUsingTeamwork is "whenever you cast a spell using teamwork".
func youCastUsingTeamwork(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && game.OptionalCostTimesPaid(spell, g.StackItemPaidForEffect(spell.InstanceID).OptionalCosts, game.TeamworkKey) > 0
}
