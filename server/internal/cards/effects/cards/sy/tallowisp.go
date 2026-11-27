package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tallowisp — Creature — Spirit {1}{W}, 1/3 (Betrayers of Kamigawa, #5602):
//
//	"Whenever you cast a Spirit or Arcane spell, you may search your library
//	 for an Aura card with enchant creature, reveal it, put it into your hand,
//	 then shuffle."
//
// The cast trigger is the ki cycle's (ChkB1SpiritOrArcane). An Aura "with
// enchant creature" is read off its catalogued enchant clause, which must be
// a creature clause (EnchantCreature); a library card has no oracle text, so
// an Aura with no catalogued clause is a manual object in this sandbox and is
// allowed, as Auratouched Mage's search allows it. The search reveals the
// card and shuffles whether or not one is found.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25187c05-ed6a-4690-b92b-89c532fdb780",
		Name:         "Tallowisp",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), "Tallowisp — search your library for an Aura card with enchant creature, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							if !c.HasSubtype("Aura") {
								return false
							}
							spec := game.TargetSpecFor(game.CatalogKey(c))
							return spec == nil || spec.Mode == "creature"
						},
						Dest:     game.ZoneHand,
						Limit:    1,
						Reveal:   true,
						Shuffle:  true,
						Optional: true,
						Reason:   "Tallowisp — an Aura card with enchant creature",
					}.Apply(effects.NewContext(g, item))
				}),
				"Tallowisp — search your library for an Aura card with enchant creature?"),
		},
	})
}
