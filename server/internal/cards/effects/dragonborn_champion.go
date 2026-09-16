package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonborn Champion — Creature — Dragon Warrior {2}{R}{G}, 5/3
// (EDHREC rank 3619):
//
//	"Trample
//	 Whenever a source you control deals 5 or more damage to a
//	 player, draw a card."
//
// The big-damage draw. Trample rides PrintedKeywords. The trigger
// reads one damage event: a source the controller controls — a
// creature in combat (the Champion itself connecting for 5
// included), a spell, an ability — dealing 5 or more to a player in
// one hit. Damage is per source per recipient, so two creatures
// hitting for 3 each do not add up, and a Fireball split three ways
// is three events, exactly as printed. Damage to a planeswalker or a
// creature does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9e51ccf9-d55d-4231-8ff5-ef3ca1485c8d",
		Name:            "Dragonborn Champion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b34SourceYouControlDealtDamageToPlayerAtLeast(ev, source, g, 5)
			}, "Dragonborn Champion — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
