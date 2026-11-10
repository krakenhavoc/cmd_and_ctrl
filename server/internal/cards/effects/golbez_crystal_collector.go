package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Golbez, Crystal Collector — Legendary Creature — Human Wizard {U}{B},
// 1/4 (Final Fantasy, #4462):
//
//	"Whenever an artifact you control enters, surveil 1.
//	 At the beginning of your end step, if you control four or more
//	 artifacts, return target creature card from your graveyard to your
//	 hand. Then if you control eight or more artifacts, each opponent
//	 loses life equal to that card's power."
//
// The first ability is Quicksmith Genius's watch (the Golbez itself is not
// an artifact) over a surveil step. The second is an intervening "if"
// (CR 603.4), checked as the end step begins and again as the ability
// resolves. Its target is chosen on the stack; if it is illegal the
// ability does nothing, so no opponent loses life however many artifacts
// there are (the ruling). The second "if" is not intervening: it is read
// as the ability resolves, after the card has come back, and the life
// lost is the returned card's power as it now is in hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "533d5bf9-0912-451f-a3f4-3842b0a75b1d",
		Name:         "Golbez, Crystal Collector",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, ArtifactEnteredUnderYourControl,
				"Golbez, Crystal Collector — surveil 1", Do(Surveil{N: 1})),
			Targeting(InterveningIf(golbezFourArtifacts{}, On(game.EventBeginEndStep,
				AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return countControlled(g, source.Controller, func(c game.Card) bool { return c.IsArtifact() }) >= 4
				}),
				"Golbez, Crystal Collector — return target creature card from your graveyard to your hand; then if you control eight or more artifacts, each opponent loses life equal to that card's power",
				golbezReturnAndDrain)),
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature())),
		},
	})
}

// golbezArtifacts counts the artifacts the ability's controller controls.
func golbezArtifacts(ctx *Context) int {
	return countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsArtifact() })
}

// golbezFourArtifacts is the intervening "if you control four or more
// artifacts".
type golbezFourArtifacts struct{}

func (golbezFourArtifacts) Holds(ctx *Context) bool { return golbezArtifacts(ctx) >= 4 }

func (golbezFourArtifacts) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondBoard, 0), true
}

func golbezReturnAndDrain(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var id = item.SourceCardID
	found := false
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			id, found = t.ID, true
			break
		}
	}
	if !found {
		return nil
	}
	if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
		return err
	}
	if golbezArtifacts(ctx) < 8 {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if p := g.PlayerByIDForEffect(c.Owner); p == nil || !p.Hand.Contains(id) {
		return nil
	}
	if n := c.CurrentPower(); n > 0 {
		return EachOpponentLosesLife{N: n}.Apply(ctx)
	}
	return nil
}
