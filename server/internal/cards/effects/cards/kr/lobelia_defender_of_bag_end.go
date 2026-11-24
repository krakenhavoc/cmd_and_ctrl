package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lobelia, Defender of Bag End — Legendary Creature — Halfling Citizen
// {2}{B}, 2/2 (Tales of Middle-earth Commander, #5948 GAP2 b07):
//
//	"When Lobelia enters, look at the top card of each opponent's library
//	 and exile those cards face down.
//	 {T}, Sacrifice an artifact: Choose one —
//	 • Until end of turn, you may play a card exiled with Lobelia without
//	   paying its mana cost.
//	 • Each opponent loses 2 life and you gain 2 life."
//
// The enters trigger exiles the top card of each opponent's library
// face down, linked to this object (CR 607.2a), the way Urianger
// Augurelt's Draw Arcanum does (ExileHiddenLinkedForEffect): the cards
// land FaceDownHidden, so Lobelia's controller, who looked at them, may
// go on looking at them (CR 406.3) and nobody else may, and each carries
// ExiledWith naming this object.
//
// The first mode is ONE play (#6091, CR 611.2c locks the set): one
// permission over the cards exiled with this object as it resolves, at
// {0} ("without paying its mana cost"), with a budget of one
// (CastsLeft), which a spell cast or a land played through it spends, as
// Kotose's does. A land is played with the turn's land drop and normal
// timing (CR 305.2a); a spell keeps its own timing. It lasts until end
// of turn. The source is the object the ability was activated from
// (SourceObjectForEffect), so a Lobelia sacrificed in response still
// reaches what it exiled. The second mode is the drain.
//
// No simplification.
func init() {
	const name = "Lobelia, Defender of Bag End"
	effects.Register(effects.Spec{
		OracleID:     "0f2e3e2e-9857-4be5-9bf0-62da1cc147fc",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Lobelia — look at the top card of each opponent's library and exile those cards face down",
				lobeliaExileEachOpponentsTop),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: Choose one — play a card exiled with Lobelia without paying its mana cost this turn; or each opponent loses 2 life and you gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Modes: effects.ChooseOne(
				effects.ModeDoing("Until end of turn, you may play a card exiled with Lobelia without paying its mana cost.", nil,
					func(item *game.StackItem, ctx *effects.Context, _ int) error {
						return lobeliaGrantPlay(ctx, item)
					}),
				effects.ModeDoing("Each opponent loses 2 life and you gain 2 life.", nil, effects.EachOpponentLosesTwoYouGainTwo),
			),
		}},
	})
}

// lobeliaExileEachOpponentsTop exiles the top card of each opponent's
// library face down, linked to this Lobelia.
func lobeliaExileEachOpponentsTop(g *game.Game, item *game.StackItem) error {
	ref, ok := g.SourceObjectForEffect(item)
	if !ok {
		return nil
	}
	src := game.PermissionCardRef(ref)
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == item.Controller || p.Library == nil || len(p.Library.Cards) == 0 {
			continue
		}
		top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
		if _, err := g.ExileHiddenLinkedForEffect(src, item.Controller, top); err != nil {
			return err
		}
	}
	return nil
}

// lobeliaGrantPlay is the first mode: one free play from among the cards
// exiled with this object now.
func lobeliaGrantPlay(ctx *effects.Context, item *game.StackItem) error {
	ref, ok := ctx.Game.SourceObjectForEffect(item)
	if !ok {
		return nil
	}
	src := game.PermissionCardRef(ref)
	var cards []game.Card
	for _, c := range ctx.Game.ExiledWithForEffect(src) {
		cards = append(cards, *c)
	}
	if len(cards) == 0 {
		return nil
	}
	ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     item.Controller,
		Zone:       game.ZoneExile,
		Cost:       "{0}",
		CastsLeft:  1,
		Source:     src.ID,
		SourceName: "Lobelia, Defender of Bag End",
		Label:      "Lobelia — play a card exiled with Lobelia without paying its mana cost",
	}, cards)
	return nil
}
