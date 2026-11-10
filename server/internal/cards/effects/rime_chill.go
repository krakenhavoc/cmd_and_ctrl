package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rime Chill — Instant {6}{U} (Lorwyn Eclipsed, #4458):
//
//	"Vivid — This spell costs {1} less to cast for each color among
//	 permanents you control.
//	 Tap up to two target creatures. Put a stun counter on each of them.
//	 Draw a card."
//
// A self cost reduction counted as the spell is priced (CR 601.2f): the
// colors among the caster's permanents then, colorless counting none, so at
// most {5} less. Vivid is an ability word with no rules meaning. The
// targets are "up to two": each one still legal as the spell resolves is
// tapped and gets a stun counter (Loch Mare's tap-and-stun, CR 122.1g's
// "if it would become untapped, remove a stun counter instead"), and the
// card is drawn even with no target chosen; if every chosen target is
// illegal the spell does not resolve, so no card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bad84deb-3f2c-46fd-9dcb-4e3ba9732769",
		Name:         "Rime Chill",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(rimeChillColorsAmongYourPermanents,
				"Vivid — This spell costs {1} less to cast for each color among permanents you control."),
		},
		Targets: TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}

// rimeChillColorsAmongYourPermanents is the number of colors among the
// caster's permanents, read as the spell is priced.
func rimeChillColorsAmongYourPermanents(q game.CostQuery) int {
	if q.Game == nil || q.Game.Battlefield == nil {
		return 0
	}
	n := 0
	for _, color := range []string{"W", "U", "B", "R", "G"} {
		for i := range q.Game.Battlefield.Cards {
			c := &q.Game.Battlefield.Cards[i]
			if c.Controller == q.Controller && c.HasColor(color) {
				n++
				break
			}
		}
	}
	return n
}
