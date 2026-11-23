package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// covetousElegyKeep is the keyed run of Covetous Elegy (#4290): every seat
// that had to choose keeps its picks, the rest are sacrificed together, and
// then the caster makes the Treasures.
var covetousElegyKeep = keepRun{
	Then: RunThen("prompt-run/covetous-elegy-keep", func(ctx *Context, r game.RunDone) error {
		if err := sacrificeTheRest(ctx, r); err != nil {
			return err
		}
		return covetousElegyTreasures(ctx)
	}),
	leg: game.ChooseCardsRunLeg,
}

// covetousElegyTreasures creates a tapped Treasure for each creature the
// caster's opponents control, counted after the sacrifice.
func covetousElegyTreasures(ctx *Context) error {
	n := 0
	for _, opp := range ctx.Opponents() {
		n += len(PermanentsControlledByMatching(ctx.Game, opp, Creature()))
	}
	if n == 0 {
		return nil
	}
	return CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: n}.Apply(ctx)
}

// covetousElegyResolve is "each player chooses up to two creatures they
// control, then sacrifices the rest. Then you create a tapped Treasure
// token for each creature your opponents control."
func CovetousElegyResolve(item *game.StackItem, ctx *Context) error {
	seats := SeatsFromActive(ctx.Game)
	asked := false
	for _, seat := range seats {
		if p := ctx.Game.PlayerByIDForEffect(seat); p == nil || p.Eliminated {
			continue
		}
		if len(PermanentsControlledByMatching(ctx.Game, seat, Creature())) > 2 {
			asked = true
		}
	}
	if !asked {
		return covetousElegyTreasures(ctx)
	}
	return KeepCreaturesThenSacrificeTheRest(ctx.Game, item, seats[:len(seats):len(seats)],
		"Covetous Elegy — choose up to two creatures to keep; you sacrifice the rest", 2, covetousElegyKeep)
}

// bnsB22EachPlayerRegainsTheirCreatures is "each player gains control of all
// creatures they own": for every creature someone else controls, an
// indefinite control effect for its owner (CR 611.2a, 613.1b).
func BnsB22EachPlayerRegainsTheirCreatures(label string) func(ctx *Context) error {
	return func(ctx *Context) error {
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if !c.IsCreature() || c.Controller == c.Owner {
				continue
			}
			if err := (GainControl{
				Target:     c.InstanceID,
				Controller: c.Owner,
				Duration:   game.IndefiniteDuration(),
				Label:      label,
			}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
		return nil
	}
}

// bnsB22ArtifactsYouControl is "artifacts you control", read post-layer.
func BnsB22ArtifactsYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.IsArtifact()
}

// bnsB22AreAlsoClues is "<artifacts you control> are Clues in addition to
// their other types" (CR 205.1b): a layer-4 subtype addition.
func BnsB22AreAlsoClues(label string, applies func(target *game.Card, g *game.Game, source *game.Card) bool) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer4Type,
		AppliesTo: applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if !EotHasType(c.Types, "Artifact") || EotHasType(c.Subtypes, "Clue") {
				return
			}
			c.Subtypes = append(append([]string(nil), c.Subtypes...), "Clue")
		},
		Label: label,
	}
}
