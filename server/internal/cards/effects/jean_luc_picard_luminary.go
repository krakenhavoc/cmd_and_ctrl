package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jean-Luc Picard, Luminary — Legendary Creature — Human Officer
// {1}{U}{R}{W}, 2/4 (Star Trek Commander, #4974):
//
//	"Whenever another historic permanent you control with mana value 3
//	 or greater enters, draw a card. (Artifacts, legendaries, and Sagas
//	 are historic.)
//	 Seek out — Whenever you draw your second card each turn, put a
//	 +1/+1 counter on up to one other target creature and two charge
//	 counters on up to one target Spacecraft."
//
// Historic (CR 700.6) and the mana value are read off the permanent
// that entered, after layers (b09IsHistoric), so an animated artifact
// counts and a token's mana value is its copied cost or zero.
//
// The seek-out trigger (trekB31SeekOut) has two independent "up to one"
// clauses (Clauses, ADR 0065): the creature is any creature other than
// this Picard as an object, the Spacecraft any Spacecraft permanent. The
// clauses need not differ (CR 601.2c), so a Spacecraft creature may be
// both. Each is re-checked as the trigger resolves (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "b95a6001-2773-43c1-a5ad-a37ef7c9f054",
		Name:         "Jean-Luc Picard, Luminary",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && b09IsHistoric(c) && c.ManaValue() >= 3
			}, "Jean-Luc Picard, Luminary — draw a card", Do(DrawCards{N: 1})),
			Targeting(
				trekB31SeekOut("Jean-Luc Picard, Luminary — a +1/+1 counter on up to one other creature and two charge counters on up to one Spacecraft",
					jeanLucPicardLuminarySeekOut),
				Clauses(
					Another(TargetCreature("up to one other target creature").WithCount(0, 1)),
					TargetPermanent("up to one target Spacecraft", HasSubtype("Spacecraft")).WithCount(0, 1),
				)),
		},
	})
}

func jeanLucPicardLuminarySeekOut(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
		if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
		return AddCounter{Target: t.ID, Kind: game.CounterCharge, N: 2}.Apply(ctx)
	}
	return nil
}
