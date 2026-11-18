package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wand of Orcus — Legendary Artifact — Equipment {2}{B} (Adventures in the
// Forgotten Realms Commander):
//
//	"Whenever equipped creature attacks or blocks, it and Zombies you control
//	 gain deathtouch until end of turn.
//	 Whenever equipped creature deals combat damage to a player, create that
//	 many 2/2 black Zombie creature tokens.
//	 Equip {3}"
//
// The first trigger fires for the equipped creature's attack, or its block
// (once however many attackers it blocks, CR 509.3a). The set that gains
// deathtouch is locked as it resolves (CR 611.2c): the creature itself, if it
// is still on the battlefield, and the Zombie creatures you control then. The
// second creates as many tokens as the combat damage the creature dealt to
// that player in that step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9fd521d5-baa3-4054-a990-d23265863c71",
		Name:         "Wand of Orcus",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attachedCreatureAttacked(ev, source) ||
						(ev.Kind == game.EventBlock && source.IsAttachedTo(ev.CardID) && ev.Target != uuid.Nil && ev.Amount <= 1)
				},
				"Wand of Orcus — it and Zombies you control gain deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					host := item.Trigger.Event.CardID
					ctx := NewContext(g, item)
					zombie := ctx.TextSubtype("Zombie")
					return GrantKeywordUntilEOT{
						Keywords: []string{"deathtouch"},
						Label:    "Wand of Orcus — deathtouch",
						Match: func(_ *game.Game, caster uuid.UUID, c game.Card) bool {
							if !c.IsCreature() {
								return false
							}
							return c.InstanceID == host || (c.Controller == caster && c.HasSubtype(zombie))
						},
					}.Apply(ctx)
				}),
			On(game.EventDealDamage,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
				},
				"Wand of Orcus — create that many 2/2 black Zombie creature tokens",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("2/2 black Zombie"), N: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
