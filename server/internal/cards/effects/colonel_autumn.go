package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colonel Autumn — Legendary Creature — Human Soldier {1}{W}{B}, 2/3
// (Fallout, #5948):
//
//	"Lifelink
//	 Exploit (When this creature enters, you may sacrifice a creature.)
//	 Other legendary creatures you control have exploit.
//	 Whenever a creature you control exploits a creature, put a +1/+1
//	 counter on each creature you control."
//
// Its own exploit is the shared constructor (Exploit); the grant is a
// bundle holding that same row, given to the other legendary creatures its
// controller controls (the Lost Monarch of Ifnir's shape), so each has
// exploit as an ability it can lose and sacrifices a creature when it
// enters. The counter trigger reads the exploited event (game.EventExploited,
// CR 702.110b), whose actor is the exploiter's controller as it exploited,
// once per exploit ability that really sacrificed a creature, Colonel
// Autumn's own and a granted one alike. Each creature you control then gets
// its counter in one placement per creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5c6ce876-781b-437d-b20e-db0b6e2769f4",
		Name:            "Colonel Autumn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Grants: []AbilityGrant{{
			Key:       colonelAutumnExploitGrant,
			Triggered: []game.TriggeredAbility{Exploit("Colonel Autumn", game.BodyRef{})},
			Text:      "Exploit",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.IsLegendary() && target.Controller == source.Controller &&
					target.InstanceID != source.InstanceID
			}, colonelAutumnExploitGrant), "Other legendary creatures you control have exploit."),
		},
		Triggered: []game.TriggeredAbility{
			Exploit("Colonel Autumn", game.BodyRef{}),
			On(game.EventExploited, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventExploited && ev.Actor == source.Controller
			}, "Colonel Autumn — put a +1/+1 counter on each creature you control", b11PutCounterOnEachCreatureYouControl),
		},
	})
}

// colonelAutumnExploitGrant is the bundle key of the exploit Colonel Autumn
// gives. On disk.
const colonelAutumnExploitGrant = "colonel-autumn/exploit"
