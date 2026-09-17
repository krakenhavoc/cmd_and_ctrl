package deck

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Real keyword-only records exercise import and coverage without relying on a
// catalog entry: adding the canonical token must enable both together.
func TestEvasionKeywordsImportWithAutomaticCoverage(t *testing.T) {
	for _, c := range []cards.Card{
		{OracleID: uuid.MustParse("f640a102-7ee8-4605-9e15-09938fb229b2"), Name: "Bladetusk Boar", TypeLine: "Creature — Boar", ManaCost: "{3}{R}", Power: "3", Toughness: "2", Keywords: []string{"Intimidate"}, OracleText: "Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)"},
		{OracleID: uuid.MustParse("34a8a658-a5f6-406f-99d8-c32ea2e26202"), Name: "Severed Legion", TypeLine: "Creature — Zombie", ManaCost: "{1}{B}{B}", Power: "2", Toughness: "2", Keywords: []string{"Fear"}, OracleText: "Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)"},
		{OracleID: uuid.MustParse("6ccb3ac8-0509-431b-a196-d3cbf4fa80ae"), Name: "Soltari Foot Soldier", TypeLine: "Creature — Soltari Soldier", ManaCost: "{W}", Power: "1", Toughness: "1", Keywords: []string{"Shadow"}, OracleText: "Shadow (This creature can block or be blocked by only creatures with shadow.)"},
		{OracleID: uuid.MustParse("6b59ffa2-5c8b-4549-aa09-2f2a5cabc5af"), Name: "Furtive Homunculus", TypeLine: "Creature — Homunculus", ManaCost: "{1}{U}", Power: "2", Toughness: "1", Keywords: []string{"Skulk"}, OracleText: "Skulk (This creature can't be blocked by creatures with greater power.)"},
		{OracleID: uuid.MustParse("c24da532-311b-4b83-82d2-7dc80409ba12"), Name: "Wu Light Cavalry", TypeLine: "Creature — Human Soldier", ManaCost: "{1}{U}", Power: "1", Toughness: "2", Keywords: []string{"Horsemanship"}, OracleText: "Horsemanship (This creature can't be blocked except by creatures with horsemanship.)"},
	} {
		t.Run(c.Name, func(t *testing.T) {
			imported := ToGameCard(c, false)
			kw, ok := game.CanonicalKeyword(c.Keywords[0])
			if !ok || !game.HasKeyword(&imported, kw) {
				t.Fatalf("import discarded %q: %v", c.Keywords[0], imported.Keywords)
			}
			if imported.NeedsEffect {
				t.Fatal("keyword-only creature still requests an effect implementation")
			}
			c.OracleText += "\nWhen this creature enters, draw a card."
			if !ToGameCard(c, false).NeedsEffect {
				t.Fatal("keyword support hid an unimplemented triggered ability")
			}
		})
	}
}
