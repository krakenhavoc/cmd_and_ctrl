package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scrapheap Scrounger — Artifact Creature — Construct {2}, 3/2 (Kaladesh,
// #5539):
//
//	"This creature can't block.
//	 {1}{B}, Exile another creature card from your graveyard: Return this card
//	 from your graveyard to the battlefield."
//
// Rot Farm Skeleton's "can't block" and graveyard activation, with Bone
// Dragon's exile cost: one other creature card named at announce (the source
// is never a legal pick, so "another" needs no predicate), exiled as a cost
// (CR 602.2b, 601.2h). Any speed: the card prints no sorcery-speed clause. If
// the Scrounger has left the graveyard by resolution the ability does nothing
// (CR 608.2b); the cost is not refunded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d309e1e9-26e6-4a10-a3f4-352c7ba3f2ff",
		Name:         "Scrapheap Scrounger",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.RestrictSelf(game.CantBlock),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Exile another creature card from your graveyard: Return this card from your graveyard to the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: effects.Plus(effects.ManaCost("{1}{B}"),
				effects.ExileFromGraveyard(1, "another creature card", func(c game.Card) bool { return c.IsCreature() })),
			Zones:  []game.ZoneKind{game.ZoneGraveyard},
			Effect: effects.ReturnThisFromGraveyardToBattlefield,
		}},
	})
}
