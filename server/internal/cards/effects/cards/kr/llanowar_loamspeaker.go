package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Llanowar Loamspeaker — Creature — Elf Druid {1}{G}, 1/3:
//
//	"{T}: Add one mana of any color.
//	 {T}: Target land you control becomes a 3/3 Elemental creature with
//	 haste until end of turn. It's still a land. Activate only as a
//	 sorcery."
//
// The target land is checked again as the ability resolves (CR 608.2b).
// The land keeps its colours (none are set) and, being a creature with
// haste, may attack the turn it arrives; its own mana ability is then
// refused only if it is summoning sick, which haste cures (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08377424-619e-48ae-bc3c-fb199cd77cf9",
		Name:         "Llanowar Loamspeaker",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Target land you control becomes a 3/3 Elemental creature with haste until end of turn. It's still a land. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit}),
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Targets:      effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
			Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.Cmdr2B13Animate(ctx, id,
					"Llanowar Loamspeaker — a 3/3 Elemental creature with haste until end of turn",
					3, 3, "Elemental", nil, "haste")
			})),
		}},
	})
}
