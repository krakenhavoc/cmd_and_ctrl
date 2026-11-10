package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abyssal Harvester — Creature — Demon Warlock {1}{B}{B}, 3/2 (Foundations,
// #4465):
//
//	"{T}: Exile target creature card from a graveyard that was put there this
//	 turn. Create a token that's a copy of it, except it's a Nightmare in
//	 addition to its other types. Then exile all other Nightmare tokens you
//	 control."
//
// The target is a creature card in a graveyard whose arrival this turn is
// read off the turn's events (PutIntoAGraveyardThisTurn), so a card that
// died, was discarded or was milled this turn is a legal target and one
// that has been there since an earlier turn is not. The exile and the copy
// are The Scarab God's: the card is exiled first, and only one that really
// was exiled is copied, from the printed values the card has there and not
// from anything it was on the battlefield (the ruling, CR 707.2). The
// exception is "in addition to its other types" (CR 707.9b), a copiable
// value of the token, so a copy of it is a Nightmare too. The token is made
// through the keyed creation door, so a copy that asks an entry question
// (a Clone-like creature) is a restore point while it waits; "then exile all
// other Nightmare tokens you control" runs once it has settled and spares
// exactly the tokens that creation made. Every Nightmare token you control
// is exiled, not only ones this ability made (the ruling). A creature
// source, so the tap waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68cc5065-0514-4adf-885f-da554e1db00a",
		Name:         "Abyssal Harvester",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Exile target creature card from a graveyard that was put there this turn. Create a token that's a copy of it, except it's a Nightmare in addition to its other types. Then exile all other Nightmare tokens you control.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
			Cost:    TapCost(),
			Targets: TargetCardInGraveyard("target creature card from a graveyard that was put there this turn",
				Creature(), PutIntoAGraveyardThisTurn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileTargetThenTokenCopy(g, item, abyssalHarvesterCopy)
			},
		}},
	})
}

// abyssalHarvesterCopy is the exile's keyed rest: the Nightmare copy, then
// the exile of the other Nightmare tokens. An on-disk key: never renamed,
// never reused.
var abyssalHarvesterCopy = CardTail("abyssal-harvester/nightmare-copy", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(ctx.Game, o.IDs[0])
	if !ok {
		return nil
	}
	// "Nightmare" is a word of the Harvester's text (CR 612.2).
	tokenCopyIsAlsoASubtype(&tmpl, ctx.Game.TextReaderForEffect(ctx.Item).TextSubtype("Nightmare"))
	return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: ctx.Controller(),
		Source:     ctx.Source(),
		Groups:     []game.TokenGroup{{Template: copiedToken(tmpl), Count: 1}},
	}, ctx.ItemTail(abyssalHarvesterExileOthers))
})

// abyssalHarvesterExileOthers is "Then exile all other Nightmare tokens you
// control": the tokens the creation made are the ones spared. An on-disk key:
// never renamed, never reused.
var abyssalHarvesterExileOthers = CardTail("abyssal-harvester/exile-other-nightmares", func(ctx *Context, o game.TailOutcome) error {
	nightmare := ctx.Game.TextReaderForEffect(ctx.Item).TextSubtype("Nightmare")
	spared := map[string]bool{}
	for _, id := range o.IDs {
		spared[id.String()] = true
	}
	var doomed []game.Card
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsToken() && c.HasSubtype(nightmare) && !spared[c.InstanceID.String()] {
			doomed = append(doomed, c)
		}
	}
	for _, c := range doomed {
		if err := (ExileTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
