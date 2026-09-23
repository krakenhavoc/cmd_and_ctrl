package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch04_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 04 (#297, `edhrec_rank` 472–577).
//
// Its own file rather than helpers.go, per the convention #231 set:
// concurrent card batches collide on shared helper files. Every
// package-level name here carries a b04 prefix for the same reason —
// batches 02 and 03 are being written at the same time.

// b04BounceLand is the Ravnica "bounce land" / karoo shape:
//
//	"This land enters tapped.
//	 When this land enters, return a land you control to its owner's
//	 hand.
//	 {T}: Add {A}{B}."
//
// Azorius Chancery (Aang deck) is the template: the real CR 614
// self-replacement (SelfEntersTapped) for the tapped entry, and
// ReturnOneYouControl for the bounce.
//
// The bounce is a CHOICE, not a target — "return a land you control",
// with no "target" in the printed text — made on resolution
// (ReturnOneYouControl, the own_permanents prompt, #1214). It used to
// be a target clause picked when the trigger went on the stack, a
// declared simplification that let opponents see and answer the
// choice before the resolution-time picker existed. The land is
// always a candidate while it is still on the battlefield, and
// returning itself is a normal, sometimes correct, line.
func b04BounceLand(oracleID, name, produced string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: produced,
			Label:    "Add " + produced,
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(name+" — return a land you control", Do(ReturnOneYouControl{
				Match:    MatchLand,
				Question: name + " — return a land you control to its owner's hand",
			})),
		},
	}
}

// b04NontokenCreature is "a nontoken creature" — Accursed Marauder's
// edict clause, the one word that separates it from Fleshbag
// Marauder: a Treasure deck's Zombie tokens cannot be fed to it.
func b04NontokenCreature(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.IsCreature() && !IsToken(c)
}

// b04LandNamesControlled counts the lands `controller` controls
// with DIFFERENT names — Field of the Dead's seven. Reads the printed
// name; two copies of the same land count once, as printed.
func b04LandNamesControlled(g *game.Game, controller uuid.UUID) int {
	seen := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsLand() {
			seen[c.Name] = true
		}
	}
	return len(seen)
}

// b04CreaturesControlled counts the creatures `controller` controls
// on the printed type line — Adeline's characteristic-defining
// power. Counts Adeline herself, as printed ("creatures you
// control", not "other"). Walks the live slice: it runs inside a
// layer recompute.
func b04CreaturesControlled(g *game.Game, controller uuid.UUID) int {
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && c.IsCreature() {
			n++
		}
	}
	return n
}

// b04CreatureIDsControlledBy snapshots the instance IDs of the
// creatures `controller` controls, for effects that mutate the
// battlefield while iterating it (Cathars' Crusade's counters).
func b04CreatureIDsControlledBy(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// b04OpponentLostLife reports whether ev is an opponent of
// `controller` losing life, and how much — Exquisite Blood's
// trigger. Two event kinds carry a life loss: EventChangeLife with a
// negative Amount (a drain, a "loses N life" effect, a Bond) and
// EventDealDamage whose Target is a player (a Bolt, combat damage),
// because the damage paths write the life total directly and emit
// no EventChangeLife of their own. Neither kind is emitted for the
// other's loss, so nothing is counted twice.
func b04OpponentLostLife(ev game.Event, controller uuid.UUID, g *game.Game) (int, bool) {
	if ev.Target == uuid.Nil || ev.Target == controller || g.PlayerByIDForEffect(ev.Target) == nil {
		return 0, false
	}
	switch ev.Kind {
	case game.EventChangeLife:
		if ev.Amount < 0 {
			return -ev.Amount, true
		}
	case game.EventDealDamage:
		if ev.Amount > 0 {
			return ev.Amount, true
		}
	}
	return 0, false
}
