package effects

// Tomb Trawler — Artifact Creature — Golem {2}, 0/4 (Outlaws of Thunder
// Junction, #4468):
//
//	"{2}: Put target card from your graveyard on the bottom of your
//	 library."
//
// Barkform Harvester's ability without the keywords: any card in your own
// graveyard (YouOwn, since a card there has no controller), tucked to the
// bottom of its owner's library by the shared body
// (trekB27TuckTargetCardToBottom). A card gone from the graveyard in
// response is left alone (CR 608.2b). The row always names a target, so it
// declares no Answers (ADR 0142).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d22d3934-8437-4a18-81c1-4e9c501846ca",
		Name:         "Tomb Trawler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}: Put target card from your graveyard on the bottom of your library.",
			Cost:    ManaCost("{2}"),
			Targets: TargetCardInGraveyard("target card from your graveyard", YouOwn()),
			Effect:  trekB27TuckTargetCardToBottom,
		}},
	})
}
