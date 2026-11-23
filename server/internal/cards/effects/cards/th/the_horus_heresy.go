package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Horus Heresy — Enchantment — Saga {3}{U}{B}{R} (Warhammer 40,000
// Commander):
//
//	I   — For each opponent, gain control of up to one target nonlegendary
//	      creature that player controls for as long as this Saga remains on
//	      the battlefield.
//	II  — Draw a card for each creature you control but don't own.
//	III — Starting with you, each player chooses a creature. Destroy each
//	      creature chosen this way.
//
// Chapter I is Age of Ultron's per-opponent "up to one" clause list, each
// pick bound to its opponent, and The Super Hero Civil War's theft keyed to
// the Saga as it is now (CR 400.7): the creatures go home when the Saga
// leaves, and a Saga already gone when the chapter resolves takes nothing
// (CR 611.2b). Chapter II counts creatures you control whose owner is
// someone else as it resolves, the ones chapter I took included. Chapter III
// asks each seat in turn from you (CR 101.4) to choose ANY creature on the
// battlefield, theirs or not, and a creature two seats chose is destroyed
// once; the destruction is one simultaneous sweep, regeneration allowed.
//
// No simplification.
func init() {
	const name = "The Horus Heresy"
	steal := effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "for each opponent, gain control of up to one target nonlegendary creature that player controls for as long as this Saga remains on the battlefield"),
		horusHeresySteal)
	steal.Targets = effects.PerOpponentClauses(game.MaxPlayers-1, "nonlegendary creature", effects.And(effects.Creature(), effects.Not(effects.Legendary())))
	effects.Register(effects.Spec{
		OracleID:     "fc315174-d063-41dd-9379-11d95c5fd4ed",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			steal,
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "draw a card for each creature you control but don't own"),
				func(g *game.Game, item *game.StackItem) error {
					n := effects.CountControlled(g, item.Controller, func(c game.Card) bool {
						return c.IsCreature() && c.Owner != item.Controller
					})
					if n == 0 {
						return nil
					}
					return effects.DrawCards{Player: item.Controller, N: n}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "starting with you, each player chooses a creature; destroy each creature chosen this way"),
				func(g *game.Game, item *game.StackItem) error {
					var legs []game.ChooseCardsPrompt
					var all []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() {
							all = append(all, c.InstanceID)
						}
					}
					if len(all) == 0 {
						return nil
					}
					for _, seat := range effects.SeatsFromController(g, item.Controller) {
						legs = append(legs, game.ChooseCardsPrompt{
							Chooser:    seat,
							FromPlayer: seat,
							Source:     item.SourceCardID,
							Question:   "The Horus Heresy — choose a creature; each creature chosen this way is destroyed",
							Cards:      all,
							Min:        1,
							Max:        1,
							Zone:       game.ZoneBattlefield,
							ThenKey:    game.ChooseCardsRunLeg,
						})
					}
					_, err := g.ChooseCardsRunThenKeyedForEffect(legs, horusHeresyDestroy, item.SourceCardID,
						game.RunData{Asked: game.NewAskedItem(item)})
					return err
				}),
		},
	})
}

// horusHeresyDestroy destroys every creature any seat chose, once each.
var horusHeresyDestroy = effects.RunThen("run/the-horus-heresy-destroy", func(ctx *effects.Context, r game.RunDone) error {
	chosen := map[uuid.UUID]bool{}
	for _, id := range r.Picks().Cards() {
		chosen[id] = true
	}
	if len(chosen) == 0 {
		return nil
	}
	return effects.DestroyAllMatching{
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return chosen[c.InstanceID] && c.IsCreature() },
	}.Apply(ctx)
})

// horusHeresySteal is chapter I: control of each target still legal, for as
// long as the Saga remains.
func horusHeresySteal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	d, ok := effects.DurationWhileSourceRemains(ctx, item.SourceCardID)
	if !ok {
		return nil
	}
	for _, t := range targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.GainControl{
			Target:     t.ID,
			Controller: item.Controller,
			Duration:   d,
			Label:      "The Horus Heresy — control for as long as the Saga remains",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
