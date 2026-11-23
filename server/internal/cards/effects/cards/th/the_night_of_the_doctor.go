package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Night of the Doctor — Enchantment — Saga {4}{W}{W} (Doctor Who):
//
//	"I — Destroy all creatures.
//	 II — Return target legendary creature card from your graveyard to the
//	 battlefield. Put your choice of a first strike, vigilance, or lifelink
//	 counter on it."
//
// I is a board wipe (Wrath of God's DestroyAllMatching, which regeneration
// and indestructible still answer). II returns the legendary creature card
// and then asks the controller which counter it gets (Aragorn, Company
// Leader's choice among keyword counters); the counter goes on the permanent
// that entered, if it did.
//
// No simplification.
func init() {
	const name = "The Night of the Doctor"
	effects.Register(effects.Spec{
		OracleID:     "a9c1fa95-34a8-42ef-af2f-1394df23cad8",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "destroy all creatures"),
					func(g *game.Game, item *game.StackItem) error {
						return effects.DestroyAllMatching{Match: effects.Creature()}.Apply(effects.NewContext(g, item))
					}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}}),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "return target legendary creature card from your graveyard to the battlefield; put your choice of a first strike, vigilance, or lifelink counter on it"),
				effects.TargetCardInGraveyard("target legendary creature card in your graveyard", effects.Creature(), effects.Legendary(), effects.YouOwn()),
				nightOfTheDoctorReturn),
		},
	})
}

var nightOfTheDoctorCounterKinds = []string{game.CounterFirstStrike, game.CounterVigilance, game.CounterLifelink}

func nightOfTheDoctorReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneBattlefield {
			return nil
		}
		options := make([]game.ChoiceOption, len(nightOfTheDoctorCounterKinds))
		for i, kind := range nightOfTheDoctorCounterKinds {
			options[i] = game.ChoiceOption{Label: "A " + kind + " counter"}
		}
		return effects.PickOption{
			Question: "The Night of the Doctor — put your choice of a counter on the returned creature",
			Options:  options,
			ThenKey:  nightOfTheDoctorCounterChosen,
			Carry:    []uuid.UUID{t.ID},
			WithItem: true,
		}.Apply(ctx)
	}
	return nil
}

var nightOfTheDoctorCounterChosen = effects.OptionPickThen("option-pick/the-night-of-the-doctor-counter", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) != 1 || r.Index < 0 || r.Index >= len(nightOfTheDoctorCounterKinds) {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(r.Carry[0]); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return ctx.Game.AddCounterByForEffect(r.Chooser, r.Carry[0], nightOfTheDoctorCounterKinds[r.Index], 1)
})
