package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grim Reaper's Scythe — Legendary Artifact {2}{B} (Marvel Super Heroes
// Commander, #4455):
//
//	"Whenever one or more creature cards leave your graveyard, create a
//	 2/2 black Zombie creature token.
//	 {3}{B}, {T}, Sacrifice two creatures: Return target creature card
//	 from your graveyard to the battlefield with a finality counter on it.
//	 Activate only as a sorcery. (If a creature with a finality counter on
//	 it would die, exile it instead.)"
//
// The trigger is a batch trigger (OncePerBatch): a mass reanimation makes
// one Zombie, and a card leaves in any direction. The activation sacrifices
// two creatures as a cost, then returns the target with the finality counter
// as an entry counter (CR 614.1c); the exile-instead is the engine's
// replacement (CR 122.1h, game/finality_counters.go). If the target is gone
// by resolution nothing returns (CR 608.2b), but the cost stays paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91b1db8c-4bce-43d1-94ab-14b1435686c7",
		Name:         "Grim Reaper's Scythe",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventZoneMove, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b02CreatureCardLeftYourGraveyard(ev, source, g)
			}, "Grim Reaper's Scythe — create a 2/2 black Zombie",
				Do(CreateToken{Template: BlackZombieToken(), N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}{B}, {T}, Sacrifice two creatures: Return target creature card from your graveyard to the battlefield with a finality counter on it. Activate only as a sorcery.",
			Purpose:      ForTargets(Reanimates(0)),
			Cost:         Plus(ManaCost("{3}{B}"), TapCost(), SacrificeN(2, "two creatures", Creature())),
			SorcerySpeed: true,
			Targets:      targetCreatureInYourGraveyard(),
			Effect:       grimReapersScytheReturn,
		}},
	})
}

func grimReapersScytheReturn(g *game.Game, item *game.StackItem) error {
	ids := legalTargetCardIDs(NewContext(g, item))
	if len(ids) == 0 {
		return nil
	}
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(ids[0], item.Controller, false,
		map[string]int{game.CounterFinality: 1}, game.KeyedTail{})
}
