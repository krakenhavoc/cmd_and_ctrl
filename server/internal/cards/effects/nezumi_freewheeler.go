package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nezumi Freewheeler // Hideous Fleshwheeler — {3}{B} Creature — Rat
// Samurai 3/3 // Creature — Phyrexian Rat 4/5 (March of the Machine):
//
//	Front: "Menace
//	        When this creature enters, each player mills three cards.
//	        {5}{W/P}: Transform this creature. Activate only as a sorcery."
//	Back:  "Menace
//	        When this creature transforms into Hideous Fleshwheeler, put
//	        target permanent card with mana value 2 or less from a
//	        graveyard onto the battlefield under your control."
//
// The mill is each player's, in turn order, through the shared
// each-player mill. The back's trigger watches the turn-over to the back
// face only (CR 701.27e); its target is any graveyard's permanent card
// with mana value 2 or less (CR 202.3), re-judged on resolution.
//
// No simplification.
func init() {
	const oracle = "e6fd0c7c-b164-4cf1-8910-0f8e680b2213"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Nezumi Freewheeler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Nezumi Freewheeler — each player mills three cards",
				func(g *game.Game, item *game.StackItem) error { return b02EachPlayerMills(g, item, 3) }),
		},
		Activated: []ActivatedAbility{{
			Label:        "{5}{W/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         ManaCost("{5}{W/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	reanimate := On(game.EventTransform, ThisTransformedInto(1),
		"Hideous Fleshwheeler — put target permanent card with mana value 2 or less from a graveyard onto the battlefield under your control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
		})
	reanimate.Targets = TargetCardInGraveyard("target permanent card with mana value 2 or less from a graveyard", Permanent(), ManaValueLE(2))
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Hideous Fleshwheeler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered:       []game.TriggeredAbility{reanimate},
	})
}
