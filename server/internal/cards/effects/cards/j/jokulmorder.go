package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jokulmorder — Creature — Leviathan {4}{U}{U}{U}, 12/12 (Coldsnap, #5597):
//
//	"Trample
//	 This creature enters tapped.
//	 When this creature enters, sacrifice it unless you sacrifice five lands.
//	 This creature doesn't untap during your untap step.
//	 Whenever you play an Island, you may untap this creature."
//
// The tapped entry is the CR 614 self-replacement and the untap-step rule Nettle
// Sentinel's. The enter trigger is Tresserhorn's Lord's two-way choice: sacrifice
// Jokulmorder or sacrifice five lands (offered only when its controller controls
// five they can sacrifice, and made together). A Jokulmorder that has already
// left, or come back as a new object, is not sacrificed (CR 400.7). "Play an
// Island" is a land played as a land drop (CR 305.1), Event.Played, not one that
// merely entered, and the untap is a "may" asked as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "fbf074b2-d650-43b2-b8b8-d7a778dfd27e",
		Name:                  "Jokulmorder",
		Completeness:          effects.CompletenessFull,
		PrintedKeywords:       []string{"trample"},
		Replacements:          []game.ReplacementEffect{effects.SelfEntersTapped()},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Jokulmorder — sacrifice it unless you sacrifice five lands", jokulmorderETB),
			effects.Optional(effects.On(game.EventZoneMove, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || !effects.B20LandPlayed(ev, g) {
					return false
				}
				land, ok := g.LookupCardForEffect(ev.CardID)
				return ok && land.HasSubtype(source.TextSubtype("Island"))
			}, "Jokulmorder — untap it", effects.B31UntapSelf), "Jokulmorder — untap this creature?"),
		},
	})
}

const (
	jokulmorderSacrificeLabel = "Sacrifice Jokulmorder"
	jokulmorderLandsLabel     = "Sacrifice five lands"
)

// jokulmorderLands lists the lands the controller may sacrifice.
func jokulmorderLands(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range effects.PermanentsControlledByMatching(g, controller, effects.Land()) {
		if !g.CantBeSacrificedForEffect(id) {
			out = append(out, id)
		}
	}
	return out
}

func jokulmorderSacrificeSelf(ctx *effects.Context) error {
	ctx.Game.SacrificeAllForEffect(ctx.Source(), []uuid.UUID{ctx.Source()})
	return nil
}

func jokulmorderETB(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	if len(jokulmorderLands(g, item.Controller)) < 5 {
		return jokulmorderSacrificeSelf(ctx)
	}
	return effects.PickOption{
		Question: "Jokulmorder — sacrifice it, or sacrifice five lands",
		Options:  []game.ChoiceOption{{Label: jokulmorderSacrificeLabel}, {Label: jokulmorderLandsLabel}},
		ThenKey:  jokulmorderChose,
		WithItem: true,
	}.Apply(ctx)
}

// jokulmorderChose is the keyed answer to the unless-cost choice. An on-disk
// key: never renamed, never reused.
var jokulmorderChose = effects.OptionPickThen("option-pick/s82-old-b568-jokulmorder-choice", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || r.Option.Label != jokulmorderLandsLabel {
		return jokulmorderSacrificeSelf(ctx)
	}
	lands := jokulmorderLands(ctx.Game, ctx.Controller())
	if len(lands) < 5 {
		return jokulmorderSacrificeSelf(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   "Jokulmorder — sacrifice five lands",
		Cards:      lands,
		Min:        5,
		Max:        5,
		ThenKey:    jokulmorderSacrificed,
		LiveZone:   game.ZoneBattlefield,
	})
	return nil
})

// jokulmorderSacrificed pays the cost: the five lands. An on-disk key: never
// renamed, never reused.
var jokulmorderSacrificed = game.RegisterChooseCardsThen("choose-cards/s82-old-b568-jokulmorder-lands", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	g.SacrificeAllForEffect(r.Source, r.Picked)
	return nil
})
