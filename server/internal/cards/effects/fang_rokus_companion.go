package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fang, Roku's Companion — Legendary Creature — Dragon {3}{R}{R}, 4/4
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"Flying
//	 Whenever Fang attacks, another target legendary creature you control
//	 gets +X/+0 until end of turn, where X is Fang's power.
//	 When Fang dies, if he wasn't a Spirit, return this card to the
//	 battlefield under your control. He's a Spirit in addition to his
//	 other types."
//
// X is Fang's power as the attack trigger resolves, or as he last existed
// on the battlefield if he has left (CR 608.2h).
//
// The dies trigger's "if" is intervening (CR 603.4) and reads Fang as he
// last existed (CR 603.10a): a Fang that died a Spirit, a changeling
// included, does not trigger. The card comes back from the graveyard only
// if it is still there, and the returned Fang is a Spirit for as long as
// he stays on the battlefield, so his next death does not return him.
//
// No simplification.
func init() {
	const pump = "Fang, Roku's Companion — another target legendary creature you control gets +X/+0 until end of turn, where X is Fang's power"
	Register(Spec{
		OracleID:        "6665bfe2-942c-4b3d-a976-776fe00c99f4",
		Name:            "Fang, Roku's Companion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks(pump, fangRokusCompanionPump),
				Another(TargetCreature("another target legendary creature you control", Legendary(), YouControl()))),
			On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source) && !lki.AllCreatureTypes && !keywordSliceContains(lki.Subtypes, "Spirit")
			}, "Fang, Roku's Companion — he wasn't a Spirit: return this card to the battlefield under your control as a Spirit",
				fangRokusCompanionReturn),
		},
	})
}

// fangRokusCompanionPump is the attack trigger's body.
func fangRokusCompanionPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	x := b17LastKnownPowerOffBattlefield(g, item.SourceCardID)
	if sourceIsStillThisPermanent(g, item) {
		if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
			x = c.CurrentPower()
		}
	}
	return BoostUntilEOT{Target: target, Power: x, Label: "Fang, Roku's Companion — +X/+0"}.Apply(ctx)
}

// fangRokusCompanionReturn is the dies trigger's body: the card, if it is
// still in a graveyard, returns under the trigger's controller's control.
func fangRokusCompanionReturn(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	ctx := NewContext(g, item)
	return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(id, item.Controller, false, ctx.ItemTail(fangRokusCompanionSpirit))
}

// fangRokusCompanionSpirit is "He's a Spirit in addition to his other
// types", pinned to the returned permanent. An on-disk key: never
// renamed, never reused.
var fangRokusCompanionSpirit = EnteredTail("fang-rokus-companion/spirit", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || !onBattlefield(g, entered) {
		return nil
	}
	return ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Spirit")...)},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Fang, Roku's Companion — he's a Spirit in addition to his other types",
	}.Apply(ctx)
})
