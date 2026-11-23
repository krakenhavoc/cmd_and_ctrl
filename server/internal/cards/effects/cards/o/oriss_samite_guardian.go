package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oriss, Samite Guardian — Legendary Creature — Human Cleric {1}{W}{W},
// 1/3:
//
//	"{T}: Prevent all damage that would be dealt to target creature this
//	 turn.
//	 Grandeur — Discard another card named Oriss, Samite Guardian: Target
//	 player can't cast spells this turn, and creatures that player
//	 controls can't attack this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the first ability is the
// not-one-use shield pinned to the target. Grandeur is an ordinary
// activated ability whose cost discards a card with this name from your
// hand; "another" is any such card, since this one is on the
// battlefield. Its two bans last the turn: a cast ban on the player
// (Mandate of Peace's record) and "can't attack" on the creatures that
// player controls, read live for the rest of the turn (a rule, not a
// characteristic change, so a creature that player gains control of
// later this turn can't attack either).
//
// No simplifications.
func init() {
	const name = "Oriss, Samite Guardian"
	effects.Register(effects.Spec{
		OracleID:     "8d6e0dab-400a-4761-8343-92c7eb7e8735",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.ShieldTargetRow("{T}: Prevent all damage that would be dealt to target creature this turn.",
				effects.TapCost(), effects.TargetCreature("target creature")),
			{
				Label:   "Grandeur — Discard another card named Oriss, Samite Guardian: Target player can't cast spells this turn, and creatures that player controls can't attack this turn.",
				Cost:    effects.DiscardCardsMatching(1, "another card named Oriss, Samite Guardian", func(c game.Card) bool { return c.Name == name }),
				Targets: effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						if err := (effects.RestrictCasting{
							Player:   t.ID,
							Rule:     game.CastBanRule{Kind: game.CastBanOutright},
							Label:    name + " — can't cast spells this turn",
							Duration: effects.DurationUntilEndOfTurn(ctx),
						}).Apply(ctx); err != nil {
							return err
						}
						g.RegisterScopedRuleEffectForEffect(item.SourceCardID, game.ScopeYourCreatures, t.ID,
							[]game.Mod{game.AddRestrictionsMod(game.CantAttack)}, effects.DurationUntilEndOfTurn(ctx),
							name+" — creatures that player controls can't attack this turn")
						return nil
					}
					return nil
				},
			},
		},
	})
}
