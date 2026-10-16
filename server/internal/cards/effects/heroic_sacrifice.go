package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heroic Sacrifice — Instant {1}{W}:
//
//	"Choose target creature you control. Until end of turn, all damage
//	 that would be dealt to you and creatures you control is dealt to the
//	 chosen creature instead (if it's still on the battlefield). When
//	 that creature dies this turn, put its counters on up to one target
//	 creature you control and draw a card."
//
// The redirection is ADR 0108 §9's (#1905): all damage this turn to you
// and to each creature you control, from every source, is dealt to the
// chosen creature instead, pinned to it as an object, so one that has
// left the battlefield or come back as a new object takes nothing and the
// damage is dealt as it was (CR 614.9, 400.7). Damage that would be dealt
// to the chosen creature itself stays where it is.
//
// "When that creature dies this turn" is a delayed triggered ability
// about that object (CR 603.7a, 603.7c; #2057), controlled by the caster
// (CR 603.7d). Its "up to one target creature you control" is chosen as
// it goes on the stack (CR 603.3d), and "its counters" are every kind the
// creature had as it last existed on the battlefield (CR 608.2h), each
// put on the target through the counter replacements. You draw a card
// whether or not you chose a target or the creature had counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03c47f1c-02a7-428c-a126-9e85325ebc71",
		Name:         "Heroic Sacrifice",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve:    heroicSacrifice,
	})
}

func heroicSacrifice(_ *game.StackItem, ctx *Context) error {
	chosen, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (RedirectDamage{
		Protect: ShieldYouAndYourCreatures,
		To:      RedirectToClause(0),
		Label:   "Heroic Sacrifice — damage to you and creatures you control is dealt to the chosen creature",
	}).Apply(ctx); err != nil {
		return err
	}
	return whenThatCreatureDiesThisTurn(ctx, chosen,
		"Heroic Sacrifice — put its counters on up to one target creature you control and draw a card",
		heroicSacrificeDiesBody)
}
