package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Assimilation Assessor — Artifact Creature — Borg {4}{B}, 5/5 (Star Trek
// Commander, #4974):
//
//	"When this creature enters, exile target opponent's graveyard.
//	 Whenever this creature attacks, assimilate a creature card exiled
//	 with it. (Put it onto the battlefield under your control with a +1/+1
//	 counter. It's a Borg artifact creature and loses all other creature
//	 types.)"
//
// The enters ability exiles every card in the target opponent's graveyard
// at once, LINKED to the Assessor as an object (CR 607.2a,
// ExileLinkedKeyedForEffect), so each card carries Card.ExiledWith. A
// target that is no longer an opponent's, or has left the game, makes the
// ability do nothing (CR 608.2b). An Assessor that left and came back
// before the ability resolved is a new object (CR 400.7): the cards are
// still exiled, but nothing links them to it.
//
// The attack trigger reads the creature cards still in exile linked to
// this object, as it resolves. It is not a target and not a "may": with
// two or more on offer the controller picks exactly one, with one it is
// taken, with none nothing happens. The card is assimilated straight out
// of exile (trekB11Assimilate's entry: a +1/+1 counter, a Borg artifact
// creature that keeps its other card types, CR 205.1b). The trigger reads
// the link off the object that attacked, so it still works if the
// Assessor has left the battlefield by the time it resolves.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Assimilation Assessor — exile target opponent's graveyard", assimilationAssessorExile)
	enters.Targets = TargetPlayer("target opponent", Opponent())
	Register(Spec{
		OracleID:     "48848570-00c4-4a89-a84c-e3df47a8b0d2",
		Name:         "Assimilation Assessor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			enters,
			WheneverThisAttacks("Assimilation Assessor — assimilate a creature card exiled with it", assimilationAssessorAssimilate),
		},
	})
}

// assimilationAssessorExile exiles the target opponent's whole graveyard
// as one exile, linked to this object.
func assimilationAssessorExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	link, _ := ctx.SourceRef()
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		p := g.PlayerByIDForEffect(t.ID)
		if p == nil || p.Graveyard == nil || len(p.Graveyard.Cards) == 0 {
			continue
		}
		ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
		for _, c := range p.Graveyard.Cards {
			ids = append(ids, c.InstanceID)
		}
		if err := g.ExileLinkedKeyedForEffect(link, ids, game.KeyedTail{}); err != nil {
			return err
		}
	}
	return nil
}

// assimilationAssessorAssimilate is the attack trigger: one creature card
// exiled with this object, chosen as it resolves, assimilated.
func assimilationAssessorAssimilate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	link, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var cands []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(link)) {
		if c.IsCreature() && !c.IsToken() {
			cands = append(cands, c.InstanceID)
		}
	}
	switch len(cands) {
	case 0:
		return nil
	case 1:
		return trekB36AssimilateTogether(ctx, game.ZoneExile, cands, assimilationAssessorLabel)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Assimilation Assessor — assimilate a creature card exiled with it",
		Cards:    cands,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  assimilationAssessorPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

const assimilationAssessorLabel = "Assimilation Assessor — assimilated: a Borg artifact creature"

// assimilationAssessorPick is the chosen card, assimilated from exile. An
// on-disk key: never renamed, never reused.
var assimilationAssessorPick = askedPick("choose-cards/assimilation-assessor-assimilate", func(ctx *Context, r game.CardsChosen) error {
	return trekB36AssimilateTogether(ctx, game.ZoneExile, r.Picked, assimilationAssessorLabel)
})
