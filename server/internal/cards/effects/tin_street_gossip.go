package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tin Street Gossip — Creature — Lizard Advisor {2}{R}{G} (Murders at
// Karlov Manor, #4469):
//
//	"Vigilance
//	 {T}: Add {R}{G}. Spend this mana only to cast face-down spells or to
//	 turn creatures face up."
//
// The restriction is ManaRestrictFaceDownSpell (a cast whose spell is
// cast face down, CR 708.4) or ManaRestrictTurnFaceUp (CR 116.2g). The
// engine's face-up special action is the only thing that spends the
// second kind, and it applies to a permanent that is a creature when
// face down (CR 708.2a), so no creature check is needed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "19dc29a7-8ccc-4e54-829b-86cc102b082d",
		Name:            "Tin Street Gossip",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}{G}",
			Label:    "Add {R}{G}. Spend this mana only to cast face-down spells or to turn creatures face up",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictFaceDownSpell},
				[]string{game.ManaRestrictTurnFaceUp},
			)},
		}},
	})
}
