package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Valentin, Dean of the Vein // Lisette, Dean of the Root — Modal double-faced
// card (Strixhaven: School of Mages, #5502):
//
//	Front: {B} Legendary Creature — Vampire Warlock 1/1
//	  "Menace, lifelink
//	   If a nontoken creature an opponent controls would die, exile it instead.
//	   When you do, you may pay {2}. If you do, create a 1/1 black and green
//	   Pest creature token with "When this token dies, you gain 1 life.""
//	Back: {2}{G}{G} Legendary Creature — Human Druid 4/4
//	  "Whenever you gain life, you may pay {1}. If you do, put a +1/+1 counter
//	   on each creature you control and those creatures gain trample until end
//	   of turn."
//
// The front's replacement is Kalitas's (OpponentNontokenCreatureWouldDie): the
// creature is exiled, so it never died. "When you do" is the CR 603.12
// reflexive trigger the replacement creates (Head of the Hunt's declaration,
// #4517), once per creature exiled, controlled by Valentin's controller; on
// resolving it asks for {2} and only a payment makes the Pest (CR 118.12).
//
// The back's trigger asks for {1} as it resolves. Paying puts a counter on
// each creature its controller controls right then, and the same creatures
// gain trample until end of turn (CR 611.2c): a creature that arrives later
// gets neither. The back face is registered under "<oracle_id>#1".
//
// No simplification.
func init() {
	const oracle = "6e1aa07a-5d9b-4daf-8dc4-c1717360855f"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Valentin, Dean of the Vein",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace", "lifelink"},
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: effects.OpponentNontokenCreatureWouldDie,
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return g.QueueReflexiveTriggerAfterReplacedMoveForEffect(ev, src, game.ReflexiveTrigger{
					Label: "Valentin, Dean of the Vein — you may pay {2}; if you do, create a Pest",
					Body:  valentinPestOffer,
				})
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Valentin, Dean of the Vein: exile instead of dying",
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Lisette, Dean of the Root",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Lisette, Dean of the Root — you may pay {1}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Pay {1}: put a +1/+1 counter on each creature you control, and they gain trample until end of turn?",
						Then:     lisettePaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// valentinPestOffer is the reflexive trigger: "you may pay {2}", keyed.
var valentinPestOffer = game.ReflexiveBody("s82-old-b046/valentin-pest-offer", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{2}",
		Question: "Pay {2}: create a 1/1 black and green Pest creature token?",
		Then:     valentinPestPaid,
	}.Apply(effects.NewContext(g, item))
}), nil)

// valentinPestPaid is the paid half: one Pest for the payer.
var valentinPestPaid = effects.PayUnlessBranches("pay-unless/s82-old-b046-valentin-pest",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Template: effects.PestToken(), N: 1}.Apply(ctx)
	}, nil)

// lisettePaid is the paid half: a +1/+1 counter on each creature the
// controller controls and trample for those same creatures.
var lisettePaid = effects.PayUnlessBranches("pay-unless/s82-old-b046-lisette-counters",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		var ids []uuid.UUID
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if c.Controller == ctx.Controller() && c.IsCreature() {
				ids = append(ids, c.InstanceID)
			}
		}
		for _, id := range ids {
			if !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Lisette, Dean of the Root — trample"}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
		return nil
	}, nil)
