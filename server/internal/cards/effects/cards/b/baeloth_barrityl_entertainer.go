package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Baeloth Barrityl, Entertainer — Legendary Creature — Elf Shaman {4}{R},
// 2/5 (Commander Legends: Battle for Baldur's Gate, S82 GAP2 b11, #6098):
//
//	"Creatures your opponents control with power less than Baeloth
//	 Barrityl's power are goaded.
//	 Whenever a goaded attacking or blocking creature dies, you create a
//	 Treasure token.
//	 Choose a Background"
//
// The goad is game.PowerGatedGoad: the static reaches each creature an
// opponent controls in layer 6, and the engine compares the two finished
// powers (layers and P/T counters) once every layer has run, because goad
// is an effect on the game rules (CR 613.11) and not a characteristic. A
// creature pumped to Baeloth's power, or a Baeloth shrunk, stops being
// goaded at once; one already declared as an attacker keeps attacking (CR
// 506.4).
//
// The trigger reads the dying creature as it last existed (CR 603.10a):
// attacking or blocking (Event.AttackingTarget / BlockingTarget), and
// goaded by anyone, by a goad that resolved or by a continuous one
// (PermanentInfo.Goaded). Baeloth's own goad counts, so a goaded creature
// that dies together with Baeloth still triggers it (CR 603.10a).
//
// Choose a Background is deck construction (CR 702.124k).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33b401b1-cf94-4c7a-8a8a-426ac3d87bd0",
		Name:         "Baeloth Barrityl, Entertainer",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Creatures your opponents control with power less than Baeloth Barrityl's power are goaded.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller != source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				c.PowerGatedGoads = append(c.PowerGatedGoads, game.PowerGatedGoad{Source: source.InstanceID, By: source.Controller})
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, baelothGoadedCombatantDied,
				"Baeloth Barrityl, Entertainer — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}

// baelothGoadedCombatantDied is "a goaded attacking or blocking creature
// dies": a creature put into a graveyard from the battlefield that was
// attacking or blocking, and goaded, as it last existed.
func baelothGoadedCombatantDied(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	if ev.AttackingTarget == uuid.Nil && ev.BlockingTarget == uuid.Nil {
		return false
	}
	if was, known := ev.WasType("creature"); known && !was {
		return false
	}
	rec, ok := g.LastKnownPermanentForEffect(ev.CardID)
	return ok && rec.Goaded
}
