package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Red Hulk — Legendary Creature — Gamma Berserker Villain {4}{R}{R}, 6/7
// (Marvel Super Heroes, #4455):
//
//	"Reach, trample
//	 Enrage — Whenever Red Hulk is dealt damage, put a +1/+1 counter on him.
//	 When you do, he deals damage equal to the number of +1/+1 counters on
//	 him to any other target."
//
// Strong, the Brutish Thespian's enrage trigger (one trigger per batch of
// damage, CR 120.2) whose body puts the counter and, only if it really
// landed, raises the printed "when you do" as a CR 603.12 reflexive trigger.
// That trigger chooses its own "any other target" as it goes on the stack
// (the source itself is not a legal pick) and reads the counters when it
// resolves; if Red Hulk has left by then it deals the count it had when the
// counter went on (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a14a838a-49f4-45e3-9ec0-3c803af45a68",
		Name:            "Red Hulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			},
			Key:        "Red Hulk — Enrage: put a +1/+1 counter on him",
			AtBatchEnd: ThatMuchDamage(),
			Effect:     mshB7RedHulkEnrage,
		}},
	})
}

func mshB7RedHulkEnrage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return ReflexiveTrigger{
		Label:  "Red Hulk — he deals damage equal to the number of +1/+1 counters on him to any other target",
		Body:   mshB7RedHulkDamageBody,
		Params: game.EffectParams{Amount: forgottenAncientCounters(g, item.SourceCardID)},
	}.Apply(ctx)
}

var mshB7RedHulkDamageBody = game.ReflexiveBody("red-hulk/damage", simpleBody(mshB7RedHulkDamage),
	constTargets(func() *game.TargetSpec { return Another(TargetAny()) }))

func mshB7RedHulkDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	amount := item.Params.Amount
	if onBattlefield(g, item.SourceCardID) {
		amount = forgottenAncientCounters(g, item.SourceCardID)
	}
	return DealDamage{Source: item.SourceCardID, Target: targets[0].ID, Amount: amount}.Apply(ctx)
}
