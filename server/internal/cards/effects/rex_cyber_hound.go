package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rex, Cyber-Hound — Legendary Artifact Creature — Robot Dog {1}{W}{U},
// 2/2:
//
//	"Whenever Rex deals combat damage to a player, they mill two cards
//	 and you get {E}{E} (two energy counters).
//	 Pay {E}{E}: Choose target creature card in a graveyard. Exile it
//	 with a brain counter on it. Activate only as a sorcery.
//	 Rex has all activated abilities of all cards in exile with brain
//	 counters on them."
//
// The damaged player is the triggering event's target. The exile puts
// the brain counter on once the card has landed in exile; a card a
// replacement sent elsewhere gets none. The borrow is not an "exiled
// with" link: every card in exile with a brain counter lends, whoever put
// it there (ExileLenders). A counter placed or removed on an exiled card
// re-reads the layers.
//
// No simplification.
func init() {
	const clause = "Rex has all activated abilities of all cards in exile with brain counters on them."
	borrow := HasActivatedAbilitiesOf(game.AbilityBorrow{
		Label:   clause,
		Lenders: ExileLenders(LenderWithCounter(rexBrainCounter)),
	})
	borrow.DependsOnExile = true
	Register(Spec{
		OracleID:     "7ad52f9c-48d8-486f-90b7-73ce5cb3b9db",
		Name:         "Rex, Cyber-Hound",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{borrow},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Rex — they mill two cards and you get {E}{E}", rexMillAndEnergy),
		},
		Activated: []ActivatedAbility{{
			Label:        "Pay {E}{E}: Choose target creature card in a graveyard. Exile it with a brain counter on it. Activate only as a sorcery.",
			Cost:         PayEnergy(2),
			SorcerySpeed: true,
			Targets:      TargetCardInGraveyard("target creature card in a graveyard", Creature()),
			Effect:       exileTargetWithCounter(rexBrainCounter),
		}},
	})
}

const rexBrainCounter = "brain"

// rexMillAndEnergy is the combat-damage trigger: the damaged player mills
// two, then Rex's controller gets two energy.
func rexMillAndEnergy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if item.Trigger != nil {
		if err := (MillCards{Player: item.Trigger.Event.Target, N: 2}).Apply(ctx); err != nil {
			return err
		}
	}
	return GetEnergy{N: 2}.Apply(ctx)
}

// exileTargetWithCounter is "exile it with a <kind> counter on it" (Rex's
// brain counter, Mairsil's cage counter): the target card is exiled, and
// the counter goes on it once it has landed in exile.
func exileTargetWithCounter(kind string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return exileCardWithCounter(ctx, t.ID, kind)
		}
		return nil
	}
}

// exileCardWithCounter exiles one card and, once it is in exile, puts one
// `kind` counter on it.
func exileCardWithCounter(ctx *Context, id uuid.UUID, kind string) error {
	return ExileTarget{Target: id, ThenTail: withWord(withCarry(ctx.ItemTail(rexCyberHoundExiled), id), kind)}.Apply(ctx)
}

// rexCyberHoundExiled is the clause after the exile (#4253 slice 9).
var rexCyberHoundExiled = CardTail("rex-cyber-hound/after-exile", func(ctx *Context, o game.TailOutcome) error {
	id, kind := tailCarry0(o), o.Tail.Word
	exiled := len(o.IDs) == 1

	if !exiled {
		return nil
	}
	return AddCounter{Target: id, Kind: kind, N: 1}.Apply(ctx)
})
