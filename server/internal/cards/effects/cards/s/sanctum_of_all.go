package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanctum of All — Legendary Enchantment — Shrine {W}{U}{B}{R}{G} (Core Set
// 2021, #5511):
//
//	"At the beginning of your upkeep, you may search your library and/or
//	 graveyard for a Shrine card and put it onto the battlefield. If you
//	 search your library this way, shuffle.
//	 If a triggered ability of another Shrine you control triggers while you
//	 control six or more Shrines, that ability triggers an additional time."
//
// The search names its zones first and shuffles only after a library search
// (Runeforge Champion's search, to the battlefield). The second ability is a
// CR 603.2d trigger doubler: it doubles a triggered ability of a Shrine that
// is not this one and that its controller controls, only while that player
// controls six or more Shrines as the ability triggers. Sanctum of All's own
// upkeep trigger is never doubled ("another Shrine").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63f60dfc-2bc4-411f-ae53-504156fe0102",
		Name:         "Sanctum of All",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Sanctum of All — you may search your library and/or graveyard for a Shrine card and put it onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     func(c game.Card) bool { return c.HasSubtype("Shrine") },
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Optional:      true,
						Shuffle:       true,
						AlsoGraveyard: true,
						Reason:        "Sanctum of All — a Shrine card",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		TriggerDoublers: []game.TriggerDoubler{
			effects.DoublesAbilitiesOf(nil, effects.DoublesAbilitiesOfOptions{SourceMatch: sanctumOfAllDoubles}),
		},
	})
}

// sanctumOfAllDoubles is "a triggered ability of another Shrine you control
// ... while you control six or more Shrines".
func sanctumOfAllDoubles(g *game.Game, q game.TriggerDoublingQuery) bool {
	if !q.SourceIsPermanent || q.Source.InstanceID == q.Doubler.InstanceID {
		return false
	}
	shrine := false
	for _, s := range q.SourceLKI.Subtypes {
		if s == "Shrine" {
			shrine = true
		}
	}
	return shrine && effects.S82OldB105ShrinesYouControl(g, q.DoublerLKI.Controller) >= 6
}
