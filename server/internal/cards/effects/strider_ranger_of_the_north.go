package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const striderLabel = "Strider, Ranger of the North — target creature gets +1/+1 until end of turn, then first strike if its power is 4 or greater"

// Strider, Ranger of the North — Legendary Creature — Human Ranger
// {2}{R}{G}, 4/4 (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Landfall — Whenever a land you control enters, target creature gets
//	 +1/+1 until end of turn. Then if that creature has power 4 or greater,
//	 it gains first strike until end of turn."
//
// Genemorph Imago's targeted landfall. "Then" orders the halves: the
// +1/+1 lands first, so it counts toward the power read, which is the
// creature's power as the ability resolves, through the layers. A target
// that left in response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "118f2cb6-2af9-4acc-b386-a8e8dded5793",
		Name:         "Strider, Ranger of the North",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(Landfall(striderLabel, striderPump), TargetCreature("target creature")),
		},
	})
}

func striderPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Strider, Ranger of the North — +1/+1"}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.PowerForComparison() < 4 {
		return nil
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: "Strider, Ranger of the North — first strike"}.Apply(ctx)
}
