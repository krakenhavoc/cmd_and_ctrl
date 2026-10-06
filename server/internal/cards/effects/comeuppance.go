package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Comeuppance — Instant {3}{W}:
//
//	"Prevent all damage that would be dealt to you and planeswalkers you
//	 control this turn by sources you don't control. If damage from a
//	 creature source is prevented this way, Comeuppance deals that much
//	 damage to that creature. If damage from a noncreature source is
//	 prevented this way, Comeuppance deals that much damage to the
//	 source's controller."
//
// The shield is #2026's controller test ("sources you don't control"),
// read as each source would deal damage (CR 609.7b), protecting you and
// the planeswalkers you control as the damage would be dealt. The rest is
// its CR 615.5 additional effect, once per source within one instance of
// damage (ThenPerSource), so three attackers stopped in one combat damage
// step are each dealt their own damage back. Whether the source was a
// creature is read as it would have dealt the damage. Comeuppance is the
// source of the new damage, which is not combat damage (the ruling). A
// creature that has left the battlefield is dealt nothing.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "01f29158-2dc9-4b9d-b726-add5d3fd5782",
		Name:         "Comeuppance",
		Completeness: CompletenessFull,
		OnResolve: sourceShieldSpell(PreventDamageFromSource{
			Protect:       ShieldYouAndYourPermanents("planeswalker"),
			Filter:        game.DamageSourceFilter{Controller: game.SourceControllerNotYou},
			Then:          preventedBackAtTheSourceBody,
			ThenPerSource: true,
		}),
	})
}
