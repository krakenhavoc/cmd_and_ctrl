package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani, Nacatl Pariah // Ajani, Nacatl Avenger — Modern Horizons 3 (#5948), a
// transforming legendary creature whose back face is a planeswalker.
//
// Front face, Legendary Creature — Cat Warrior {1}{W}, 1/2:
//
//	"When Ajani enters, create a 2/1 white Cat Warrior creature token.
//	 Whenever one or more other Cats you control die, you may exile Ajani,
//	 then return him to the battlefield transformed under his owner's
//	 control."
//
// Back face, Legendary Planeswalker — Ajani, loyalty 3:
//
//	"+2: Put a +1/+1 counter on each Cat you control.
//	 0: Create a 2/1 white Cat Warrior creature token. When you do, if you
//	    control a red permanent other than Ajani, he deals damage equal to
//	    the number of creatures you control to any target.
//	 −4: Each opponent chooses an artifact, a creature, an enchantment, and
//	     a planeswalker from among the nonland permanents they control, then
//	     sacrifices the rest."
//
// The back face registers under "<oracle_id>#1" (game.CatalogKey). The flip
// trigger is one per batch of deaths (OncePerBatch) reading each dead Cat's
// last-known subtypes and controller (CR 603.10a), with the "may" as the
// trigger's optional prompt. The 0 is a token creation whose tail creates the
// CR 603.12 reflexive trigger, only once the token has really been made; its
// "if" is checked as it is created and again as it resolves (CR 603.4), and
// the damage counts creatures as the trigger resolves. The −4 is Cataclysmic
// Gearhulk's choose-one-of-each-type, for the opponents only.
//
// No simplification.
func init() {
	flip := effects.OncePerBatch(effects.On(game.EventLTB, ajaniOtherCatYouControlDied,
		"Ajani, Nacatl Pariah — exile him, then return him transformed",
		func(g *game.Game, item *game.StackItem) error {
			return effects.ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
		}))
	flip.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Ajani, Nacatl Pariah — exile Ajani, then return him to the battlefield transformed?"}

	effects.Register(effects.Spec{
		OracleID:     ajaniNacatlPariahOracle,
		Name:         "Ajani, Nacatl Pariah",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ajani, Nacatl Pariah — create a 2/1 white Cat Warrior creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/1 white Cat Warrior"), N: 1})),
			flip,
		},
	})

	effects.Register(effects.Spec{
		OracleID:        ajaniNacatlPariahOracle + "#1",
		Name:            "Ajani, Nacatl Avenger",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: Put a +1/+1 counter on each Cat you control.",
				Cost:  effects.LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					g.RecomputeLayersIfStaleLocked()
					var cats []uuid.UUID
					cat := ctx.TextSubtype("Cat")
					for _, c := range g.Battlefield.Cards {
						if c.Controller == ctx.Controller() && c.IsCreature() && c.HasSubtype(cat) {
							cats = append(cats, c.InstanceID)
						}
					}
					for _, id := range cats {
						if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:   "0: Create a 2/1 white Cat Warrior creature token. When you do, if you control a red permanent other than Ajani, he deals damage equal to the number of creatures you control to any target.",
				Cost:    effects.LoyaltyCost(0),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: item.Controller,
						Source:     item.SourceCardID,
						Groups: []game.TokenGroup{{
							Template: ctx.TextToken(effects.TokenCard("2/1 white Cat Warrior")),
							Count:    1,
						}},
					}, ctx.ItemTail(ajaniAvengerTokenMade))
				},
			},
			{
				Label: "−4: Each opponent chooses an artifact, a creature, an enchantment, and a planeswalker from among the nonland permanents they control, then sacrifices the rest.",
				Cost:  effects.LoyaltyCost(-4),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.GearhulkKeepOneOfEach(g, item, true)
				},
			},
		},
	})
}

// ajaniNacatlPariahOracle is shared by both faces.
const ajaniNacatlPariahOracle = "2588f348-d7a3-46c8-9ace-dca53ed5ef99"

// ajaniOtherCatYouControlDied is "one or more other Cats you control die": a
// creature other than the source, a Cat as it last existed, under your control.
func ajaniOtherCatYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := effects.DiedCreature(ev, g)
	return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
		effects.LeftAsSubtype(ev, dead, source.TextSubtype("Cat"))
}

// ajaniControlsRedOther is "you control a red permanent other than Ajani".
func ajaniControlsRedOther(g *game.Game, player, ajani uuid.UUID) bool {
	g.RecomputeLayersIfStaleLocked()
	for _, c := range g.Battlefield.Cards {
		if c.Controller != player || c.InstanceID == ajani {
			continue
		}
		for _, col := range c.EffectiveColors() {
			if col == "R" {
				return true
			}
		}
	}
	return false
}

// ajaniAvengerTokenMade is "When you do": once the token has been created, a
// reflexive trigger with its own "any target", only if a red permanent other
// than Ajani is controlled. An on-disk key.
var ajaniAvengerTokenMade = effects.CardTail("ajani-nacatl-avenger/token-made", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || !ajaniControlsRedOther(ctx.Game, ctx.Controller(), ctx.Source()) {
		return nil
	}
	return effects.WhenYouDo("Ajani, Nacatl Avenger — he deals damage equal to the number of creatures you control to any target",
		ajaniAvengerDamageBody).Apply(ctx)
})

// ajaniAvengerDamageBody is the damage: as the trigger resolves, the number of
// creatures you control, if you still control a red permanent other than Ajani.
var ajaniAvengerDamageBody = game.ReflexiveBody("ajani-nacatl-avenger/damage", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !ajaniControlsRedOther(g, ctx.Controller(), ctx.Source()) {
		return nil
	}
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok {
		return nil
	}
	n := 0
	g.RecomputeLayersIfStaleLocked()
	for _, c := range g.Battlefield.Cards {
		if c.Controller == ctx.Controller() && c.IsCreature() {
			n++
		}
	}
	return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
}), effects.ConstTargets(effects.TargetAny))
