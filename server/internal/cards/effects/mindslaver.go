package effects

// Mindslaver — Legendary Artifact {6}:
//
//	"{4}, {T}, Sacrifice Mindslaver: You control target player during
//	 that player's next turn. (You see all cards that player could see
//	 and make all decisions for them.)"
//
// Turn control (CR 723.1, ADR 0167, #3857): the sacrifice is a cost, so
// the control survives Mindslaver being destroyed in response, and it
// waits for the target's next turn that actually begins (CR 723.1b).
// The controller plays that turn as the target, seeing their hand
// (CR 723.4) and paying only from their resources (CR 723.5a). There is
// no extra turn afterwards (controlTargetPlayerNextTurn). Two controls
// over one player stack: the newest is the one that applies
// (CR 723.1a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a806f4ee-f48a-46c3-8772-5aaabebe4c7e",
		Name:         "Mindslaver",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice Mindslaver: You control target player during that player's next turn.",
			Cost:    Plus(ManaCost("{4}"), TapCost(), SacrificeThis()),
			Targets: TargetPlayer("target player"),
			Effect:  controlTargetPlayerNextTurn("Mindslaver"),
		}},
	})
}
