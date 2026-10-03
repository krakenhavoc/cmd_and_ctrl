package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unbreathing Horde — Creature — Zombie {2}{B}, 0/0:
//
//	"This creature enters with a +1/+1 counter on it for each other Zombie you control and each Zombie card in your graveyard.
//	 If this creature would be dealt damage, prevent that damage and remove a +1/+1 counter from it."
//
// ADR 0108 §8 (#1906): a prevention static with an additional effect, one
// application per recipient in a damage instance, so only one counter is
// removed however much damage is prevented, and with no counter left the
// damage is still prevented (the rulings). Damage that can't be prevented
// still removes one (CR 615.12).
//
// The entry count is read as it enters, before the move: other Zombies on
// the battlefield under its controller, and Zombie cards in its
// controller's graveyard — itself included when it enters from there (the
// ruling).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "e6cd9203-e4d3-4d9f-b59f-4e454fc5a477",
		Name:         "Unbreathing Horde",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
					return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
						src != nil && ev.CardID == src.InstanceID
				},
				Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
					ev.AddCounterAtETB(game.CounterPlusOne, unbreathingHordeCount(g, src))
					return nil
				},
				Label: "Unbreathing Horde: enters with a +1/+1 counter for each other Zombie you control and Zombie card in your graveyard",
			},
			PreventDamageDealtTo(PreventionStatic{
				To:    ToThisCreature,
				Then:  removeACounterFromThisBody,
				Label: "Unbreathing Horde — prevent damage to it and remove a +1/+1 counter",
			}),
		},
	})
}

// unbreathingHordeCount is the entry count: other Zombies `src`'s
// controller controls, plus Zombie cards in their graveyard.
func unbreathingHordeCount(g *game.Game, src *game.Card) int {
	you := src.Controller
	if you == uuid.Nil {
		you = src.Owner
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID != src.InstanceID && c.Controller == you && c.HasSubtype("Zombie") {
			n++
		}
	}
	if p := g.PlayerByIDForEffect(you); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			if c.HasSubtype("Zombie") {
				n++
			}
		}
	}
	return n
}
