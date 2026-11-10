package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Belligerent — Legendary Artifact — Vehicle {2}{U}{R}, 5/5 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Whenever The Belligerent attacks, create a Treasure token. Until end
//	 of turn, you may look at the top card of your library any time, and
//	 you may play lands and cast spells from the top of your library.
//	 Crew 3"
//
// The attack trigger makes the Treasure, then grants Gwenom, Remorseless's
// turn-long permission without its price: one stored standing permission
// over the top of the controller's own library (CR 401.5) that carries the
// look (SeesLibraryTop), so a land is played from the top using the
// turn's land play and a spell is cast from the top for its own costs.
// The grant's zero duration is "until end of turn", and it outlasts the
// Vehicle: once the trigger has resolved, the permission is the player's.
// Crew is the shared CrewCost / CrewEffect.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a0c9023-9431-4f98-b167-ab4c7215319c",
		Name:         "The Belligerent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("The Belligerent — create a Treasure; until end of turn, look at and play cards from the top of your library",
				theBelligerentAttacks),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("The Belligerent"),
		}},
	})
}

// theBelligerentAttacks is the attack trigger's body, in printed order.
func theBelligerentAttacks(g *game.Game, item *game.StackItem) error {
	if err := (CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}).Apply(NewContext(g, item)); err != nil {
		return err
	}
	g.GrantCastPermissionForEffect(game.CastPermission{
		Player:           item.Controller,
		Zone:             game.ZoneLibrary,
		Scope:            game.ScopeStanding,
		TopOfLibraryOnly: true,
		SeesLibraryTop:   true,
		Source:           item.SourceCardID,
		Label:            "Play a land or cast a spell from the top of your library (The Belligerent)",
	})
	return nil
}
