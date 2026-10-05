package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hideous Taskmaster — Creature — Eldrazi {6}{R}, 7/2:
//
//	"Devoid
//	 When you cast this spell, for each opponent, gain control of up to
//	 one target creature that player controls until end of turn. Untap
//	 those creatures. They gain trample, haste, and annihilator 1 until
//	 end of turn.
//	 Trample, haste, annihilator 1"
//
// Devoid is declared in PrintedKeywords beside the other three, and the
// engine reads it as CR 702.114a's colour-defining ability (#2152), so
// the card is colourless in every zone. The cast trigger is Molten
// Primordial's shape
// on a spell: one "up to one target creature that player controls"
// clause per opponent of the caster, in seat order, each pick bound to
// its player and re-checked at resolution (CR 608.2b). It resolves
// above the spell, so it happens even if the Taskmaster is countered
// (the 2024-06-07 ruling). Each creature taken untaps and gains
// trample, haste and annihilator 1 until end of turn — the keyword the
// engine turns into an attack trigger (ADR 0113 §2).
//
// No simplification.
func init() {
	cast := WhenYouCastThisSpell("Hideous Taskmaster — gain control of up to one creature each opponent controls",
		hideousTaskmasterSteal)
	// ADR 0041 P9: the clause list reads only the caster (off the
	// trigger's carried cast event) and the seat list, neither of which
	// a restore can change, so the row is not TargetsFromReadsBoard.
	cast.TargetsFrom = hideousTaskmasterClauses
	Register(Spec{
		OracleID:        "8f48c43e-fa70-4c4f-bb88-be0526303453",
		Name:            "Hideous Taskmaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "trample", "haste", "annihilator 1"},
		Triggered:       []game.TriggeredAbility{cast},
	})
}

// hideousTaskmasterClauses is one "up to one target creature that
// player controls" clause per opponent of the caster, in seat order.
func hideousTaskmasterClauses(tc game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	caster := tc.Event.Actor
	if caster == uuid.Nil {
		caster = source.Controller
	}
	var clauses []*game.TargetSpec
	for _, p := range g.Seats {
		if p == nil || p.ID == caster {
			continue
		}
		clauses = append(clauses, TargetCreature("up to one target creature "+p.Name+" controls",
			ControlledBy(p.ID)).WithCount(0, 1))
	}
	if len(clauses) == 0 {
		return nil
	}
	return Clauses(clauses[0], clauses[1:]...)
}

// hideousTaskmasterSteal takes each still-legal pick until end of turn,
// untaps it and gives it trample, haste and annihilator 1.
func hideousTaskmasterSteal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := threatenAndGrant(ctx, t.ID, "Hideous Taskmaster", "trample", "haste", "annihilator 1"); err != nil {
			return err
		}
	}
	return nil
}
