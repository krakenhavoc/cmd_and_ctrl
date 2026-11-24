package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lord of the Pit — Creature — Demon {4}{B}{B}{B}, 7/7 (Secret Lair Drop,
// #5462):
//
//	"Flying, trample
//	 At the beginning of your upkeep, sacrifice a creature other than this
//	 creature. If you can't, this creature deals 7 damage to you."
//
// Demonic Taskmaster's edict on yourself over every creature but the Lord:
// the choice is made as the trigger resolves. "If you can't" is the edict
// finding nothing to offer (no other creature, or none that can be sacrificed,
// CR 701.21a), and then the Lord deals 7 damage to its controller, as it last
// existed if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ea152809-85a2-4fde-8251-3b1f267e4443",
		Name:            "Lord of the Pit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Lord of the Pit — sacrifice a creature other than this creature; if you can't, it deals 7 damage to you",
				func(g *game.Game, item *game.StackItem) error {
					source := item.SourceCardID
					return g.PlayerSacrificesThenKeyedForEffect(source, item.Controller,
						effects.SacrificeSpec("a creature other than this creature", effects.Creature(), effects.NotSelf(source)),
						"Lord of the Pit — sacrifice a creature other than this creature", 1,
						lordOfThePitCouldntSacrifice, []uuid.UUID{item.Controller})
				}),
		},
	})
}

// lordOfThePitCouldntSacrifice is "If you can't, this creature deals 7 damage
// to you": the run's answer says the controller sacrificed nothing, whether
// there was no other creature to offer or none could be sacrificed. The
// damage is dealt by the Lord as it last existed (CR 608.2h). An on-disk key:
// never renamed, never reused.
var lordOfThePitCouldntSacrifice = game.RegisterPromptRunThen("prompt-run/lord-of-the-pit-cant-sacrifice", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 || r.Sacrifices().Sacrificed(r.Carry[0]) {
		return nil
	}
	return effects.DealDamage{Source: r.Source, Target: r.Carry[0], Amount: 7}.Apply(effects.KeyedContext(g, r.Carry[0], r.Source))
})
