package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kukemssa Serpent — Creature — Serpent {3}{U}, 4/3:
//
//	"This creature can't attack unless defending player controls an
//	 Island.
//	 {U}, Sacrifice an Island: Target land an opponent controls becomes
//	 an Island until end of turn.
//	 When you control no Islands, sacrifice this creature."
//
// The restriction is ADR 0107 §2's (#1879, CR 508.1c), read per defending
// player (CR 508.5), and the sacrifice is ADR 0107 §1's CR 603.8 state
// trigger over its controller's Islands (#1858): it triggers as soon as
// the last one is gone, which the ability's own sacrifice can cause. The
// ability is ADR 0109 §1's (#1881) CR 305.7 type set: until end of turn
// the opponent's land is an Island (its other subtypes stay, CR 205.1a),
// loses its rules-text abilities and taps for {U} (CR 305.6).
//
// No simplification.
func init() {
	q := effects.QuerySubtype("Island")
	effects.Register(effects.Spec{
		OracleID:     "49e3ac82-7c22-4dec-b61a-b581148c0419",
		Name:         "Kukemssa Serpent",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouControlNo(q, "Kukemssa Serpent — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, Sacrifice an Island: Target land an opponent controls becomes an Island until end of turn.",
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.SacrificeN(1, "an Island", effects.HasSubtype("Island"))),
			Targets: effects.TargetPermanent("target land an opponent controls", effects.Land(), effects.OpponentControls()),
			Effect:  effects.TargetLandBecomesUntilEOT("Kukemssa Serpent", "Island"),
		}},
	})
}
