package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Storm — Enchantment {3}{U}:
//
//	"Creature spells can't be cast.
//	 Pay 4 life: Destroy this enchantment. It can't be regenerated.
//	 Any player may activate this ability."
//
// ADR 0106 PR 6 (#1793). Two lines, both existing shapes:
//
//   - "Creature spells can't be cast" names no player, so it binds
//     every player, its controller included: a CastRestriction
//     (PlayersCantCast, CR 101.2) over creature spells, judged by the
//     spell's own characteristics as it is announced. It stops CASTING
//     only — a creature put onto the battlefield some other way is
//     untouched, which is the printed card.
//   - The any-player row (CR 602.2, 602.1b). The 4 life is the
//     ACTIVATOR's to pay (CR 602.1a, 119.4), and the destruction is the
//     effect, so it resolves like any other ability and can be answered.
//     "It can't be regenerated" is DestroyOptions.CantBeRegenerated
//     (CR 701.19c).
//
// No purpose for the bot: the effect helps whoever wants creatures
// cast again, which no Purpose field describes, so a bot
// never pays life to destroy another player's Aether Storm.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ff4297d3-3d96-4bd6-a606-1bdc20a6df2b",
		Name:         "Aether Storm",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			PlayersCantCast("Creature spells can't be cast.", Creature()),
		},
		Activated: []ActivatedAbility{{
			Label:     "Pay 4 life: Destroy this enchantment. It can't be regenerated. Any player may activate this ability.",
			Purpose:   game.Purpose{Answers: game.AnswerRestrict},
			Cost:      game.AbilityCost{Life: 4},
			AnyPlayer: true,
			Effect:    destroyThisPermanent(true),
		}},
	})
}
