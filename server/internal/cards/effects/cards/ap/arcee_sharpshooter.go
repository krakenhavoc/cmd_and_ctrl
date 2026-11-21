package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arcee, Sharpshooter // Arcee, Acrobatic Coupe (#3195, ADR 0152) — a
// More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {1}{R}{W}, 2/2:
//
//	"More Than Meets the Eye {R}{W}
//	 First strike
//	 {1}, Remove one or more +1/+1 counters from Arcee: It deals that
//	 much damage to target creature. Convert Arcee."
//
// Back face, Legendary Artifact — Vehicle, 2/2:
//
//	"Living metal
//	 Whenever you cast a spell that targets one or more creatures or
//	 Vehicles you control, put that many +1/+1 counters on Arcee.
//	 Convert Arcee."
//
// "One or more" counters is the variable counter-removal cost with a
// floor of one: the activator names how many as the cost is paid, and
// the damage is that number, read off the paid-cost record. The back
// face counts the DISTINCT creatures and Vehicles you control that the
// spell targets, as it is cast (a spell that targets one creature twice
// still targets one), and carries that number on the trigger, so a
// target that dies before the trigger resolves is still counted.
//
// No simplification.
// arceeVictimRecord files the creature the damage was dealt to.
const arceeVictimRecord = "arcee-sharpshooter/victim"

const ArceeOracleID = "9fba3097-d53f-4ea4-b5d2-8adddaa09524"

func init() {
	effects.Register(effects.Spec{
		OracleID:         ArceeOracleID,
		Name:             "Arcee, Sharpshooter",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"first strike"},
		AlternativeCosts: []game.AlternativeCost{effects.MoreThanMeetsTheEye("{R}{W}")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Remove one or more +1/+1 counters from Arcee: It deals that much damage to target creature. Convert Arcee.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.RemoveCountersXFromThis(game.CounterPlusOne, 1)),
			Targets: effects.TargetCreature("target creature"),
			// The conversion waits for a prompt the damage asks (a CR 616 order; CR 608.2c, #4155).
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					var victim uuid.UUID
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							victim = t.ID
							break
						}
					}
					if victim == uuid.Nil {
						return nil // the target is gone: the whole ability does nothing (CR 608.2b)
					}
					ctx.RecordStep(arceeVictimRecord, []uuid.UUID{victim})
					return effects.DealDamage{Source: ctx.Item.SourceCardID, Target: victim, Amount: ctx.CountersRemoved()}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					if len(ctx.StepRecord(arceeVictimRecord).IDs) == 0 {
						return nil
					}
					return effects.ConvertThis{}.Apply(ctx)
				}),
			),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     ArceeOracleID + "#1",
		Name:         "Arcee, Acrobatic Coupe",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.LivingMetal()},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && arceeTargetsYouControl(g, ev, source.Controller) > 0
			},
			Key: "Arcee, Acrobatic Coupe — put that many +1/+1 counters on Arcee, then convert",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Arcee, Acrobatic Coupe — put that many +1/+1 counters on Arcee, then convert")
				item.Params.Amount = arceeTargetsYouControl(g, ev, source.Controller)
				return item
			},
			// The conversion waits for a prompt the counters ask (a CR 616 order; CR 608.2c, #4155).
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					item := ctx.Item
					if effects.OnBattlefield(ctx.Game, item.SourceCardID) && !effects.SourceIsNewObject(ctx.Game, item) {
						return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: item.Params.Amount}.Apply(ctx)
					}
					return nil
				}),
				effects.ConvertThis{},
			),
		}},
	})
}

// arceeTargetsYouControl is how many distinct creatures and Vehicles
// `you` control that the spell this EventCast announced targets.
func arceeTargetsYouControl(g *game.Game, ev game.Event, you uuid.UUID) int {
	item := g.StackItemForEffect(ev.CardID)
	if item == nil || item.Kind != game.StackItemSpell {
		return 0
	}
	seen := map[uuid.UUID]bool{}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || seen[t.ID] {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || c.Controller != you || !effects.OnBattlefield(g, t.ID) {
			continue
		}
		if c.IsCreature() || c.HasSubtype("Vehicle") {
			seen[t.ID] = true
		}
	}
	return len(seen)
}
