package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Covetous Urge — Sorcery {U/B}{U/B}{U/B}{U/B} (Throne of Eldraine, S82 old
// sets):
//
//	"Target opponent reveals their hand. You choose a nonland card from that
//	 player's graveyard or hand and exile it. You may cast that card for as
//	 long as it remains exiled, and you may spend mana as though it were mana
//	 of any color to cast that spell."
//
// Agonizing Remorse's revealed-hand pick with the graveyard (ADR 0116), with
// both candidate sets narrowed to nonland cards, and Cruelclaw's Heist's
// continuation: the card that landed in exile gets a cast-only permission for
// the chooser, for as long as it remains exiled, with mana of any color. Its
// own mana cost is paid and timing is not changed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f52008f-3365-4cf9-9cae-8e761d657902",
		Name:         "Covetous Urge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseFromRevealedHand{
				Player:          effects.TargetedPlayer(ctx),
				Filter:          effects.Nonland(),
				Label:           "nonland card",
				Exile:           true,
				FromGraveyard:   true,
				GraveyardFilter: effects.Nonland(),
				Then:            covetousUrgeMayCast,
			}.Apply(ctx)
		},
	})
}

// covetousUrgeMayCast is "you may cast that card for as long as it remains
// exiled, and you may spend mana as though it were mana of any color to cast
// that spell", given to the chooser once the card has landed in exile. An
// on-disk key: never renamed, never reused.
var covetousUrgeMayCast = effects.RevealedPickThen("revealed-pick/covetous-urge-may-cast",
	func(ctx *effects.Context, pick game.RevealedPick) error {
		g := ctx.Game
		for _, chosen := range pick.Chosen {
			if z := g.FindCardZoneForEffect(chosen.InstanceID); z == nil || z.Kind != game.ZoneExile {
				continue
			}
			c, ok := g.LookupCardForEffect(chosen.InstanceID)
			if !ok {
				continue
			}
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{
				Player:     pick.Chooser,
				Zone:       game.ZoneExile,
				CastOnly:   true,
				AnyColor:   true,
				Duration:   game.WhileInZoneDuration(),
				Source:     pick.Source,
				SourceName: "Covetous Urge",
				Label:      "You may cast it for as long as it remains exiled; mana of any color can be spent to cast it",
			}, []game.Card{c})
		}
		return nil
	})
