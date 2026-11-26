package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ravenous Demon // Archdemon of Greed — Creature — Demon {3}{B}{B} 4/4 //
// Creature — Demon 9/9:
//
//	Ravenous Demon: "Sacrifice a Human: Transform this creature. Activate only
//	 as a sorcery."
//	Archdemon of Greed: "Flying, trample
//	 At the beginning of your upkeep, sacrifice a Human. If you can't, tap this
//	 creature and it deals 9 damage to you."
//
// The upkeep sacrifice is a prompt (Lord of the Pit's shape): if no Human was
// sacrificed, the Archdemon taps and deals 9 damage to you, as it last existed
// if it has left (CR 608.2h).
func init() {
	const oracle = "18932bb9-e054-4ad6-80c8-d2f177312da0"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Ravenous Demon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "Sacrifice a Human: Transform this creature. Activate only as a sorcery.",
			Cost:         effects.SacrificeN(1, "a Human", effects.HasSubtype("Human")),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Archdemon of Greed",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Archdemon of Greed — sacrifice a Human; if you can't, tap this creature and it deals 9 damage to you",
				func(g *game.Game, item *game.StackItem) error {
					return g.PlayerSacrificesThenKeyedForEffect(item.SourceCardID, item.Controller,
						effects.SacrificeSpec("a Human", effects.Creature(), effects.HasSubtype("Human")),
						"Archdemon of Greed — sacrifice a Human", 1,
						archdemonOfGreedCouldntSacrifice, []uuid.UUID{item.Controller})
				}),
		},
	})
}

var archdemonOfGreedCouldntSacrifice = game.RegisterPromptRunThen("prompt-run/archdemon-of-greed-cant-sacrifice", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 || r.Sacrifices().Sacrificed(r.Carry[0]) {
		return nil
	}
	ctx := effects.KeyedContext(g, r.Carry[0], r.Source)
	if err := effects.NothingIfGone(g.TapTargetForEffect(r.Source)); err != nil {
		return err
	}
	return effects.DealDamage{Source: r.Source, Target: r.Carry[0], Amount: 9}.Apply(ctx)
})
