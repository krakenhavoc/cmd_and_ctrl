package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Champion of Stray Souls — Creature — Skeleton Warrior {4}{B}{B}, 4/4
// (Commander 2019):
//
//	"{3}{B}{B}, {T}, Sacrifice X other creatures: Return X target creature
//	 cards from your graveyard to the battlefield.
//	 {5}{B}{B}: Put this card from your graveyard on top of your library."
//
// The first ability is Ruthless Technomancer's variable sacrifice cost
// (SacrificeX, whose count is the announced X, CR 107.3a) with Back in
// Town's exactly-X graveyard targets (CountFromX); the creatures come back
// together under your control, each target still legal as it resolves
// (CR 608.2b). The sacrifice is "other": the Champion is never one of the
// X. The second is Bookwurm's graveyard ability with a depth of one: it does
// nothing if the card has left the graveyard by resolution (CR 602.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4edbaac4-dcb3-40b4-abd7-ff2683bd366d",
		Name:         "Champion of Stray Souls",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{B}{B}, {T}, Sacrifice X other creatures: Return X target creature cards from your graveyard to the battlefield.",
				Cost:    effects.Plus(effects.ManaCost("{3}{B}{B}"), effects.TapCost(), championOfStraySoulsSacrifice()),
				Targets: effects.TargetsCountedByX(effects.TargetCardInGraveyard("X target creature cards from your graveyard", effects.YouOwn(), effects.Creature())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Cmdr2S33ReturnTargetedCardsTogether(item, effects.NewContext(g, item))
				},
			},
			{
				Label:   "{5}{B}{B}: Put this card from your graveyard on top of your library.",
				Cost:    effects.ManaCost("{5}{B}{B}"),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Effect:  championOfStraySoulsToTop,
			},
		},
	})
}

// championOfStraySoulsSacrifice is "Sacrifice X other creatures": SacrificeX's
// variable count over "another" creature, so the Champion is never one of them.
func championOfStraySoulsSacrifice() game.AbilityCost {
	spec := effects.Another(effects.SacrificeSpec("X other creatures", effects.Creature()))
	spec.CountFromX = true
	return game.AbilityCost{SacrificeOther: spec}
}

// championOfStraySoulsToTop puts the card from the graveyard on top of its
// owner's library; nothing if it has left the graveyard (CR 608.2b).
func championOfStraySoulsToTop(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return effects.PutInLibraryInAnyOrder{
		Chooser:   item.Controller,
		Cards:     []uuid.UUID{id},
		From:      game.ZoneGraveyard,
		Placement: game.LibraryPlaceTop,
		TopDepth:  1,
		Label:     "Champion of Stray Souls — put it on top of your library",
	}.Apply(effects.NewContext(g, item))
}
