package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Screaming Swarm — Creature — Bird Horror {5}{U}, 4/4 (Innistrad: Crimson
// Vow):
//
//	"Flying
//	 Whenever you attack with one or more creatures, target player mills that
//	 many cards.
//	 {2}{U}: Put this card from your graveyard into your library second from
//	 the top."
//
// "That many" is the attack batch's count (Armasaur Guide's, Cryptcaller
// Chariot's ThatManyEvents): one EventAttack per declared attacker, settled
// as one trigger whose target player is chosen as it goes on the stack. The
// graveyard ability functions from the graveyard (CR 113.6), does nothing if
// the card has left it, and puts the card into its owner's library second
// from the top (Commit // Memory's depth).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5ea5d171-a7b8-4269-9405-d2f4ccf0cbcb",
		Name:            "Screaming Swarm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WithAtBatchEnd(effects.Targeting(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Screaming Swarm — target player mills that many cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return effects.MillCards{Player: t.ID, N: effects.SettledEventAmount(item)}.Apply(ctx)
						}
					}
					return nil
				}),
				effects.TargetPlayer("target player")),
				effects.ThatManyEvents()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Put this card from your graveyard into your library second from the top.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{U}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  screamingSwarmTuck,
		}},
	})
}

// screamingSwarmTuck puts the card from the graveyard second from the top of
// its owner's library, if it is still in the graveyard (CR 602.5, 608.2b).
func screamingSwarmTuck(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return effects.PutInLibraryInAnyOrder{
		Chooser:   item.Controller,
		Cards:     []uuid.UUID{id},
		From:      game.ZoneGraveyard,
		Placement: game.LibraryPlaceTop,
		TopDepth:  2,
		Label:     "Screaming Swarm — put it into your library second from the top",
	}.Apply(effects.NewContext(g, item))
}
