package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Throne of Eldraine — Legendary Artifact {5}:
//
//	"As Throne of Eldraine enters, choose a color.
//	 {T}: Add four mana of the chosen color. Spend this mana only to
//	 cast monocolored spells of that color.
//	 {3}, {T}: Draw two cards. Spend only mana of the chosen color to
//	 activate this ability."
//
// Two spend restrictions pointing in opposite directions (#1600, ADR
// 0040's 2026-10-03 amendment):
//
//   - The MANA is restricted: the four tokens carry "cast", "monocolored"
//     and "color:<chosen>" (MonocoloredSpellsOfTheChosenColor), so they
//     pay for a spell that is exactly the chosen colour (CR 105.2a) and
//     nothing else — not a hybrid or gold spell of that colour (CR
//     202.2d makes a {W/U} spell both colours), not a colourless one, and
//     not an activated ability, the Throne's own draw included.
//   - The COST is restricted: the draw ability's {3} may be paid only
//     with mana of the chosen colour (SpendOnlyManaOfTheChosenColor). The
//     engine folds it into three symbols of that colour at payment, so
//     the auto-tapper taps only sources of that colour for it and the
//     bot is offered it only when such sources exist.
//
// The colour is the as-enters prompt every chosen-colour card uses
// (ChooseColorAsEnters, #742), declared ColorForMana. Until it is
// answered the Throne makes no mana and its draw ability cannot be
// paid for.
//
// Under Chromatic Orrery the two halves part ways, and that is the
// rules: "spend mana as though it were mana of any color" changes how a
// COST may be paid (CR 609.4b), so any mana pays the draw ability; it
// does not lift a restriction on the MANA (the Orrery and Mycosynth
// Lattice rulings), so the four mana still cast only monocolored spells
// of the chosen colour.
//
// The auto-tapper never taps the Throne for a spell — no restricted
// mana source is planned (ADR 0040 §7) — so its four mana is a manual
// tap, which is also how the card is played on paper.
func init() {
	Register(Spec{
		OracleID:     "5b9a2b81-a645-43be-8001-03817ac210ca",
		Name:         "Throne of Eldraine",
		Completeness: CompletenessFull,
		AsEnters:     ChooseColorAsEnters(game.ColorForMana, "Throne of Eldraine"),
		ManaAbilities: []ManaAbility{{
			Cost:             ManaAbilityCost{Tap: true},
			ProducedFunc:     ProducedChosenColorAmount(4),
			RestrictionsFunc: MonocoloredSpellsOfTheChosenColor(),
			Label:            "Add four mana of the chosen color. Spend this mana only to cast monocolored spells of that color",
		}},
		Activated: []ActivatedAbility{{
			Label: "{3}, {T}: Draw two cards. Spend only mana of the chosen color to activate this ability.",
			Cost:  Plus(ManaCost("{3}"), TapCost(), SpendOnlyManaOfTheChosenColor()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
