package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brutal Hordechief — Creature — Orc Warrior {3}{B}, 3/3:
//
//	"Whenever a creature you control attacks, defending player loses 1
//	 life and you gain 1 life.
//	 {3}{R/W}{R/W}: Creatures your opponents control block this turn if
//	 able, and you choose how those creatures block."
//
// The drain is one trigger per attacking creature (CR 508.3a), the
// defending player read off the attack as it is declared (Generous
// Plunderer's b17DefendingPlayer), so a creature attacking a
// planeswalker drains its controller.
//
// The activated ability is two effects for the turn. "Block this turn if
// able" is a block requirement (CR 509.1c) on every creature the
// ability's controller's opponents control, a rule-scoped record
// (ScopeOpponentsCreatures, read live, CR 611.2c: it changes no
// characteristic, so a creature that arrives later this turn is covered
// too). "You choose how those creatures block" is a DeclarationControl
// for the turn, blockers only (game/declaration_control.go, #4154): in
// each declare blockers step the ability's controller declares every
// other defending player's blocks, all of them opponents. With the
// requirement, every creature able to block must (the requirements are
// maximised), so what is left to choose is what each one blocks, which
// is the card's "how".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e70cac5e-674e-4df7-94c7-77f8801e3941",
		Name:         "Brutal Hordechief",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			},
			Key: "Brutal Hordechief — defending player loses 1 life and you gain 1 life",
			// The defending player is read as the attack is declared
			// (ADR 0041 P9's fill-in Build), as Generous Plunderer's is.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Brutal Hordechief — defending player loses 1 life and you gain 1 life")
				item.Params.Player = b17DefendingPlayer(g, ev)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if defender := item.Params.Player; defender != uuid.Nil && defendingPlayerStillIn(g, defender) {
					if err := (GainLife{Player: defender, Amount: -1}).Apply(ctx); err != nil {
						return err
					}
				}
				return GainLife{Player: item.Controller, Amount: 1}.Apply(ctx)
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{3}{R/W}{R/W}: Creatures your opponents control block this turn if able, and you choose how those creatures block.",
			Cost:  ManaCost("{3}{R/W}{R/W}"),
			// It changes a fight, and only in a combat window: the
			// blocks are forced and chosen (ADR 0142's combat tier).
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				g.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeOpponentsCreatures, item.Controller,
					[]game.Mod{game.AddBlockRequirementMod(game.BlockRequirementBlocks)}, DurationUntilEndOfTurn(ctx),
					"Brutal Hordechief — creatures your opponents control block this turn if able")
				g.BeginDeclarationControlForEffect(game.DeclarationControl{
					Declarer:   item.Controller,
					Source:     item.SourceCardID,
					SourceName: "Brutal Hordechief",
					Blockers:   true,
				})
				return nil
			},
		}},
	})
}
