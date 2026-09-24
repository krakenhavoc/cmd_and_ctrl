package effects

// Tome Shredder — Creature — Wolf {2}{R}, 2/2 (EDHREC rank 25532):
//
//	"Haste
//	 {T}, Exile an instant or sorcery card from your graveyard: Put a
//	 +1/+1 counter on this creature."
//
// The filtered form of #1297's graveyard exile cost: only an instant or
// sorcery card in your own graveyard pays, through the one candidate
// walk every reader shares. Haste is printed, so the {T} works the turn
// it arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b145952b-52e3-4a66-b47e-f08a489f9443",
		Name:            "Tome Shredder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:  "{T}, Exile an instant or sorcery card from your graveyard: Put a +1/+1 counter on this creature.",
			Cost:   Plus(TapCost(), ExileFromGraveyard(1, "an instant or sorcery card", MatchInstantOrSorcery)),
			Effect: b35PutCounterOnSelf,
		}},
	})
}
