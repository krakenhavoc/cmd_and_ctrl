package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// U.S.S. Reliant, From Hell's Heart — Legendary Artifact — Spacecraft
// {2}{B}{B}, 4/3 (Star Trek, #4973):
//
//	"Whenever U.S.S. Reliant enters or deals combat damage to a player,
//	 return target creature card with mana value 4 or less from your
//	 graveyard to the battlefield with a finality counter on it.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 10+.)
//	 10+ | Flying"
//
// One printed trigger with two conditions (Aerial Extortionist's
// b26SelfEnteredOrDealtCombatDamageToPlayer). It can only deal combat
// damage once it is a creature, at ten charge counters, but it enters at
// any count. The return is Grim Reaper's Scythe's: the target enters with
// a finality counter as an entry counter, and the exile-instead is the
// engine's (CR 122.1h). A target gone by resolution returns nothing (CR
// 608.2b); with no legal target the trigger is removed (CR 603.3d).
// Station, the 10+ type change and P/T box, and the flying grant are The
// Seriema's pieces (CR 702.184a, CR 721.2a).
//
// No simplification.
func init() {
	t := effects.OnAny([]game.EventKind{game.EventETB, game.EventDealDamage},
		func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return effects.B26SelfEnteredOrDealtCombatDamageToPlayer(ev, source, g)
		},
		"U.S.S. Reliant — return target creature card with mana value 4 or less with a finality counter",
		effects.GrimReapersScytheReturn)
	t.Targets = effects.TargetCardInGraveyard("target creature card with mana value 4 or less from your graveyard",
		effects.YouOwn(), effects.Creature(), effects.ManaValueLE(4))
	effects.Register(effects.Spec{
		OracleID:     "edaa05ba-8b15-4156-b725-1e0c418a9ae5",
		Name:         "U.S.S. Reliant, From Hell's Heart",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
		Activated:    []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(10, 4, 3),
			effects.ThresholdKeywords(10, "flying"),
		),
	})
}
