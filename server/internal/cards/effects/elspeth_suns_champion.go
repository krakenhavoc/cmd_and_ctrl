package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elspeth, Sun's Champion — Legendary Planeswalker — Elspeth for
// {4}{W}{W}, starting loyalty 4:
//
//	"+1: Create three 1/1 white Soldier creature tokens.
//	 −3: Destroy all creatures with power 4 or greater.
//	 −7: You get an emblem with 'Creatures you control get +2/+2 and
//	     have flying.'"
//
// COMPLETE since S40 (#623). The −7 was the catalog's first emblem and
// the worked example for ADR 0064: the whole card-side declaration is
// an `Emblem` slot holding two ordinary `game.StaticAbility` values —
// layer 7c for the +2/+2, layer 6 for the flying — and an ability
// whose Effect is `CreateEmblem{}`, which names nothing because the
// emblem it makes is this card's. The emblem object then reaches the
// layer pass through the same source gather a Glorious Anthem does.
//
// The two statics are separate because the layers are: CR 613 applies
// ability grants (6) before power/toughness modifications (7c), and
// one StaticAbility declares one layer. This is the same split
// Craterhoof Behemoth's file describes.
//
// The −3 reads CurrentPower, so it catches a 2/2 wearing +1/+1
// counters and an anthem, and spares a 6/6 that something shrank.
// That is the printed card: "power 4 or greater" is the power it has
// when the ability resolves.
//
// History: #418 shipped this card with no `Completeness` declaration,
// so the catalog page published it as "unreviewed" while the omitted
// ultimate sat in a comment where no player reads it; #621 declared it
// `caveats`. Both are now moot — the card is whole.
func init() {
	Register(Spec{
		OracleID:     "05e6b243-48a6-4a42-bc5f-413441de9c33",
		Name:         "Elspeth, Sun's Champion",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Elspeth, Sun's Champion emblem",
			Text:  "Creatures you control get +2/+2 and have flying.",
			Static: []game.StaticAbility{
				{
					Layer:     game.Layer7PT,
					SubLayer:  game.SubLayer7C_Modify,
					AppliesTo: creaturesTheEmblemsOwnerControls,
					Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
						c.Power += 2
						c.Toughness += 2
					},
				},
				KeywordGrant(creaturesTheEmblemsOwnerControls, "flying"),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Create three 1/1 white Soldier creature tokens.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("1/1 white Soldier"),
						N:          3,
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−3: Destroy all creatures with power 4 or greater.",
				Cost:  LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					// Snapshot first: destroying moves cards out of
					// the slice this would otherwise be walking, and
					// the power that matters is the power at
					// resolution, before anything has died.
					var doomed []game.Card
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.CurrentPower() >= 4 {
							doomed = append(doomed, c)
						}
					}
					for _, c := range doomed {
						if err := (DestroyTarget{Target: c.InstanceID}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label: "−7: You get an emblem with \"Creatures you control get +2/+2 and have flying.\"",
				Cost:  LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// creaturesTheEmblemsOwnerControls is "creatures you control" read
// from an emblem: the emblem's controller is its owner (CR 114.5) and
// never changes, so this is the same predicate an anthem uses with the
// same meaning. Shared by the emblem's two statics so the layer-6 half
// and the layer-7c half can never drift apart.
func creaturesTheEmblemsOwnerControls(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller
}
