package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Namor, Scourge of the Seas — Legendary Creature — Mutant Merfolk Villain
// {3}{U}{U}, 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Each other Merfolk you control with a +1/+1 counter on it has flying.
//	 At the beginning of combat on your turn, target creature you control
//	 connives."
//
// The flying grant is Bushmaster's layer 6 grant over Merfolk (a changeling
// counts), following the counter: it appears when a +1/+1 counter lands and
// goes when the last one leaves. The combat trigger targets a creature its
// controller controls and, as it resolves, that creature connives for its
// controller (connive.go's ConniveEach): draw, discard, and a counter for a
// nonland card. A creature that became illegal makes nothing connive
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5278728e-9203-4d18-97a5-1c4f751a5acf",
		Name:            "Namor, Scourge of the Seas",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID &&
					target.Controller == source.Controller &&
					target.IsCreature() && target.HasSubtype(source.TextSubtype("Merfolk")) &&
					target.Counters[game.CounterPlusOne] > 0
			}, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Namor — target creature you control connives",
					func(g *game.Game, item *game.StackItem) error {
						id, ok := b16FirstLegalTargetCard(NewContext(g, item))
						if !ok {
							return nil
						}
						return ConniveEach{Controller: item.Controller, Creatures: []uuid.UUID{id}}.Apply(NewContext(g, item))
					}),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
