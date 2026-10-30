package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hercules, Olympian Hero — {3}{W}{W} Legendary Creature — Demigod Warrior
// Hero 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever Hercules attacks, put a +1/+1 counter on him. He gains
//	 indestructible until end of turn.
//	 Whenever Hercules is dealt damage for the first time each turn, put
//	 that many +1/+1 counters on him. (He must survive the damage to get
//	 the counters.)"
//
// The attack trigger puts the counter on, then grants indestructible, both
// only while he is still the same permanent. The damage trigger is Red
// Hulk's shape: it waits for its damage batch to settle (CR 120.3, CR
// 603.2c), so damage from several blockers at once is one trigger and "that
// many" is the batch total. "For the first time each turn" is the per-object
// tally of this ability's own announcements: once it has gone on the stack
// this turn, later damage does not trigger it. A Hercules that died to the
// damage never gets the counters (he is not the same permanent), matching
// the reminder text.
//
// No simplification.
func init() {
	const damageLabel = "Hercules — put that many +1/+1 counters on him"
	Register(Spec{
		OracleID:     "ee3dfdb2-07ef-4ce2-9806-0a5b7f19aa67",
		Name:         "Hercules, Olympian Hero",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Hercules — put a +1/+1 counter on him and he gains indestructible until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !b09SourceStillOnBattlefield(g, item) {
						return nil
					}
					if err := plusOneCountersOnThis(1)(g, item); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{
						Target:   item.SourceCardID,
						Keywords: []string{"indestructible"},
						Label:    "Hercules, Olympian Hero",
					}.Apply(NewContext(g, item))
				}),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b35SelfWasDealtDamage(ev, source) && !b11TriggeredThisTurn(g, source.InstanceID, damageLabel)
				},
				Key:        damageLabel,
				AtBatchEnd: ThatMuchDamage(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || !b09SourceStillOnBattlefield(g, item) {
						return nil
					}
					return AddCounter{
						Target: item.SourceCardID, Kind: game.CounterPlusOne, N: item.Trigger.Event.Amount,
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
