package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Doctor's Tomb — Plane — Trenzalore (Doctor Who, #4870):
//
//	"If a creature would die, instead exile it and that creature's
//	 controller loses 2 life.
//	 Whenever chaos ensues, redistribute any number of players' life
//	 totals. (Each of those players gets one life total back.)"
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// static is a CR 614 replacement on any creature moving from the battlefield to
// the graveyard, read from the command zone: the creature is exiled, so it
// never died, and its controller as it left loses 2 life when the move
// finishes (ChangeLifeAfterReplacedMoveForEffect). The chaos ability
// (CR 311.7) is planesB2RedistributeLife: the planar controller chooses the
// players and which of their life totals each gets, a total never split (the
// rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc621e94-e7c7-49ea-9547-4ab041ea39f3",
		Name:         "The Doctor's Tomb",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.OldZone != game.ZoneBattlefield || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				c, _ := g.LookupCardForEffect(ev.CardID)
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return g.ChangeLifeAfterReplacedMoveForEffect(ev, src.InstanceID, c.Controller, -2)
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "The Doctor's Tomb: exile instead of dying, and its controller loses 2 life",
		}},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("The Doctor's Tomb — redistribute any number of players' life totals", Do(planesB2RedistributeLife{})),
		},
	})
}
