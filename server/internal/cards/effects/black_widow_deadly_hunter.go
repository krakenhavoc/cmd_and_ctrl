package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Black Widow, Deadly Hunter — Legendary Creature — Human Assassin Hero
// {2}{B}, 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Deathtouch
//	 Whenever a creature you control with deathtouch deals combat damage to a
//	 player, you draw a card and you lose 1 life."
//
// Reconnaissance Mission's per-creature combat-damage trigger plus a keyword
// check on the dealer, read as it is now (game.HasKeyword), so a creature
// that has deathtouch from any source counts, Black Widow herself included.
// One event per creature, so two deathtouch creatures that connect are two
// triggers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "12a7bf3c-5194-48ad-be0b-553b8ad67671",
		Name:            "Black Widow, Deadly Hunter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !combatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				return ok && game.HasKeyword(&dealer, "deathtouch")
			}, "Black Widow, Deadly Hunter — draw a card and lose 1 life",
				Do(DrawCards{N: 1}, GainLife{Amount: -1})),
		},
	})
}
