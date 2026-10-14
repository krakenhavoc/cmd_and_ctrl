package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elesh Norn // The Argent Etchings — a transforming Praetor whose back
// face is a Saga that makes Incubators (March of the Machine).
//
// Front face, Legendary Creature — Phyrexian Praetor {2}{W}{W}, 3/5:
//
//	"Vigilance
//	 Whenever a source an opponent controls deals damage to you or a
//	 permanent you control, that source's controller loses 2 life unless
//	 they pay {1}.
//	 {2}{W}, Sacrifice three other creatures: Exile Elesh Norn, then
//	 return it to the battlefield transformed under its owner's control.
//	 Activate only as a sorcery."
//
// Back face, Enchantment — Saga:
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I — Incubate 2 five times, then transform all Incubator tokens you
//	     control.
//	 II — Creatures you control get +1/+1 and gain double strike until
//	      end of turn.
//	 III — Destroy all other permanents except for artifacts, lands, and
//	       Phyrexians. Exile this Saga, then return it to the battlefield
//	       (front face up)."
//
// The flip is ExileAndReturnTransformed (ADR 0079's second verb, as Jin-
// Gitaxias): two zone changes and a new object, so the Saga enters with
// its first lore counter and chapter I. The back face registers under
// "<oracle_id>#1", and a Saga back face is nothing new (The Great
// Synthesis).
//
// The damage ability is one trigger per damage event, the source's
// controller read as it fires (b17DamageSourceController: a spell on the
// stack is its caster's). It watches damage to its controller or to any
// permanent that controller controls, and only a source another player
// controls. That player is asked to pay {1}; declining, or being unable,
// costs them 2 life.
//
// Chapter I is five separate incubates (Incubate.Times) and then the
// transform of every Incubator token its controller has, the ones from
// earlier turns included, front face up (IsIncubator). Chapter II is a
// locked-in set (CR 611.2c): a creature that enters later this turn gets
// nothing. Chapter III destroys every permanent that is not an artifact,
// a land or a Phyrexian, the Saga itself excepted ("all OTHER
// permanents"), through the destroy pipeline, so indestructible
// permanents survive; a transformed Incubator is an artifact and a
// Phyrexian both. The flicker follows the destruction as its
// continuation, since a commander caught in it stops to answer a
// replacement's prompt, and brings the card back front face up as a new
// object under its owner's control.
//
// No simplification.
func init() {
	const damageLabel = "Elesh Norn — that source's controller loses 2 life unless they pay {1}"
	Register(Spec{
		OracleID:        eleshNornArgentEtchingsOracleID,
		Name:            "Elesh Norn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Amount <= 0 || !eleshNornDamagedYourSide(ev, source.Controller, g) {
					return false
				}
				dealer := b17DamageSourceController(ev, g)
				return dealer != uuid.Nil && dealer != source.Controller
			},
			Key: damageLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, damageLabel)
				item.Params.Player = b17DamageSourceController(ev, g)
				return item
			},
			Effect: eleshNornPayOrLose,
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}{W}, Sacrifice three other creatures: Exile Elesh Norn, then return it to the battlefield " +
				"transformed under its owner's control. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{W}"), SacrificeAnotherN(3, "three other creatures", Creature())),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})

	Register(Spec{
		OracleID:     eleshNornArgentEtchingsOracleID + "#1",
		Name:         "The Argent Etchings",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel("The Argent Etchings", 1, "incubate 2 five times, then transform all Incubator tokens you control"),
				func(g *game.Game, item *game.StackItem) error {
					return Incubate{N: 2, Times: 5, Then: TransformAllIncubators}.Apply(NewContext(g, item))
				}),
			ChapterTrigger(2, SagaChapterLabel("The Argent Etchings", 2, "creatures you control get +1/+1 and gain double strike until end of turn"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					yours := And(Creature(), YouControl())
					if err := (BoostUntilEOT{
						Match:     yours,
						Power:     1,
						Toughness: 1,
						Label:     "The Argent Etchings — +1/+1",
					}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{
						Match:    yours,
						Keywords: []string{"double strike"},
						Label:    "The Argent Etchings — double strike",
					}.Apply(ctx)
				}),
			TriggerWithPurpose(ChapterTrigger(3, SagaChapterLabel("The Argent Etchings", 3, "destroy all other permanents except artifacts, lands, and Phyrexians, then exile this Saga and return it front face up"),
				argentEtchingsChapterThree), game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}}),
		},
	})
}

// eleshNornArgentEtchingsOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const eleshNornArgentEtchingsOracleID = "650ed75c-20b0-45b8-a4d4-813e0369aaf1"

// eleshNornDamagedYourSide is "deals damage to you or a permanent you
// control": the event's target is the controller, or a card on the
// battlefield that controller controls.
func eleshNornDamagedYourSide(ev game.Event, controller uuid.UUID, g *game.Game) bool {
	if ev.Target == controller {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && c.Controller == controller
}

// eleshNornPayOrLose is the damage ability's resolution: the source's
// controller (read when it fired, on Params.Player) pays {1} or loses
// 2 life. A player who has left the game is asked nothing.
func eleshNornPayOrLose(g *game.Game, item *game.StackItem) error {
	victim := item.Params.Player
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	source := item.SourceCardID
	return PayUnless{
		Chooser:  victim,
		Cost:     "{1}",
		Question: "Elesh Norn — pay {1}, or lose 2 life?",
		OnDecline: func(ctx *Context) error {
			return ctx.Game.ChangePlayerLifeForEffect(source, victim, -2)
		},
	}.Apply(NewContext(g, item))
}

// argentEtchingsChapterThree destroys everything that is neither an
// artifact, a land, a Phyrexian nor the Saga itself, then flickers the
// Saga. The flicker is the sweep's continuation, so a commander's prompt
// holds it back instead of racing it.
func argentEtchingsChapterThree(g *game.Game, item *game.StackItem) error {
	self := item.SourceCardID
	return DestroyAllMatching{
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.InstanceID != self && !c.IsArtifact() && !c.IsLand() && !c.HasSubtype("Phyrexian")
		},
		Then: func(ctx *Context, _ []game.Card, _ int) error {
			return Flicker{Target: self}.Apply(ctx)
		},
	}.Apply(NewContext(g, item))
}
