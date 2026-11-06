package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Rise of Sozin // Fire Lord Sozin — Enchantment — Saga {4}{B}{B},
// transforming into a Legendary Creature — Human Noble, 5/5:
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I — Destroy all creatures.
//	 II — Choose a card name. Search target opponent's graveyard, hand,
//	      and library for up to four cards with that name and exile
//	      them. Then that player shuffles.
//	 III — Exile this Saga, then return it to the battlefield
//	       transformed under your control."
//
//	Fire Lord Sozin: "Menace, firebending 3 (Whenever this creature
//	 attacks, add {R}{R}{R}. This mana lasts until end of combat.)
//	 Whenever Fire Lord Sozin deals combat damage to a player, you may
//	 pay {X}. When you do, put any number of target creature cards with
//	 total mana value X or less from that player's graveyard onto the
//	 battlefield under your control."
//
// The FRONT face is a transforming Saga (The Legend of Roku's shape).
// Chapter I is The Last Ronin's wrath. Chapter II targets the opponent
// as it is put on the stack and names the card only as it resolves (the
// ruling), with Demonic Consultation's card-name prompt; the search is
// the extraction family's three-zone search of that player's graveyard,
// hand and library (AlsoGraveyardAndHand), capped at four and allowed to
// find fewer (the ruling), and the library is that player's to shuffle.
// Permanents on the battlefield are not touched (the ruling). Chapter
// III is the exile-and-return verb (ADR 0079).
//
// The BACK face's firebending is Avatar Roku's trigger with N = 3. Its
// combat damage trigger is a "you may pay {X}" on resolution (MayPayX,
// CR 608.2d) whose paid branch creates a CR 603.12 reflexive trigger
// carrying X and the damaged player. That trigger's clause is "any
// number of target creature cards" in that player's graveyard bounded
// by a TargetTotal of X (the picks' summed mana value), so creature
// cards with mana value 0 always fit (the ruling). It puts every target
// still legal onto the battlefield under its controller's control as
// one entry.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	const name = "The Rise of Sozin"
	Register(Spec{
		OracleID:     theRiseOfSozinOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(ChapterTrigger(1, SagaChapterLabel(name, 1, "destroy all creatures"), func(g *game.Game, item *game.StackItem) error {
				return DestroyAllMatching{Match: Creature()}.Apply(NewContext(g, item))
			}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}}),
			ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "choose a card name; search target opponent's graveyard, hand, and library for up to four cards with that name and exile them"),
				TargetPlayer("target opponent", Opponent()), theRiseOfSozinChapterTwo),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				ChapterExileAndReturnTransformed),
		},
	})
	Register(Spec{
		OracleID:        theRiseOfSozinOracleID + "#1",
		Name:            "Fire Lord Sozin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Fire Lord Sozin — firebending 3", Do(firebending(3))),
			WheneverThisDealsCombatDamageToAPlayer("Fire Lord Sozin — you may pay {X}; when you do, put creature cards with total mana value X or less from that player's graveyard onto the battlefield",
				fireLordSozinCombatDamage),
		},
	})
}

// theRiseOfSozinOracleID is shared with the back face, which registers
// under it plus "#1" (game.CatalogKey).
const theRiseOfSozinOracleID = "9b0b217d-cc9d-4187-89cb-f13160ed2983"

// theRiseOfSozinChapterTwo asks for the name as the chapter resolves.
func theRiseOfSozinChapterTwo(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim := firstLegalTargetPlayer(ctx)
	if victim == uuid.Nil {
		return nil
	}
	g.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
		"The Rise of Sozin — choose a card name",
		withCarry(ctx.ItemTail(theRiseOfSozinNamed), victim))
	return nil
}

// theRiseOfSozinNamed is the search once the name (the tail's Word) is
// known. An on-disk key.
var theRiseOfSozinNamed = CardTail("the-rise-of-sozin/named", func(ctx *Context, o game.TailOutcome) error {
	victim, name := tailCarry0(o), o.Tail.Word
	if victim == uuid.Nil {
		return nil
	}
	return SearchLibrary{
		Player:       ctx.Controller(),
		LibraryOwner: victim,
		Predicate: func(c game.Card) bool {
			return name != "" && game.CardNameMatches(c, name)
		},
		Dest:                 game.ZoneExile,
		Limit:                4,
		AlsoGraveyardAndHand: true,
		Shuffle:              true,
		Reason:               "The Rise of Sozin — search their graveyard, hand and library for up to four cards named " + name,
	}.Apply(ctx)
})

// fireLordSozinCombatDamage is the combat damage trigger's resolution:
// "you may pay {X}", the damaged player carried to the paid branch.
func fireLordSozinCombatDamage(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Cost:  "{X}",
		Label: "Fire Lord Sozin",
		Buys:  "put creature cards with total mana value X or less from that player's graveyard onto the battlefield",
		Goal:  fireLordSozinGoal(item.Trigger.Event.Target),
		Then:  fireLordSozinPaid,
		Carry: []uuid.UUID{item.Trigger.Event.Target},
	}.Apply(NewContext(g, item))
}

// fireLordSozinPaid is "When you do, …": the reflexive trigger, with X
// and the damaged player as its plain data.
var fireLordSozinPaid = PayUnlessBranches("pay-unless/fire-lord-sozin-reanimate",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 || r.Carry[0] == uuid.Nil {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Fire Lord Sozin — put any number of target creature cards with total mana value X or less from that player's graveyard onto the battlefield under your control",
			Body:   fireLordSozinReanimateBody,
			Params: game.EffectParams{Player: r.Carry[0], Amount: r.Number},
		}.Apply(ctx)
	}, nil)

// fireLordSozinReanimateBody is the reflexive trigger: its clause is
// built from the X paid and the player dealt damage.
var fireLordSozinReanimateBody = game.ReflexiveBody("fire-lord-sozin/reanimate", fireLordSozinReanimate,
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetCardInGraveyard("any number of target creature cards with total mana value X or less from that player's graveyard",
			Creature(), OwnedBy(p.Player)).WithCount(0, 0).WithTotal(TotalManaValueAtMost(p.Amount))
	})

// fireLordSozinReanimate puts every target still legal onto the
// battlefield under the trigger's controller's control, together.
func fireLordSozinReanimate(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return putCardsFromZoneTogether(ctx, game.ZoneGraveyard, ids, item.Controller, false)
}

// fireLordSozinGoal is the X a bot pays: the most it can, when that
// player's graveyard holds a creature card that fits; otherwise 0.
func fireLordSozinGoal(player uuid.UUID) func(ctx *Context, ceiling int) int {
	return func(ctx *Context, ceiling int) int {
		p := ctx.Game.PlayerByIDForEffect(player)
		if p == nil || p.Graveyard == nil {
			return 0
		}
		best := 0
		for _, c := range p.Graveyard.Cards {
			if !c.IsCreature() {
				continue
			}
			if mv, ok := ctx.Game.ManaValueForEffect(c); ok && mv <= ceiling && mv > best {
				best = mv
			}
		}
		return best
	}
}
