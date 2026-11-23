package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eye for an Eye — Instant {W}{W}:
//
//	"The next time a source of your choice would deal damage to you this
//	 turn, instead that source deals that much damage to you and Eye for
//	 an Eye deals that much damage to that source's controller."
//
// ADR 0108 §9 decision 3 (#1905): a redirection record with no
// destination. The damage to you is dealt as it was, by the original
// source (the ruling), and its follow-up has Eye for an Eye deal that much
// damage to the source's controller, as the source was when it dealt the
// damage. Nothing is dealt instead to another player, so it is not a
// redirection (CR 614.9) and "can't be dealt instead" does not stop it.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22647b1a-5a7c-41b5-b820-b2e9f49c7aad",
		Name:         "Eye for an Eye",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.RedirectSpell(effects.RedirectDamage{Choose: true, Protect: effects.ShieldYou, Next: true, Then: effects.ThatMuchToTheSourcesControllerBody}),
	})
}
