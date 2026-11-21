package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Koth, Fire of Resistance — Legendary Planeswalker — Koth {2}{R}{R},
// starting loyalty 4 (Phyrexia: All Will Be One, #5479):
//
//	"+2: Search your library for a basic Mountain card, reveal it, put it
//	     into your hand, then shuffle.
//	 −3: Koth deals damage to target creature equal to the number of
//	     Mountains you control.
//	 −7: You get an emblem with "Whenever a Mountain you control enters,
//	     this emblem deals 4 damage to any target.""
//
// The −3 counts Mountains as it resolves; Koth is the damage source and a
// Koth that has left still deals it (CR 608.2h). The emblem is the source
// of its own damage and its trigger takes its target as it goes on the
// stack (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c75b870-263b-41ab-9ece-6e3845bc3d8d",
		Name:            "Koth, Fire of Resistance",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &effects.EmblemSpec{
			Label: "Koth, Fire of Resistance emblem",
			Text:  "Whenever a Mountain you control enters, this emblem deals 4 damage to any target.",
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventETB},
				Key:     "Koth, Fire of Resistance emblem — 4 damage to any target",
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B12MountainYouControlEntered(ev, source, g)
				},
				Targets: effects.TargetAny(),
				Effect:  effects.DamageToFirstTarget(4),
			}},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Search your library for a basic Mountain card, reveal it, put it into your hand, then shuffle.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: game.Purpose{Tutors: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLandWithSubtype("Mountain"),
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Koth, Fire of Resistance — a basic Mountain card, into your hand",
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−3: Koth deals damage to target creature equal to the number of Mountains you control.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					n := effects.S82OneB20LandsControlled(g, item.Controller, ctx.TextSubtype("Mountain"))
					if n <= 0 {
						return nil
					}
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
				},
			},
			{
				Label:  "−7: You get an emblem with \"Whenever a Mountain you control enters, this emblem deals 4 damage to any target.\"",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}
