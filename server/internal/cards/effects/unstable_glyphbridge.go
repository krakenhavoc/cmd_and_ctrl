package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unstable Glyphbridge // Sandswirl Wanderglyph — a transforming
// artifact (#2719, ADR 0137 for craft, ADR 0063 and ADR 0066's
// 2026-10-10 amendments for the back face):
//
//	Unstable Glyphbridge — Artifact {3}{W}{W}
//	  "When this artifact enters, if you cast it, for each player,
//	   choose a creature with power 2 or less that player controls.
//	   Then destroy all creatures except creatures chosen this way.
//	   Craft with artifact {3}{W}{W}"
//	Sandswirl Wanderglyph — Artifact Creature — Golem, 5/3
//	  "Flying
//	   Whenever an opponent casts a spell during their turn, they can't
//	   attack you or planeswalkers you control this turn.
//	   Each opponent who attacked you or a planeswalker you control this
//	   turn can't cast spells."
//
// The front face is The Eternal Wanderer's −4 with a power bound and a
// destroy: the Glyphbridge's controller picks on every board, APNAP
// from the active player, one creature with power 2 or less where the
// player has one (the ruling: "you must choose … if you can"), and then
// every other creature is destroyed as one event. "If you cast it" is
// the intervening-if every cast-gated entry reads (b16EnteredFromStack).
// Requested for the Sami Whammy deck (#2190).
//
// The back face's trigger is a this-turn grant on the casting opponent,
// scoped to you and your planeswalkers (a battle you control stays
// open, per the ruling). It lasts the turn even if the Wanderglyph
// leaves. The cast ban is a static on the Wanderglyph, read off this
// turn's attack record, so it ends the moment the Wanderglyph leaves
// (the other ruling).
//
// No simplification.
const unstableGlyphbridgeOracleID = "fcd54631-ea47-49a7-ad5f-9a9a51a815ba"

func init() {
	Register(Spec{
		OracleID:     unstableGlyphbridgeOracleID,
		Name:         "Unstable Glyphbridge",
		Completeness: CompletenessFull,
		// ADR 0126 §6: one creature with power 2 or less survives on each board.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && b16EnteredFromStack(g, source.InstanceID)
			}, "Unstable Glyphbridge — choose a creature with power 2 or less for each player, destroy the rest", unstableGlyphbridgeEnters),
		},
		Activated: []ActivatedAbility{
			Craft("Craft with artifact {3}{W}{W}", "{3}{W}{W}", CraftWith("artifact")),
		},
	})

	Register(Spec{
		OracleID:        unstableGlyphbridgeOracleID + "#1",
		Name:            "Sandswirl Wanderglyph",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15OpponentCastSpell(ev, source) && isActivePlayer(g, ev.Actor)
			}, "Sandswirl Wanderglyph — they can't attack you or planeswalkers you control this turn", sandswirlWanderglyphRestrict),
		},
		CastRestrictions: []game.CastRestriction{
			OpponentsWhoAttackedYouCantCast("Each opponent who attacked you or a planeswalker you control this turn can't cast spells"),
		},
	})
}

// unstableGlyphbridgeEnters is the choice and the destruction.
func unstableGlyphbridgeEnters(g *game.Game, item *game.StackItem) error {
	return ChoosePermanents{
		Question:   "Unstable Glyphbridge — choose the creature with power 2 or less that player keeps",
		Of:         seatsFromActive(g),
		Candidates: unstableGlyphbridgeCandidates,
		Then:       destroyCreaturesNotChosen,
	}.Apply(NewContext(g, item))
}

// unstableGlyphbridgeCandidates is one player's creatures with power 2
// or less, exactly one to be chosen. A player with none is skipped.
//
// Caller holds g.mu.
func unstableGlyphbridgeCandidates(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var ids []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == of && c.IsCreature() && c.CurrentPower() <= 2 {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil, 0, 0
	}
	return ids, 1, 1
}

// destroyCreaturesNotChosen is "then destroy all creatures except
// creatures chosen this way", as one event.
//
// Caller holds g.mu.
func destroyCreaturesNotChosen(ctx *Context, picked game.PromptedPicks) error {
	g := ctx.Game
	var doomed []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsCreature() && !picked.Contains(c.InstanceID) {
			doomed = append(doomed, c.InstanceID)
		}
	}
	if len(doomed) > 0 {
		g.DestroyPermanentsForEffect(doomed)
	}
	return nil
}

// sandswirlWanderglyphRestrict is the trigger's resolution: the
// opponent who cast the spell can't attack the Wanderglyph's controller
// or their planeswalkers for the rest of this turn.
func sandswirlWanderglyphRestrict(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	g.GrantCantAttackPlayerThisTurnForEffect(item.Trigger.Event.Actor, item.Controller,
		game.CantAttackScope{PlaneswalkersOnly: true},
		"Sandswirl Wanderglyph", item.SourceCardID)
	return nil
}
