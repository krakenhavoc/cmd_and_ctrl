package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invasion of New Capenna — Battle — Siege, defense 4, for {W}{B} (March
// of the Machine, #5477):
//
//	"(As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, you may sacrifice an artifact or creature.
//	 When you do, exile target artifact or creature an opponent
//	 controls."
//
// The reminder text is engine behaviour keyed on the card type
// (game/battle.go). This Spec carries the sacrifice, its reflexive
// exile (CR 603.12, target chosen as the reflexive trigger goes on the
// stack) and the defeated trigger that hands over Holy Frazzle-Cannon.
//
// No simplification.
func init() {
	const name = "Invasion of New Capenna"
	Register(Spec{
		OracleID:     invasionOfNewCapennaOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 4,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(name+" — you may sacrifice an artifact or creature; when you do, exile target artifact or creature an opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, name+" — sacrifice an artifact or creature to exile target artifact or creature an opponent controls?",
						Or(Artifact(), Creature()), false, invasionOfNewCapennaSacrificed)
				}),
			DefeatedTrigger(name+" — defeated: exile it, then cast Holy Frazzle-Cannon", SiegeDefeated()),
		},
	})
}

var invasionOfNewCapennaSacrificed = RegisterSacrificedOne("prompt-run/invasion-of-new-capenna-sacrificed", func(ctx *Context, _ uuid.UUID) error {
	return ReflexiveTrigger{
		Label: "Invasion of New Capenna — exile target artifact or creature an opponent controls",
		Body:  invasionOfNewCapennaExileBody,
	}.Apply(ctx)
})

var invasionOfNewCapennaExileBody = game.ReflexiveBody("invasion-of-new-capenna/exile", simpleBody(exileEveryLegalCardTarget),
	constTargets(func() *game.TargetSpec {
		return TargetPermanent("target artifact or creature an opponent controls", Or(Artifact(), Creature()), OpponentControls())
	}))

// invasionOfNewCapennaOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfNewCapennaOracleID = "530dc119-e266-4525-abe9-7fab599e9de2"
