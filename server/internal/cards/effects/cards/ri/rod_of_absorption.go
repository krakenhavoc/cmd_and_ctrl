package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rod of Absorption — Artifact {2}{U} (Adventures in the Forgotten Realms
// Commander):
//
//	"Whenever a player casts an instant or sorcery spell, exile it instead of
//	 putting it into a graveyard as it resolves.
//	 {X}, {T}, Sacrifice this artifact: You may cast any number of spells from
//	 among cards exiled with this artifact with total mana value X or less
//	 without paying their mana costs."
//
// The first ability is Forger's Foundry's exit without the "you may": the
// trigger marks the spell on the stack (Game.MarkSpellExilesLinkedForEffect)
// and the spell's own resolution exit (CR 608.2n) exiles it "exiled with" the
// Rod (CR 607.2a) instead of putting it into the graveyard. A countered or
// fizzled spell goes to the graveyard as usual, and a Rod that has left the
// battlefield by then leaves no link, so the spell goes to the graveyard.
//
// The second ability's cost sacrifices the Rod, so it reads the cards exiled
// with the object it was (SourceRef, last-known) as it resolves, and grants
// one free flash-speed cast permission over all of them sharing a budget of X
// total mana value (CastPermission.ManaValueBudget, Uldaros Theorix's shape),
// offered during the resolution (CR 608.2g). A spell is judged at its own mana
// value with X as 0 (CR 107.3b); the budget is spent by each cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d32345c1-ed20-4bfc-a5fa-6ac7b99542ec",
		Name:         "Rod of Absorption",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, anInstantOrSorceryWasCast,
				"Rod of Absorption — exile that spell instead of putting it into a graveyard as it resolves",
				rodOfAbsorptionMark),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{X}, {T}, Sacrifice this artifact: You may cast any number of spells from among cards exiled with this artifact with total mana value X or less without paying their mana costs.",
			Cost:  effects.Plus(effects.ManaCost("{X}"), effects.TapCost(), effects.SacrificeThis()),
			// Casts spells from exile: nothing a stop on an opponent's item answers.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return rodOfAbsorptionCast(g, item)
			},
		}},
	})
}

// anInstantOrSorceryWasCast is "whenever a player casts an instant or sorcery
// spell": any player's cast, read off the spell as it sits on the stack.
func anInstantOrSorceryWasCast(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor == uuid.Nil {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (spell.IsInstant() || spell.IsSorcery())
}

// rodOfAbsorptionMark marks the spell the trigger fired on to be exiled
// with the Rod as it resolves.
func rodOfAbsorptionMark(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	spell := ctx.Trigger().Event.CardID
	ref, ok := ctx.SourceRef()
	if !ok || spell == uuid.Nil {
		return nil
	}
	g.MarkSpellExilesLinkedForEffect(spell, game.PermissionCardRef(ref))
	return nil
}

func rodOfAbsorptionCast(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	budget := ctx.X()
	controller := ctx.Controller()
	var cards []game.Card
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsLand() {
			continue
		}
		cards = append(cards, *c)
		ids = append(ids, c.InstanceID)
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:          controller,
		Zone:            game.ZoneExile,
		Cost:            "{0}",
		Timing:          game.TimingFlash,
		CastOnly:        true,
		LapseOnPass:     game.LapseStaysInExile,
		ManaValueBudget: &budget,
		Source:          ctx.Source(),
		SourceName:      "Rod of Absorption",
		Label:           "Cast any number of spells exiled with Rod of Absorption with total mana value X or less without paying their mana costs",
	}, cards)
	effects.OfferResolutionCastOver(g, controller, ctx.Source(), "Rod of Absorption", game.MayCastKeywordFreeCast, effects.DeclineLeaveInExile, ids...)
	return nil
}
