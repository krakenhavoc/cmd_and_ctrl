package effects

// Loyal Retainers — Creature — Human Advisor {2}{W}, 1/1 (#4062):
//
//	"Sacrifice this creature: Return target legendary creature card from your graveyard to the battlefield.
//	 Activate only during your turn, before attackers are declared."
//
// The window is DuringYourTurnBeforeAttackersAreDeclared, an activation
// condition (CR 602.1b): your own turn, and only until the first declare
// attackers step of the turn begins - the card's ruling reads it from the
// first combat, so a turn with several combat phases does not open it
// again. Not a sorcery-speed gate: it opens your beginning, combat and
// main-phase steps up to that point, with or without an item on the stack,
// and the view and the legal enumerator read the same closure, so the
// ability greys out when the window shuts.
//
// The sacrifice is a cost, so the Retainers are in the graveyard when the
// ability resolves - they can return a legend that was there already,
// never themselves (they are not legendary).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55b88d43-ec45-4435-a316-1c8aee249ff0",
		Name:         "Loyal Retainers",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "Sacrifice this creature: Return target legendary creature card from your graveyard to the battlefield. Activate only during your turn, before attackers are declared.",
			Cost:      SacrificeThis(),
			Targets:   TargetCardInGraveyard("target legendary creature card from your graveyard", Creature(), Legendary(), YouOwn()),
			Purpose:   ForTargets(Reanimates(0)),
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    reanimateChosenUnderYourControl("Loyal Retainers"),
		}},
	})
}
