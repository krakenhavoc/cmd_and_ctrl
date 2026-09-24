package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// God-Eternal Bontu — Legendary Creature — Zombie God {3}{B}{B}, 5/6
// (EDHREC rank 3546):
//
//	"Menace
//	 When God-Eternal Bontu enters, sacrifice any number of other
//	 permanents, then draw that many cards.
//	 When God-Eternal Bontu dies or is put into exile from the
//	 battlefield, you may put it into its owner's library third from
//	 the top."
//
// The aristocrats deck's refill that will not stay dead. The return
// is God-Eternal Oketra's, unchanged — the same two-condition
// EventLTB, the same "you may", the same tuck under the top two.
//
// "Sacrifice any number of other permanents, then draw that many
// cards" is a CHOICE, not a target — there is no "target" in the
// printed text — made on resolution (CR 608.2, ChoosePermanents's
// Sacrifice mode, #1214, #1337). It used to be modelled as "any
// number of other TARGET permanents you control" so the controller
// got the existing board picker; the resolution-time picker retired
// that. Bontu himself is excluded by instance ID rather than by name,
// which the ChoosePermanents Candidates closure can do directly —
// TargetsFrom exists for a trigger's target clause, and this is not
// one any more.
//
// With nothing else on the battlefield, ChoosePermanents skips the
// controller's leg entirely (CR 608.2c) and Then draws zero cards,
// which is the same outcome "the printed card would put it on the
// stack to do nothing" described under the old model.
func init() {
	Register(Spec{
		OracleID:        "183891b0-b5ec-47f4-8d09-b9d3cfc4e7f1",
		Name:            "God-Eternal Bontu",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "God-Eternal Bontu — sacrifice any number of other permanents, then draw that many cards",
						bontuSacrificeThenDraw)
				},
			},
			Optional(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b22SelfDiedOrWasExiledFromBattlefield(ev, source)
			}, "God-Eternal Bontu — put it into its owner's library third from the top", tuckSelfThirdFromTop), "God-Eternal Bontu — put it into its owner's library third from the top?"),
		},
	})
}

// bontuSacrificeThenDraw is the entry ability's resolution: every
// OTHER permanent the controller controls is a candidate, "any
// number" of them (0..all) may be chosen, the chosen set is
// sacrificed together, and the controller draws one card per
// permanent actually sacrificed.
func bontuSacrificeThenDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	self := item.SourceCardID
	controller := item.Controller
	return ChoosePermanents{
		Question:  "God-Eternal Bontu — sacrifice any number of other permanents",
		Sacrifice: true,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.InstanceID != self {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 0 // "any number" — 0 min, uncapped max
		},
		Then: func(ctx *Context, picked game.PromptedPicks) error {
			n := picked.Count()
			if n == 0 {
				return nil
			}
			return DrawCards{Player: controller, N: n}.Apply(ctx)
		},
	}.Apply(ctx)
}
