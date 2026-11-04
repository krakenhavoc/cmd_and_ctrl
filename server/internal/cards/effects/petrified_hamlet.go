package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Petrified Hamlet — Land (Secrets of Strixhaven, #4456):
//
//	"When this land enters, choose a land card name.
//	 Activated abilities of sources with the chosen name can't be activated
//	 unless they're mana abilities.
//	 Lands with the chosen name have "{T}: Add {C}."
//	 {T}: Add {C}."
//
// Three pieces, all on the shared machinery of the card-name family. The
// choice is a triggered ability (not an "as enters" choice, so Pithing
// Needle's entry hook is not used): its resolution queues the ordinary
// card-name prompt (socB31ChooseALandCardName) and the answer is stored on
// this permanent, as the Needle's is. Activated abilities of sources with
// the name are refused unless they are mana abilities
// (ChosenNameCantActivate's reading, Pithing Needle), and lands with the
// name get "{T}: Add {C}" in layer 6 (Chromatic Lantern's grant, Alpine
// Moon's name test).
//
// "A land card name" is a restriction on the CHOICE, and the engine takes a
// card name as free text with no vocabulary to check it against
// (game/choose_card_name.go; Alpine Moon, Phyrexian Revoker). So it is kept
// where the name is read: the ban reaches only a source with a land face of
// that name (socB31LandFaceNamed, CR 201.2b), and only a land gains the
// mana ability. A name no legal choice could make then reaches nothing,
// exactly as a legal name nothing on the table has does; nothing is ever
// stronger than the card. A basic land's name is a legal choice. Until the
// prompt is answered the name is empty and names nothing.
//
// No simplification.
const petrifiedHamletColorless = "petrified-hamlet/colorless"

func init() {
	Register(Spec{
		OracleID:     "78a2972c-14f4-41f3-99f9-167948bdd73a",
		Name:         "Petrified Hamlet",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "{T}: Add {C}."},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Petrified Hamlet — choose a land card name",
				Do(socB31ChooseALandCardName{Label: "Petrified Hamlet — choose a land card name"})),
		},
		ActivationRestrictions: []game.ActivationRestriction{{
			Label: "Petrified Hamlet — activated abilities of sources with the chosen name can't be activated unless they're mana abilities.",
			Forbids: func(q game.ActivationQuery) bool {
				return !q.Ability.Mana && socB31LandFaceNamed(q.Card, q.Source.ChosenName)
			},
		}},
		Grants: []AbilityGrant{TapForManaGrant(petrifiedHamletColorless, "{C}", "{T}: Add {C}.", "{T}: Add {C}.")},
		Static: []game.StaticAbility{
			Labelled(GrantAbilities(petrifiedHamletNamedLand, petrifiedHamletColorless), "Lands with the chosen name have \"{T}: Add {C}.\""),
		},
	})
}

// petrifiedHamletNamedLand is "lands with the chosen name": a land, of any
// controller, whose name is the one chosen for this Hamlet (the face that is
// up, CR 712.8f).
func petrifiedHamletNamedLand(target *game.Card, _ *game.Game, source *game.Card) bool {
	return source.ChosenName != "" && target.IsLand() && game.PermanentHasName(*target, source.ChosenName)
}
