package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Reaper, King No More — Legendary Artifact Creature — Scarecrow
// {2/B}{2/R}{2/G}, 3/3 (Lorwyn Eclipsed Commander, #5368):
//
//	"When The Reaper enters, put a -1/-1 counter on each of up to two
//	 target creatures.
//	 Whenever a creature an opponent controls with a -1/-1 counter on it
//	 dies, you may put that card onto the battlefield under your control.
//	 Do this only once each turn."
//
// The enters trigger is one clause of up to two creatures; each still legal as
// it resolves takes a counter (CR 608.2b). The second reads the -1/-1 counter
// the creature last had and the controller it died under (Necroskitter's
// shape), so a creature of mine an opponent stole does not count. The "you
// may" is asked as it resolves (Optional), the card comes back only while it
// is still the graveyard object that died (CR 400.7), under the Reaper's
// controller. "Only once each turn" limits the DOING (b28DoneThisTurn), as
// Baron Strucker's does: a declined "may" leaves the turn's one return for a
// later creature. A token has ceased to exist (CR 111.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39b67a4d-6a87-41f0-a86f-b66671ccc20d",
		Name:         "The Reaper, King No More",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("The Reaper — put a -1/-1 counter on each of up to two target creatures", cmdr2B25CountersOnEachTarget),
				Harmful(TargetCreature("up to two target creatures").WithCount(0, 2))),
			Optional(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return cmdr2B25OpponentCreatureDiedWithACounter(ev, source, g, true) && !b28DoneThisTurn(g, source.InstanceID, cmdr2B25ReaperDone)
			}, "The Reaper — you may put that card onto the battlefield under your control", cmdr2B25ReaperReturn),
				"The Reaper — put that card onto the battlefield under your control?"),
		},
	})
}

const cmdr2B25ReaperDone = "The Reaper, King No More — a creature was put onto the battlefield this turn"

func cmdr2B25ReaperReturn(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || b28DoneThisTurn(g, item.SourceCardID, cmdr2B25ReaperDone) {
		return nil
	}
	ctx := NewContext(g, item)
	dead := item.Trigger.Event.CardID
	if !diedCardStillInGraveyard(ctx, dead) {
		return nil
	}
	b28NoteDoneThisTurn(g, item.SourceCardID, cmdr2B25ReaperDone)
	return ReturnFromGraveyard{Target: dead, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
}
