package effects

// Sanctum of Eternity — Land (EDHREC rank 3753):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Return target commander you own from the battlefield
//	 to your hand. Activate only during your turn."
//
// The commander-blink land: a repeatable way to re-cast your
// commander for its ETB, or to save it from a wipe on your turn. The
// colourless mana is an ordinary tap ability. The bounce is a real
// activated ability with a target clause narrowed to permanents
// flagged as commanders that the activator OWNS — a stolen commander
// of yours is still yours to return, an opponent's commander you
// happen to control is not. The chosen commander goes to its owner's
// hand — and since #539 the bounce runs through the shared exit
// primitive, so CR 903.9 offers its owner the command zone instead,
// exactly as it does for a commander that dies. Declined, it is in
// hand to be re-cast for the printed cost, which is the point.
//
// Sandbox simplification, declared: "activate only during your turn"
// has no shape on an activated ability — the engine's one timing
// gate is sorcery speed — so the ability is sorcery-speed instead.
// Every window that allows is a window the printed card allows;
// the printed card also allows your combat, your end step, and
// responding on your own turn, which this cannot. Weaker than
// printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "c7d9ff27-f1fc-42e4-a47b-d2e6d68e4035",
		Name:         "Sanctum of Eternity",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The bounce can only be activated at sorcery speed, not at any time during your turn."},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{2}, {T}: Return target commander you own from the battlefield to your hand. Activate only during your turn.",
			Cost:         Plus(ManaCost("{2}"), TapCost()),
			Targets:      TargetPermanent("target commander you own", b36CommanderYouOwn()),
			SorcerySpeed: true,
			Effect:       bounceChosenTarget,
		}},
	})
}
