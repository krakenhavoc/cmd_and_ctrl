package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pain Distributor — Creature — Devil Citizen {2}{R}, 2/3 (March of the
// Machine Commander, #5477):
//
//	"Menace
//	 Whenever a player casts their first spell each turn, they create a
//	 Treasure token.
//	 Whenever an artifact an opponent controls is put into a graveyard
//	 from the battlefield, this creature deals 1 damage to that player."
//
// Menace is printed. The first ability is Rodeo Pyromancers' first-spell
// clock on every player's cast (the per-turn tally is bumped before
// EventCast, so a total of one is the first), and the Treasure goes to the
// caster (s82MomB14TreasureForTheCaster). The second is Party Dude's
// artifact-put-into-a-graveyard read (b23ArtifactPutIntoGraveyardFromBattlefield),
// a token included, for an artifact whose controller as it left is an
// opponent of the Distributor's controller; the damage goes to that
// player (s82MomB14DamageTheArtifactsController), so a Treasure the
// caster sacrifices off their own side does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "435df16b-3dd4-446e-a643-f9478aaa47c0",
		Name:            "Pain Distributor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor != uuid.Nil && g.CastTallyFor(ev.Actor).Total == 1
			}, "Pain Distributor — that player creates a Treasure token",
				Do(s82MomB14TreasureForTheCaster{CreateToken{Template: TreasureToken(), N: 1}})),
			On(game.EventLTB, s82MomB14OpponentArtifactPutIntoGraveyard, "Pain Distributor — 1 damage to the artifact's controller",
				Do(s82MomB14DamageTheArtifactsController{DamageEachOpponent{N: 1}})),
		},
	})
}
