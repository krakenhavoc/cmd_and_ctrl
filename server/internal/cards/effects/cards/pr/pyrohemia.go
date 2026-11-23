package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyrohemia — Enchantment {2}{R}{R} (EDHREC rank 2470):
//
//	"At the beginning of the end step, if no creatures are on the
//	 battlefield, sacrifice this enchantment.
//	 {R}: This enchantment deals 1 damage to each creature and each
//	 player."
//
// Pestilence in red. The activation is a CR 602 ability with a
// {R} mana cost and no target: one damage from the enchantment to
// every creature and every player — the controller and their own
// creatures included, as printed — snapshotted before the first
// point lands, the deaths left to the state-based sweep. Activate it
// as many times as there is red mana.
//
// The sacrifice is "the end step", every player's, with an
// intervening-if (CR 603.4): the trigger is checked when the step
// begins and again when it resolves, so a creature that arrives in
// response keeps the enchantment, and an enchantment already gone by
// then sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ac57a10-3402-4656-9079-f713884cde35",
		Name:         "Pyrohemia",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Pyrohemia deals 1 damage to each creature and each player.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
			Cost:    effects.ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B23DamageEachCreatureAndEachPlayer(effects.NewContext(g, item), 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B23NoCreaturesOnBattlefield(g)
			}, "Pyrohemia — no creatures: sacrifice it", effects.B23SacrificeSelfWhenNoCreatures),
		},
	})
}
