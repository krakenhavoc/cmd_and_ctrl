package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rubblebelt Maverick — Creature — Human Detective {G}, 1/1 (Murders at
// Karlov Manor, #4469):
//
//	"When this creature enters, surveil 2.
//	 {G}, Exile this card from your graveyard: Put a +1/+1 counter on
//	 target creature. Activate only as a sorcery."
//
// The enters trigger is the shared surveil. The graveyard ability is
// Scavenge's shape (ExileThis from the graveyard zone, sorcery speed): the
// card is exiled as a cost, so it is gone whatever happens to the target,
// and a target that has become illegal gets no counter (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ba59349-a162-4444-b3e0-cddfc01ef03b",
		Name:         "Rubblebelt Maverick",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rubblebelt Maverick — surveil 2", effects.Do(effects.Surveil{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{G}, Exile this card from your graveyard: Put a +1/+1 counter on target creature. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{G}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return effects.AddCounter{Target: ts[0].ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			},
		}},
	})
}
