package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Izzet Chemister — Creature — Goblin Wizard {2}{R}, 1/3:
//
//	"Haste
//	 {R}, {T}: Exile target instant or sorcery card from your graveyard.
//	 {1}{R}, {T}, Sacrifice this creature: Cast any number of cards exiled
//	 with this creature without paying their mana costs."
//
// The first ability links each card it exiles to this object (CR 607.2a,
// Game.ExileLinkedForEffect). The second reads that link from the sacrificed
// Chemister's object (Rod of Absorption's way): the cards stay linked to it
// after it has left, and a new Chemister has no claim on them (CR 400.7). The
// casts are offered as the ability resolves (CR 608.2g), each for free, any
// number of them, and a card not cast stays in exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e4a3d6f3-36ba-4e6d-a254-351bc8886f20",
		Name:            "Izzet Chemister",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}, {T}: Exile target instant or sorcery card from your graveyard.",
				Cost:    game.AbilityCost{Mana: "{R}", Tap: true},
				Targets: effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
				Effect:  izzetChemisterExile,
			},
			{
				Label:   "{1}{R}, {T}, Sacrifice this creature: Cast any number of cards exiled with this creature without paying their mana costs.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost(), effects.SacrificeThis()),
				Effect:  izzetChemisterCast,
			},
		},
	})
}

func izzetChemisterExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		ref, ok := ctx.SourceRef()
		if !ok {
			return g.ExileCardForEffect(t.ID)
		}
		return g.ExileLinkedForEffect(ref, t.ID)
	}
	return nil
}

func izzetChemisterCast(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	controller := ctx.Controller()
	var cards []game.Card
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		cards = append(cards, *c)
		ids = append(ids, c.InstanceID)
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      controller,
		Zone:        game.ZoneExile,
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		AnyNumber:   true,
		LapseOnPass: game.LapseStaysInExile,
		Source:      ctx.Source(),
		SourceName:  "Izzet Chemister",
		Label:       "Cast any number of cards exiled with Izzet Chemister without paying their mana costs",
	}, cards)
	effects.OfferResolutionCastOver(g, controller, ctx.Source(), "Izzet Chemister", game.MayCastKeywordFreeCast, effects.DeclineLeaveInExile, ids...)
	return nil
}
