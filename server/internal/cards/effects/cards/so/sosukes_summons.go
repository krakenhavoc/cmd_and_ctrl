package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sosuke's Summons — Sorcery {2}{G} (Betrayers of Kamigawa, S82 OLD,
// #5602):
//
//	"Create two 1/1 green Snake creature tokens.
//	 Whenever a nontoken Snake you control enters, you may return this card
//	 from your graveyard to your hand."
//
// The spell is Bestial Menace's Snake tokens. The second sentence is a
// triggered ability that functions from the graveyard (Zones, CR 113.6k),
// Spellpyre Phoenix's shape: a Snake card you control that enters (a token
// does not count, nor does a Snake an opponent controls) offers the
// return, and a Summons that has left the graveyard by the time the
// ability resolves is not returned (CR 400.7). The Summons's own Snake
// tokens are tokens, so casting it never recurs it.
//
// No simplification.
func init() {
	const name = "Sosuke's Summons"
	back := effects.Optional(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
		return ok && !c.IsToken() && c.HasSubtypeWord(source.TextSubtype("Snake"))
	}, name+" — return this card from your graveyard to your hand", effects.ReturnThisCardFromYourGraveyardToYourHand),
		name+" — return this card from your graveyard to your hand?")
	back.Zones = []game.ZoneKind{game.ZoneGraveyard}
	effects.Register(effects.Spec{
		OracleID:     "6be663a5-70e2-435a-91ae-16168402787b",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("1/1 green Snake"), N: 2}),
		Triggered:    []game.TriggeredAbility{back},
	})
}
