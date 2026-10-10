package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_reprint_a_helpers.go — helpers for slice
// fra-reprint-a (tracker #2795): older cards reprinted in Reality
// Fracture. Every name carries the rfReprintA prefix.

// rfReprintABrainsurgePutBack is Brainsurge's "then put two cards from
// your hand on top of your library in any order", registered under its
// own key so the prompt names the right card.
var rfReprintABrainsurgePutBack = game.RegisterDrawThen("brainsurge-put-back", func(g *game.Game, d game.DrawThen) error {
	ctx := NewContext(g, &game.StackItem{Controller: d.Player, SourceCardID: d.Source})
	return PutFromHandOnTopInAnyOrder{
		Player: d.Player,
		N:      d.N,
		Label:  "Brainsurge — put two cards from your hand on top of your library",
	}.Apply(ctx)
})

// rfReprintALandscape builds the Landscape cycle's shape: "{T}: Add
// {C}. {T}, Sacrifice this land: Search your library for a basic A, B,
// or C card, put it onto the battlefield tapped, then shuffle. Cycling
// {cost}". The fetch is the Panorama's without the {1}.
func rfReprintALandscape(oracleID, name, cyclingCost string, subtypes ...string) Spec {
	match := IsBasicLandOfAnySubtype(subtypes...)
	label := "{T}, Sacrifice this land: Search your library for a basic " +
		subtypes[0] + ", " + subtypes[1] + ", or " + subtypes[2] +
		" card, put it onto the battlefield tapped, then shuffle."
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   label,
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     match,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Reveal:        true,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Choose a basic " + subtypes[0] + ", " + subtypes[1] + ", or " + subtypes[2] + " to put onto the battlefield tapped",
					}.Apply(NewContext(g, item))
				},
			},
			Cycling(cyclingCost),
		},
	}
}

// rfReprintAOccultEpiphanyDiscard is Occult Epiphany's "then discard X
// cards. Create a 1/1 white Spirit creature token with flying for each
// card type among cards discarded this way", registered as the draw's
// continuation so the discard prompt opens only once the draws are done.
var rfReprintAOccultEpiphanyDiscard = game.RegisterDrawThen("occult-epiphany-discard", func(g *game.Game, d game.DrawThen) error {
	player := d.Player
	return g.PlayerDiscardsThenForEffect(game.DiscardPrompt{
		Player:   player,
		Source:   d.Source,
		N:        d.N,
		Question: "Occult Epiphany — discard X cards",
	}, func(g *game.Game, discarded game.PromptedDiscards) error {
		n := rfReprintACardTypesAmong(g, discarded.By(player))
		if n == 0 {
			return nil
		}
		return g.CreateTokenForEffect(player, TokenCard("1/1 white Spirit with flying"), n)
	})
})

// rfReprintACardTypesAmong counts the distinct card types among the
// given cards, read off their printed type lines (a card in a
// graveyard has only its printed characteristics, CR 109.3).
func rfReprintACardTypesAmong(g *game.Game, ids []uuid.UUID) int {
	seen := map[string]bool{}
	for _, id := range ids {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		_, types, _ := game.ParseTypeLine(c.TypeLine)
		for _, t := range types {
			seen[t] = true
		}
	}
	return len(seen)
}

// rfReprintAMassPolymorphReveal is Mass Polymorph's second sentence:
// "reveal cards from the top of your library until you reveal that many
// creature cards. Put all creature cards revealed this way onto the
// battlefield, then shuffle the rest of the revealed cards into your
// library." A library that runs out first reveals itself entirely.
// Tokens are not cards (CR 108.2), so they never count.
func rfReprintAMassPolymorphReveal(ctx *Context, want int) error {
	player := ctx.Controller()
	p := ctx.PlayerByID(player)
	var run []uuid.UUID
	hit := map[uuid.UUID]bool{}
	if p != nil && p.Library != nil && want > 0 {
		for i := len(p.Library.Cards) - 1; i >= 0 && len(hit) < want; i-- {
			c := p.Library.Cards[i]
			run = append(run, c.InstanceID)
			if !c.IsToken() && c.IsCreature() {
				hit[c.InstanceID] = true
			}
		}
	}
	if len(run) > 0 {
		ctx.Game.RevealForEffect(game.RevealSpec{
			Player: player,
			Source: ctx.Source(),
			Reason: "Mass Polymorph — reveal until that many creature cards",
			Cards:  run,
		})
	}
	return PutFromLibraryOntoBattlefield{
		Player: player,
		Cards:  run,
		Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit[c.InstanceID] },
		All:    true,
		Then: func(g *game.Game, res PutFromLibraryResult) error {
			return g.ShuffleLibraryForEffect(res.Player)
		},
	}.Apply(ctx)
}
