package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venom, Deadly Devourer — Legendary Creature — Symbiote Villain
// {2}{B}{G}, 4/4 (Marvel's Spider-Man, #4460):
//
//	"Vigilance, menace
//	 {3}: Exile target creature card from a graveyard. When you do, put X
//	 +1/+1 counters on target Symbiote, where X is the exiled card's
//	 toughness."
//
// The exile is the activated ability's effect; "when you do" is a CR
// 603.12 reflexive trigger queued by the exile's keyed tail, so it only
// happens if the card actually reached exile (a replacement that sends it
// elsewhere, or a target that left, makes no trigger). The Symbiote is
// chosen as the trigger goes on the stack (CR 603.3d) and X is the exiled
// card's toughness as it sits in exile, carried as Params.Amount.
func init() {
	Register(Spec{
		OracleID:        "f064fc04-28fb-423f-a4f3-bb567c71168f",
		Name:            "Venom, Deadly Devourer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
		Activated: []ActivatedAbility{{
			Label:   "{3}: Exile target creature card from a graveyard. When you do, put X +1/+1 counters on target Symbiote, where X is the exiled card's toughness.",
			Cost:    ManaCost("{3}"),
			Targets: TargetCardInGraveyard("target creature card from a graveyard", Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(venomDeadlyDevourerExiled)}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

var venomDeadlyDevourerExiled = CardTail("exile-then/venom-deadly-devourer-counters", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	exiled, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	return ReflexiveTrigger{
		Label:  "Venom, Deadly Devourer — put X +1/+1 counters on target Symbiote",
		Body:   venomDeadlyDevourerBody,
		Params: game.EffectParams{Amount: exiled.CurrentToughness()},
	}.Apply(ctx)
})

var venomDeadlyDevourerBody = game.ReflexiveBody("venom-deadly-devourer/counters", venomDeadlyDevourerCounters,
	constTargets(func() *game.TargetSpec {
		return TargetPermanent("target Symbiote", Subtype("Symbiote"))
	}))

func venomDeadlyDevourerCounters(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: p.Amount}.Apply(ctx)
		}
	}
	return nil
}
