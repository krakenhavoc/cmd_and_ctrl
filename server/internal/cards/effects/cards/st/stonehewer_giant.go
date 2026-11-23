package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stonehewer Giant — Creature — Giant Warrior {3}{W}{W}, 4/4 (Tales of
// Middle-earth Commander):
//
//	"Vigilance
//	 {1}{W}, {T}: Search your library for an Equipment card, put it onto the
//	 battlefield, attach it to a creature you control, then shuffle."
//
// Honored Knight-Captain's search, with the Equipment's entry followed by
// Beatrix's attach: the creature is chosen from those the found Equipment
// can legally be attached to (CR 301.5c), as the search finishes, and an
// Equipment with no such creature stays unattached on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b3d99ea0-d61f-4394-8913-b4057c38ff3e",
		Name:            "Stonehewer Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: Search your library for an Equipment card, put it onto the battlefield, attach it to a creature you control, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtype("Equipment") },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Stonehewer Giant — an Equipment card",
					ThenTail:  ctx.Tail(s82GapB118StonehewerFound),
				}.Apply(ctx)
			},
		}},
	})
}

var s82GapB118StonehewerFound = effects.CardTail("stonehewer-giant/attach-found-equipment", func(ctx *effects.Context, o game.TailOutcome) error {
	g := ctx.Game
	for _, equip := range o.IDs {
		if !g.Battlefield.Contains(equip) {
			continue
		}
		var hosts []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == ctx.Controller() && c.IsCreature() && c.InstanceID != equip && g.AttachLegalForEffect(equip, c.InstanceID) {
				hosts = append(hosts, c.InstanceID)
			}
		}
		if len(hosts) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: "Stonehewer Giant — attach the Equipment to a creature you control",
			Cards:    hosts,
			Min:      1,
			Max:      1,
			ThenKey:  s82GapB118StonehewerAttach,
			LiveZone: game.ZoneBattlefield,
			Carry:    []uuid.UUID{equip},
		})
	}
	return nil
})

var s82GapB118StonehewerAttach = game.RegisterChooseCardsThen("choose-cards/stonehewer-giant-attach", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 || len(r.Picked) == 0 || !g.Battlefield.Contains(r.Carry[0]) || !g.Battlefield.Contains(r.Picked[0]) {
		return nil
	}
	return g.AttachForEffect(r.Carry[0], game.TargetRef{Kind: game.TargetCard, ID: r.Picked[0]})
})
