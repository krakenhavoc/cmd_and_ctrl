package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cruelclaw's Heist — Sorcery {B}{B} (Bloomburrow, #4467):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 Target opponent reveals their hand. You choose a nonland card from
//	 it. Exile that card. If the gift was promised, you may cast that
//	 card for as long as it remains exiled, and mana of any type can be
//	 spent to cast it."
//
// Gift a card is the keyword's one field (gift.go, CR 702.174): the chosen
// opponent draws before the rest of the spell (CR 702.174j), so a card
// they draw from the gift can be the one taken. The reveal and the pick are
// Null Summoner's exiling revealed-hand pick (ADR 0116): the whole table
// sees the hand, you choose a nonland card, and it is exiled.
//
// The promise is read as the spell resolves and picks the continuation:
// promised, the card that landed in exile gets Black Cat, Cunning Thief's
// permission (cast only, mana of any type, for as long as it remains
// exiled); not promised, it is simply exiled. "May cast" leaves timing
// alone, and its mana cost is paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a0bc265-6b40-48d6-a588-c9624df7d1bc",
		Name:         "Cruelclaw's Heist",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		Gift:         GiftACard(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			pick := ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: Nonland(),
				Label:  "nonland card",
				Exile:  true,
			}
			if ctx.GiftPromised() {
				pick.Then = cruelclawsHeistMayCast
			}
			return pick.Apply(ctx)
		},
	})
}

// cruelclawsHeistMayCast is the promised gift's "you may cast that card for
// as long as it remains exiled, and mana of any type can be spent to cast
// it", given to the chooser once the card has landed in exile. An on-disk
// key: never renamed, never reused.
var cruelclawsHeistMayCast = RevealedPickThen("revealed-pick/cruelclaws-heist-may-cast",
	func(ctx *Context, pick game.RevealedPick) error {
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
				AnyType:    true,
				Duration:   game.WhileInZoneDuration(),
				Source:     pick.Source,
				SourceName: "Cruelclaw's Heist",
				Label:      "You may cast it for as long as it remains exiled; mana of any type can be spent to cast it",
			}, []game.Card{c})
		}
		return nil
	})
