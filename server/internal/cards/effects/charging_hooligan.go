package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charging Hooligan — Creature — Human Peasant {3}{R}, 3/3 (Wilds of
// Eldraine, #4471):
//
//	"Whenever this creature attacks, it gets +1/+0 until end of turn for
//	 each attacking creature. If a Rat is attacking, this creature gains
//	 trample until end of turn."
//
// Goblin Rabblemaster's attack pump, counted at resolution over every
// attacking creature on the battlefield, the Hooligan included (the
// printed line has no "other"), as a one-time until-end-of-turn boost.
// The trample half reads the same attackers for a Rat, whoever controls
// it; "Rat" is a word of the card's text (CR 612.2). A Hooligan that has
// left the battlefield gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c808892a-83d5-4b09-b0da-118b0d6fe58e",
		Name:         "Charging Hooligan",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Charging Hooligan — +1/+0 for each attacking creature, and trample if a Rat is attacking",
				chargingHooliganPump),
		},
	})
}

// chargingHooliganPump is the attack trigger's body.
func chargingHooliganPump(g *game.Game, item *game.StackItem) error {
	if !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	ctx := NewContext(g, item)
	rat := ctx.TextSubtype("Rat")
	attackers := b26AttackingCreatures(g)
	ratAttacking := false
	for _, c := range attackers {
		if c.HasSubtype(rat) {
			ratAttacking = true
		}
	}
	if err := (BoostUntilEOT{Target: item.SourceCardID, Power: len(attackers), Label: "Charging Hooligan — +1/+0 for each attacking creature"}).Apply(ctx); err != nil {
		return err
	}
	if !ratAttacking {
		return nil
	}
	return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"trample"}, Label: "Charging Hooligan — trample"}.Apply(ctx)
}
