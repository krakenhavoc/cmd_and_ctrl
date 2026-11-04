package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Last Ronin — Enchantment — Saga {4}{B}{G} (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Destroy all creatures.
//	 II — Mill four cards. When you do, return target creature card from
//	 your graveyard to your hand.
//	 III — Whenever a creature you control attacks alone this turn, put
//	 three +1/+1 counters on it. It gains trample, lifelink, and
//	 indestructible until end of turn."
//
// Chapter I is the Wrath sweep. Chapter II is Primal Witchstalker's
// mill-then-reflexive (CR 603.12): the creature card is chosen as the
// reflexive trigger goes on the stack, so a card the mill just put there
// is a legal pick; an empty library mills nothing, nothing was done and
// nothing triggers.
//
// Chapter III creates a repeating delayed trigger for the rest of the
// turn (CR 603.7b, Benefactor's Draught's shape) on the attack event:
// one of the controller's creatures was declared as an attacker and it
// is the only creature declared (CR 506.5, game.AttackedAlone, Agent 13's
// read). Each additional combat's lone attacker triggers it again. The
// delayed ability resolves on the attacker it names (CR 603.7c): the
// counters and the three keywords, nothing when it has left.
//
// No simplification.
func init() {
	const name = "The Last Ronin"
	Register(Spec{
		OracleID:     "e1865bec-1744-40c2-9031-7d6363be5333",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "destroy all creatures"), func(g *game.Game, item *game.StackItem) error {
				return DestroyAllMatching{Match: Creature()}.Apply(NewContext(g, item))
			}),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "mill four cards; when you do, return target creature card from your graveyard to your hand"), Do(
				Instruction(func(ctx *Context) error {
					p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
					if p == nil || p.Library == nil || p.Library.Size() == 0 {
						return nil
					}
					ctx.RecordStep(theLastRoninMilledRecord, nil, 1)
					return MillCards{N: 4}.Apply(ctx)
				}),
				Instruction(func(ctx *Context) error {
					if ctx.StepRecord(theLastRoninMilledRecord).N() == 0 {
						return nil
					}
					return WhenYouDo("The Last Ronin — return target creature card from your graveyard to your hand",
						theLastRoninReturnBody).Apply(ctx)
				}),
			)),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "this turn, whenever a creature you control attacks alone, it gets three +1/+1 counters, trample, lifelink and indestructible"),
				func(g *game.Game, item *game.StackItem) error {
					return DelayedOnEvent{
						Label:     "The Last Ronin — put three +1/+1 counters on the creature attacking alone; it gains trample, lifelink, and indestructible until end of turn",
						On:        []game.EventKind{game.EventAttack},
						Condition: theLastRoninAttacksAloneCondition,
						Body:      theLastRoninAttacksAloneBody,
						Repeats:   true,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// theLastRoninMilledRecord files that chapter II's library had a card to
// mill.
const theLastRoninMilledRecord = "the-last-ronin/milled"

// theLastRoninReturnBody is chapter II's reflexive trigger: the chosen
// creature card goes to its owner's hand. An on-disk key.
var theLastRoninReturnBody = game.ReflexiveBody("the-last-ronin/return",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
			}
		}
		return nil
	}),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature())
	}))

// theLastRoninAttacksAloneCondition is "whenever a creature you control
// attacks alone": the declaration is the delayed ability's controller's,
// and it declared exactly one attacker (CR 506.5). An on-disk key.
var theLastRoninAttacksAloneCondition = game.DelayedCondition("attack/a-creature-you-control-attacks-alone",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		return attackDeclaredByYou(ev, dt.Controller) && game.AttackedAlone(g)
	})

// theLastRoninAttacksAloneBody is chapter III's delayed ability: three
// +1/+1 counters on the lone attacker, then trample, lifelink and
// indestructible until end of turn. An on-disk key.
var theLastRoninAttacksAloneBody = game.SimpleDelayedBody("the-last-ronin/attacks-alone", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range ctx.PayloadCards() {
		if !onBattlefield(g, id) {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 3}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"trample", "lifelink", "indestructible"},
			Label:    "The Last Ronin — trample, lifelink, and indestructible",
		}.Apply(ctx)
	}
	return nil
})
