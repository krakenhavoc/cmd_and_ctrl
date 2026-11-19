package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oko, Thief of Crowns — Legendary Planeswalker — Oko {1}{G}{U}, loyalty 4
// (Breaking News, #5929):
//
//	"+2: Create a Food token.
//	 +1: Target artifact or creature loses all abilities and becomes a green
//	 Elk creature with base power and toughness 3/3.
//	 −5: Exchange control of target artifact or creature you control and
//	 target creature an opponent controls with power 3 or less."
//
// The +2 is the shared Food template. The +1 is one scoped effect with no
// duration (CR 611.2a: no stated duration, so it lasts as long as the object
// stays on the battlefield) pinned to the target object (CR 400.7): layer 4
// replaces its card types and creature types (CR 205.1a) with "creature —
// Elk", layer 5 makes it green, layer 6 removes every ability, layer 7b
// sets its base power and toughness to 3/3, so counters and pumps still
// apply on top. An artifact that is not a creature becomes one, summoning
// sick if it came under its controller's control this turn. The −5 is
// Switcheroo's exchange (CR 701.12a) between the first target, an artifact
// or creature you control, and the second, a creature an opponent controls
// whose power is 3 or less; both must be
// legal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "60c60923-ff1b-43f7-8768-731499fcffc9",
		Name:            "Oko, Thief of Crowns",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label: "+2: Create a Food token.",
				Cost:  LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: FoodToken(), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "+1: Target artifact or creature loses all abilities and becomes a green Elk creature with base power and toughness 3/3.",
				Cost:    LoyaltyCost(1),
				Targets: TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					mods := []game.Mod{
						game.SetTypesMod([]string{"Creature"}, ctx.TextSubtypes("Elk")),
						game.SetColorsMod(ctx.TextColors("G")...),
						game.LoseAllAbilitiesMod(),
					}
					mods = append(mods, game.SetBasePTMods(3, 3)...)
					return ScopedEffectFor{
						Target:   id,
						Mods:     mods,
						Duration: game.IndefiniteDuration(),
						Label:    "Oko, Thief of Crowns — loses all abilities, becomes a green Elk 3/3",
					}.Apply(ctx)
				},
			},
			{
				Label: "−5: Exchange control of target artifact or creature you control and target creature an opponent controls with power 3 or less.",
				Cost:  LoyaltyCost(-5),
				Targets: Clauses(
					TargetPermanent("target artifact or creature you control", Or(Artifact(), Creature()), YouControl()),
					TargetCreature("target creature an opponent controls with power 3 or less", OpponentControls(), PowerLE(3)),
				),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					mine, ok := ctx.ClauseTarget(0)
					if !ok {
						return nil
					}
					theirs, ok := ctx.ClauseTarget(1)
					if !ok {
						return nil
					}
					return ExchangeControl{A: mine.ID, B: theirs.ID, Label: "Oko, Thief of Crowns — exchange control"}.Apply(ctx)
				},
			},
		},
	})
}
