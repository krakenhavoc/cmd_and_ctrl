package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boseiju Reaches Skyward // Branch of Boseiju — a transforming Saga
// (Kamigawa: Neon Dynasty, #5493).
//
//	I   — Search your library for up to two basic Forest cards, reveal
//	      them, put them into your hand, then shuffle.
//	II  — Put up to one target land card from your graveyard on top of
//	      your library.
//	III — Exile this Saga, then return it to the battlefield transformed
//	      under your control.
//
// Branch of Boseiju, Enchantment Creature — Plant 0/0: "Reach. This
// creature gets +1/+1 for each land you control."
//
// Chapter III is Fable of the Mirror-Breaker's exile-and-return: the
// Branch is a new object with no lore counters (CR 400.7, 712.14a).
func init() {
	effects.Register(effects.Spec{
		OracleID:     BoseijuReachesSkywardOracleID,
		Name:         "Boseiju Reaches Skyward",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel("Boseiju Reaches Skyward", 1, "search for up to two basic Forest cards"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLandWithSubtype("forest"),
						Dest:      game.ZoneHand,
						Limit:     2,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Boseiju Reaches Skyward — up to two basic Forest cards",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel("Boseiju Reaches Skyward", 2, "put up to one target land card from your graveyard on top of your library"),
				effects.TargetCardInGraveyard("up to one target land card from your graveyard", effects.Land(), effects.YouOwn()).WithCount(0, 1),
				effects.PutChosenTargetOnTopOfLibrary),
			effects.ChapterTrigger(3, effects.SagaChapterLabel("Boseiju Reaches Skyward", 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})

	effects.Register(effects.Spec{
		OracleID:        BoseijuReachesSkywardOracleID + "#1",
		Name:            "Branch of Boseiju",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static: []game.StaticAbility{
			effects.S82NeoB18SelfPT("This creature gets +1/+1 for each land you control.",
				func(g *game.Game, source *game.Card) int { return effects.B03LandsControlled(g, source.Controller) }),
		},
	})
}

// boseijuReachesSkywardOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const BoseijuReachesSkywardOracleID = "ec08aeb3-bba7-4982-9160-68d25bd411d6"
