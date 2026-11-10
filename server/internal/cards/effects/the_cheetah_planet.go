package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Cheetah Planet — Plane — Outside Mutter's Spiral (Doctor Who, #4870):
//
//	"When you planeswalk to The Cheetah Planet and at the beginning of
//	 your upkeep, put two +1/+1 counters on target non-Cat creature you
//	 control. It becomes a Cat in addition to its other types.
//	 Whenever chaos ensues, all Cats gain provoke until end of turn.
//	 (Whenever a Cat attacks, its controller may have target creature
//	 defending player controls untap and block it if able.)"
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// first ability is one ability with two trigger conditions and a target
// chosen as it goes on the stack; the Cat type has no stated duration, so it
// lasts while the creature stays on the battlefield (CR 611.2a). The chaos
// ability (CR 311.7) grants the Provoke triggered ability (CR 702.39) to
// every Cat on the battlefield as it resolves, whoever controls it (CR 611.2c);
// each one triggers separately and its controller may use it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c4679a1-f92f-44a9-9a44-066a137c8cd7",
		Name:         "The Cheetah Planet",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       theCheetahPlanetProvoke,
			Triggered: []game.TriggeredAbility{Provoke()},
			Text:      "Provoke (Whenever this creature attacks, you may have target creature defending player controls untap and block it if able.)",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenYouPlaneswalkToThisAndAtYourUpkeep("The Cheetah Planet — put two +1/+1 counters on target non-Cat creature you control; it becomes a Cat",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (AddCounter{Target: t, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
						return err
					}
					return ScopedEffectFor{
						Target:   t,
						Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtype("Cat"))},
						Duration: g.PinnedTo(game.IndefiniteDuration(), t),
						Label:    "The Cheetah Planet — a Cat in addition to its other types",
					}.Apply(ctx)
				}),
				TargetCreature("target non-Cat creature you control", YouControl(), Not(HasSubtype("Cat")))),
			WheneverChaosEnsues("The Cheetah Planet — all Cats gain provoke until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return GrantAbilitiesFor{
						Match:    And(Creature(), HasSubtype("Cat")),
						Keys:     []string{theCheetahPlanetProvoke},
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "The Cheetah Planet — provoke until end of turn",
					}.Apply(ctx)
				}),
		},
	})
}

// theCheetahPlanetProvoke is the granted bundle's key. An on-disk key.
const theCheetahPlanetProvoke = "the-cheetah-planet/provoke"
