package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gisa, Glorious Resurrector — Legendary Creature — Human Wizard {2}{B}{B},
// 4/4:
//
//	"If a creature an opponent controls would die, exile it instead.
//	 At the beginning of your upkeep, put all creature cards exiled with
//	 Gisa onto the battlefield under your control. They gain decayed."
//
// The replacement is Liesa's (a CR 614 exile in place of the death, so
// the creature's dies-triggers never fire, CR 700.4) with the exile
// linked to Gisa (ReplacementEvent.ExiledWith, CR 607.2a), as The Darkness
// Crystal does. It exiles tokens too, as printed; they cease to exist and
// are never "creature cards exiled with Gisa".
//
// The upkeep trigger reads the link as it resolves, takes the creature
// cards still in exile, and puts them onto the battlefield together
// under Gisa's controller's control (CR 603.6a). Decayed is then
// granted to each new object with no duration (CR 611.2a), pinned to it
// (CR 400.7): a creature that leaves and returns loses it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1354b1b-896c-4492-b0c0-3dd07c6d3917",
		Name:         "Gisa, Glorious Resurrector",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: effects.RfOpponentCreatureWouldDie,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				ev.ExiledWith = game.PermissionCardRef{ID: src.InstanceID, Epoch: src.ObjectEpoch}
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Gisa, Glorious Resurrector: exile instead of dying",
		}},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Gisa, Glorious Resurrector — put all creature cards exiled with Gisa onto the battlefield under your control; they gain decayed", gisaUpkeep),
		},
	})
}

// gisaDecayed gives each creature that entered decayed for good.
var gisaDecayed = effects.TogetherTail("gisa/decayed", func(ctx *effects.Context, ids []uuid.UUID, _ int) error {
	for _, id := range ids {
		if err := (effects.ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.AddKeywordsMod(game.KeywordDecayed)},
			Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
			Label:    "Gisa, Glorious Resurrector — gains decayed",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

func gisaUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ids := effects.BorgCubeCreatureCardsExiledWith(ctx)
	if len(ids) == 0 {
		return nil
	}
	return effects.ReturnFromExileTogether{
		Targets:    ids,
		Controller: ctx.Controller(),
		ThenTail:   gisaDecayed,
	}.Apply(ctx)
}
