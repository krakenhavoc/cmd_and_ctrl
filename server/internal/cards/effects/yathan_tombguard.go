package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yathan Tombguard — Creature — Human Warrior {2}{B}, 2/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Menace
//	 Whenever a creature you control with a counter on it deals combat
//	 damage to a player, you draw a card and you lose 1 life."
//
// Black Widow, Deadly Hunter's trigger with a counter check in place of the
// keyword: one event per creature, so two creatures with counters that
// connect are two triggers. "A counter on it" is any kind of counter,
// read off the dealer as it stands when the damage is dealt, and the draw
// comes before the life loss, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2321b4fc-6b73-4d1a-84cc-13524adb0899",
		Name:            "Yathan Tombguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, creatureWithACounterYouControlHitAPlayer, "Yathan Tombguard — draw a card and lose 1 life",
				Do(DrawCards{N: 1}, GainLife{Amount: -1})),
		},
	})
}

// creatureWithACounterYouControlHitAPlayer is the condition of "whenever a
// creature you control with a counter on it deals combat damage to a
// player": one event per creature.
func creatureWithACounterYouControlHitAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	if !ok {
		return false
	}
	for _, n := range dealer.Counters {
		if n > 0 {
			return true
		}
	}
	return false
}
