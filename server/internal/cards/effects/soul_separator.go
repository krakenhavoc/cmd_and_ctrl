package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul Separator — Artifact {3} (Innistrad Remastered, #5948):
//
//	"{5}, {T}, Sacrifice this artifact: Exile target creature card from your
//	 graveyard. Create a token that's a copy of that card, except it's 1/1, it's
//	 a Spirit in addition to its other types, and it has flying. Create a black
//	 Zombie creature token with power equal to that card's power and toughness
//	 equal to that card's toughness."
//
// The Separator is sacrificed as a cost, so the rest happens whatever becomes
// of it. The exile is a keyed continuation (Kaya, Intangible Slayer's shape),
// told the card that reached exile: the token copy takes the card's copiable
// values (CR 707.2) with Kaya's exceptions (1/1, Spirit in addition, flying), and
// the Zombie is the table's black Zombie, sized from the card as it is in exile
// (its power and toughness there, a characteristic-defining ability included).
// If the card left the graveyard in response, the ability is countered on
// resolution for want of its one target (CR 608.2b) and nothing is made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1da52a4-08a0-4427-a461-a1dc6627b98e",
		Name:         "Soul Separator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}, Sacrifice this artifact: Exile target creature card from your graveyard. Create a token that's a copy of that card, except it's 1/1, it's a Spirit in addition to its other types, and it has flying. Create a black Zombie creature token with power equal to that card's power and toughness equal to that card's toughness.",
			Purpose: game.Purpose{Tokens: 2, CreatureTokens: 2},
			Cost:    Plus(ManaCost("{5}"), TapCost(), SacrificeThis()),
			Targets: TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
			Effect:  gapB07SoulSeparatorExile,
		}},
	})
}

func gapB07SoulSeparatorExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := firstLegalTarget(ctx)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return ExileTarget{
		Target: t.ID,
		ThenTail: game.KeyedTail{
			Ref: gapB07SoulSeparatorTokens, Player: item.Controller, Source: item.SourceCardID,
			Carry: []uuid.UUID{t.ID},
		},
	}.Apply(ctx)
}

// gapB07SoulSeparatorTokens makes the 1/1 flying Spirit copy and the sized
// Zombie from the card that reached exile. An on-disk key.
var gapB07SoulSeparatorTokens = CardTail("then/gap-b07-soul-separator-tokens", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	id := o.Tail.Carry[0]
	card, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	spirit := ctx.TextSubtype("Spirit")
	if err := (CreateTokenCopy{
		Controller: o.Tail.Player,
		Copy:       id,
		N:          1,
		Except: func(t *game.Card) {
			t.Power, t.Toughness = 1, 1
			t.VariableToughness = false
			t.PrintedPTKnown = true
			t.TypeLine = socB19WithCreatureAndSubtypeInAddition(t.TypeLine, spirit)
			if !hasKeywordWord(t.Keywords, "flying") {
				t.Keywords = append(t.Keywords, "flying")
			}
		},
	}).Apply(ctx); err != nil {
		return err
	}
	zombie := ctx.TextToken(TokenCard("0/0 black Zombie"))
	zombie.Power, zombie.Toughness = max(card.PowerForComparison(), 0), max(card.CurrentToughness(), 0)
	return ctx.Game.CreateTokenForEffect(o.Tail.Player, zombie, 1)
})
