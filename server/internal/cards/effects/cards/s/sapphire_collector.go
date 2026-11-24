package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sapphire Collector — Creature — Human Mercenary {2}{R}, 3/3 (Mystery
// Booster Commander, #5948):
//
//	"Prowess
//	 When you cast your second noncreature spell in a turn, conjure a
//	 card named Mox Sapphire into your hand. This ability triggers only
//	 once.
//	 {2}{U}: Target instant or sorcery card in your graveyard gains
//	 flashback until end of turn. The flashback cost is equal to its mana
//	 cost."
//
// "Your second noncreature spell in a turn" is the cast that makes this
// turn's noncreature count two (CastTally.Noncreature). The conjured Mox
// Sapphire is a real card (ADR 0099 amendment 2026-11-24). The flashback
// grant is Snapcaster Mage's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7de34763-6d81-4bec-92cf-61a0d15ad45f",
		Name:            "Sapphire Collector",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnce(effects.On(game.EventCast, YouCastYourSecondNoncreatureSpellInATurn,
				"Sapphire Collector — conjure a card named Mox Sapphire into your hand",
				effects.Do(effects.CheckConjure("Sapphire Collector", effects.Conjure{Names: []string{"Mox Sapphire"}, Zone: game.ZoneHand})))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Target instant or sorcery card in your graveyard gains flashback until end of turn. The flashback cost is equal to its mana cost.",
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.TargetCardInGraveyard("target instant or sorcery card in your graveyard", effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ids := effects.LegalTargetCardIDs(ctx)
				if len(ids) == 0 {
					return nil
				}
				return effects.GrantFlashbackToCard{Target: ids[0], Label: "Flashback — its mana cost (Sapphire Collector)"}.Apply(ctx)
			},
		}},
	})
}

// YouCastYourSecondNoncreatureSpellInATurn is "when you cast your second
// noncreature spell in a turn".
func YouCastYourSecondNoncreatureSpellInATurn(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return effects.YouCast(effects.Noncreature())(ev, source, lki, g) && g.CastTallyFor(ev.Actor).Noncreature == 2
}
