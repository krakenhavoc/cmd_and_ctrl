package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thundering Mightmare — {4}{G} Creature — Horse Spirit 3/3
// (Innistrad: Crimson Vow Commander, #5675):
//
//	"Soulbond (You may pair this creature with another unpaired creature
//	 when either enters. They remain paired for as long as you control
//	 both of them.)
//	 As long as Thundering Mightmare is paired with another creature,
//	 each of those creatures has "Whenever an opponent casts a spell,
//	 put a +1/+1 counter on this creature.""
//
// Soulbond is the keyword's two triggers (CR 702.95a, soulbond.go). The
// second line is a layer-6 grant to both creatures of the pair, so each
// has its own instance of the ability: one spell from an opponent
// triggers both, each under its creature's controller, and each puts a
// counter on the creature it is on. The pair ends, and the grant with
// it, when either creature changes controller, stops being a creature
// or leaves the battlefield (CR 702.95e).
//
// No simplification.
const thunderingMightmareGrant = "thundering-mightmare/counter"

func init() {
	Register(Spec{
		OracleID:        "bf9a6242-b608-46ef-a059-7e757dd52e9b",
		Name:            "Thundering Mightmare",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       Soulbond("Thundering Mightmare"),
		Grants: []AbilityGrant{{
			Key: thunderingMightmareGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b15OpponentCastSpell(ev, source)
				}, "Thundering Mightmare — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			},
			Text: "Whenever an opponent casts a spell, put a +1/+1 counter on this creature.",
		}},
		Static: []game.StaticAbility{
			PairedCreaturesHave("As long as Thundering Mightmare is paired with another creature, each of those creatures has \"Whenever an opponent casts a spell, put a +1/+1 counter on this creature.\"",
				thunderingMightmareGrant),
		},
	})
}
