package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Extraplanar Lens — Artifact {3} (Mirrodin, #5606):
//
//	"Imprint — When this artifact enters, you may exile target land you
//	 control.
//	 Whenever a land with the same name as the exiled card is tapped for
//	 mana, its controller adds one mana of any type that land produced."
//
// The imprint is Duplicant's: an enters trigger that targets as it goes
// on the stack, asks its "you may" as it resolves, and exiles the land
// linked to this Lens (CR 607.2a), so a Lens that leaves and returns is
// a new object with nothing exiled (CR 400.7).
//
// The second ability is a triggered mana ability (CR 605.1b), Heartbeat
// of Spring's, narrowed to lands that share a name with a card exiled
// with this Lens. It fires for every player's land, and the extra mana
// goes to whoever tapped it, one mana of a type the land produced. With
// nothing exiled, or once the exiled card has left exile, it never
// fires.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e48f9fc7-72cf-4c3a-b3b3-d237edc01310",
		Name:         "Extraplanar Lens",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WhenThisEnters("Extraplanar Lens — exile target land you control", exileLinkedTargetsEffect),
				TargetPermanent("target land you control", Land(), YouControl())),
				"Extraplanar Lens — exile target land you control?"),
		},
		ManaTriggers: []game.ManaTrigger{extraplanarLensMana()},
	})
}

// extraplanarLensMana is "Whenever a land with the same name as the
// exiled card is tapped for mana, its controller adds one mana of any
// type that land produced". Read-only, as every mana trigger is: it
// reads the tapped land through prod.Source and the cards in exile
// linked to this Lens.
func extraplanarLensMana() game.ManaTrigger {
	t := WheneverAPlayerTapsALandForMana(
		"Extraplanar Lens — add one mana of any type that land produced",
		AddsOneManaOfAnyTypeProduced())
	anyLand := t.AppliesTo
	t.AppliesTo = func(prod game.ManaProduced, source *game.Card, g *game.Game) bool {
		if !anyLand(prod, source, g) {
			return false
		}
		for _, c := range g.ExiledWithForEffect(game.ObjectRefOf(*source)) {
			if c.Name == prod.Source.Name {
				return true
			}
		}
		return false
	}
	return t
}
