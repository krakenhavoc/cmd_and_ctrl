package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Judgment of Alexander — Instant {2}{W}:
//
//	"Prevent all damage that would be dealt to you this turn by sources
//	 your opponents control. Whenever damage from a creature is prevented
//	 this way, each commander creature you control deals damage equal to
//	 its power to that creature."
//
// The shield is #2026's controller test, read as each source would deal
// damage (CR 609.7b). The second sentence is a triggered ability of the
// shield (Samite Ministration's shape): it triggers when the shield
// prevents some damage from a creature (CR 615.13), and once for each
// creature whose damage it prevents at the same time — "that creature"
// names one, and one event with several occurrences triggers once per
// occurrence (CR 603.2c) — so the follow-up runs per source
// (ThenPerSource). Whether the source was a creature is read as it would
// have dealt the damage. On resolution, each commander creature you
// control then deals damage equal to its power then, all at once, to the
// creature, if it is still on the battlefield.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "eb226c0d-a71b-4f9c-9fce-47d61dcfa8f8",
		Name:         "Judgment of Alexander",
		Completeness: CompletenessFull,
		OnResolve: sourceShieldSpell(PreventDamageFromSource{
			Protect:       ShieldYou,
			Filter:        game.DamageSourceFilter{Controller: game.SourceControllerOpponents},
			Then:          preventedCommandersStrikeTriggerBody,
			ThenPerSource: true,
		}),
	})
}
