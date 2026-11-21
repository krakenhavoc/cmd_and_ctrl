package effectstest

// fog_test.go covers the S17 sub-PR 5 damage-prevention scaffold:
// Fog cancels combat damage (both creature-on-creature and
// attacker-on-player) via the turn-scoped replacement slot.
// Non-combat damage (spell / effect) still lands. The replacement
// clears at StepCleanup so next turn's combat damage resolves
// normally.

const FogOracle = "27e9db49-7af7-4bef-ad4c-bf5dfb92030d"
