package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Demonlord of Ashmouth — Creature — Demon {2}{B}{B}, 5/4:
//
//	"Flying
//	 When this creature enters, exile it unless you sacrifice another
//	 creature.
//	 Undying"
//
// The trigger fires on every entry, the undying return included
// (ruling). You may always choose the exile, even with another creature
// to sacrifice (ruling); with none the exile happens without a
// question. A Demonlord that has already left, or come back as a new
// object, is not exiled (CR 400.7). Flying and undying are
// PrintedKeywords (#2075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "049777d0-910d-46d2-9ca1-10277b2fb845",
		Name:            "Demonlord of Ashmouth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordUndying},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Demonlord of Ashmouth — exile it unless you sacrifice another creature", demonlordExileUnlessSacrifice),
		},
	})
}

func demonlordExileUnlessSacrifice(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	hasOther := false
	for _, id := range effects.PermanentsControlledByMatching(g, item.Controller, effects.Creature()) {
		if id != item.SourceCardID {
			hasOther = true
			break
		}
	}
	if !hasOther {
		return demonlordExileItself(ctx)
	}
	return effects.PickOption{
		Question: "Demonlord of Ashmouth — sacrifice another creature, or exile Demonlord of Ashmouth",
		Options:  []game.ChoiceOption{{Label: demonlordExileLabel}, {Label: demonlordSacrificeLabel}},
		ThenKey:  DemonlordChose,
	}.Apply(ctx)
}

const (
	demonlordExileLabel     = "Exile Demonlord of Ashmouth"
	demonlordSacrificeLabel = "Sacrifice another creature"
)

// demonlordChose is the keyed answer to "sacrifice another creature, or exile
// Demonlord" (#4253): the sacrifice pick is a keyed choose_cards prompt, so
// a table waiting on either prompt is a restore point. An on-disk key:
// never renamed, never reused.
var DemonlordChose = effects.OptionPickThen("option-pick/demonlord-of-ashmouth-choice", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || r.Option.Label != demonlordSacrificeLabel {
		return demonlordExileItself(ctx)
	}
	var others []uuid.UUID
	for _, id := range effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.Creature()) {
		// #3275, CR 101.2: a permanent that can't be sacrificed is not a
		// legal choice.
		if id != ctx.Source() && !ctx.Game.CantBeSacrificedForEffect(id) {
			others = append(others, id)
		}
	}
	if len(others) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   "Demonlord of Ashmouth — sacrifice another creature",
		Cards:      others,
		Min:        1,
		Max:        1,
		ThenKey:    DemonlordSacrificed,
		LiveZone:   game.ZoneBattlefield,
	})
	return nil
})

// demonlordSacrificed sacrifices the pick; nothing more is printed after the
// sacrifice. An on-disk key: never renamed, never reused.
var DemonlordSacrificed = game.RegisterChooseCardsThen("choose-cards/demonlord-of-ashmouth-sacrifice", func(g *game.Game, r game.CardsChosen) error {
	g.SacrificeAllForEffect(uuid.Nil, r.Picked)
	return nil
})

// demonlordExileItself exiles the trigger's source, the Demonlord.
func demonlordExileItself(ctx *effects.Context) error {
	return effects.ExileTarget{Target: ctx.Item.SourceCardID}.Apply(ctx)
}
