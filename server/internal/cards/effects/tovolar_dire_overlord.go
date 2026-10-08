package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tovolar, Dire Overlord // Tovolar, the Midnight Scourge — {1}{R}{G}
// Legendary Creature — Human Werewolf 3/3 // Legendary Creature —
// Werewolf 4/4 (#2586, ADR 0132):
//
//	Front: "Whenever a Wolf or Werewolf you control deals combat damage
//	        to a player, draw a card.
//	        At the beginning of your upkeep, if you control three or more
//	        Wolves and/or Werewolves, it becomes night. Then transform
//	        any number of Human Werewolves you control.
//	        Daybound"
//	Back:  "Whenever a Wolf or Werewolf you control deals combat damage
//	        to a player, draw a card.
//	        {X}{R}{G}: Target Wolf or Werewolf you control gets +X/+0 and
//	        gains trample until end of turn.
//	        Nightbound"
//
// The upkeep trigger is an intervening if (CR 603.4): checked as the
// upkeep begins and again on resolution. "It becomes night" is the
// designation, which already turns every daybound permanent over (CR
// 702.145); the "then transform" half is therefore for Human Werewolves
// that are NOT daybound, and it is a real choice of any number, asked as
// a Scapeshift-style pick (ChoosePermanents) with a floor of zero.
// A Human Werewolf with daybound is not offered: the day/night rules have
// just turned it over and no effect may (CR 702.145b / 702.145e).
//
// The draw is one trigger per creature dealing damage, as printed ("a
// Wolf or Werewolf ... deals").
//
// No simplification.
func init() {
	const oracle = "45d49831-548a-4a0e-9a18-9f7397913895"
	draw := func(name string) game.TriggeredAbility {
		return On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if !combatDamageToPlayerBy(ev, source.Controller, g) {
				return false
			}
			dealer, ok := g.LookupCardForEffect(ev.Source)
			return ok && (dealer.HasSubtype("Wolf") || dealer.HasSubtype("Werewolf"))
		}, name+" — draw a card", Do(DrawCards{N: 1}))
	}
	Register(Spec{
		OracleID:        oracle,
		Name:            "Tovolar, Dire Overlord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Triggered: []game.TriggeredAbility{
			draw("Tovolar, Dire Overlord"),
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && tovolarControlsThreeWolves(g, source.Controller)
			}, "Tovolar, Dire Overlord — it becomes night, then transform any number of Human Werewolves", tovolarUpkeep),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Tovolar, the Midnight Scourge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		XMatters:        true,
		Triggered:       []game.TriggeredAbility{draw("Tovolar, the Midnight Scourge")},
		Activated: []ActivatedAbility{{
			Label:   "{X}{R}{G}: Target Wolf or Werewolf you control gets +X/+0 and gains trample until end of turn.",
			Cost:    ManaCost("{X}{R}{G}"),
			Targets: TargetCreature("target Wolf or Werewolf you control", YouControl(), Or(OfCreatureType("Wolf"), OfCreatureType("Werewolf"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (BoostUntilEOT{Target: target, Power: ctx.X(), Label: "Tovolar, the Midnight Scourge — +X/+0"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: target, Keywords: []string{"trample"}, Label: "Tovolar, the Midnight Scourge — trample"}.Apply(ctx)
			},
		}},
	})
}

// tovolarControlsThreeWolves is the intervening if: three or more Wolves
// and/or Werewolves under the controller's control.
func tovolarControlsThreeWolves(g *game.Game, controller uuid.UUID) bool {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsCreature() && (c.HasSubtype("Wolf") || c.HasSubtype("Werewolf")) {
			n++
		}
	}
	return n >= 3
}

// tovolarUpkeep re-checks the condition (CR 603.4), makes it night, then
// offers the Human Werewolves you control for the transform.
func tovolarUpkeep(g *game.Game, item *game.StackItem) error {
	if !tovolarControlsThreeWolves(g, item.Controller) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (BecomeNight{}).Apply(ctx); err != nil {
		return err
	}
	return ChoosePermanents{
		Question: "Tovolar, Dire Overlord — transform any number of Human Werewolves you control",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.Battlefield.Cards {
				if c.Controller == of && c.IsCreature() && c.HasSubtype("Human") && c.HasSubtype("Werewolf") &&
					!game.HasKeyword(&c, "daybound") && !game.HasKeyword(&c, "nightbound") {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, len(out)
		},
		Then: func(ctx *Context, picked game.PromptedPicks) error {
			for _, id := range picked.Cards() {
				if err := (Transform{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}.Apply(ctx)
}
