package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nazgûl — Creature — Wraith Knight {2}{B}, 1/2:
//
//	"Deathtouch
//	 When this creature enters, the Ring tempts you.
//	 Whenever the Ring tempts you, put a +1/+1 counter on each Wraith
//	 you control.
//	 A deck can have up to nine cards named Nazgûl."
//
// The tempt is the keyword action (ADR 0114, CR 701.54). The second
// ability triggers on every temptation, including one in which no
// creature could be chosen (CR 701.54d), so each Nazgûl on the
// battlefield triggers once per temptation and every Wraith gets one
// counter from each of them. "Each Wraith you control" is every
// permanent with the Wraith subtype, read as the trigger resolves.
//
// The last line is deck construction (CR 113.6n), read by
// deck.Validate from the card's own text (ADR 0114 §6, #2087).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "48a62778-7c11-486f-a0e1-020c283a7ef9",
		Name:            "Nazgûl",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Nazgûl — the Ring tempts you", Do(TheRingTemptsYou{})),
			WheneverTheRingTemptsYou("Nazgûl — put a +1/+1 counter on each Wraith you control",
				putAPlusOneCounterOnEachYouControl(Subtype("Wraith"))),
		},
	})
}
