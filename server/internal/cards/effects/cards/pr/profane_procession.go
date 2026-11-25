package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Profane Procession // Tomb of the Dusk Rose (Rivals of Ixalan, S82 OLD batch
// 218, #5531).
//
// Front face, Legendary Enchantment {1}{W}{B}:
//
//	"{3}{W}{B}: Exile target creature. Then if there are three or more cards
//	 exiled with Profane Procession, transform it."
//
// Back face, Legendary Land:
//
//	"(Transforms from Profane Procession.)
//	 {T}: Add one mana of any color.
//	 {2}{W}{B}, {T}: Put a creature card exiled with this permanent onto the
//	 battlefield under your control."
//
// The exile is linked to this object (CR 607.2a, Card.ExiledWith), and the link
// is to the permanent, which is the same object after it transforms (CR 712.18),
// so the land finds the cards the enchantment exiled. The "then if" counts every
// card still exiled with it, in a registered tail, after the new card has landed
// in exile; three or more transforms it. The land's ability offers the creature
// cards exiled with it as it resolves and puts the picked one onto the
// battlefield under your control, whoever owns it; with none it does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     profaneProcessionOracleID,
		Name:         "Profane Procession",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}{B}: Exile target creature. Then if there are three or more cards exiled with Profane Procession, transform it.",
			Cost:    effects.ManaCost("{3}{W}{B}"),
			Targets: effects.TargetCreature("target creature"),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  profaneProcessionExile,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     profaneProcessionOracleID + "#1",
		Name:         "Tomb of the Dusk Rose",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{B}, {T}: Put a creature card exiled with this permanent onto the battlefield under your control.",
			Cost:    effects.Plus(effects.ManaCost("{2}{W}{B}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  tombOfTheDuskRoseOffer,
		}},
	})
}

const profaneProcessionOracleID = "a656ad7f-133f-4d93-919a-43bcf1f815f3"

func profaneProcessionExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.ExileTarget{Target: id, Linked: true, ThenTail: ctx.ItemTail(profaneProcessionThen)}.Apply(ctx)
}

// profaneProcessionThen is "Then if there are three or more cards exiled with
// Profane Procession, transform it". An on-disk key: never renamed, never reused.
var profaneProcessionThen = effects.CardTail("then/profane-procession-transform", func(ctx *effects.Context, _ game.TailOutcome) error {
	ref, ok := ctx.SourceRef()
	if !ok || len(ctx.Game.ExiledWithForEffect(game.PermissionCardRef(ref))) < 3 {
		return nil
	}
	return effects.TransformThis{}.Apply(ctx)
})

func tombOfTheDuskRoseOffer(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var cands []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsCreature() {
			cands = append(cands, c.InstanceID)
		}
	}
	if len(cands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Tomb of the Dusk Rose — put a creature card exiled with it onto the battlefield under your control",
		Cards:    cands,
		Min:      1,
		Max:      1,
		ThenKey:  tombOfTheDuskRosePicked,
		LiveZone: game.ZoneExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// tombOfTheDuskRosePicked puts the chosen card onto the battlefield under the
// controller's control. An on-disk key: never renamed, never reused.
var tombOfTheDuskRosePicked = effects.ItemChoice("choose-cards/tomb-of-the-dusk-rose-return", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ReturnFromExile{Target: r.Picked[0], Controller: ctx.Controller()}.Apply(ctx)
})
