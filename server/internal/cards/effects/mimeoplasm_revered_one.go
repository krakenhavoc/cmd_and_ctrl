package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mimeoplasm, Revered One — Legendary Creature — Ooze {X}{B}{G}{U}, 0/0
// (Aetherdrift, #4464):
//
//	"As Mimeoplasm enters, exile up to X creature cards from your
//	 graveyard. It enters with three +1/+1 counters on it for each creature
//	 card exiled this way.
//	 {2}: Mimeoplasm becomes a copy of target creature card exiled with it,
//	 except it's 0/0 and has this ability."
//
// X is the value announced for the spell (CR 107.3m, Card.CastX); an
// Ooze that wasn't cast has X = 0 and exiles nothing. The "as it enters"
// choice is Phyrexian Processor's entry slot (Spec.AsEnters, CR 614.12):
// a choose-cards prompt over the creature cards in your graveyard, none
// to X of them. Each card named is exiled LINKED to this object
// (ExileLinkedForEffect, CR 607.2a), so it is "exiled with it" for the
// ability, a Mimeoplasm that left and came back being a new object with no
// claim on them (CR 400.7); and three +1/+1 counters go on for each card
// really exiled. Naming none enters with no counters, a 0/0 that dies.
//
// The ability is Lazav, the Multifarious's copy over a card (a copy with no
// stated duration, CR 611.2a), with the exception "it's 0/0"
// (PrintedValues.SetPowerToughness) so the counters, not the copied card,
// size it, and "has this ability" as a granted bundle, so a Mimeoplasm that
// has become another creature can copy again. The target is a creature card
// exiled with this object, judged as the ability is put on the stack and as
// it resolves (CR 608.2b).
//
// No simplification.
const mimeoplasmGrant = "mimeoplasm-revered-one/this-ability"

var mimeoplasmCopyAbility = ActivatedAbility{
	Label:   "{2}: Mimeoplasm becomes a copy of target creature card exiled with it, except it's 0/0 and has this ability.",
	Cost:    ManaCost("{2}"),
	Targets: ExiledWithThis(TargetCardInExile("target creature card exiled with this", Creature())),
	Effect: func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		targets := ctx.LegalTargets()
		if len(targets) == 0 {
			return nil
		}
		return BecomeCopy{
			Targets:  []uuid.UUID{ctx.Source()},
			Of:       targets[0].ID,
			Duration: CopyIndefinite,
			Except: func(v *game.PrintedValues) {
				v.SetPowerToughness(0, 0)
				v.GrantAbility(mimeoplasmGrant)
			},
			Label: "Mimeoplasm, Revered One — becomes a copy",
		}.Apply(ctx)
	},
}

func init() {
	Register(Spec{
		OracleID:     "2b8c2882-dbca-466a-a0af-dd1d65edfe7c",
		Name:         "Mimeoplasm, Revered One",
		Completeness: CompletenessFull,
		AsEnters:     mimeoplasmExileAsEnters,
		Grants: []AbilityGrant{{
			Key:       mimeoplasmGrant,
			Text:      "{2}: Mimeoplasm becomes a copy of target creature card exiled with it, except it's 0/0 and has this ability.",
			Activated: []ActivatedAbility{mimeoplasmCopyAbility},
		}},
		Activated: []ActivatedAbility{mimeoplasmCopyAbility},
	})
}

// mimeoplasmExileAsEnters queues "exile up to X creature cards from your
// graveyard" for the controller, with nothing to ask when X is 0 or the
// graveyard holds no creature card.
func mimeoplasmExileAsEnters(card *game.Card, ctx *Context) error {
	x := card.CastX()
	if x <= 0 {
		return nil
	}
	p := ctx.Game.PlayerByIDForEffect(card.Controller)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  card.Controller,
		Source:   card.InstanceID,
		Question: "Mimeoplasm, Revered One — exile up to X creature cards from your graveyard",
		Cards:    candidates,
		Min:      0,
		Max:      min(x, len(candidates)),
		ThenKey:  mimeoplasmExiled,
		LiveZone: game.ZoneGraveyard,
	})
	return nil
}

// mimeoplasmExiled is the keyed answer (#4253): each card named is exiled
// linked to the Ooze, and it gets three +1/+1 counters for each. An on-disk
// key: never renamed, never reused.
var mimeoplasmExiled = game.RegisterChooseCardsThen("choose-cards/mimeoplasm-exile-creature-cards", func(g *game.Game, r game.CardsChosen) error {
	ref, ok := g.PermanentRefForEffect(r.Source)
	if !ok || len(r.Picked) == 0 {
		return nil
	}
	exiled := 0
	for _, id := range r.Picked {
		if err := g.ExileLinkedForEffect(ref, id); err != nil {
			return err
		}
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			exiled++
		}
	}
	if exiled == 0 {
		return nil
	}
	return g.AddCounterForEffect(r.Source, game.CounterPlusOne, 3*exiled)
})
