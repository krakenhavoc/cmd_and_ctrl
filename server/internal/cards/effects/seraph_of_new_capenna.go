package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        oracle,
		Name:            "Seraph of New Capenna",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:        "{4}{B/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         ManaCost("{4}{B/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Seraph of New Phyrexia",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Seraph of New Phyrexia — you may sacrifice another creature or artifact to get +2/+1",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Seraph of New Phyrexia — sacrifice another creature or artifact to get +2/+1 until end of turn?",
						Or(Creature(), Artifact()), true, seraphOfNewPhyrexiaSacrificed)
				}),
		},
	})
}

var seraphOfNewPhyrexiaSacrificed = RegisterSacrificedOne("prompt-run/seraph-of-new-phyrexia-pump", func(ctx *Context, _ uuid.UUID) error {
	if sourceIsNewObject(ctx.Game, ctx.Item) || !onBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 1, Label: "Seraph of New Phyrexia — +2/+1"}.Apply(ctx)
})
