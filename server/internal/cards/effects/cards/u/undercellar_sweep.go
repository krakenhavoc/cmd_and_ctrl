package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undercellar Sweep — Enchantment {4}{W} (Commander Legends: Battle for
// Baldur's Gate, #5948):
//
//	"When this enchantment enters, you take the initiative.
//	 Whenever you attack, if you or a player you're attacking has the
//	 initiative, you create two 1/1 white Soldier creature tokens that are
//	 tapped and attacking."
//
// The initiative is Goliath Paladin's shared entry trigger. "Whenever you
// attack" is Battlesong Berserker's once-per-declaration trigger with an
// intervening "if" (CR 603.4, read as the attack is declared and again as it
// resolves): you hold the initiative, or the owner of an attacked planeswalker
// or the player you attack does, read off the creatures you have attacking.
// The Soldiers enter tapped and attacking, and their controller chooses what
// each attacks (CR 508.4; Basri Ket's CreateTokensAttackingYourChoice, the
// first attacker's target offered first); they were never declared as
// attackers, so they fire no attack trigger, this one included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a747cdb-afbe-451d-8b1b-9458401a5419",
		Name:         "Undercellar Sweep",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouTakeTheInitiative("Undercellar Sweep"),
			effects.OncePerBatch(effects.InterveningIfYou("you or a player you're attacking has the initiative",
				effects.YouOrAPlayerYoureAttackingHasTheInitiative,
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Undercellar Sweep — create two 1/1 white Soldier creature tokens that are tapped and attacking",
					func(g *game.Game, item *game.StackItem) error {
						return effects.CreateTokensAttackingYourChoice{
							Template: effects.TokenCard("1/1 white Soldier"),
							N:        2,
							Tapped:   true,
							Prefer:   effects.TriggerAttackTarget(item),
							Label:    "Undercellar Sweep",
						}.Apply(effects.NewContext(g, item))
					}))),
		},
	})
}
