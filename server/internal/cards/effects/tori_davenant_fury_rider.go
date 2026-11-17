package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tori D'Avenant, Fury Rider — Legendary Creature — Human Knight
// {1}{R}{R}{W}, 3/3 (Dominaria United):
//
//	"Vigilance, trample
//	 Whenever Tori D'Avenant attacks, all other attacking creatures you
//	 control get +1/+1 until end of turn. Other red attacking creatures you
//	 control gain trample until end of turn. Untap each other white
//	 attacking creature you control."
//
// One trigger, three instructions in printed order. Each one reads the
// OTHER attacking creatures its controller controls as the ability resolves,
// and the sets are fixed then (CR 611.2c): the boost for all of them, the
// trample for the red ones and the untap for the tapped white ones. A white
// creature put in combat tapped and attacking is untapped too, and an untap
// never removes a creature from combat (CR 506.4). Red and white are read
// through the layered view.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6d51ebd1-ceb9-42c0-8c1e-34dbbbc7bf8a",
		Name:            "Tori D'Avenant, Fury Rider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Tori D'Avenant — other attacking creatures you control get +1/+1, red ones gain trample, white ones untap",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					others := And(AttackingCreature(), YouControl(), OtherThan(item.SourceCardID))
					if err := (BoostUntilEOT{Match: others, Power: 1, Toughness: 1, Label: "Tori D'Avenant — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					if err := (GrantKeywordUntilEOT{
						Match:    And(others, OfColor("R")),
						Keywords: []string{"trample"},
						Label:    "Tori D'Avenant — trample",
					}).Apply(ctx); err != nil {
						return err
					}
					white := ctx.TextColor("W")
					return b16UntapAllYouControlMatching(ctx, item.Controller, func(c game.Card) bool {
						return c.IsCreature() && c.AttackingTarget != uuid.Nil && c.InstanceID != item.SourceCardID && c.HasColor(white)
					})
				}),
		},
	})
}
