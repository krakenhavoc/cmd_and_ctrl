package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scion of Darkness — Creature — Avatar {5}{B}{B}{B}, 6/6 (Foundations
// Jumpstart, S82 BNS):
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, you may put
//	 target creature card from that player's graveyard onto the battlefield
//	 under your control.
//	 Cycling {3}"
//
// Ink-Eyes' clause: the card is targeted in the damaged player's graveyard
// as the trigger goes on the stack, "you may" is asked as it resolves, and
// the card comes back under your control, not its owner's. Cycling is the
// keyword's own hand ability.
//
// No simplification.
func init() {
	damage := effects.WheneverThisDealsCombatDamageToAPlayer(
		"Scion of Darkness — put target creature card from that player's graveyard onto the battlefield under your control",
		effects.S82Mom16PutTargetOntoTheBattlefield)
	damage.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		return effects.TargetCardInGraveyard("target creature card from that player's graveyard",
			effects.Creature(), effects.OwnedBy(tc.Event.Target))
	}
	effects.Register(effects.Spec{
		OracleID:        "bd940cab-d23a-44bf-ab96-8aa1dcd3f0ff",
		Name:            "Scion of Darkness",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(damage, "Scion of Darkness — put target creature card from that player's graveyard onto the battlefield under your control?"),
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{3}")},
	})
}
