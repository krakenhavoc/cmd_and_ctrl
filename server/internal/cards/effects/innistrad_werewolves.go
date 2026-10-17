package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// innistrad_werewolves.go — the vocabulary of the original Innistrad
// werewolves (ADR 0132 amendment 2026-10-16, #4071). They are not
// daybound: each face prints its own upkeep trigger, and nothing in the
// day/night rules turns them over.
//
//	Front: "At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// Both are CR 603.4 intervening-if triggers on EVERY player's upkeep,
// checked as the upkeep begins and again as the trigger resolves. "Last
// turn" counts every player's spells (game.NoSpellsWereCastLastTurnForEffect,
// game.APlayerCastTwoOrMoreSpellsLastTurnForEffect); the game's first
// turn has no last turn, so neither is true then. The transform is
// TransformThis, so CR 701.27f applies (a face that has already turned
// over since the trigger went on the stack is left alone) and a
// "can't transform" effect (Immerwolf) stops it.
//
// A face's "Whenever this creature transforms into <face>" is
// ThisTransformedInto.

// TransformAtUpkeepIfNoSpellsLastTurn is the front face's trigger.
func TransformAtUpkeepIfNoSpellsLastTurn(name string) game.TriggeredAbility {
	return transformAtUpkeepIf(name, (*game.Game).NoSpellsWereCastLastTurnForEffect)
}

// TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn is the back face's.
func TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn(name string) game.TriggeredAbility {
	return transformAtUpkeepIf(name, (*game.Game).APlayerCastTwoOrMoreSpellsLastTurnForEffect)
}

func transformAtUpkeepIf(name string, cond func(*game.Game) bool) game.TriggeredAbility {
	return On(game.EventBeginUpkeep, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		return cond(g)
	}, name+" — transform this creature", func(g *game.Game, item *game.StackItem) error {
		if !cond(g) {
			return nil // CR 603.4: the condition is checked again as it resolves
		}
		return TransformThis{}.Apply(NewContext(g, item))
	})
}

// ThisTransformedInto is "whenever this creature transforms into <the
// face at index face>" (CR 701.27e): the source turned over in place and
// that face is now up. A permanent that ENTERS on a face never
// transformed into it (CR 712.18), so an "enters or transforms into"
// ability adds EventETB itself (Huntmaster of the Fells).
func ThisTransformedInto(face int) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.Kind == game.EventTransform && ev.CardID == source.InstanceID && ev.Amount == face
	}
}

// NonHumanWerewolvesYouControlCantTransform is Immerwolf's "Non-Human
// Werewolves you control can't transform" (CR 101.2). It is read through
// the layers, so it asks of each permanent's face as it is now: a front
// face that is a Human Werewolf can still turn over, and the back face it
// turns into, a Werewolf that is not a Human, cannot turn back. The two
// types are words of the source's text (CR 612.2, #3984). The engine
// half, and what it does and does not stop, is game/cant_transform.go.
func NonHumanWerewolvesYouControlCantTransform() game.StaticAbility {
	return game.StaticAbility{
		Label: "Non-Human Werewolves you control can't transform.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.Controller == source.Controller &&
				target.HasSubtype(source.TextSubtype("Werewolf")) &&
				!target.HasSubtype(source.TextSubtype("Human"))
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.CantTransform = true
		},
	}
}
