package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fall of the First Civilization — Enchantment — Saga {2}{W}:
//
//	"I — You and target opponent each draw two cards.
//	 II — Exile target artifact an opponent controls.
//	 III — Each player chooses three nonland permanents they control.
//	 Destroy all other nonland permanents."
//
// Chapter I is Flumph's draw, two cards each, the opponent's waiting on
// any prompt yours asks (CR 608.2c). Chapter III is Planetary
// Annihilation's keep-set shape over nonland permanents: each seat with
// more than three (APNAP, CR 101.4) is asked once for the three it keeps,
// a seat with three or fewer is not asked (all of its permanents are
// kept), and only when every seat has answered do all the others go in
// one simultaneous destroy (indestructible survives, CR 702.12b).
func init() {
	const name = "Fall of the First Civilization"
	effects.Register(effects.Spec{
		OracleID:     "dd0494f5-58bf-46ae-ad69-90b063f6f660",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			func() game.TriggeredAbility {
				t := effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "you and target opponent each draw two cards"),
					effects.TargetPlayer("target opponent", effects.Opponent()),
					effects.Do(
						effects.DrawCards{N: 2},
						effects.Instruction(func(ctx *effects.Context) error {
							if p, ok := effects.FirstLegalPlayerTarget(ctx); ok {
								return effects.DrawCards{Player: p, N: 2}.Apply(ctx)
							}
							return nil
						}),
					))
				return effects.TriggerWithPurpose(t, game.Purpose{Draws: 2, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2})})
			}(),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "exile target artifact an opponent controls"),
				effects.TargetPermanent("target artifact an opponent controls", effects.Artifact(), effects.OpponentControls()),
				effects.ExileFirstTarget),
			effects.TriggerWithPurpose(effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "each player chooses three nonland permanents they control; destroy all other nonland permanents"),
				fallOfTheFirstCivilizationChoose),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}}),
		},
	})
}

// fallOfTheFirstCivilizationChoose asks each seat with more than three
// nonland permanents for the three it keeps.
func fallOfTheFirstCivilizationChoose(g *game.Game, item *game.StackItem) error {
	var legs []game.ChooseCardsPrompt
	var asked []uuid.UUID
	for _, seat := range effects.ApnapPlayers(g) {
		perms := effects.PermanentsControlledByMatching(g, seat, effects.Nonland())
		if len(perms) <= 3 {
			continue
		}
		asked = append(asked, seat)
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    seat,
			FromPlayer: seat,
			Source:     item.SourceCardID,
			Question:   "Fall of the First Civilization — choose three nonland permanents you control; all other nonland permanents are destroyed",
			Cards:      perms,
			Min:        3,
			Max:        3,
			Zone:       game.ZoneBattlefield,
			ThenKey:    game.ChooseCardsRunLeg,
		})
	}
	if len(legs) == 0 {
		return nil
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, fallOfTheFirstCivilizationRun, item.SourceCardID, game.RunData{Carry: asked})
	return err
}

// fallOfTheFirstCivilizationRun destroys the nonland permanents the asked
// seats did not keep. An on-disk key.
var fallOfTheFirstCivilizationRun = effects.RunThen("prompt-run/fall-of-the-first-civilization-keep", func(ctx *effects.Context, r game.RunDone) error {
	kept := map[uuid.UUID]bool{}
	for _, id := range r.Picks().Cards() {
		kept[id] = true
	}
	asked := map[uuid.UUID]bool{}
	for _, seat := range r.Carry {
		asked[seat] = true
	}
	return effects.DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return asked[c.Controller] && !c.IsLand() && !kept[c.InstanceID]
	}}.Apply(ctx)
})
