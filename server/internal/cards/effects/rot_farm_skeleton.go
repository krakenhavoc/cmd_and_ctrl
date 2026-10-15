package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rot Farm Skeleton — Creature — Plant Skeleton {2}{B}{G}, 4/1:
//
//	"This creature can't block.
//	 {2}{B}{G}, Mill four cards: Return this card from your graveyard
//	 to the battlefield. Activate only as a sorcery."
//
// Carrion Feeder's "can't block" (RestrictSelf, enforced by the combat
// engine and never offered by the enumerator) plus Reassembling
// Skeleton's graveyard activation (Zones, CR 113.6) with a mill cost.
// "Mill four cards" is MillCardsCost(4) (#3845, ADR 0020's 2026-10-15
// amendment): validated before anything is paid, so a library of fewer
// than four cards refuses the activation (CR 701.17b, CR 118.3), and
// paid last (CR 601.2h). SorcerySpeed is CR 602.5d: the activator's
// main phase, empty stack.
//
// Unlike Reassembling Skeleton it returns untapped. If the card has
// left the graveyard by resolution (Bojuka Bog), the ability does
// nothing (CR 608.2b); the cost is not refunded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f7b2c135-1b99-4b69-89d9-3fef25b509d4",
		Name:         "Rot Farm Skeleton",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			RestrictSelf(game.CantBlock),
		},
		Activated: []ActivatedAbility{{
			Label:        "{2}{B}{G}, Mill four cards: Return this card from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{B}{G}"), MillCardsCost(4)),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       returnThisFromGraveyardToBattlefield,
		}},
	})
}
