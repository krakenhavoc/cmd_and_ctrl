package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost-Lit Stalker — Creature — Spirit {B}, 1/1 (Saviors of Kamigawa,
// S82 OLD b584, #5601):
//
//	"{4}{B}, {T}: Target player discards two cards. Activate only as a
//	 sorcery.
//	 Channel — {5}{B}{B}, Discard this card: Target player discards four
//	 cards. Activate only as a sorcery."
//
// Aggressive Sabotage's discard of a chosen player; the target player picks
// the cards. Both rows are sorcery-speed, the channel row from the hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90a4c5ed-d268-4216-9b55-0a38406efd74",
		Name:         "Ghost-Lit Stalker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:        "{4}{B}, {T}: Target player discards two cards. Activate only as a sorcery.",
				Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 2}),
				Cost:         effects.Plus(effects.ManaCost("{4}{B}"), effects.TapCost()),
				SorcerySpeed: true,
				Targets:      effects.TargetPlayer("target player"),
				Effect:       ghostLitStalkerDiscard(2),
			},
			{
				Label:        "Channel — {5}{B}{B}, Discard this card: Target player discards four cards. Activate only as a sorcery.",
				Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 4}),
				Cost:         effects.Plus(effects.ManaCost("{5}{B}{B}"), effects.DiscardThis()),
				SorcerySpeed: true,
				Zones:        []game.ZoneKind{game.ZoneHand},
				Targets:      effects.TargetPlayer("target player"),
				Effect:       ghostLitStalkerDiscard(4),
			},
		},
	})
}

func ghostLitStalkerDiscard(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetPlayer {
				g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: t.ID, Source: item.SourceCardID, N: n})
				return nil
			}
		}
		return nil
	}
}
