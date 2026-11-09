package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terra, Magical Adept // Esper Terra (Final Fantasy, #4462) — a
// transforming legendary creature whose back face is a Saga creature.
//
// Front face, Legendary Creature — Human Wizard Warrior {1}{R}{G}, 4/2:
//
//	"When Terra enters, mill five cards. Put up to one enchantment card
//	 milled this way into your hand.
//	 Trance — {4}{R}{G}, {T}: Exile Terra, then return it to the
//	 battlefield transformed under its owner's control. Activate only as a
//	 sorcery."
//
// Back face, Esper Terra — Legendary Enchantment Creature — Saga Wizard,
// 6/6, flying (esper_terra.go).
//
// The mill's pick is over the enchantment cards that reached the graveyard,
// up to one. Trance is Jin-Gitaxias' flip with a mana-and-tap cost
// (ExileAndReturnTransformed, a new object under its owner's control).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     terraMagicalAdeptOracleID,
		Name:         "Terra, Magical Adept",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Terra, Magical Adept — mill five cards, then put up to one enchantment card milled this way into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 5, ThenTail: ctx.ItemTail(terraMagicalAdeptMilled)}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "Trance — {4}{R}{G}, {T}: Exile Terra, then return it to the battlefield transformed under its owner's control. " +
				"Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{4}{R}{G}"), TapCost()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}

// terraMagicalAdeptOracleID is shared with the back face, which registers
// under it plus "#1" (game.CatalogKey).
const terraMagicalAdeptOracleID = "54ca5904-9099-497a-9684-101656025487"
