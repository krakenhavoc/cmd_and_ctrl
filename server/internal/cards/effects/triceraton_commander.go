package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Triceraton Commander — {X}{X}{W}{W} Creature — Dinosaur Soldier 2/2
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Flying
//	 When this creature enters, create X 2/2 white Dinosaur Soldier
//	 creature tokens.
//	 Whenever this creature attacks, Dinosaurs you control other than
//	 this creature get +1/+1 and gain flying until end of turn."
//
// X is the value announced for the cast (CR 107.3m; {X}{X} costs two mana
// per X, CR 107.3a), read off the entering permanent as the trigger is put on
// the stack (Champions from Beyond's XFromCast); put onto the battlefield
// without being cast, X is 0 and nothing is made. At X=0 it is still a 2/2
// flier, so it takes an x_matters_guard_test allowlist entry rather than
// declaring XMatters (the bot may cast it for {W}{W}). The attack trigger's
// Dinosaurs are the ones its controller controls as it resolves, locked then
// (CR 611.2c), the Commander itself left out; the creature type is read
// through the card's own text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "45d40cdb-21cc-4569-a894-83dd3026c947",
		Name:            "Triceraton Commander",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       "Triceraton Commander — create X 2/2 white Dinosaur Soldier tokens",
				XFromCast: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return CreateToken{Template: TokenCard("2/2 white Dinosaur Soldier"), N: ctx.X()}.Apply(ctx)
				},
			},
			WheneverThisAttacks("Triceraton Commander — other Dinosaurs you control get +1/+1 and gain flying until end of turn",
				triceratonCommanderRally),
		},
	})
}

// triceratonCommanderRally pumps and grants flying to every other Dinosaur
// the item's controller controls as it resolves.
func triceratonCommanderRally(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	dinosaur := g.TextReaderForEffect(item).TextSubtype("Dinosaur")
	self := item.SourceCardID
	others := func(_ *game.Game, caster uuid.UUID, c game.Card) bool {
		return c.InstanceID != self && c.Controller == caster && c.IsCreature() && c.HasSubtype(dinosaur)
	}
	if err := (BoostUntilEOT{Match: others, Power: 1, Toughness: 1, Label: "Triceraton Commander — +1/+1"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Match: others, Keywords: []string{"flying"}, Label: "Triceraton Commander — flying"}.Apply(ctx)
}
