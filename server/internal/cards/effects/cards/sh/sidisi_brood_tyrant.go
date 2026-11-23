package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// b33SidisiZombieLabel is the stack label of Sidisi's mill trigger —
// the "one or more" dedup keys on it.
const b33SidisiZombieLabel = "Sidisi, Brood Tyrant — create a 2/2 black Zombie"

// Sidisi, Brood Tyrant — Legendary Creature — Snake Shaman
// {1}{B}{G}{U}, 3/3 (EDHREC rank 3524):
//
//	"Whenever Sidisi enters or attacks, mill three cards.
//	 Whenever one or more creature cards are put into your graveyard
//	 from your library, create a 2/2 black Zombie creature token."
//
// The self-mill commander. The first trigger is Sun Titan's "enters
// or attacks" shape with a three-card mill; the second is Colossal
// Grave-Reaver's: it watches every library-to-graveyard arrival, a
// mill's EventMill and a surveil's or a search's EventZoneMove (#2809,
// library_to_graveyard.go), one per card, so the first creature card
// of a batch fires it and the rest are declined as later events of the
// same batch (OncePerBatch; see docs/adding-cards.md). Two separate
// mills in one turn make two Zombies, as printed, and a creature card
// surveilled or Entombed into your graveyard counts too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3fad7072-21e3-446e-a28f-615038c8bfea",
		Name:         "Sidisi, Brood Tyrant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventAttack}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B21SelfEnteredOrAttacked(ev, source)
			}, "Sidisi, Brood Tyrant — mill three cards", effects.Do(effects.MillCards{N: 3})),
			effects.OncePerBatch(effects.OnAny(effects.LibraryToGraveyardKinds, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CardPutIntoYourGraveyardFromYourLibrary(ev, source, g, effects.IsCreatureCard)
			}, b33SidisiZombieLabel, effects.Do(effects.CreateToken{Template: effects.BlackZombieToken(), N: 1}))),
		},
	})
}
