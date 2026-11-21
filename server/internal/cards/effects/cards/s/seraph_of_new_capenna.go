package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seraph of New Capenna // Seraph of New Phyrexia — {2}{W} Creature — Angel
// Soldier 2/2 // Creature — Phyrexian Angel 3/3 (March of the Machine):
//
//	Front: "Flying
//	        {4}{B/P}: Transform this creature. Activate only as a sorcery."
//	Back:  "Flying
//	        Whenever this creature attacks, you may sacrifice another
//	        creature or artifact. If you do, this creature gets +2/+1
//	        until end of turn."
//
// The attack trigger is Comet Crawler's shape (`maySacrificeOne`, asked at
// resolution over what you control then, the source excluded). The pump
// lands only while this is still the object that attacked (CR 400.7).
//
// No simplification.
func init() {
	const oracle = "08574d1b-4b76-42dd-884b-29c5a61cfdda"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Seraph of New Capenna",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{4}{B/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{4}{B/P}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Seraph of New Phyrexia",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Seraph of New Phyrexia — you may sacrifice another creature or artifact to get +2/+1",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Seraph of New Phyrexia — sacrifice another creature or artifact to get +2/+1 until end of turn?",
						effects.Or(effects.Creature(), effects.Artifact()), true, seraphOfNewPhyrexiaSacrificed)
				}),
		},
	})
}

var seraphOfNewPhyrexiaSacrificed = effects.RegisterSacrificedOne("prompt-run/seraph-of-new-phyrexia-pump", func(ctx *effects.Context, _ uuid.UUID) error {
	if effects.SourceIsNewObject(ctx.Game, ctx.Item) || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 1, Label: "Seraph of New Phyrexia — +2/+1"}.Apply(ctx)
})
