package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82BnsB28CountersOnYourCreatures totals the +1/+1 counters on the
// creatures the ability's controller controls (Deepwood Denizen).
func s82BnsB28CountersOnYourCreatures(q game.CostQuery) int {
	if q.Game == nil || q.Game.Battlefield == nil {
		return 0
	}
	n := 0
	for _, c := range q.Game.Battlefield.Cards {
		if c.Controller == q.Source.Controller && c.IsCreature() {
			n += c.Counters["+1/+1"]
		}
	}
	return n
}

var s82BnsB28DraugrToGraveyard = MayYes("s82-bns-b28/draugr-to-graveyard", putCarriedLibraryCardIntoGraveyard)

// s82BnsB28DraugrLook is "look at the top card of target player's
// library. You may put that card into their graveyard."
func s82BnsB28DraugrLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var victim uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			victim = t.ID
		}
	}
	if victim == uuid.Nil {
		return nil
	}
	looked := g.LookAtTopOfPlayersLibraryForEffect(item.Controller, victim, 1)
	if len(looked) == 0 {
		return nil
	}
	return MayChoice{
		Player:   ctx.Controller(),
		Question: "Draugr Thought-Thief — put the top card of that library into its owner's graveyard?",
		YesLabel: "Put it into the graveyard",
		NoLabel:  "Leave it on top",
		Then:     s82BnsB28DraugrToGraveyard,
		Carry:    []uuid.UUID{looked[0]},
	}.Apply(ctx)
}
