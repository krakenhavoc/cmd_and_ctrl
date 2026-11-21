package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unshakable Tail — Creature — Zombie Detective {2}{B}, 3/2 (Murders at
// Karlov Manor Commander, #5368):
//
//	"When this creature enters and at the beginning of your upkeep,
//	 surveil 1.
//	 Whenever one or more creature cards are put into your graveyard
//	 from your library, investigate.
//	 {2}, Sacrifice a Clue: Return this card from your graveyard to your
//	 hand."
//
// The first is Obsessive Pursuit's two-event trigger (thisEntersOrYourUpkeep).
// The second is Colossal Grave-Reaver's: every library-to-graveyard
// arrival is seen, a mill's and a surveil's alike (library_to_graveyard.go),
// so Tail's own surveil can set it off, and one Clue per batch. The third
// is Metalwork Colossus's graveyard ability with a Clue as its cost; any
// Clue, token or not, pays it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d28e89d-d5e2-46b1-8dd2-d808b78f08fe",
		Name:         "Unshakable Tail",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventBeginUpkeep}, effects.ThisEntersOrYourUpkeep,
				"Unshakable Tail — surveil 1", effects.Do(effects.Surveil{N: 1})),
			{
				OncePerBatch: true,
				Watches:      effects.LibraryToGraveyardKinds,
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.CardPutIntoYourGraveyardFromYourLibrary(ev, source, g, effects.IsCreatureCard)
				},
				Key:    "Unshakable Tail — investigate",
				Effect: effects.Do(effects.Investigate{N: 1}),
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice a Clue: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeN(1, "a Clue", effects.OfSubtype("Clue"))),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
