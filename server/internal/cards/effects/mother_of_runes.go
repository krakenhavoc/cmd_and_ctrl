package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mother of Runes — Creature — Human Cleric, {W}, 1/1:
//
//	"{T}: Target creature you control gains protection from the color
//	 of your choice until end of turn."
//
// The card the protection seam row was named after, and it needed
// nothing new to ship once #662 landed: the colour prompt is #742's
// existing `choose_color` kind (ChooseColorThen), the grant is the
// ordinary layer-6 GrantKeywordUntilEOT, and the token it grants is
// minted by game.ProtectionFromColor — the one place a colour PICK
// becomes a protection token, so this file never spells one.
//
// THE ORDER MATTERS AND IT IS THE PRINTED ONE. The target is chosen
// at activation (CR 602.2b) and the colour at RESOLUTION, which is
// why the colour is not a mode and not an announce-time pick: an
// opponent who responds to the activation has to commit before
// knowing which colour they are being blanked out of. Protecting a
// creature from a removal spell already on the stack works for the
// reason every protection-style keyword does — the CR 608.2b
// re-check runs the same predicate the announce gate did, so the
// spell is countered by game rules when it resolves.
//
// Protection from the creature's OWN controller's colour is a legal
// and sometimes useful choice (it does not stop your own auras from
// staying on, it stops new ones), and nothing here narrows the five.
// "The color of your choice" is exactly five options; CR 702.16 has
// no "colorless" quality and the prompt does not offer one.
func init() {
	Register(Spec{
		OracleID:     "60433b48-d27f-413c-905c-43839b1943f1",
		Name:         "Mother of Runes",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature you control gains protection from the color of your choice until end of turn.",
			Cost:    game.AbilityCost{Tap: true},
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var target game.TargetRef
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						target = t
					}
				}
				if target.Kind != game.TargetCard {
					// The creature left in response: CR 608.2b
					// counters the ability and nobody is asked for a
					// colour they cannot spend.
					return nil
				}
				// #780: the colour named here is the one the creature
				// is being protected FROM, so the prompt says so and
				// a bot names the biggest threat on the board rather
				// than its own main colour.
				ChooseColorThen(game.ColorForProtection, g, item.Controller, item.SourceCardID,
					"Mother of Runes — choose a color",
					func(g *game.Game, color string) error {
						token := game.ProtectionFromColor(color)
						if token == "" {
							return nil
						}
						return GrantKeywordUntilEOT{
							Target:   target.ID,
							Keywords: []string{token},
							Label:    "Mother of Runes — " + token,
						}.Apply(NewContext(g, item))
					})
				return nil
			},
		}},
	})
}
