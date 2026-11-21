package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liberator, Urza's Battlethopter — Legendary Artifact Creature —
// Thopter {3}, 1/2 (The Brothers' War):
//
//	"Flash
//	 Flying
//	 You may cast colorless spells and artifact spells as though they had
//	 flash.
//	 Whenever you cast a spell, if the amount of mana spent to cast that
//	 spell is greater than Liberator's power, put a +1/+1 counter on
//	 Liberator."
//
// The timing grant is Mystic Forge's filter ("artifact spells and
// colorless spells", lands excluded) on Blur of Heroism's flash grant.
// The trigger reads the mana spent as the spell is cast and carries it
// (Aberrant Manawurm's Build); the intervening "if" (CR 603.4) is
// checked when it triggers, against Liberator's power as it is then, and
// again as it resolves. Liberator must still be the same permanent.
func init() {
	const label = "Liberator, Urza's Battlethopter — put a +1/+1 counter on Liberator"
	t := effects.WheneverYouCast(nil, label, effects.Do(effects.If{
		Cond: effects.CondBoard{Name: "the mana spent is greater than this creature's power", Test: liberatorBeatenNow},
		Then: []effects.Applier{effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}},
	}))
	inner := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		if inner != nil && !inner(ev, source, ch, g) {
			return false
		}
		return liberatorBeaten(g, source.InstanceID, g.StackItemPaidForEffect(ev.CardID).Spent().Total())
	}
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	effects.Register(effects.Spec{
		OracleID:        "798a4a6f-7443-4e8e-afcb-0a694fdfcbb3",
		Name:            "Liberator, Urza's Battlethopter",
		Completeness:    effects.CompletenessFull,
		ReadsManaSpent:  game.ReadsYourCasts,
		PrintedKeywords: []string{"flash", "flying"},
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{NonLandOnly: true, ArtifactOrColorlessOnly: true},
				"You may cast colorless spells and artifact spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{t},
	})
}

// liberatorBeaten is the intervening "if": the mana spent is greater than
// Liberator's power as it is now, read through the layers.
func liberatorBeaten(g *game.Game, liberator uuid.UUID, spent int) bool {
	c, ok := g.LookupCardForEffect(liberator)
	return ok && spent > c.PowerForComparison()
}

// liberatorBeatenNow is the intervening "if" as the item resolves: the mana
// the trigger carried against Liberator's power now, Liberator being the
// same permanent that triggered.
func liberatorBeatenNow(ctx *effects.Context) bool {
	return ctx.Item != nil && effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) &&
		liberatorBeaten(ctx.Game, ctx.Item.SourceCardID, ctx.Item.Params.Amount)
}
