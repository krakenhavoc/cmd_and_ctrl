package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Null Chamber — World Enchantment {3}{W}:
//
//	"As this enchantment enters, you and an opponent each choose a card
//	 name other than a basic land card name.
//	 Spells with the chosen names can't be cast and lands with the chosen
//	 names can't be played."
//
// #1997, the last of the world enchantments ADR 0109 PR 7 left out
// (ADR 0109 amendment 2026-10-23, second part). The two names are the
// ordinary as-enters card-name prompt (game.QueueCardNamePromptForEffect,
// Pithing Needle's), asked twice: the controller's answer is the
// permanent's ChosenName, the opponent's its ChosenSecondName. "An
// opponent" is one the controller chooses, and it does not target them
// (the 2013 ruling): with one opponent left nobody is asked, with more
// the controller picks one through a keyed choose-a-player prompt, so a
// table waiting on any of the three questions is a restore point. The
// two names are then asked in APNAP order (CR 101.4), so the later
// chooser knows the earlier name (CR 101.4b). With no opponent left only
// the controller names a card.
//
// "Other than a basic land card name" is a rule on the choice (CR
// 201.4a): both prompts refuse one (PendingChoice.NotBasicLandName), and
// neither the suggestion list nor the bot offers one.
//
// The bans read both names live off this permanent, so the window
// before an answer bans nothing. A spell is judged by the face being
// cast (CR 709.3b, CR 712.11c) and a land by the face being played
// (CR 712.12, game.SpellOrLandHasName). Casting and playing are refused for every
// player, this enchantment's controller included (CR 101.2). A world
// permanent: the world rule (CR 704.5k) puts it into its owner's
// graveyard when a newer one enters.
//
// No simplifications.
const (
	nullChamberCastLabel = "Spells with the chosen names can't be cast."
	nullChamberLandLabel = "Lands with the chosen names can't be played."
	nullChamberYourName  = "Null Chamber — choose a card name other than a basic land card name"
)

// nullChamberOpponentChosen is the keyed answer to "which opponent
// chooses a name". Assigned in init: the key is an on-disk identity,
// never renamed, never reused.
var nullChamberOpponentChosen game.OptionPickThen

func init() {
	nullChamberOpponentChosen = OptionPickThen("option-pick/null-chamber-opponent", nullChamberOpponentPicked)
	Register(Spec{
		OracleID:     "cb75aca6-dfc7-460d-b90c-80b7988f39a8",
		Name:         "Null Chamber",
		Completeness: CompletenessFull,
		AsEnters:     nullChamberAsEnters,
		CastRestrictions: []game.CastRestriction{{
			Label: nullChamberCastLabel,
			Forbids: func(q game.CastQuery) bool {
				return nullChamberNames(q.Source, q.Card)
			},
		}},
		LandPlayRestrictions: []game.LandPlayRestriction{{
			Label: nullChamberLandLabel,
			Forbids: func(q game.LandPlayQuery) bool {
				return nullChamberNames(q.Source, q.Card)
			},
		}},
	})
}

// nullChamberNames reports whether `c` has one of the names chosen for
// the Null Chamber `source`.
func nullChamberNames(source, c game.Card) bool {
	return slices.ContainsFunc(game.ChosenNamesOn(&source), func(n string) bool {
		return game.SpellOrLandHasName(c, n)
	})
}

// nullChamberAsEnters asks the controller which opponent chooses, when
// there is more than one, and otherwise asks for both names at once.
func nullChamberAsEnters(card *game.Card, ctx *Context) error {
	g := ctx.Game
	you := card.Controller
	var opponents []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.ID != you {
			opponents = append(opponents, p.ID)
		}
	}
	switch len(opponents) {
	case 0:
		nullChamberAskNames(g, card.InstanceID, you, uuid.Nil)
		return nil
	case 1:
		nullChamberAskNames(g, card.InstanceID, you, opponents[0])
		return nil
	}
	if g.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  you,
		Among:    opponents,
		Source:   card.InstanceID,
		Question: "Null Chamber — choose an opponent to choose a card name",
		ThenKey:  nullChamberOpponentChosen,
		Carry:    []uuid.UUID{you},
	}) == uuid.Nil {
		// Nobody could be asked: the controller still names a card.
		nullChamberAskNames(g, card.InstanceID, you, uuid.Nil)
	}
	return nil
}

// nullChamberOpponentPicked is the keyed continuation: the chosen
// opponent and the controller name their cards. Nobody chosen (the
// prompt was dropped) leaves the controller's name alone to ask.
func nullChamberOpponentPicked(ctx *Context, r game.OptionPicked) error {
	opponent := uuid.Nil
	if r.Option != nil {
		opponent = r.Option.Player
	}
	nullChamberAskNames(ctx.Game, r.Source, ctx.Controller(), opponent)
	return nil
}

// nullChamberAskNames queues the two name prompts in APNAP order (CR
// 101.4): the controller's fills ChosenName, the opponent's
// ChosenSecondName. A seat that has left the game is not asked.
func nullChamberAskNames(g *game.Game, source, you, opponent uuid.UUID) {
	for _, p := range apnapPlayers(g) {
		switch p {
		case you:
			g.QueueCardNamePromptForEffect(game.CardNamePrompt{
				Chooser: you, Source: source, Reason: nullChamberYourName, NotBasicLandName: true,
			})
		case opponent:
			g.QueueCardNamePromptForEffect(game.CardNamePrompt{
				Chooser: opponent, Source: source, Reason: nullChamberYourName, NotBasicLandName: true, SecondName: true,
			})
		}
	}
}
