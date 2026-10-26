package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "049777d0-910d-46d2-9ca1-10277b2fb845",
		Name:            "Demonlord of Ashmouth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordUndying},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Demonlord of Ashmouth — exile it unless you sacrifice another creature", demonlordExileUnlessSacrifice),
		},
	})
}

func demonlordExileUnlessSacrifice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	hasOther := false
	for _, id := range permanentsControlledByMatching(g, item.Controller, Creature()) {
		if id != item.SourceCardID {
			hasOther = true
			break
		}
	}
	if !hasOther {
		return demonlordExileItself(ctx)
	}
	return PickOption{
		Question: "Demonlord of Ashmouth — sacrifice another creature, or exile Demonlord of Ashmouth",
		Options:  []game.ChoiceOption{{Label: demonlordExileLabel}, {Label: demonlordSacrificeLabel}},
		ThenKey:  demonlordChose,
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
var demonlordChose = OptionPickThen("option-pick/demonlord-of-ashmouth-choice", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || r.Option.Label != demonlordSacrificeLabel {
		return demonlordExileItself(ctx)
	}
	var others []uuid.UUID
	for _, id := range permanentsControlledByMatching(ctx.Game, ctx.Controller(), Creature()) {
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
		ThenKey:    demonlordSacrificed,
		LiveZone:   game.ZoneBattlefield,
	})
	return nil
})

// demonlordSacrificed sacrifices the pick; nothing more is printed after the
// sacrifice. An on-disk key: never renamed, never reused.
var demonlordSacrificed = game.RegisterChooseCardsThen("choose-cards/demonlord-of-ashmouth-sacrifice", func(g *game.Game, r game.CardsChosen) error {
	g.SacrificeAllForEffect(uuid.Nil, r.Picked)
	return nil
})

// demonlordExileItself exiles the trigger's source, the Demonlord.
func demonlordExileItself(ctx *Context) error {
	return ExileTarget{Target: ctx.Item.SourceCardID}.Apply(ctx)
}
