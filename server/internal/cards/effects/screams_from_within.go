package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Screams from Within — Enchantment — Aura {1}{B}{B} (Secrets of
// Strixhaven Commander, #4834):
//
//	"Enchant creature
//	 Enchanted creature gets -1/-1.
//	 When enchanted creature dies, return this card from your graveyard to
//	 the battlefield."
//
// The static is the layer-7c -1/-1 on the attached creature. The dies
// trigger is Angelic Destiny's and Ghoulish Impetus's: it triggers while
// the Aura is still attached (the dies event comes before the state-based
// action that puts the fallen-off Aura into the graveyard), and resolves
// with the Aura one move later in a graveyard (sourceCardInGraveyardAfterLeaving),
// which is the object it means by "this card". It returns only from the
// graveyard of the ability's controller ("your graveyard", the 2004
// ruling): a Screams from Within another player took control of goes to
// its owner's graveyard and does nothing.
//
// The Aura returns without being cast, so its controller chooses what it
// enchants as it enters (CR 303.4f, the 2005 ruling): any creature that
// can be enchanted, including the controller's own, and with none it
// stays in the graveyard (CR 303.4g). The creature that died is gone, so
// the Aura lands on a different creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3b3b9ae-1e48-481d-80a5-7e2a26cf2f50",
		Name:         "Screams from Within",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{PumpAttached(-1, -1)},
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Screams from Within — return this card from your graveyard to the battlefield",
				screamsFromWithinReturn),
		},
	})
}

func screamsFromWithinReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !sourceCardInGraveyardAfterLeaving(ctx) {
		return nil
	}
	if c, ok := g.LookupCardForEffect(item.SourceCardID); !ok || c.Owner != item.Controller {
		return nil
	}
	return ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneBattlefield}.Apply(ctx)
}
