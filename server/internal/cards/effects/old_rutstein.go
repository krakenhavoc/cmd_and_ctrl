package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Old Rutstein — {1}{B}{G} Legendary Creature — Human Peasant 1/4
// (Innistrad: Crimson Vow, S82):
//
//	"When Old Rutstein enters and at the beginning of your upkeep, mill
//	 a card. If a land card is milled this way, create a Treasure
//	 token. If a creature card is milled this way, create a 1/1 green
//	 Insect creature token. If a noncreature, nonland card is milled
//	 this way, create a Blood token."
//
// One trigger on two events (Gathering Stone's enters-or-upkeep). The
// milled card is read back from the mill's own tail (Dune Chanter's
// "milled this way"), where it is in the graveyard; a card a
// replacement sent elsewhere was not milled and makes nothing. The
// three sentences are independent ifs on the one card, so a land
// creature (Dryad Arbor) makes both a Treasure and an Insect, and a
// noncreature nonland card makes only the Blood.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f5ffb0b-0585-4a37-8048-eda53990d1dd",
		Name:         "Old Rutstein",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventBeginUpkeep}, thisEntersOrYourUpkeep,
				"Old Rutstein — mill a card; a land makes a Treasure, a creature an Insect, a noncreature nonland card a Blood token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 1, ThenTail: ctx.ItemTail(oldRutsteinMilled)}.Apply(ctx)
				}),
		},
	})
}

// oldRutsteinMilled makes what the milled card earns: Treasure for a land
// card, an Insect for a creature card, Blood for a noncreature, nonland
// card. An on-disk key.
var oldRutsteinMilled = CardTail("old-rutstein/after-mill", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if c.IsLand() {
			if err := (CreateToken{Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		if c.IsCreature() {
			if err := (CreateToken{Template: TokenCard("1/1 green Insect"), N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		if !c.IsCreature() && !c.IsLand() {
			if err := (CreateToken{Template: BloodToken(), N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
})
