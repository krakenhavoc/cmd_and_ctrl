package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB008AddManaCumulativeUpkeep is "Cumulative upkeep—Add {R}" (Braid of
// Fire): the payment is adding the mana, once per age counter (CR 702.24a),
// so "yes" puts that much red mana in the controller's pool (to be spent in
// the upkeep step) and "no" sacrifices the permanent. `symbol` is the colour
// letter ("R").
func OldB008AddManaCumulativeUpkeep(label, symbol string) game.TriggeredAbility {
	branches := PayUnlessBranches("pay-unless/old-b008-cumulative-upkeep-add-"+strings.ToLower(symbol),
		func(ctx *Context, r game.PayUnlessAnswer) error {
			age := CounterCountOn(ctx.Game, r.Source, game.CounterAge)
			return ctx.Game.AddManaForEffect(r.Chooser, r.Source, strings.Repeat(symbol, age))
		},
		func(ctx *Context, r game.PayUnlessAnswer) error {
			return SacrificePermanent{Target: r.Source}.Apply(ctx)
		})
	ref := CardTail("old-b008/cumulative-upkeep-add-"+strings.ToLower(symbol), func(ctx *Context, _ game.TailOutcome) error {
		age := CounterCountOn(ctx.Game, ctx.Item.SourceCardID, game.CounterAge)
		if age <= 0 {
			return nil
		}
		return UpkeepPayUnless{
			Chooser:  ctx.Controller(),
			Cost:     "Add " + strings.Repeat("{"+symbol+"}", age),
			Action:   &game.PayAction{Kind: game.PayActionNothing},
			Question: label,
			Else:     branches,
		}.Apply(ctx)
	})
	return cumulativeUpkeepTrigger(label, ref)
}

// OldB008OliviaBoost is Olivia, Mobilized for War's "if you do": a +1/+1
// counter on the entering creature, haste until end of turn, and it becomes
// a Vampire in addition to its other types. A creature that has left the
// battlefield, or is a new object (CR 400.7), is not touched.
func OldB008OliviaBoost(ctx *Context, id uuid.UUID, epoch int) error {
	c, ok := BattlefieldCardForEffect(ctx.Game, id)
	if !ok || c.ObjectEpoch != epoch {
		return nil
	}
	if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Olivia, Mobilized for War — haste until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	return ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Vampire")...)},
		Duration: game.IndefiniteDuration(),
		Label:    "Olivia, Mobilized for War — becomes a Vampire in addition",
	}.Apply(ctx)
}
