package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Joshua, Phoenix's Dominant // Phoenix, Warden of Fire (Final Fantasy,
// #4462) — a transforming legendary creature whose back face is a Saga
// creature.
//
// Front face, Legendary Creature — Human Noble Wizard {1}{R}{W}, 3/4:
//
//	"When Joshua enters, discard up to two cards, then draw that many
//	 cards.
//	 {3}{R}{W}, {T}: Exile Joshua, then return it to the battlefield
//	 transformed under its owner's control. Activate only as a sorcery."
//
// Back face, Phoenix, Warden of Fire — Legendary Enchantment Creature —
// Saga Phoenix, 4/4, flying and lifelink (phoenix_warden_of_fire.go).
//
// The enters trigger is Greasewrench Goblin's rummage: a genuine "up to",
// including zero, drawing as many as were really discarded. The flip is
// Jin-Gitaxias' (ExileAndReturnTransformed, a new object under its owner's
// control) with a mana-and-tap cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     joshuaPhoenixsDominantOracleID,
		Name:         "Joshua, Phoenix's Dominant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Joshua, Phoenix's Dominant — discard up to two cards, then draw that many cards",
				func(g *game.Game, item *game.StackItem) error {
					return b39MayDiscardThenDraw(2, false,
						"Joshua, Phoenix's Dominant — discard up to two cards, then draw that many",
						drawAsManyAsDiscarded)(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}{R}{W}, {T}: Exile Joshua, then return it to the battlefield transformed under its owner's control. " +
				"Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}{R}{W}"), TapCost()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}

// joshuaPhoenixsDominantOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const joshuaPhoenixsDominantOracleID = "f4629329-a79b-4460-a1bc-9ae0f59adf5a"
