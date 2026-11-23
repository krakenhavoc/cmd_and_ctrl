package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_jou_b1_helpers.go — bodies shared by the Journey into Nyx batch 1
// cards (#5558, S82): the strive pumps (Phalanx Formation, Rouse the Mob,
// Cruel Feeding, Desperate Stand, Aerial Formation, Ajani's Presence) and
// Pheres-Band Thunderhoof's two counters.

// jouStrivePump is "any number of target creatures each get +P/+T and gain
// <keywords> until end of turn". Each target still legal as the spell
// resolves gets the boost and the keywords (CR 608.2b); one that left in
// response gets nothing, and zero targets does nothing.
func JouStrivePump(power, toughness int, label string, keywords ...string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard || !OnBattlefield(ctx.Game, t.ID) {
				continue
			}
			if err := jouPumpAndGrant(ctx, t.ID, power, toughness, label, keywords); err != nil {
				return err
			}
		}
		return nil
	}
}

// jouPumpAndGrant gives one creature the boost and the keywords until end
// of turn.
func jouPumpAndGrant(ctx *Context, id uuid.UUID, power, toughness int, label string, keywords []string) error {
	if err := (BoostUntilEOT{Target: id, Power: power, Toughness: toughness, Label: label}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: keywords, Label: label}.Apply(ctx)
}

// thisGetsTwoPlusOneCounters is "put two +1/+1 counters on this creature"
// for the source of a trigger (Pheres-Band Thunderhoof's heroic, Weird
// Angel Flame's sticker), nothing if the source is a new object (CR 400.7).
func ThisGetsTwoPlusOneCounters(g *game.Game, item *game.StackItem) error {
	if SourceIsNewObject(g, item) || !OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}.Apply(NewContext(g, item))
}
