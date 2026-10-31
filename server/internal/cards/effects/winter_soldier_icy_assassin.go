package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winter Soldier, Icy Assassin — Legendary Creature — Human Assassin
// Villain {W}{B}, 2/2 (Marvel Super Heroes, #4455):
//
//	"Vigilance, menace
//	 Winter Soldier gets +2/+0 for each Equipment attached to him.
//	 {3}{W}{B}: Return this card from your graveyard to the battlefield
//	 with a finality counter on him. Then you may attach an Equipment you
//	 control to him. (If a creature with a finality counter on it would
//	 die, exile it instead.)"
//
// The pump is a layer 7c static that counts the Equipment attached to
// him as each recompute runs. The graveyard ability puts him onto the
// battlefield with the counter as an entry counter (CR 614.1c), through
// the keyed entry door, so the optional attach is a tail that runs once
// he has landed: a choose_cards prompt over the Equipment the controller
// controls with a floor of zero. The finality replacement itself is the
// engine's (CR 122.1h, game/finality_counters.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6f4c94b6-f829-4646-a574-083997b78dfb",
		Name:            "Winter Soldier, Icy Assassin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
		Static: []game.StaticAbility{{
			Label:     "Winter Soldier gets +2/+0 for each Equipment attached to him.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				if g == nil || target == nil {
					return
				}
				for _, a := range g.BattlefieldCardsForEffect() {
					if a.HasSubtype("Equipment") && a.IsAttachedTo(target.InstanceID) {
						c.Power += 2
					}
				}
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}{B}: Return this card from your graveyard to the battlefield with a finality counter on him. Then you may attach an Equipment you control to him.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{3}{W}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  mshB11WinterSoldierReturn,
		}},
	})
}

func mshB11WinterSoldierReturn(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	ctx := NewContext(g, item)
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(id, item.Controller, false,
		map[string]int{game.CounterFinality: 1}, ctx.ItemTail(mshB11WinterSoldierAttach))
}

// mshB11WinterSoldierAttach is "then you may attach an Equipment you
// control to him", keyed: an optional pick among the Equipment, the
// returned creature carried to the answer.
var mshB11WinterSoldierAttach = EnteredTail("winter-soldier-icy-assassin/attach", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || !onBattlefield(g, entered) {
		return nil
	}
	var equipment []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.HasSubtype("Equipment") {
			equipment = append(equipment, c.InstanceID)
		}
	}
	if len(equipment) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Winter Soldier, Icy Assassin — you may attach an Equipment you control to him",
		Cards:    equipment,
		Min:      0,
		Max:      1,
		ThenKey:  mshB11WinterSoldierAttachPick,
		LiveZone: game.ZoneBattlefield,
		Carry:    []uuid.UUID{entered},
	})
	return nil
})

// mshB11WinterSoldierAttachPick attaches the chosen Equipment to the
// carried Winter Soldier, if he is still on the battlefield.
var mshB11WinterSoldierAttachPick = game.RegisterChooseCardsThen("choose-cards/winter-soldier-icy-assassin-attach", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 || !onBattlefield(g, r.Carry[0]) {
		return nil
	}
	return g.AttachForEffect(r.Picked[0], game.TargetRef{Kind: game.TargetCard, ID: r.Carry[0]})
})
