package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// White Widow, Yelena Belova — Legendary Creature — Human Assassin Villain
// {1}{B}, 1/2 (Marvel Super Heroes Commander, #4455):
//
//	"Deathtouch
//	 Whenever a creature you control with deathtouch deals combat damage to
//	 a player, put a +1/+1 counter on it."
//
// Black Widow, Deadly Hunter's trigger (a per-creature combat-damage event
// whose dealer is checked for deathtouch as it is now, game.HasKeyword, so a
// granted deathtouch counts and White Widow herself qualifies). The counter
// goes on the creature that dealt the damage, not on White Widow; if it has
// left the battlefield by resolution nothing happens (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7aff7cb6-aa56-47ae-8813-75d850917320",
		Name:            "White Widow, Yelena Belova",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, deathtouchCreatureYouControlHitAPlayer, "White Widow — put a +1/+1 counter on the creature with deathtouch that dealt the damage",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.Trigger.Event.Source, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
