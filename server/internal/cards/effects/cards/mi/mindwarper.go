package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindwarper — Creature — Spirit {2}{B}{B}, 0/0 (Stronghold, S82 OLD batch 748,
// #5627):
//
//	"This creature enters with three +1/+1 counters on it.
//	 {2}{B}, Remove a +1/+1 counter from this creature: Target player discards
//	 a card. Activate only as a sorcery."
//
// Monoskelion's entry replacement and counter-removal cost, at sorcery timing
// (CR 602.5d). The target player chooses the card to discard (CR 701.9a) from a
// real prompt over their hand, once the target is read again as the ability
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efb42ccc-6fed-47f0-84b4-e2a9a42e4dfa",
		Name:         "Mindwarper",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB031SelfEntersWithCounters(game.CounterPlusOne, 3)},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{B}, Remove a +1/+1 counter from this creature: Target player discards a card. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{B}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets:      effects.TargetPlayer("target player"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				player := effects.TargetedPlayer(effects.NewContext(g, item))
				if player == uuid.Nil {
					return nil
				}
				g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: player, Source: item.SourceCardID, N: 1})
				return nil
			},
		}},
	})
}
