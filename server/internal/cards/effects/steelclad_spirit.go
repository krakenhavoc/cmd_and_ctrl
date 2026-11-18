package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steelclad Spirit — Creature — Spirit {1}{U}, 3/3 (Innistrad: Crimson
// Vow, S82):
//
//	"Defender
//	 Whenever an enchantment you control enters, this creature can attack
//	 this turn as though it didn't have defender."
//
// The waiver is Hightide Hermit's CanAttackThisTurnAsThoughNoDefender: a
// until-end-of-turn waiver on the Spirit that leaves the keyword on it
// (CR 702.3b, CR 609.4). The trigger is Ajani's Chosen's "an enchantment
// you control enters" and also fires for the Spirit entering if it is an
// enchantment (it is not), so it is another enchantment's arrival.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3dc225a2-61b6-4ac9-a86b-df0c2b0b3898",
		Name:            "Steelclad Spirit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsEnchantment()
			}, "Steelclad Spirit — can attack this turn as though it didn't have defender",
				CanAttackThisTurnAsThoughNoDefender("Steelclad Spirit — can attack this turn as though it didn't have defender")),
		},
	})
}
