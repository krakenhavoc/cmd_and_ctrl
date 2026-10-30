package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doombot Harbinger — Artifact Creature — Robot Villain {2}{B}, 2/1 (Marvel
// Super Heroes Commander, #4455):
//
//	"Flying
//	 When this creature enters, you may mill four cards.
//	 When this creature dies, you may exile this card. When you do, return
//	 target creature card from your graveyard to your hand."
//
// The mill is a "you may" on the trigger (CR 603.5). The dies trigger is
// Undead Butler's: a CR 603.12 reflexive trigger whose target is chosen when
// it goes on the stack, after the Harbinger has left the graveyard, so it
// cannot return itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "37042a3e-c377-4282-80f9-068e4955f8a1",
		Name:            "Doombot Harbinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters("Doombot Harbinger — mill four cards", Do(MillCards{N: 4})),
				"Doombot Harbinger — mill four cards?"),
			Optional(
				WhenThisDies("Doombot Harbinger — exile it", exileItThenRebuyACreatureCard("Doombot Harbinger")),
				"Doombot Harbinger — exile it to return a creature card from your graveyard to your hand?"),
		},
	})
}
