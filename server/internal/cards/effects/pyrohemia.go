package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9ac57a10-3402-4656-9079-f713884cde35",
		Name:         "Pyrohemia",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{R}: Pyrohemia deals 1 damage to each creature and each player.",
			Cost:  ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b23DamageEachCreatureAndEachPlayer(NewContext(g, item), 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b23NoCreaturesOnBattlefield(g)
			}, "Pyrohemia — no creatures: sacrifice it", func(g *game.Game, item *game.StackItem) error {
				if !b23NoCreaturesOnBattlefield(g) || !onBattlefield(g, item.SourceCardID) {
					return nil
				}
				return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}),
		},
	})
}
