package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kharasha Foothills — Plane — Mongseng (Commander Masters, #4870):
//
//	"Whenever a creature you control attacks a player, for each other
//	 opponent, you may create a token that's a copy of that creature,
//	 tapped and attacking that opponent. Exile those tokens at the
//	 beginning of the next end step.
//	 Whenever chaos ensues, you may sacrifice any number of creatures. If
//	 you do, Kharasha Foothills deals that much damage to target creature."
//
// A plane (ADR 0216 §6). The first ability triggers for each of the planar
// controller's creatures that attacks a player (CR 508.3a), never one that
// attacks a planeswalker (the ruling), and is myriad's shape over the
// planar controller's other opponents: each is asked in turn, the copies are
// made together from the attacker's copiable values (or its last-known
// information if it left, CR 707.2), each tapped and attacking that
// opponent as a player (never a planeswalker), and exiled at the beginning
// of the next end step. The chaos ability (CR 311.7) chooses its target as it
// goes on the stack and sacrifices nothing until it resolves (the ruling); the
// number sacrificed is the damage, dealt by the plane.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "256770fe-c17d-45dd-bf05-36eeee240324",
		Name:         "Kharasha Foothills",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			kharashaAttackTrigger(),
			Targeting(WheneverChaosEnsues("Kharasha Foothills — you may sacrifice any number of creatures; if you do, it deals that much damage to target creature",
				func(g *game.Game, item *game.StackItem) error {
					return ChoosePermanents{
						Question:   "Kharasha Foothills — sacrifice any number of creatures",
						Keyed:      kharashaSacrificed,
						Candidates: yourCreaturesAnyNumber, // "any number": no floor, no ceiling
					}.Apply(NewContext(g, item))
				}), TargetCreature("target creature")),
		},
	})
}

const kharashaAttackLabel = "Kharasha Foothills — for each other opponent, you may create a token copy of that creature, tapped and attacking them; exile it at the beginning of the next end step"

// kharashaSacrificed is the sacrifice and "if you do, deal that much
// damage to target creature". An on-disk key: never renamed, never reused.
var kharashaSacrificed = RegisterPermanentsPick("prompt-run/kharasha-foothills-sacrifice", PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		n := picked.Count()
		if n == 0 {
			return nil
		}
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
		}
		return nil
	},
})

// kharashaAttackTrigger is the plane's first ability: one trigger per
// attacking creature of the planar controller that attacks a player.
func kharashaAttackTrigger() game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventAttack},
		Key:     kharashaAttackLabel,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return attackDeclaredByYou(ev, source.Controller) && attackedPlayer(g, ev) != uuid.Nil
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, kharashaAttackLabel)
			item.Params.Object, _ = g.PermanentRefForEffect(ev.CardID)
			item.Params.Player = attackedPlayer(g, ev)
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			tmpl, ok := g.TokenCopyTemplateForObjectForEffect(item.Params.Object)
			if !ok {
				return nil
			}
			return kharashaChooseOpponents(NewContext(g, item), item.Params.Object, tmpl, b27OtherOpponents(g, item.Controller, item.Params.Player), nil)
		},
	}
}

// kharashaChooseOpponents asks, for each opponent still to ask, whether to
// create a copy attacking them; `chosen` are the opponents who have said
// yes. The attacker rides the prompt so the copy's template is read again
// when the answer arrives. With none left to ask, the copies are made
// together and exiled at the beginning of the next end step.
func kharashaChooseOpponents(ctx *Context, obj game.ObjectRef, tmpl game.Card, opponents, chosen []uuid.UUID) error {
	if len(opponents) == 0 {
		var groups []game.TokenGroup
		for _, opponent := range chosen {
			groups = append(groups, game.TokenGroup{Template: copiedToken(tmpl), Count: 1, Entry: game.TokenEntryOptions{Tapped: true}, Attacking: opponent})
		}
		return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
			Controller: ctx.Controller(), Source: ctx.Source(), Groups: groups, Together: true,
		}, ctx.ItemTail(kharashaExileTokens))
	}
	opponent := opponents[0]
	p := ctx.Game.PlayerByIDForEffect(opponent)
	if p == nil || p.Eliminated {
		return kharashaChooseOpponents(ctx, obj, tmpl, opponents[1:], chosen)
	}
	return (MayChoice{Question: "Kharasha Foothills — create a token copy of that creature, tapped and attacking " + p.Name + "?",
		YesLabel: "Create copy", NoLabel: "Skip opponent",
		Then:    kharashaAnswer,
		Carry:   joinCarry(chosen, opponents),
		Objects: []game.ObjectRef{obj},
	}).Apply(ctx)
}

// kharashaExileTokens is "exile those tokens at the beginning of the next
// end step" as a keyed tail. An on-disk key.
var kharashaExileTokens = CardTail("kharasha-foothills/exile-tokens-at-next-end-step", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return (ScheduleDelayedTrigger{At: game.StepEnd, Controller: ctx.Item.Controller,
		Label: "Kharasha Foothills — exile the token copies", Cards: o.IDs, Body: exileListedCardsBody}).Apply(ctx)
})

// kharashaAnswer is one opponent's offer of a token copy: a yes adds them
// to the opponents chosen, either way the next is asked. Registered in
// init because it asks the next opponent through kharashaChooseOpponents,
// which names it. Keyed (#4291).
var kharashaAnswer game.ConfirmThen

func init() {
	kharashaAnswer = MayAnswer("confirm/kharasha-foothills-copy", func(ctx *Context, r game.ConfirmAnswer) error {
		if len(r.Objects) == 0 {
			return nil
		}
		obj := r.Objects[0]
		tmpl, ok := ctx.Game.TokenCopyTemplateForObjectForEffect(obj)
		if !ok {
			return nil
		}
		chosen, queue := splitCarry(r)
		if len(queue) == 0 {
			return nil
		}
		asked, rest := queue[0], queue[1:]
		if r.Accepted {
			chosen = append(chosen, asked)
		}
		return kharashaChooseOpponents(ctx, obj, tmpl, rest, chosen)
	})
}
