package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana, Death's Majesty — Legendary Planeswalker — Liliana {3}{B}{B},
// loyalty 5:
//
//	"+1: Create a 2/2 black Zombie creature token. Mill two cards.
//	 −3: Return target creature card from your graveyard to the
//	 battlefield. That creature is a black Zombie in addition to its
//	 other colors and types.
//	 −7: Destroy all non-Zombie creatures."
//
// The +1 is the token, then the mill, in the printed order. The −3 is a
// targeted return from your own graveyard; "that creature is a black
// Zombie in addition to its other colors and types" is a continuous
// effect with no duration (CR 611.2a), so it is a scoped effect pinned
// to the returned object: it lasts exactly as long as that permanent
// stays on the battlefield and is gone the moment it leaves (CR 400.7).
// The two halves are game.AddSubtypesMod (layer 4, "Zombie") and
// game.AddColorsMod (layer 5, black), each read through the ability's
// text (CR 612.2, #3984), so the creature keeps every other
// colour and type. It is applied from the return's continuation, after
// any entry question has been answered.
//
// The −7 destroys every creature that is not a Zombie right now,
// including your own. "Zombie" is read through OfCreatureType, which
// asks Card.HasSubtype, so a changeling counts as a Zombie and the
// creature the −3 returned (and every Zombie token) survives. Indestructible
// creatures survive it, as printed.
//
// Starting loyalty is printed card data stamped by the deck importer
// (ADR 0032), not declared here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fca595e-1902-40fd-a359-622caa612b6a",
		Name:         "Liliana, Death's Majesty",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Create a 2/2 black Zombie creature token. Mill two cards.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("2/2 black Zombie"),
						N:          1,
					}).Apply(ctx); err != nil {
						return err
					}
					return MillCards{N: 2}.Apply(ctx)
				},
			},
			{
				Label:   "−3: Return target creature card from your graveyard to the battlefield. That creature is a black Zombie in addition to its other colors and types.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature()),
				Effect:  lilianaDeathsMajestyReanimate,
			},
			{
				Label: "−7: Destroy all non-Zombie creatures.",
				Cost:  LoyaltyCost(-7),
				Purpose: game.Purpose{Sweep: game.Sweep{
					Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DestroyAllMatching{
						Match: And(Creature(), Not(OfCreatureType("Zombie"))),
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

func lilianaDeathsMajestyReanimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetCardIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets: ids,
		Then: func(ctx *Context, entered []uuid.UUID) error {
			for _, id := range entered {
				if err := (ScopedEffectFor{
					Target:   id,
					Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Zombie")...), game.AddColorsMod(ctx.TextColors("B")...)},
					Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
					Label:    "Liliana, Death's Majesty — it is a black Zombie in addition to its other colors and types",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}.Apply(ctx)
}
