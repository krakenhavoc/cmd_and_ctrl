package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cherished Hatchling — Creature — Dinosaur {1}{G}, 2/1 (Foundations
// Jumpstart):
//
//	"When this creature dies, you may cast Dinosaur spells this turn as
//	 though they had flash, and whenever you cast a Dinosaur spell this
//	 turn, it gains "When this creature enters, you may have it fight
//	 another target creature.""
//
// One dies trigger, two effects. The flash is Yeva's cast-timing grant
// narrowed to Dinosaur spells for the turn (GrantCastTiming). The second
// half is a repeating delayed trigger (CR 603.7b: "whenever … this turn"),
// controlled by the Hatchling's controller, that fires on every Dinosaur
// spell they cast until cleanup. As it resolves, the spell gains the bundle
// below, which Rivaz of the Claw's grant shape hands to the permanent the
// spell becomes (CR 400.7a); a spell countered or gone before the trigger
// resolves gains nothing.
//
// The granted enters trigger is the creature's own: optional ("you may have
// it fight"), targeting another creature (never itself), and the fight is
// skipped when the target is gone or either creature has left (CR 701.14).
//
// No simplification.
const cherishedHatchlingFightGrant = "cherished-hatchling/enters-fight"

// cherishedHatchlingYouCastADinosaur is "whenever you cast a Dinosaur
// spell": an EventCast by the delayed trigger's controller whose spell has
// the Dinosaur subtype.
var cherishedHatchlingCastCondition = game.DelayedCondition("s82-bns-b27/you-cast-a-dinosaur", func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
	if ev.Kind != game.EventCast || ev.Actor != dt.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	// "Dinosaur" is a word of the Hatchling's text (CR 612.2).
	return ok && c.HasSubtype(g.TextReaderOfForEffect(dt.SourceCardID).TextSubtype("Dinosaur"))
})

// cherishedHatchlingGainsFight is the delayed trigger's body: the spell
// the event named gains the fight bundle for the permanent it becomes.
var cherishedHatchlingGainsFightBody = game.SimpleDelayedBody("s82-bns-b27/dinosaur-spell-gains-fight", func(g *game.Game, item *game.StackItem) error {
	cast := NewContext(g, item).PayloadCards()
	if len(cast) == 0 {
		return nil
	}
	spell := cast[0]
	if spell == uuid.Nil || g.StackItemForEffect(spell) == nil {
		return nil
	}
	g.RegisterScopedEffectForEffect(item.SourceCardID,
		[]game.AffectedObject{{ID: spell}},
		[]game.Mod{game.GrantAbilitiesMod(cherishedHatchlingFightGrant)},
		game.ForThePermanentThisSpellBecomesDuration(spell),
		"Cherished Hatchling — it gains \"When this creature enters, you may have it fight another target creature.\"")
	return nil
})

func init() {
	Register(Spec{
		OracleID:     "bd7111bc-9b8f-4c4b-b61c-7841a857ce6b",
		Name:         "Cherished Hatchling",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: cherishedHatchlingFightGrant,
			Triggered: []game.TriggeredAbility{
				func() game.TriggeredAbility {
					t := Targeting(
						WhenThisEnters("Cherished Hatchling — you may have it fight another target creature", itFightsTheFirstTarget),
						Another(TargetCreature("another target creature")))
					t.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Have it fight another target creature?"}
					return t
				}(),
			},
			Text: "When this creature enters, you may have it fight another target creature.",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Cherished Hatchling — you may cast Dinosaur spells this turn as though they had flash, and whenever you cast a Dinosaur spell this turn, it gains \"When this creature enters, you may have it fight another target creature.\"",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (GrantCastTiming{
						Timing: game.TimingFlash,
						Filter: game.PermissionFilter{CreatureType: ctx.TextSubtype("Dinosaur")},
						Label:  "You may cast Dinosaur spells this turn as though they had flash.",
					}).Apply(ctx); err != nil {
						return err
					}
					return DelayedOnEvent{
						Label:     "Cherished Hatchling — that Dinosaur spell gains \"When this creature enters, you may have it fight another target creature.\"",
						On:        []game.EventKind{game.EventCast},
						Condition: cherishedHatchlingCastCondition,
						Body:      cherishedHatchlingGainsFightBody,
						Repeats:   true,
					}.Apply(ctx)
				}),
		},
	})
}
