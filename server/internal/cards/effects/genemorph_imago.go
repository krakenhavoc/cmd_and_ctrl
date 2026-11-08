package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Genemorph Imago — Creature — Insect Druid {G}{U}, 1/3 (Edge of
// Eternities, #4461):
//
//	"Flying
//	 Landfall — Whenever a land you control enters, target creature has
//	 base power and toughness 3/3 until end of turn. If you control six or
//	 more lands, that creature has base power and toughness 6/6 until end
//	 of turn instead."
//
// A targeted landfall trigger with Restless Vinestalk's base power and
// toughness set (a layer 7b effect, CR 613.4b, so counters and pumps
// still apply on top). The six-land check is made as the trigger
// resolves (CR 608.2h), counting the land that triggered it. A target no
// longer legal is unaffected (CR 608.2b). Flying is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bf3a0585-3976-4078-8beb-6afa79f97027",
		Name:            "Genemorph Imago",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(Landfall(genemorphImagoLabel, genemorphImagoSetBase), TargetCreature("target creature")),
		},
	})
}

const genemorphImagoLabel = "Genemorph Imago — target creature has base power and toughness 3/3 (6/6 with six or more lands) until end of turn"

// genemorphImagoSetBase sets the target's base power and toughness to
// 3/3, or 6/6 when the controller has six or more lands, until end of
// turn.
func genemorphImagoSetBase(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	n := 3
	if countControlled(g, item.Controller, func(c game.Card) bool { return c.IsLand() }) >= 6 {
		n = 6
	}
	return untilEndOfTurn(ctx, target, nil, genemorphImagoLabel, game.SetBasePTMods(n, n)...)
}
