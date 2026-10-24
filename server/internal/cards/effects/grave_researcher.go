package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grave Researcher // Reanimate — a preparation card (CR 722, ADR 0090):
//
//	Grave Researcher {2}{B}, Creature — Troll Warlock 3/3:
//	"At the beginning of your upkeep, surveil 1. Then if there are three
//	 or more creature cards in your graveyard, this creature becomes
//	 prepared."
//	Reanimate {B}, Sorcery:
//	"Put target creature card from a graveyard onto the battlefield
//	 under your control. You lose life equal to that card's mana value."
//
// The count is read after the surveil, in its Then, so a creature card
// surveilled into the graveyard counts toward the three. The prepare
// spell is Reanimate's body (reanimate.go); the engine makes the copy,
// the cast permission and the unprepare.
//
// No simplification. Mass Casualty deck request (#3245).
const graveResearcherOracleID = "6cb8b8c4-0674-4f14-9d89-010969fbb80e"

func init() {
	Register(Spec{
		OracleID:     graveResearcherOracleID,
		Name:         "Grave Researcher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Grave Researcher — surveil 1, then maybe become prepared", graveResearcherUpkeep),
		},
	})
	Register(Spec{
		OracleID:     graveResearcherOracleID + "#1",
		Name:         "Reanimate",
		Completeness: CompletenessFull,
		Targets:      targetCreatureInAnyGraveyard(),
		Purpose:      ForTargets(Reanimates(0)),
		OnResolve:    reanimateLosingItsManaValue,
	})
}

func graveResearcherUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return Surveil{N: 1, ThenTail: ctx.ItemTail(graveResearcherPrepare)}.Apply(ctx)
}

// graveResearcherPrepare is "if there are three or more creature cards in
// your graveyard, this creature becomes prepared". An on-disk key.
var graveResearcherPrepare = CardTail("grave-researcher/become-prepared", func(ctx *Context, o game.TailOutcome) error {
	p := ctx.Game.PlayerByIDForEffect(o.Tail.Player)
	if p == nil {
		return nil
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() {
			n++
		}
	}
	if n < 3 {
		return nil
	}
	return BecomePrepared{Target: o.Tail.Source}.Apply(ctx)
})
