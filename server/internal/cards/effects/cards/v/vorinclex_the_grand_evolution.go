package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vorinclex // The Grand Evolution — a transforming Praetor whose back face
// is a Saga (March of the Machine).
//
// Front face, Legendary Creature — Phyrexian Praetor {3}{G}{G}, 6/6:
//
//	"Reach, trample
//	 When Vorinclex enters, search your library for up to two Forest cards,
//	 reveal them, put them into your hand, then shuffle.
//	 {6}{G}{G}: Exile Vorinclex, then return it to the battlefield
//	 transformed under its owner's control. Activate only as a sorcery."
//
// Back face, Enchantment — Saga:
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I — Mill ten cards. Put up to two creature cards from among the milled
//	     cards onto the battlefield.
//	 II — Distribute seven +1/+1 counters among any number of target
//	      creatures you control.
//	 III — Until end of turn, creatures you control gain "{1}: This creature
//	       fights target creature you don't control." Exile this Saga, then
//	       return it to the battlefield (front face up)."
//
// The flip is ExileAndReturnTransformed (ADR 0079's second verb, as Jin-
// Gitaxias): two zone changes and a new object, so the Saga enters with its
// first lore counter and chapter I. The back face registers under
// "<oracle_id>#1".
//
// Chapter I mills ten, then asks for up to two of the creature cards that
// really reached the graveyard and puts them onto the battlefield as one
// simultaneous event, the milled-card list being MillToZone.ThenTail's
// (Smuggler's Surprise's shape). Chapter II divides seven counters among
// the targets it chose (Divide), any number of them. Chapter III gives every
// creature its controller has right now the fight ability until end of turn
// (a locked-in set, CR 611.2c): a creature that enters later this turn gets
// nothing. The ability is a bundle (vorinclexFightGrant), and each use
// makes the granted creature fight a creature its controller does not
// control. The Saga then flickers, returning front face up as a new object
// (CR 712.14a) after the grant is made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        VorinclexOracleID,
		Name:            "Vorinclex",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Vorinclex — search your library for up to two Forest cards, reveal them, put them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsLandWithSubtype("Forest"),
						Dest:      game.ZoneHand,
						Limit:     2,
						Reveal:    true,
						Shuffle:   true,
						Source:    item.SourceCardID,
						Reason:    "Vorinclex — search for up to two Forest cards",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{6}{G}{G}: Exile Vorinclex, then return it to the battlefield transformed under its owner's control. " +
				"Activate only as a sorcery.",
			Cost:         effects.ManaCost("{6}{G}{G}"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})

	chapterTwo := effects.ChapterTrigger(2, effects.SagaChapterLabel("The Grand Evolution", 2, "distribute seven +1/+1 counters among any number of target creatures you control"),
		func(g *game.Game, item *game.StackItem) error {
			return effects.PutDividedCounters(effects.NewContext(g, item), game.CounterPlusOne)
		})
	chapterTwo.Targets = effects.TargetCreature("any number of target creatures you control", effects.YouControl()).WithCount(0, 0).Dividing(effects.Divide(7))

	effects.Register(effects.Spec{
		OracleID:     VorinclexOracleID + "#1",
		Name:         "The Grand Evolution",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: vorinclexFightGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}: This creature fights target creature you don't control.",
				Cost:    effects.ManaCost("{1}"),
				Targets: effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
				Effect:  effects.B30SourceFightsFirstLegalTarget,
			}},
			Text: "{1}: This creature fights target creature you don't control.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel("The Grand Evolution", 1, "mill ten cards, then put up to two creature cards from among them onto the battlefield"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{Player: ctx.Controller(), N: 10, ThenTail: ctx.ItemTail(grandEvolutionMilled)}.Apply(ctx)
				}),
			chapterTwo,
			effects.ChapterTrigger(3, effects.SagaChapterLabel("The Grand Evolution", 3, "creatures you control gain the fight ability until end of turn, then exile this Saga and return it front face up"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GrantAbilitiesFor{
						Match: effects.And(effects.Creature(), effects.YouControl()),
						Keys:  []string{vorinclexFightGrant},
						Label: "The Grand Evolution — {1}: this creature fights target creature you don't control",
					}).Apply(ctx); err != nil {
						return err
					}
					return effects.Flicker{Target: item.SourceCardID}.Apply(ctx)
				}),
		},
	})
}

// vorinclexOracleID is shared with the back face, which registers under it
// plus "#1" (game.CatalogKey).
const VorinclexOracleID = "166e5680-b8cf-47da-bcae-3591367c5fb4"

// vorinclexFightGrant is the bundle chapter III hands out. An on-disk key.
const vorinclexFightGrant = "the-grand-evolution/fight"

// grandEvolutionMilled is chapter I's "put up to two creature cards from
// among the milled cards onto the battlefield", once the mill is done. An
// on-disk key.
var grandEvolutionMilled = effects.CardTail("the-grand-evolution/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	g, you := ctx.Game, ctx.Controller()
	var candidates []uuid.UUID
	for _, id := range o.IDs {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !c.IsCreature() {
			continue
		}
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    you,
		FromPlayer: you,
		Source:     ctx.Source(),
		Question:   "The Grand Evolution — put up to two creature cards from among the milled cards onto the battlefield",
		Cards:      candidates,
		Min:        0,
		Max:        2,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    grandEvolutionPut,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
})

// grandEvolutionPut puts the chosen creature cards onto the battlefield as
// one event. An on-disk key.
var grandEvolutionPut = effects.AskedPick("choose-cards/the-grand-evolution-put", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	entries := make([]game.BatchEntry, 0, len(r.Picked))
	for _, id := range r.Picked {
		entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneGraveyard})
	}
	return ctx.Game.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{Controller: ctx.Controller()}, game.KeyedTail{})
})
