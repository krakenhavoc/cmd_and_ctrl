package effects

import (
	"github.com/google/uuid"

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
	t := WheneverYouCast(nil, label, func(g *game.Game, item *game.StackItem) error {
		if !sourceIsStillThisPermanent(g, item) || !liberatorBeaten(g, item.SourceCardID, item.Params.Amount) {
			return nil
		}
		return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
	})
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
	Register(Spec{
		OracleID:        "798a4a6f-7443-4e8e-afcb-0a694fdfcbb3",
		Name:            "Liberator, Urza's Battlethopter",
		Completeness:    CompletenessFull,
		ReadsManaSpent:  game.ReadsYourCasts,
		PrintedKeywords: []string{"flash", "flying"},
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{NonLandOnly: true, ArtifactOrColorlessOnly: true},
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
