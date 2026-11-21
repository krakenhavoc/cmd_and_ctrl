package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jolene, the Plunder Queen — Legendary Creature — Human Warrior
// {2}{R}{G}, 2/2 (Streets of New Capenna Commander):
//
//	"Whenever a player attacks one or more of your opponents, that
//	 attacking player creates a Treasure token.
//	 If you would create one or more Treasure tokens, instead create those
//	 tokens plus an additional Treasure token.
//	 Sacrifice five Treasures: Put five +1/+1 counters on Jolene."
//
// The attack trigger fires once per attacking player per declaration
// (CR 603.2c), however many creatures and however many of Jolene's
// controller's opponents they aim at: the engine's once-per-batch guard
// keyed on the attacker. An attack at a planeswalker or a battle is not an
// attack on an opponent (CR 508.3e), and the opponent attacking Jolene's
// controller is not attacking "one of your opponents". The attacker may be
// Jolene's controller. The Treasure is the ATTACKER's, so it is the
// replacement that matters when that is Jolene's controller: their
// Treasure is a pair, exactly as Xorn's. An opponent of Jolene's
// controller who attacks gets one Treasure.
//
// The replacement is Xorn's, a CR 614 token-creation row that appends one
// Treasure to a group that has any, once per instruction, only for tokens
// Jolene's controller creates.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8215f4a2-7131-426b-ad9c-6427caabd715",
		Name:         "Jolene, the Plunder Queen",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatchPerAttacker(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				opp := effects.AttackedPlayer(g, ev)
				if opp == uuid.Nil || opp == source.Controller {
					return false
				}
				p := g.PlayerByIDForEffect(opp)
				return p != nil && !p.Eliminated
			}, "Jolene, the Plunder Queen — the attacking player creates a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Trigger.Event.Actor, Template: effects.TreasureToken(), N: 1}.Apply(effects.NewContext(g, item))
				})),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventTokenCreated},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCreateTokens &&
					ev.TokenController == src.Controller &&
					ev.TokenTemplatesMatch(func(t game.Card) bool { return t.HasSubtype("Treasure") })
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.TokenGroups = append(ev.TokenGroups, game.TokenGroup{Template: effects.RuleDefinedToken(effects.TreasureToken()), Count: 1})
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Jolene, the Plunder Queen: an additional Treasure",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice five Treasures: Put five +1/+1 counters on Jolene.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.SacrificeN(5, "five Treasures", effects.Subtype("Treasure")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 5}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
