package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Molten Primordial — Creature — Avatar {5}{R}{R}, 6/4:
//
//	"Haste
//	 When this creature enters, for each opponent, gain control of up
//	 to one target creature that player controls until end of turn.
//	 Untap those creatures. They gain haste until end of turn."
//
// A seven-mana Insurrection that only takes one creature from each
// opponent — in a four-player game, three Act of Treasons at once.
//
// Filed against #1559 as a second set-level target rule, and triaged
// OUT of it (2026-09-24): "for each opponent, up to one target
// creature that player controls" is not a rule about the set, it is
// one CLAUSE PER OPPONENT, each binding its pick to that player. The
// machinery for that already existed — TriggeredAbility.TargetsFrom
// builds the clause list when the trigger is put on the stack
// (CR 603.3d) — so this card needed no engine change, and it is
// shipped beside #1559's cards because the Edea deck (#1565) wants it.
//
// The binding is the reason for the shape. Each clause's predicate
// names its opponent, and the item remembers the clauses it was
// announced under, so the CR 608.2b re-check asks "does THAT player
// still control it?". A creature that changed hands in response is
// an illegal target and is not taken, even if it went to another
// opponent. A set rule over one clause (EachDifferentController, as
// Windgrace's Judgment uses) would judge who controls it now instead.
//
// The clause list covers every seat but the controller's, in seat
// order, eliminated seats included (their permanents are gone, so
// their clause offers nothing and the picker skips it). Keeping the
// list's length independent of eliminations is what keeps each
// pick's clause index stable across the several times the dispatch
// asks TargetsFrom for it.
//
// Every clause is "up to one", so the trigger always goes on the
// stack and may take nothing.
//
// No simplifications.
func init() {
	enters := WhenThisEnters("Molten Primordial — gain control of up to one creature each opponent controls", moltenPrimordialEffect)
	enters.TargetsFrom = moltenPrimordialClauses
	Register(Spec{
		OracleID:        "8d8c9f7b-92c7-4284-ad9c-304ce42edba5",
		Name:            "Molten Primordial",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}

// moltenPrimordialClauses is one "up to one target creature that
// player controls" clause per opponent seat, in seat order. The label
// names the player, because the picker asks them one at a time and
// "up to one target creature that player controls" three times in a
// row would not say which player.
//
// Runs under g.mu (TargetsFrom's contract): reads g.Seats only.
func moltenPrimordialClauses(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	var clauses []*game.TargetSpec
	for _, p := range g.Seats {
		if p == nil || p.ID == source.Controller {
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

// moltenPrimordialEffect takes each still-legal pick until end of
// turn, untaps it and gives it haste — Act of Treason's three
// primitives, once per creature, in announce order.
func moltenPrimordialEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (GainControl{
			Target:   t.ID,
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "Molten Primordial — gain control until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
		if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{
			Target:   t.ID,
			Keywords: []string{"haste"},
			Label:    "Molten Primordial — haste until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
