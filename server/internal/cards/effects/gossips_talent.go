package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gossip's Talent — Enchantment — Class {1}{U} (Bloomburrow, #4467):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 Whenever a creature you control enters, surveil 1.
//	 {1}{U}: Level 2
//	 Whenever you attack, target attacking creature with power 3 or less
//	 can't be blocked this turn.
//	 {3}{U}: Level 3
//	 Whenever a creature you control deals combat damage to a player, you may
//	 exile it, then return it to the battlefield under its owner's control."
//
// ADR 0071 Class machinery (Bandit's Talent): LevelUp for each level, the
// level-2 and level-3 triggers gated with AtLevel (CR 716.2a).
//
//   - Level 1 is a surveil on each creature entering under your control.
//   - Level 2 is one trigger per attack declaration (OncePerBatch). Its
//     target, an attacking creature with power 3 or less, is checked as it is
//     announced and again as it resolves, so a creature pumped to power 4 in
//     between is no longer a legal target and nothing happens (the ruling);
//     one pumped afterwards stays unblockable.
//   - Level 3 is one trigger for each creature that dealt combat damage to a
//     player (the ruling), with the creature as the object it was when it
//     dealt the damage (Build stamps it). Each asks its "you may" as it
//     resolves and flickers it (Flicker, stdB27FlickerObject): the creature
//     returns as a new object, so a double striker is no longer attacking,
//     counters and Auras are gone, and a token does not come back (the
//     rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "96e6d220-87db-4dab-abaa-e80c9577d89a",
		Name:         "Gossip's Talent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			}, "Gossip's Talent — surveil 1", Do(Surveil{N: 1})),
			AtLevel(2, Targeting(OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Gossip's Talent — target attacking creature with power 3 or less can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked,
						Label: "Gossip's Talent — can't be blocked this turn"}.Apply(ctx)
				})), TargetCreature("target attacking creature with power 3 or less", AttackingCreature(), PowerLE(3)))),
			AtLevel(3, gossipsTalentFlicker()),
		},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{1}{U}")),
			LevelUp(3, ManaCost("{3}{U}")),
		},
	})
}

// gossipsTalentFlicker is the level-3 trigger: one per creature you control
// that dealt combat damage to a player, optional as it resolves.
func gossipsTalentFlicker() game.TriggeredAbility {
	const label = "Gossip's Talent — you may exile that creature, then return it to the battlefield under its owner's control"
	return Optional(game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return combatDamageToPlayerBy(ev, source.Controller, g)
		},
		Key: label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			if c, ok := g.LookupCardForEffect(ev.Source); ok {
				item.Params.Object = game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}
			}
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			return stdB27FlickerObject(NewContext(g, item), item.Params.Object)
		},
	}, "Gossip's Talent — exile that creature, then return it to the battlefield under its owner's control?")
}
