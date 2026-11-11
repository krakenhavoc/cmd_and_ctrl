package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Diluvian Primordial — Creature — Avatar {5}{U}{U}, 5/5:
//
//	"Flying
//	 When this creature enters, for each opponent, you may cast up to one
//	 target instant or sorcery card from that player's graveyard without
//	 paying its mana cost. If a spell cast this way would be put into a
//	 graveyard, exile it instead."
//
// One "up to one target ... from that player's graveyard" clause per
// opponent (Molten Primordial's TargetsFrom shape), each bound to its
// opponent's graveyard, so a card that moved to another graveyard in
// response is an illegal target (CR 608.2b). Each pick gets its own free,
// exile-instead permission, and the casts are offered during the
// resolution (CR 608.2g), one at a time.
func init() {
	enters := WhenThisEnters("Diluvian Primordial — cast up to one instant or sorcery card from each opponent's graveyard", diluvianPrimordialEffect)
	enters.TargetsFrom = diluvianPrimordialClauses
	Register(Spec{
		OracleID:        "f21ce158-2925-4658-ab7b-b73718d31965",
		Name:            "Diluvian Primordial",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}

// diluvianPrimordialClauses is one clause per opponent seat, in seat order
// (moltenPrimordialClauses' contract: the list reads only the constant seat
// list and the controller as the Primordial entered).
func diluvianPrimordialClauses(tc game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	controller := source.Controller
	if obj := tc.Object; obj != nil && obj.ID == source.InstanceID && obj.Controller != uuid.Nil {
		controller = obj.Controller
	}
	var clauses []*game.TargetSpec
	for _, p := range g.Seats {
		if p == nil || p.ID == controller {
			continue
		}
		clauses = append(clauses, TargetCardInGraveyard(
			"up to one target instant or sorcery card from "+p.Name+"'s graveyard",
			Or(Instant(), Sorcery()), OwnedBy(p.ID)).WithCount(0, 1))
	}
	if len(clauses) == 0 {
		return nil
	}
	return Clauses(clauses[0], clauses[1:]...)
}

func diluvianPrimordialEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	cmdr2S42FreeGraveyardCasts(ctx, "Diluvian Primordial", freeInstantOrSorcery(), true, false, ids)
	return nil
}
