package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kytheon, Hero of Akros // Gideon, Battle-Forged (#2567) — a
// transforming legendary creature whose back face is a planeswalker.
//
// Front face, Legendary Creature — Human Soldier {W}, 2/1:
//
//	"At end of combat, if Kytheon and at least two other creatures
//	 attacked this combat, exile Kytheon, then return him to the
//	 battlefield transformed under his owner's control.
//	 {2}{W}: Kytheon gains indestructible until end of turn."
//
// Back face, Legendary Planeswalker — Gideon, loyalty 3:
//
//	"+2: Up to one target creature an opponent controls attacks Gideon
//	     during its controller's next turn if able.
//	 +1: Until your next turn, target creature gains indestructible.
//	     Untap that creature.
//	 0: Until end of turn, Gideon becomes a 4/4 Human Soldier creature
//	    with indestructible that's still a planeswalker. Prevent all
//	    damage that would be dealt to him this turn."
//
// The back face registers under "<oracle_id>#1" (game.CatalogKey).
//
// # The flip
//
// "At end of combat" triggers as every end of combat step begins
// (CR 511.2). "If Kytheon and at least two other creatures attacked this
// combat" is an intervening if (CR 603.4), checked then and again as the
// trigger resolves, against this combat's declared attackers
// (game.AttackersThisCombatForEffect): a creature that attacked and has
// since died still counts, one put onto the battlefield attacking never
// does (the rulings), and an additional combat phase is a combat of its
// own. Kytheon is the object the trigger came from, so a Kytheon that
// left and came back did not attack (CR 400.7). The flip is ADR 0079's
// ExileAndReturnTransformed, as Sorin of House Markov's: a new object
// that enters with its printed loyalty (CR 306.5b), under its owner's
// control.
//
// # The +2
//
// CreatureAttacksThisDuringItsControllersNextTurn
// (attack_this_permanent.go): a requirement pinned to the creature,
// obeyed only by attacking this Gideon as the object he was when the
// ability was activated, binding in a turn after this one and ending at
// the end of its controller's next turn, with "its controller" read
// live — a creature that changes controllers first is bound during its
// new controller's next turn (the ruling). With no target chosen it
// does nothing.
//
// THE +1 is indestructible until your next turn, then the untap. THE 0
// is animateGideon (gideon_animate.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     kytheonOracleID,
		Name:         "Kytheon, Hero of Akros",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "Kytheon and at least two other creatures attacked this combat",
				Test: func(ctx *effects.Context) bool {
					ref, ok := ctx.SourceRef()
					return ok && kytheonAndTwoOthersAttacked(ctx.Game, ref)
				},
			}, effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepEndCombat, false), kytheonAttackedWithTwoOthers),
				"Kytheon, Hero of Akros — exile him, then return him transformed",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Kytheon gains indestructible until end of turn.",
			Cost:    effects.ManaCost("{2}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Kytheon, Hero of Akros — indestructible until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})

	effects.Register(effects.Spec{
		OracleID:     kytheonOracleID + "#1",
		Name:         "Gideon, Battle-Forged",
		Completeness: effects.CompletenessFull,
		// Printed loyalty reaches a card through deck import (ADR 0032
		// §1); this is the fallback for fixtures and the dev spawner.
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Up to one target creature an opponent controls attacks Gideon during its controller's next turn if able.",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.TargetCreature("up to one target creature an opponent controls", effects.OpponentControls()).WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.CreatureAttacksThisDuringItsControllersNextTurn{
						Creature: id,
						Label:    "Gideon, Battle-Forged — that creature attacks Gideon during its controller's next turn if able",
					}.Apply(ctx)
				},
			},
			{
				Label:   "+1: Until your next turn, target creature gains indestructible. Untap that creature.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.ScopedEffectFor{
						Target:   id,
						Mods:     []game.Mod{game.AddKeywordsMod("indestructible")},
						Duration: effects.DurationUntilYourNextTurn(ctx, item.Controller),
						Label:    "Gideon, Battle-Forged — indestructible until your next turn",
					}).Apply(ctx); err != nil {
						return err
					}
					return effects.UntapTarget{Target: id}.Apply(ctx)
				},
			},
			{
				Label: "0: Until end of turn, Gideon becomes a 4/4 Human Soldier creature with indestructible that's still a planeswalker. Prevent all damage that would be dealt to him this turn.",
				Cost:  effects.LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.AnimateGideon(g, item, effects.GideonAnimation{
						Label:          "Gideon, Battle-Forged — a 4/4 Human Soldier creature with indestructible until end of turn",
						Subtypes:       []string{"Human", "Soldier"},
						Power:          4,
						Toughness:      4,
						Indestructible: true,
					})
				},
			},
		},
	})
}

// kytheonOracleID is shared by both faces.
const kytheonOracleID = "e49d745c-5130-4b54-bcbb-643b08528456"

// kytheonAndTwoOthersAttacked is the flip's condition: the object `self`
// names, and at least two other objects, were declared as attackers in
// the combat phase in progress.
func kytheonAndTwoOthersAttacked(g *game.Game, self game.ObjectRef) bool {
	him, others := false, 0
	for _, a := range g.AttackersThisCombatForEffect() {
		if a == self {
			him = true
			continue
		}
		others++
	}
	return him && others >= 2
}

// kytheonAttackedWithTwoOthers is the same condition as the trigger
// predicate, read as the end of combat step begins (CR 603.4).
func kytheonAttackedWithTwoOthers(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return kytheonAndTwoOthersAttacked(g, game.ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch})
}
