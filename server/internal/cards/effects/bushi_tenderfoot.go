package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bushi Tenderfoot // Kenzo the Hardhearted — {W} Creature — Human
// Soldier 1/1 // Legendary Creature — Human Samurai 3/4, a flip card
// (#3216, ADR 0151):
//
//	Top:     "When a creature dealt damage by this creature this turn
//	          dies, flip this creature."
//	Flipped: "Double strike; bushido 2 (Whenever this creature blocks or
//	          becomes blocked, it gets +2/+2 until end of turn.)"
//
// Audited against the engine (the issue's "check"): "dealt damage by
// this creature this turn" is read off the turn's damage events, which
// name the source and the damaged creature, as DealtDamageThisTurn does
// for the victim alone. The source and the victim each start a new
// record when they re-enter the battlefield (CR 400.7), so a creature
// that was dealt damage, blinked and died is not the one that was
// damaged. The trigger is "a creature dies", not "another", but a
// Tenderfoot that dies with its victim is gone and has nothing to flip.
//
// Kenzo is white and {W} (CR 710.1c), and has double strike and bushido
// 2 as its own half's text (ADR 0151 decision 3).
//
// No simplification.
const bushiTenderfootOracleID = "82959ca2-cd96-4cca-9ce0-afb8db209860"

// damagedBySinceEntering reports whether `source` dealt damage to
// `victim` this turn, counting only the objects they are now: either
// one entering the battlefield again starts over (CR 400.7). The damage
// is the amount that was really dealt, after prevention.
func damagedBySinceEntering(g *game.Game, source, victim uuid.UUID) bool {
	dealt := false
	for _, ev := range g.EventsThisTurn() {
		switch {
		case ev.Kind == game.EventZoneMove && ev.NewZone == game.ZoneBattlefield && (ev.CardID == source || ev.CardID == victim):
			dealt = false
		case ev.Kind == game.EventDealDamage && ev.Source == source && ev.Target == victim && ev.Amount > 0:
			dealt = true
		}
	}
	return dealt
}

func init() {
	Register(Spec{
		OracleID:     bushiTenderfootOracleID,
		Name:         "Bushi Tenderfoot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := diedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && damagedBySinceEntering(g, source.InstanceID, ev.CardID)
			}, "Bushi Tenderfoot — flip it", Do(FlipThis{})),
		},
	})
	Register(Spec{
		OracleID:        bushiTenderfootOracleID + "#1",
		Name:            "Kenzo the Hardhearted",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered:       []game.TriggeredAbility{Bushido(2, "Kenzo the Hardhearted")},
	})
}
