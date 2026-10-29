package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bifur, Melodic Rider — Legendary Creature — Dwarf Bard
// {4}{R/W}{R/W}, 4/5:
//
//	"Storied (If you control three or more artifacts, legendaries,
//	 and/or Sagas, you have an enduring story for the rest of the game.)
//	 Whenever Bifur enters or attacks, put a +1/+1 counter on target
//	 creature.
//	 As long as you have an enduring story, if a triggered ability of a
//	 Dwarf you control triggers, that ability triggers an additional
//	 time."
//
// Storied is the engine's (CR 702.195a). The doubling is Roaming
// Throne's CR 603.2d shape (DoublesAbilitiesOf) over Dwarves you control,
// Bifur himself among them, asked as the ability triggers: a Dwarf's
// ability that triggers before the story is told triggers once.
//
// No simplification. The Hobbit (#4475).
func init() {
	counter := WhenThisEntersOrAttacks("Bifur, Melodic Rider — a +1/+1 counter on target creature",
		PlusOneCounterOnTheTarget)
	counter.Targets = TargetCreature("target creature")
	doubler := DoublesAbilitiesOf(nil, DoublesAbilitiesOfOptions{SourceMatch: bifurDwarfWithAStory})
	doubler.Label = "Bifur, Melodic Rider"
	doubler.RowLabel = "As long as you have an enduring story, abilities of Dwarves you control trigger an additional time."
	Register(Spec{
		OracleID:        "b8d563e4-e2bc-4e8b-8841-6655beff9138",
		Name:            "Bifur, Melodic Rider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordStoried},
		Triggered:       []game.TriggeredAbility{counter},
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}

// bifurDwarfWithAStory is "a triggered ability of a Dwarf you control"
// while Bifur's controller has an enduring story: the source as it is (or
// last was) is a Dwarf, the type word read through Bifur's own text (CR
// 612.2). DoublesAbilitiesOf has already checked that the source is
// Bifur's controller's.
func bifurDwarfWithAStory(g *game.Game, q game.TriggerDoublingQuery) bool {
	src := characteristicCard(q.Source.InstanceID, q.Source, q.SourceLKI)
	return src.HasSubtype(q.Doubler.TextSubtype("Dwarf")) && YouHaveAnEnduringStory(g, q.DoublerLKI.Controller)
}
