package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dead Man's Chest — Enchantment — Aura {1}{B}:
//
//	"Enchant creature an opponent controls
//	 When enchanted creature dies, exile cards equal to its power from the
//	 top of its owner's library. You may cast spells from among those
//	 cards for as long as they remain exiled, and mana of any type can be
//	 spent to cast them."
//
// The power is the creature's last-known power as it died (CR 608.2h), and
// the library is its OWNER's, which can differ from the opponent who
// controlled it. The Aura's controller gets the permission: cast only (a
// land among the cards is stranded), for as long as each card stays in
// exile, with mana of any type (Court of Locthwain's grant, which pays {C}
// for coloured costs too).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5fdd6fd-6492-47ec-8b74-9ff22ff9849a",
		Name:         "Dead Man's Chest",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.OpponentControls()),
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Dead Man's Chest — exile cards equal to its power from the top of its owner's library; you may cast them",
				deadMansChestResolve),
		},
	})
}

func deadMansChestResolve(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	dead := item.Trigger.Event.CardID
	info, ok := g.LastKnownPermanentForEffect(dead)
	if !ok || info.Power <= 0 {
		return nil
	}
	owner := info.Controller
	if c, found := g.LookupCardForEffect(dead); found {
		owner = c.Owner
	}
	return effects.ExileTopWithPermission{
		From:        owner,
		GrantTo:     item.Controller,
		N:           info.Power,
		CastOnly:    true,
		AnyType:     true,
		WhileExiled: true,
	}.Apply(effects.NewContext(g, item))
}
