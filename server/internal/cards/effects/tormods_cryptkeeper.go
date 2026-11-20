package effects

// Tormod's Cryptkeeper — Artifact Creature — Golem {3}, 3/2 (Modern
// Horizons 2, #5501):
//
//	"Vigilance
//	 {T}, Sacrifice this creature: Exile target player's graveyard."
//
// Tormod's Crypt's ability on a creature: the {T} waits out summoning
// sickness (CR 302.6), and the whole graveyard leaves through Bojuka Bog's
// exileTargetPlayersGraveyard. A player with an empty graveyard is still a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8c846841-1b44-4bc5-b803-76baaa217548",
		Name:            "Tormod's Cryptkeeper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Exile target player's graveyard.",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetPlayer("target player"),
			Effect:  exileTargetPlayersGraveyard,
		}},
	})
}
