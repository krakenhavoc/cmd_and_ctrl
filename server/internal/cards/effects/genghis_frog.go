package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Genghis Frog — {G}{U} Legendary Creature — Frog Mutant Rogue 1/3
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Trample
//	 Whenever Genghis Frog or another Mutant you control enters, create a
//	 Mutagen token. (It's an artifact with "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target
//	 creature. Activate only as a sorcery.")"
//
// One trigger with Lossarnach Captain's condition: the Frog itself, or
// another permanent its controller controls that has the creature type
// Mutant as it enters (a word of the Frog's text, CR 612.2). The Mutagen token is the catalog's CR 111.10v template (MutagenToken, #4676).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1d6d0348-46ad-46ed-ae7b-d7510f0ebe70",
		Name:            "Genghis Frog",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: genghisFrogSelfOrAnotherMutantEntered,
			Key:       "Genghis Frog — create a Mutagen token",
			Effect:    Do(CreateToken{Template: MutagenToken(), N: 1}),
		}},
	})
}

// genghisFrogSelfOrAnotherMutantEntered is "Genghis Frog or another Mutant
// you control enters".
func genghisFrogSelfOrAnotherMutantEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind == game.EventETB && ev.CardID == source.InstanceID {
		return true
	}
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype(source.TextSubtype("Mutant"))
}
