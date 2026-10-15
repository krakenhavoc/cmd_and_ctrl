package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sepulchral Primordial — Creature — Avatar {5}{B}{B}, 5/4 (deck
// request #3655):
//
//	"Intimidate
//	 When this creature enters, for each opponent, you may put up to one
//	 target creature card from that player's graveyard onto the
//	 battlefield under your control."
//
// Molten Primordial's per-opponent clause list, aimed at graveyards:
// one "up to one" clause per opponent seat, each bound to that player's
// graveyard by ownership, so the CR 608.2b re-check asks "is it still in
// THAT player's graveyard?". Every clause is optional, so the trigger
// always goes on the stack and takes what it can. The picks enter
// together under the controller's control, in announce order.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Sepulchral Primordial — put up to one creature card from each opponent's graveyard onto the battlefield under your control",
		sepulchralPrimordialEffect)
	// Like Molten Primordial the list reads only the seat list and the
	// entering object's controller, neither of which can change while
	// the trigger waits, so the row is a restore point.
	enters.TargetsFrom = sepulchralPrimordialClauses
	Register(Spec{
		OracleID:        "480454a0-fb10-432b-a963-8b0c0ffc6680",
		Name:            "Sepulchral Primordial",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}

// sepulchralPrimordialClauses is one clause per opponent seat, in seat
// order, naming the player so the picker can tell the graveyards apart.
func sepulchralPrimordialClauses(tc game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	controller := source.Controller
	if obj := tc.Object; obj != nil && obj.ID == source.InstanceID && obj.Controller != uuid.Nil {
		controller = obj.Controller
	}
	var clauses []*game.TargetSpec
	for _, p := range g.Seats {
		if p == nil || p.ID == controller {
			continue
		}
		clauses = append(clauses, TargetCardInGraveyard("up to one target creature card in "+p.Name+"'s graveyard",
			Creature(), OwnedBy(p.ID)).WithCount(0, 1))
	}
	if len(clauses) == 0 {
		return nil
	}
	return Clauses(clauses[0], clauses[1:]...)
}

// sepulchralPrimordialEffect puts every still-legal pick onto the
// battlefield under the controller's control.
func sepulchralPrimordialEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
