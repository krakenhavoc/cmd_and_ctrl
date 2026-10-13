package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ensnared by the Mara — Sorcery {2}{R}{R}:
//
//	"Each opponent faces a villainous choice — They exile cards from the
//	 top of their library until they exile a nonland card, then you may
//	 cast that card without paying its mana cost, or that player exiles
//	 the top four cards of their library and Ensnared by the Mara deals
//	 damage equal to the total mana value of those exiled cards to that
//	 player."
//
// A villainous choice (CR 701.55, #3143) asked of each opponent in APNAP
// order, one at a time.
//
// Option A exiles from the FACING player's library until a nonland card
// lands (MillToZone with an Until, the cascade-style loop, so a mill
// replacement or a commander's command-zone prompt is handled by the one
// primitive that owns it), then grants the spell's CONTROLLER a free cast
// of that card (grantFreeCasts, the window Emergent Ultimatum's "cast them
// as part of this" uses): price {0}, cast only, flash timing because the
// cast is part of this resolution (CR 608.2g), and it closes on the
// caster's next priority pass, leaving an uncast card in exile (nothing
// says it goes anywhere). The lands exiled on the way stay exiled. A
// library with no nonland card exiles all of it and grants nothing.
//
// Option B exiles the top four of the facing player's library (fewer from
// a shorter one) and Ensnared by the Mara deals damage to that player
// equal to the total mana value of the cards that were actually exiled;
// lands and a card whose cost reads as zero add nothing, and no cards
// means no damage. Both options are always on offer (CR 701.55b).
//
// No simplifications.
var maraChoice = NewVillainousChoice("ensnared-by-the-mara",
	"Ensnared by the Mara — exile until a nonland card for the caster to cast free, or exile four and take damage?",
	VillainousOption{
		Label: "Exile cards from the top of your library until you exile a nonland card; the controller of Ensnared by the Mara may cast it without paying its mana cost",
		Harm:  2,
		Body: func(ctx *Context, f Facing) error {
			return MillToZone{
				Player: f.Player,
				To:     game.ZoneExile,
				Until:  UntilCard(func(c game.Card) bool { return !c.IsLand() }),
				Then: func(ctx *Context, exiled []uuid.UUID) error {
					var hit []uuid.UUID
					for _, id := range exiled {
						if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsLand() {
							hit = append(hit, id)
						}
					}
					grantFreeCasts(ctx.Game, ctx.Controller(), ctx.Source(), "Ensnared by the Mara",
						game.TimingFlash, game.LapseStaysInExile, hit)
					return nil
				},
			}.Apply(ctx)
		},
	},
	VillainousOption{
		Label: "Exile the top four cards of your library; Ensnared by the Mara deals damage to you equal to their total mana value",
		Harm:  1,
		Body: func(ctx *Context, f Facing) error {
			return MillToZone{
				Player: f.Player,
				N:      4,
				To:     game.ZoneExile,
				Then: func(ctx *Context, exiled []uuid.UUID) error {
					total := 0
					for _, id := range exiled {
						if c, ok := ctx.Game.LookupCardForEffect(id); ok {
							total += c.ManaValue()
						}
					}
					return DealDamage{Source: ctx.Source(), Target: f.Player, Amount: total}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})

func init() {
	Register(Spec{
		OracleID:     "16090270-0ac5-404c-b0c8-a6c8a4b508a4",
		Name:         "Ensnared by the Mara",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return FaceVillainousChoice{Players: ctx.Opponents(), Choice: maraChoice}.Apply(ctx)
		},
	})
}
