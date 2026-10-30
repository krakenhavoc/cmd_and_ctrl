package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Costume Closet — Artifact {1}{W} (Marvel Super Heroes Commander,
// #4455):
//
//	"This artifact enters with two +1/+1 counters on it.
//	 {T}: Move a +1/+1 counter from this artifact onto target creature
//	 you control. Activate only as a sorcery.
//	 Whenever a modified creature you control leaves the battlefield, put
//	 a +1/+1 counter on this artifact. (Equipment, Auras you control, and
//	 counters are modifications.)"
//
// The move is Simic Fluxmage's (the counter lands first, then leaves the
// Closet only if it was put on, CR 701.34). "Modified" is read as the
// creature last existed on the battlefield (CR 603.10a): its last-known
// counters, plus an Equipment or an Aura its controller controlled that
// was attached to it. The Closet may leave with the creature (the
// trigger is a leaves-the-battlefield look-back for itself too), but a
// Closet that has already gone puts no counter on anything (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "072d05c4-e0d7-47af-b2a9-16b90810fd30",
		Name:         "Costume Closet",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 2, "Costume Closet: enters with two +1/+1 counters"),
		},
		Activated: []ActivatedAbility{{
			Label:        "{T}: Move a +1/+1 counter from this artifact onto target creature you control. Activate only as a sorcery.",
			Cost:         TapCost(),
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature you control", YouControl()),
			Effect:       simicFluxmageMove,
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && leftAsType(ev, c, "creature") && leftUnderControlOf(ev, c) == source.Controller &&
					msc17WasModified(g, ev.CardID)
			}, "Costume Closet — put a +1/+1 counter on this artifact", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// msc17WasModified is "modified" for a permanent that has just left the
// battlefield (CR 700.9, 603.10a): it had a counter, or an Equipment, or
// an Aura its controller controlled, attached to it as it last existed.
func msc17WasModified(g *game.Game, cardID uuid.UUID) bool {
	ref, ok := g.PermanentRefForEffect(cardID)
	if !ok {
		return false
	}
	info, ok := g.PermanentForEffect(ref)
	if !ok {
		return false
	}
	for _, n := range info.Counters {
		if n > 0 {
			return true
		}
	}
	for _, att := range info.Attachments {
		a, ok := g.LookupCardForEffect(att.ID)
		if !ok {
			continue
		}
		if a.HasSubtype("Equipment") || (a.IsAura() && a.Controller == info.Controller) {
			return true
		}
	}
	return false
}
