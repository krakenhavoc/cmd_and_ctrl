package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pithing Needle — {1} Artifact:
//
//	"As this artifact enters, choose a card name.
//	 Activated abilities of sources with the chosen name can't be
//	 activated unless they're mana abilities."
//
// The card ADR 0073 §7 said the engine could not write: "a ban on a
// chosen card NAME … needs a choose-a-card-name prompt, and the
// engine has NamedTribe and ChosenColor but no name." #1210 built the
// prompt, and this is it in use.
//
// TWO ENGINE PIECES AND NO CARD LOGIC. The entry hook queues the CR
// 614.12 choice (ChooseCardNameAsEnters); the restriction reads the
// answer off THIS permanent every time an ability is announced
// (ChosenNameCantActivate). Two Needles name two different cards
// because the name is per instance, and a Clone of a Needle names its
// own as it enters, because a chosen name is not a copiable value
// (CR 706.2).
//
// "UNLESS THEY'RE MANA ABILITIES" is the half Cursed Totem does not
// print, and it is why the mana exemption lives on the restriction
// rather than on the call site: name a Gaea's Cradle and it still
// taps for mana, name a Sensei's Divining Top and the {1}: Draw is
// gone. Both go through the same gate; the clause decides.
//
// "SOURCES", not permanents, so there is no zone test — the ban
// reaches a named card's ability wherever that ability functions
// from, which since #660 includes a cycling ability in hand. That is
// the printed word and the one that costs nothing to obey.
//
// The name is matched by game.CardNameMatches, which asks every FACE
// (CR 201.2b): naming "Brutal Cathar" also stops the Moonrage Brute
// it transforms into, and naming either half of a split card names
// the card.
//
// The WINDOW between entering and the choice being answered is the
// as-enters family's declared simplification (game/choose_card_name.go
// argues it): the Needle is on the battlefield with an empty name for
// as long as the prompt is open, and an empty name restricts nobody.
// Nothing can act in that window — an open PendingChoice stops
// priority — and the direction is the safe one.
//
// One caveat, and it is the prompt's rather than the card's: the
// engine cannot check a named card against a real card name, because
// CR 201.2 admits any name at all and no such list exists here. A
// player who types a name with a typo has named a card that does not
// exist and the Needle stops nothing — which is also true at a paper
// table, where the name has to be announced correctly.
func init() {
	Register(Spec{
		OracleID:     "a188fe7e-68de-4c7c-806c-bfe8fc7b44bf",
		Name:         "Pithing Needle",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCardNameAsEnters("Pithing Needle — choose a card name"),
		ActivationRestrictions: []game.ActivationRestriction{
			ChosenNameCantActivate(
				"Pithing Needle — activated abilities of sources with the chosen name can't be activated unless they're mana abilities.",
				true),
		},
	})
}
