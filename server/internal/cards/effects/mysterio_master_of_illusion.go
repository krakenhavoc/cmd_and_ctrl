package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mysterio, Master of Illusion — Legendary Creature — Human Villain {3}{U},
// 3/3 (Marvel's Spider-Man promos, #4460):
//
//	"When Mysterio enters, create a 3/3 blue Illusion Villain creature
//	 token for each nontoken Villain you control. Exile those tokens when
//	 Mysterio leaves the battlefield."
//
// The count is taken as the trigger resolves: the nontoken permanents the
// controller controls that are Villains on the layered view, Mysterio
// himself among them while he is still there. The tokens are one creation
// (CR 701.7b's window sees them together), and the creation's continuation
// is told which tokens landed.
//
// "Exile those tokens when Mysterio leaves the battlefield" is a delayed
// triggered ability the resolution creates (CR 603.7a), watching Mysterio
// as the object he is then: Animate Dead's "when this leaves" condition,
// over a duration pinned to him so a Mysterio that comes back is a new
// object the old trigger never watches (CR 400.7). It exiles each of those
// tokens still on the battlefield (CR 603.7c). A Mysterio that has already
// left as the enters trigger resolves creates the tokens and no trigger:
// its leaving happened before the delayed ability existed (CR 603.7a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "40f664d8-4c73-44aa-8fb1-d38771a42520",
		Name:         "Mysterio, Master of Illusion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mysterio, Master of Illusion — a 3/3 Illusion Villain for each nontoken Villain you control; exile them when Mysterio leaves",
				mysterioMasterOfIllusionTokens),
		},
	})
}

func mysterioMasterOfIllusionTokens(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	villain := ctx.TextSubtype("Villain")
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && !c.IsToken() && c.HasSubtype(villain) {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("3/3 blue Illusion Villain")), Count: n}},
	}, ctx.ItemTail(mysterioMasterOfIllusionExileLater))
}

// mysterioMasterOfIllusionExileLater is the creation's continuation: the
// delayed "exile those tokens when Mysterio leaves the battlefield" over
// the tokens that landed, while Mysterio is still the object that entered.
// An on-disk key: never renamed, never reused.
var mysterioMasterOfIllusionExileLater = CardTail("mysterio-master-of-illusion/exile-tokens-when-he-leaves",
	func(ctx *Context, o game.TailOutcome) error {
		g, item := ctx.Game, ctx.Item
		if len(o.IDs) == 0 || item == nil {
			return nil
		}
		self := item.SourceCardID
		if g.AbilitySourceLeftBattlefieldForEffect(item) || g.AbilitySourceIsNewObjectForEffect(item) || !onBattlefield(g, self) {
			return nil
		}
		c, ok := g.LookupCardForEffect(self)
		if !ok {
			return nil
		}
		d := g.PinnedTo(game.IndefiniteDuration(), self)
		g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller:   item.Controller,
			SourceCardID: self,
			Label:        "Mysterio, Master of Illusion — exile those tokens",
			On:           []game.EventKind{game.EventLTB},
			Condition:    thisAuraLeftCondition,
			CondParams:   game.EffectParams{Object: game.ObjectRef{ID: self, Epoch: c.ObjectEpoch}},
			Cards:        append([]uuid.UUID(nil), o.IDs...),
			Duration:     &d,
			Body:         exileListedCardsBody,
		})
		return nil
	})
