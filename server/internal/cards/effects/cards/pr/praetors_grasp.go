package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Praetor's Grasp — Sorcery {1}{B}{B} (Secret Lair Drop):
//
//	"Search target opponent's library for a card and exile it face down. Then
//	 that player shuffles. You may play that card for as long as it remains
//	 exiled."
//
// Bribery's search of an opponent's library (LibraryOwner is the named opponent,
// the searcher is the caster; that library is shuffled) with Beseech the
// Mirror's face-down exile, then Kheru Spellsnatcher's standing permission over
// the one card: the caster may cast it, or play it if it is a land, at its normal
// cost, from exile for as long as it stays there (CR 406.3). The mana can't be
// spent as though it were any type: only the printed cost is paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d56aeb1-0a50-46b6-abdb-cd6575a98dc3",
		Name:         "Praetor's Grasp",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			victim, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return effects.SearchLibrary{
				Player:       ctx.Controller(),
				LibraryOwner: victim,
				Limit:        1,
				Dest:         game.ZoneExile,
				FaceDown:     game.FaceDownPermitted,
				Shuffle:      true,
				Reason:       "Praetor's Grasp — search target opponent's library for a card",
				ThenTail:     ctx.ItemTail(praetorsGraspPlayIt),
			}.Apply(ctx)
		},
	})
}

// praetorsGraspPlayIt is "you may play that card for as long as it remains
// exiled": a standing permission over the card the search put into exile. An
// on-disk key.
var praetorsGraspPlayIt = effects.CardTail("tail/praetors-grasp-play-it", func(ctx *effects.Context, o game.TailOutcome) error {
	var cards []game.Card
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			cards = append(cards, c)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     ctx.Controller(),
		Zone:       game.ZoneExile,
		Duration:   game.WhileInZoneDuration(),
		Source:     ctx.Source(),
		SourceName: "Praetor's Grasp",
		Label:      "Praetor's Grasp — you may play that card for as long as it remains exiled",
	}, cards)
	return nil
})
