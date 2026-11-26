package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strata Scythe — Artifact — Equipment {3} (Commander 2014):
//
//	"Imprint — When this Equipment enters, search your library for a land
//	 card, exile it, then shuffle.
//	 Equipped creature gets +1/+1 for each land on the battlefield with the
//	 same name as the exiled card.
//	 Equip {3}"
//
// The exile is linked to this Equipment object (CR 607.2a), so a Scythe
// that leaves and returns has forgotten it (CR 400.7). The count is every
// land on the battlefield, whoever controls it, whose name is the exiled
// card's. With no card exiled (an empty library, or the Scythe is a new
// object) the bonus is +0/+0.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a500b724-bf28-47a3-b434-6efdebbcc04c",
		Name:         "Strata Scythe",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Strata Scythe — imprint: search your library for a land card, exile it, then shuffle", strataScytheImprint),
		},
		Static:    []game.StaticAbility{effects.PumpAttachedPer(1, 1, strataScytheLands)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}

func strataScytheImprint(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	return effects.SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: func(c game.Card) bool { return c.IsLand() },
		Dest:      game.ZoneExile,
		Limit:     1,
		Shuffle:   true,
		LinkTo:    game.PermissionCardRef(ref),
		Reason:    "Strata Scythe — a land card to exile",
	}.Apply(ctx)
}

func strataScytheLands(g *game.Game, source *game.Card) int {
	exiled := g.ExiledWithForEffect(game.ObjectRefOf(*source))
	if len(exiled) == 0 || g.Battlefield == nil {
		return 0
	}
	name := exiled[len(exiled)-1].Name
	n := 0
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.IsLand() && c.Name == name {
			n++
		}
	}
	return n
}
