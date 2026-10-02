package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// O-Kagachi, Vengeful Kami — Legendary Creature — Dragon Spirit
// {1}{W}{U}{B}{R}{G}, 6/6:
//
//	"Flying, trample
//	 Whenever O-Kagachi deals combat damage to a player, if that player
//	 attacked you during their last turn, exile target nonland permanent
//	 that player controls."
//
// A mandatory trigger on its own combat damage to a player. The "if" is
// an intervening condition (CR 603.4): ADR 0108 §6's record, asked of the
// player who was hit, checked as the damage is dealt (the trigger doesn't
// exist otherwise) and again as the ability resolves. "That player" is read
// off the trigger's event (CR 603.10) for the target clause, so in a
// four-player game the pick is the permanents of the one player who was
// hit, never anyone else's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "300715a0-4f95-4212-9e13-558434c9d1a4",
		Name:            "O-Kagachi, Vengeful Kami",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			Key:     "O-Kagachi — exile target nonland permanent that player controls",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g) &&
					g.AttackedYouDuringTheirLastTurn(ev.Target, source.Controller)
			},
			TargetsFrom: TargetNonlandPermanentOfThePlayerHit,
			Effect:      oKagachiExile,
		}},
	})
}

// oKagachiExile rechecks the "if" as the ability resolves (CR 603.4), then
// exiles the chosen permanent if it is still one the hit player controls
// (CR 608.2b).
func oKagachiExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !g.AttackedYouDuringTheirLastTurn(item.Trigger.Event.Target, ctx.Controller()) {
		return nil
	}
	return ExileFirstTarget(g, item)
}
