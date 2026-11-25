package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rowan, Scholar of Sparks // Will, Scholar of Frost — modal double-faced
// planeswalker card (Strixhaven: School of Mages, #5502):
//
//	Rowan, Scholar of Sparks — Legendary Planeswalker — Rowan {2}{R}, loyalty 2
//	 "Instant and sorcery spells you cast cost {1} less to cast.
//	  +1: Rowan deals 1 damage to each opponent. If you've drawn three or more
//	  cards this turn, she deals 3 damage to each opponent instead.
//	  −4: You get an emblem with "Whenever you cast an instant or sorcery
//	  spell, you may pay {2}. If you do, copy that spell. You may choose new
//	  targets for the copy.""
//	Will, Scholar of Frost — Legendary Planeswalker — Will {4}{U}, loyalty 4
//	 "Instant and sorcery spells you cast cost {1} less to cast.
//	  +1: Up to one target creature has base power and toughness 0/2 until your
//	  next turn.
//	  −3: Draw two cards.
//	  −7: Exile up to five target permanents. For each permanent exiled this
//	  way, its controller creates a 4/4 blue and red Elemental creature token."
//
// # Two keys, one card
//
// Rowan keeps the bare oracle ID and Will takes "<oracle_id>#1" (Arlinn Kord's
// keyspace for a planeswalker back face); each has its own starting loyalty.
//
// # Both faces
//
// The discount is Thryx's reduction of one generic mana, for instants and
// sorceries its controller casts; it is a static of the permanent, so it works
// while the planeswalker is on the battlefield.
//
// # Rowan
//
// The +1 counts the cards its controller has drawn this turn when it resolves
// (Knowledge Is Power's tally): three or more makes the damage 3. The −4 emblem
// is Will Kenrith's emblem with a {2} payment: the trigger asks its controller
// for {2} as it resolves, and the copy (of the spell as last known, CR 707.10)
// may take new targets.
//
// # Will
//
// The +1 is Will Kenrith's scoped effect with base 0/2 until its controller's
// next turn and no loss of abilities. The −7 exiles up to five targets as one
// event and creates one 4/4 Elemental for each that was exiled, for the player
// who controlled it as it left (CR 603.10a's last-known information).
//
// No simplification on either face.
func init() {
	const id = "dd06c11c-c4fd-414e-ac63-5556f31491f5"
	discount := func() game.CostModifier {
		return effects.CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.", effects.YourSpell(),
			func(q game.CostQuery) bool { return q.Card.IsInstant() || q.Card.IsSorcery() })
	}
	effects.Register(effects.Spec{
		OracleID:        id,
		Name:            "Rowan, Scholar of Sparks",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 2,
		CostModifiers:   []game.CostModifier{discount()},
		Emblem: &effects.EmblemSpec{
			Label: "Rowan, Scholar of Sparks emblem",
			Text:  "Whenever you cast an instant or sorcery spell, you may pay {2}. If you do, copy that spell. You may choose new targets for the copy.",
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventCast, effects.YouCast(effects.Or(effects.Instant(), effects.Sorcery())),
					"Rowan, Scholar of Sparks emblem — you may pay {2} to copy that spell",
					effects.MayPayTwoToCopyThatSpell("Rowan, Scholar of Sparks — pay {2} to copy that spell?")),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Rowan deals 1 damage to each opponent. If you've drawn three or more cards this turn, she deals 3 damage to each opponent instead.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					n := 1
					if effects.B15CardsDrawnThisTurn(g)[item.Controller] >= 3 {
						n = 3
					}
					return effects.DamageToEachOpponent(g, item, n)
				},
			},
			{
				Label:  "−4: You get an emblem with \"Whenever you cast an instant or sorcery spell, you may pay {2}. If you do, copy that spell. You may choose new targets for the copy.\"",
				Cost:   effects.LoyaltyCost(-4),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
	effects.Register(effects.Spec{
		OracleID:        id + "#1",
		Name:            "Will, Scholar of Frost",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		CostModifiers:   []game.CostModifier{discount()},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Up to one target creature has base power and toughness 0/2 until your next turn.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.ScopedEffectFor{
							Target:   t.ID,
							Mods:     game.SetBasePTMods(0, 2),
							Duration: effects.DurationUntilYourNextTurn(ctx, item.Controller),
							Label:    "Will, Scholar of Frost — base power and toughness 0/2 until your next turn",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:   "−3: Draw two cards.",
				Cost:    effects.LoyaltyCost(-3),
				Purpose: game.Purpose{Draws: 2},
				Effect:  effects.Do(effects.DrawCards{N: 2}),
			},
			{
				Label:   "−7: Exile up to five target permanents. For each permanent exiled this way, its controller creates a 4/4 blue and red Elemental creature token.",
				Cost:    effects.LoyaltyCost(-7),
				Targets: effects.TargetPermanent("up to five target permanents", effects.Permanent()).WithCount(0, 5),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							ids = append(ids, t.ID)
						}
					}
					if len(ids) == 0 {
						return nil
					}
					return g.ExileCardsKeyedForEffect(ids, ctx.ItemTail(oldB51WillElementals))
				},
			},
		},
	})
}

// oldB51WillElementals is the rest of Will's −7: one 4/4 Elemental for each
// permanent exiled, created by the player who controlled that permanent as it
// left the battlefield.
var oldB51WillElementals = effects.MassTail("tail/old-b051-will-elementals", func(ctx *effects.Context, swept []game.Card, _ int) error {
	for _, c := range swept {
		if err := (effects.CreateToken{Controller: c.Controller, Template: effects.TokenCard("4/4 blue and red Elemental"), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
