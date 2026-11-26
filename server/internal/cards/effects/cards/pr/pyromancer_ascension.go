package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyromancer Ascension — Enchantment {1}{R} (Zendikar, S82 OLD b493):
//
//	"Whenever you cast an instant or sorcery spell that has the same name as a
//	 card in your graveyard, you may put a quest counter on this enchantment.
//	 Whenever you cast an instant or sorcery spell while this enchantment has two
//	 or more quest counters on it, you may copy that spell. You may choose new
//	 targets for the copy."
//
// Two triggers. The first compares the spell's name with the cards in the
// caster's graveyard as it is cast (the spell is on the stack, not in the
// graveyard). The second reads the quest counters as the spell is cast, so the
// counter the first trigger adds for the same spell does not count for it
// (the counter arrives after the second trigger has triggered or not), and
// copies it from last-known information with the re-target prompt (CR 707.10).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85845b6c-abb5-4591-b987-a375972cf25b",
		Name:         "Pyromancer Ascension",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, effects.Old493SpellSharesNameWithYourGraveyard,
				"Pyromancer Ascension — put a quest counter",
				effects.Do(effects.CounterOnThis{Kind: "quest", N: 1})), "Pyromancer Ascension — put a quest counter on it?"),
			effects.Optional(effects.On(game.EventCast, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return source.Counters["quest"] >= 2 && effects.YouCast(effects.Or(effects.Instant(), effects.Sorcery()))(ev, source, ch, g)
			}, "Pyromancer Ascension — copy that spell", effects.Old493CopyThatSpell),
				"Pyromancer Ascension — copy that spell?"),
		},
	})
}
