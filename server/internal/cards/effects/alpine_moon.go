package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alpine Moon — Enchantment {R}:
//
//	"As this enchantment enters, choose a nonbasic land card name.
//	 Lands your opponents control with the chosen name lose all land
//	 types and abilities, and they gain '{T}: Add one mana of any
//	 color.'"
//
// ADR 0109 §2 (#1604): the static form of "loses all land types"
// (LosesAllLandTypes, layer 4) with "lose all abilities and gain …" in
// layer 6 (LosesAllAbilitiesAndHas), over the lands the predicate below
// picks. The name is chosen as it enters (CR 614.12) through the
// ordinary card-name prompt (ChooseCardNameAsEnters, Pithing Needle's),
// and the window before it is answered names nothing, as it does for the
// Needle.
//
// "A nonbasic land card name" is a restriction on the CHOICE (CR
// 201.4a), and the engine takes a name as free text with no vocabulary
// to check it against (game/choose_card_name.go). So the restriction is
// kept where the name is read: only a nonbasic land can be affected, and
// it is matched by the name it has now (game.PermanentHasName, the face
// that is up, CR 712.8f), not by every face's. A name no legal choice
// could make (a basic land's, an instant front face's) then reaches no
// land, which is exactly what a legal name no opponent's land has does.
// Nothing is ever stronger than the card. "Your opponents" is read
// against this enchantment's controller.
//
// No simplification.
const alpineMoonAnyColor = "alpine-moon/any-color"

func init() {
	Register(Spec{
		OracleID:     "8b46b50c-f824-4c4e-86de-38065c6f9a64",
		Name:         "Alpine Moon",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCardNameAsEnters("Alpine Moon — choose a nonbasic land card name"),
		Grants:       []AbilityGrant{AnyColorManaGrant(alpineMoonAnyColor)},
		Static: []game.StaticAbility{
			LosesAllLandTypes(alpineMoonApplies),
			LosesAllAbilitiesAndHas(alpineMoonApplies, alpineMoonAnyColor),
		},
	})
}

// alpineMoonApplies is "lands your opponents control with the chosen
// name": a nonbasic land, controlled by an opponent of this
// enchantment's controller, whose name is the one chosen for this
// enchantment.
func alpineMoonApplies(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsLand() && !target.HasSupertype("basic") &&
		target.Controller != source.Controller &&
		game.PermanentHasName(*target, source.ChosenName)
}
