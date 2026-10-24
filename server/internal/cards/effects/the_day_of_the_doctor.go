package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Day of the Doctor — Enchantment — Saga {3}{R}{W}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I, II, III — Exile cards from the top of your library until you
//	 exile a legendary card. You may play that card for as long as this
//	 Saga remains on the battlefield. Put the rest of those exiled cards
//	 on the bottom of your library in a random order.
//	 IV — Choose up to three Doctors. You may exile all other creatures.
//	 If you do, this Saga deals 13 damage to you."
//
// Chapters I to III exile one card at a time until a legendary card
// lands (MillToZone with an Until, so a commander offers its owner the
// command zone and a replacement is honoured). The legendary card gets a
// play permission whose window is "for as long as this Saga remains on
// the battlefield" (#3192, ADR 0066's second 2026-10-13 amendment:
// game.ForAsLongAs + WhileSourceOnBattlefield, read by the one liveness
// test), measured against the Saga object the chapter came from, so the
// card stays exiled and unplayable once the Saga is gone (it is
// sacrificed after IV). The rest go to the bottom in a random order
// through the game's keyed RNG. A library with no legendary card is
// exiled whole and put back on the bottom.
//
// Chapter IV: the Doctors are creatures with the creature type Doctor,
// anyone's, chosen by the Saga's controller (zero to three). The
// exile is then a "you may"; choosing it exiles every other creature
// and the Saga deals 13 damage to its controller (as it last existed,
// if it has gone, CR 608.2h). Declining does neither.
//
// No simplification.
func init() {
	const name = "The Day of the Doctor"
	dig := func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		controller := item.Controller
		window := forAsLongAsThisRemainsWindow(g, item)
		return MillToZone{
			Player: controller,
			To:     game.ZoneExile,
			Until:  UntilCard(func(c game.Card) bool { return c.IsLegendary() }),
			Then: func(ctx *Context, exiled []uuid.UUID) error {
				var rest []uuid.UUID
				granted := false
				for _, id := range exiled {
					c, ok := ctx.Game.LookupCardForEffect(id)
					z := ctx.Game.FindCardZoneForEffect(id)
					if !granted && ok && c.IsLegendary() && z != nil && z.Kind == game.ZoneExile {
						granted = true
						ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
							Player:     controller,
							Zone:       game.ZoneExile,
							Duration:   window,
							Source:     item.SourceCardID,
							SourceName: name,
							Label:      "You may play that card for as long as this Saga remains on the battlefield",
						}, []game.Card{c})
						continue
					}
					rest = append(rest, id)
				}
				return ctx.Game.PutOnBottomInRandomOrderForEffect(controller, game.ZoneExile, rest)
			},
		}.Apply(ctx)
	}
	digLabel := "exile until a legendary card; you may play it while this Saga remains"
	Register(Spec{
		OracleID:     "c2ee17b2-6119-4ade-b806-1add64513c3a",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, digLabel), dig),
			ChapterTrigger(2, SagaChapterLabel(name, 2, digLabel), dig),
			ChapterTrigger(3, SagaChapterLabel(name, 3, digLabel), dig),
			// A board wipe that spares the chosen Doctors (ADR 0126 §6).
			TriggerWithPurpose(
				ChapterTrigger(4, SagaChapterLabel(name, 4, "choose up to three Doctors; you may exile all other creatures"), dayOfTheDoctorFinale),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}}),
		},
	})
}

// dayOfTheDoctorFinale is chapter IV: the pick of up to three Doctors,
// then the "you may" exile and its 13 damage. Captures only scalars.
func dayOfTheDoctorFinale(g *game.Game, item *game.StackItem) error {
	var doctors []uuid.UUID
	// "Doctor" is a word of the Saga's text (CR 612.2, #3984).
	doctor := g.TextReaderForEffect(item).TextSubtype("Doctor")
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.HasSubtype(doctor) {
			doctors = append(doctors, c.InstanceID)
		}
	}
	if len(doctors) == 0 {
		return dayOfTheDoctorMayExileOthers(g, item, nil)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "The Day of the Doctor — choose up to three Doctors",
		Cards:    doctors,
		Min:      0,
		Max:      3,
		Zone:     game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			return dayOfTheDoctorMayExileOthers(g, item, picked)
		},
	})
	return nil
}

// dayOfTheDoctorMayExileOthers is "You may exile all other creatures. If
// you do, this Saga deals 13 damage to you." The damage follows the
// choice to exile, not the number of creatures that reached exile.
func dayOfTheDoctorMayExileOthers(g *game.Game, item *game.StackItem, chosen []uuid.UUID) error {
	return MayChoice{
		Question: "The Day of the Doctor — exile all other creatures? If you do, it deals 13 damage to you.",
		Then:     dayOfTheDoctorAnswer,
		Carry:    chosen,
	}.Apply(NewContext(g, item))
}

// dayOfTheDoctorAnswer is the "yes": exile every creature but the ones kept
// (the carried IDs), then take 13 damage. Keyed (#4291).
var dayOfTheDoctorAnswer = MayYes("confirm/the-day-of-the-doctor-exile-others", func(ctx *Context, r game.ConfirmAnswer) error {
	keep := make(map[uuid.UUID]bool, len(r.Carry))
	for _, id := range r.Carry {
		keep[id] = true
	}
	if err := (ExileAllMatching{Match: And(Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return !keep[c.InstanceID]
	})}).Apply(ctx); err != nil {
		return err
	}
	return DamageController{N: 13}.Apply(ctx)
})
