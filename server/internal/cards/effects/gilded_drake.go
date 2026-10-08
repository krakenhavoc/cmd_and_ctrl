package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gilded Drake - 2/2 Creature - Drake for {1}{U}:
//
//	"Flying
//	 When this creature enters, exchange control of this creature and
//	 up to one target creature an opponent controls. If you don't or
//	 can't make an exchange, sacrifice this creature. This ability
//	 still resolves if its target becomes illegal."
//
// The seam is the last sentence (#2182, ADR 0019's 2026-10-08
// amendment): the target clause is marked StillResolves, so
// spellAllTargetsIllegalLocked no longer removes the trigger under CR
// 608.2b when its only target is gone. The trigger then runs, the
// exchange cannot happen (CR 701.12b, and the explicit legality check
// below for a target that is still on the battlefield but is no longer
// a legal pick, e.g. it gained hexproof), and the Drake is sacrificed.
//
// "Up to one" means a Drake that enters with no opposing creature, or
// whose controller picks none, also goes on the stack and is
// sacrificed, which is the printed behaviour. A Drake flickered in
// response is a new object and is left alone (SacrificePermanent's
// source guard).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "7f06c098-6482-4bf3-a9a1-110d6d5b5703",
		Name:            "Gilded Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetCreature("up to one target creature an opponent controls", OpponentControls()).
				WithCount(0, 1).StillResolves(),
			Key: "Gilded Drake - exchange control with up to one target creature, else sacrifice",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				exchanged := false
				if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard &&
					ctx.IsTargetLegal(item.Targets[0]) {
					var err error
					exchanged, err = ExchangeControl{
						A:     item.SourceCardID,
						B:     item.Targets[0].ID,
						Label: "Gilded Drake - exchange control",
					}.ApplyAndReport(ctx)
					if err != nil {
						return err
					}
				}
				if exchanged {
					return nil
				}
				return SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}
