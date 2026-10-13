package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jin-Gitaxias // The Great Synthesis — a transforming Praetor whose
// back face is a Saga.
//
// Front face, Legendary Creature — Phyrexian Praetor {3}{U}{U}, 5/5:
//
//	"Ward {2}
//	 Whenever you cast a noncreature spell with mana value 3 or greater,
//	 draw a card.
//	 {3}{U}: Exile Jin-Gitaxias, then return it to the battlefield
//	 transformed under its owner's control. Activate only as a sorcery
//	 and only if you have seven or more cards in hand."
//
// Back face, Enchantment — Saga:
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I — Draw cards equal to the number of cards in your hand. You have
//	     no maximum hand size for as long as you control this Saga.
//	 II — Return all non-Phyrexian creatures to their owners' hands.
//	 III — You may cast any number of spells from your hand without
//	     paying their mana costs. Exile this Saga, then return it to the
//	     battlefield (front face up)."
//
// The flip is ExileAndReturnTransformed (ADR 0079's second verb, as Crystal
// Fragments): two zone changes and a new object, so the Saga enters with its
// first lore counter and chapter I. The back face registers under
// "<oracle_id>#1". The cast trigger reads the spell's mana value with its
// announced X (CR 202.3e), through ManaValueGE.
//
// Chapter I draws as many cards as are in hand when it resolves. Chapter II
// bounces every creature that is not a Phyrexian, yours included, tokens
// ceasing to exist; "Phyrexian" is OfCreatureType, so a changeling is one.
// Chapter III grants each nonland card in hand a free cast ("cast", so a land
// is left out) for as long as the holder keeps priority, the engine's
// standing model for "cast it during the resolution" (ADR 0066), and then
// exiles the Saga and returns it to the battlefield front face up as a new
// object under its owner's control, which is Jin-Gitaxias with no counters.
//
// Chapter I's second sentence is a continuous effect the resolving ability
// CREATES, not a static of the Saga (ADR 0113's amendment of 2026-10-13,
// #3247): game.GrantHandSizeForEffect with the source-bound duration "for as
// long as you control this Saga" (CR 611.2b, DurationWhileYouControlSource).
// So a chapter I that is countered or stifled leaves nothing behind, a Saga
// that left before the chapter resolved starts no effect at all, and the
// effect ends the moment the Saga leaves or changes controller. Chapter III's
// flicker returns the Saga as a new object (CR 400.7), which also ends it.
func init() {
	Register(Spec{
		OracleID:     jinGitaxiasOracleID,
		Name:         "Jin-Gitaxias",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Jin-Gitaxias — ward {2}"),
			WheneverYouCast(And(Noncreature(), ManaValueGE(3)),
				"Jin-Gitaxias — draw a card",
				Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}{U}: Exile Jin-Gitaxias, then return it to the battlefield transformed under its owner's control. " +
				"Activate only as a sorcery and only if you have seven or more cards in hand.",
			Cost:         ManaCost("{3}{U}"),
			SorcerySpeed: true,
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				p := g.PlayerByIDForEffect(controller)
				return p != nil && p.Hand != nil && p.Hand.Size() >= 7
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})

	Register(Spec{
		OracleID:     jinGitaxiasOracleID + "#1",
		Name:         "The Great Synthesis",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel("The Great Synthesis", 1, "draw cards equal to the number of cards in your hand"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := 0
					if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
						n = p.Hand.Size()
					}
					if err := (DrawCards{Player: item.Controller, N: n}).Apply(ctx); err != nil {
						return err
					}
					// "for as long as you control this Saga": with no
					// duration (the Saga gone, or a new object) the effect
					// never begins (CR 611.2b).
					d, ok := DurationWhileYouControlSource(ctx, item.SourceCardID, item.Controller)
					if !ok {
						return nil
					}
					return g.GrantHandSizeForEffect(item.Controller, game.HandSizeNoMaximum, 0,
						"The Great Synthesis — no maximum hand size", item.SourceCardID, d)
				}),
			TriggerWithPurpose(ChapterTrigger(2, SagaChapterLabel("The Great Synthesis", 2, "return all non-Phyrexian creatures to their owners' hands"),
				func(g *game.Game, item *game.StackItem) error {
					return BounceAllMatching{Match: And(Creature(), Not(OfCreatureType("Phyrexian")))}.Apply(NewContext(g, item))
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}}),
			ChapterTrigger(3, SagaChapterLabel("The Great Synthesis", 3, "cast any number of spells from your hand free, then exile this Saga and return it front face up"),
				greatSynthesisChapterThree),
		},
	})
}

// greatSynthesisKey is the alternative-cost key the chapter III grant
// synthesises (CastPermission.AlternativeCostFor): an on-disk identity,
// never renamed.
const greatSynthesisKey = "great-synthesis-free"

const jinGitaxiasOracleID = "b70c3adb-64ba-49ea-bcf6-6406a58cb453"

// greatSynthesisChapterThree grants a free cast over every nonland card in
// the controller's hand, then flickers the Saga. The flicker returns a
// double-faced card front face up (CR 712.14a), so it needs no
// "transformed" flag.
func greatSynthesisChapterThree(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range allHandCardIDs(g, item.Controller) {
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player:      item.Controller,
			AltCostKey:  greatSynthesisKey,
			Cost:        "{0}",
			CastOnly:    true,
			Timing:      game.TimingFlash,
			LapseOnPass: game.LapseStaysInExile,
			Source:      item.SourceCardID,
			SourceName:  "The Great Synthesis",
			Label:       "The Great Synthesis — cast it without paying its mana cost",
		})
	}
	return Flicker{Target: item.SourceCardID, Controller: item.Controller}.Apply(ctx)
}
