package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daxos's Torment — Enchantment {3}{B} (Commander 2015, #5546):
//
//	"Constellation — Whenever this enchantment or another enchantment you
//	 control enters, this enchantment becomes a 5/5 Demon creature with
//	 flying and haste in addition to its other types until end of turn."
//
// Constellation is Doomwake Giant's condition (its own entry, or another
// enchantment entering under its controller's control). The animation is
// Reverberating Summons': a layer 4 Creature type with the Demon subtype in
// addition to Enchantment, a layer 7b 5/5, and flying and haste, all until
// end of turn, on this object only; one that has left or become a new
// object is not animated (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93ef3098-d1c0-4b05-9d28-1afa133bfc4f",
		Name:         "Daxos's Torment",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b30SelfOrAnotherEnchantmentYouControlEntered(ev, source, g)
			}, "Daxos's Torment — becomes a 5/5 Demon creature with flying and haste until end of turn (constellation)",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := NewContext(g, item)
					if err := (BecomeCreatureUntilEOT{
						Target:       item.SourceCardID,
						Types:        []string{"Creature"},
						Subtypes:     []string{"Demon"},
						SetPower:     5,
						SetToughness: 5,
						Label:        "Daxos's Torment — becomes a 5/5 Demon creature",
					}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{
						Target:   item.SourceCardID,
						Keywords: []string{"flying", "haste"},
						Label:    "Daxos's Torment — flying and haste until end of turn",
					}.Apply(ctx)
				}),
		},
	})
}
