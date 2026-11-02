package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Interdimensional Web Watch — Artifact {4} (Marvel's Spider-Man, #4460):
//
//		"When this artifact enters, exile the top two cards of your library.
//		 Until the end of your next turn, you may play those cards.
//		 {T}: Add two mana in any combination of colors. Spend this mana only
//		 to cast spells from exile."
//
//	  - The exile and its permission are Reckless Impulse's
//	    (b19ExileTopTwoUntilEndOfNextTurn): "play", so a land is not
//	    stranded, and the duration is ADR 0063's "until the end of your next
//	    turn". The cards are exiled whether or not the Watch is still there.
//	  - The mana ability is Relic of Sauron's two any-colour slots, each its
//	    own colour pick, carrying the new cast-from-exile restriction
//	    (game.ManaRestrictCastFromExile): it pays only for a spell cast from
//	    the exile zone (the cards this Watch exiled, a foretold or plotted
//	    card, an adventure on an exile), never for an activation or for a
//	    cast from anywhere else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "456dbec4-1b0e-48c4-abae-5c8205651872",
		Name:         "Interdimensional Web Watch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Interdimensional Web Watch — exile the top two cards of your library; you may play them until the end of your next turn",
				b19ExileTopTwoUntilEndOfNextTurn),
		},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{W|U|B|R|G}{W|U|B|R|G}",
			Label:        "Add two mana in any combination of colors (spend only to cast spells from exile)",
			Restrictions: []string{ManaRestrictCastFromExile},
		}},
	})
}
