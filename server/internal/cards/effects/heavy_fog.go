package effects

// Heavy Fog — Instant {1}{G}:
//
//	"Cast this spell only during the declare attackers step and only if
//	 you've been attacked this step.
//	 Prevent all damage that would be dealt to you this turn by attacking
//	 creatures."
//
// The restriction is a CastCondition: the declare attackers step, with a
// creature declared attacking you in this combat (an attack on your
// planeswalker is not an attack on you, the ruling). The shield is
// #2026's combat status, read as each creature would deal damage (CR
// 609.7b): all damage to you, combat or not, and none to your creatures
// or planeswalkers. A creature that has left the battlefield is read as
// it last existed there (CR 608.2h): its noncombat damage is prevented
// "if it was an attacking creature at the time it left" (the ruling).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:           "a2006755-8812-4aad-8567-e8df6e8923da",
		Name:               "Heavy Fog",
		Completeness:       CompletenessFull,
		CastCondition:      youWereAttackedThisStep,
		CastConditionLabel: attackedThisStepLabel,
		OnResolve:          sourceShieldSpell(damageToYouFromAttackers()),
	})
}
