package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b207_helpers.go — shared bodies for Dominaria, batch 207 (#5530).

// OldB207EachPlayerReturnsTwoLands is "each player returns two land cards from
// their graveyard to the battlefield" (Fall of the Thran, chapters II and III)
// as a chapter body: the players are asked in turn order from the controller,
// each for up to two land cards of their own graveyard (every one they have
// when fewer, nobody asked when none), seeing the choices before theirs, and
// the chosen lands then enter together as one entry (CR 603.6a), each under
// its owner's control. Exhume's shape (B38EachPlayerReanimatesOne) with a count.
func OldB207EachPlayerReturnsTwoLands(name string) func(g *game.Game, item *game.StackItem) error {
	question := name + " — return two land cards from your graveyard to the battlefield"
	return func(g *game.Game, item *game.StackItem) error {
		return oldB207AskNext(g, item, question, ApnapPlayers(g), nil)
	}
}

// oldB207AskNext asks the first player in `order` with a land card in their
// graveyard, carrying the players still to ask and the cards chosen so far, and
// puts the chosen cards onto the battlefield together once nobody is left.
func oldB207AskNext(g *game.Game, item *game.StackItem, question string, order, chosen []uuid.UUID) error {
	ctx := NewContext(g, item)
	for i, owner := range order {
		candidates := GraveyardCardIDs(ctx, owner, func(c game.Card) bool { return c.IsLand() })
		if len(candidates) == 0 {
			continue
		}
		n := min(2, len(candidates))
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    owner,
			FromPlayer: owner,
			Source:     item.SourceCardID,
			Question:   question,
			Cards:      candidates,
			Min:        n,
			Max:        n,
			LiveZone:   game.ZoneGraveyard,
			ThenKey:    oldB207LandsPicked,
			Carry:      PackIDLists(order[i+1:], chosen),
			CarryWord:  question,
			Asked:      game.NewAskedItem(item),
		})
		return nil
	}
	return g.PutByEnteringPlayerForEffect(func() error {
		return ReturnFromGraveyardTogether{Targets: chosen}.Apply(ctx)
	})
}

// oldB207LandsPicked is one player's pick answered: the next player is asked,
// or the chosen lands enter together. It is set where the package initialises
// because it asks the next player, which names the key again.
var oldB207LandsPicked game.ChooseCardsThen

var _ = func() bool {
	oldB207LandsPicked = AskedPick("choose-cards/s82-old-b207-each-player-returns-lands", func(ctx *Context, r game.CardsChosen) error {
		rest, chosen := UnpackIDPair(r.Carry)
		next := append(append([]uuid.UUID(nil), chosen...), r.Picked...)
		return oldB207AskNext(ctx.Game, ctx.Item, r.Word, rest, next)
	})
	return true
}()

// OldB207NightmareHorror is Chainer's Torment's chapter III: create an X/X black
// Nightmare Horror token, X being half the controller's life total rounded up
// (read as the chapter resolves), then the token deals X damage to its
// controller. A 0/0 template (the printed token is */*) sized as it is created.
func OldB207NightmareHorror(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil {
		return nil
	}
	x := 0
	if p.Life > 0 {
		x = (p.Life + 1) / 2
	}
	tok := ctx.TextToken(TokenCard("0/0 black Nightmare Horror"))
	tok.Power, tok.Toughness = x, x
	ids, err := g.CreateTokensForEffect(item.Controller, tok, 1, game.TokenEntryOptions{})
	if err != nil || len(ids) == 0 || x == 0 {
		return err
	}
	return DealDamage{Source: ids[0], Target: item.Controller, Amount: x}.Apply(ctx)
}

// OldB207BelzenlokDemonToken is Rite of Belzenlok's Demon: a 6/6 black Demon with
// flying, trample and "At the beginning of your upkeep, sacrifice another
// creature. If you can't, this token deals 6 damage to you."
func OldB207BelzenlokDemonToken() game.Card {
	return TokenFromCatalog(printedOldB207BelzenlokDemonToken)
}

// printedOldB207BelzenlokDemonToken is that token as PRINTED. Its slug is an
// on-disk identity.
func printedOldB207BelzenlokDemonToken() TokenTemplate {
	return TokenTemplate{
		Slug: "demon-rite-of-belzenlok",
		Card: game.Card{
			Name:      "Demon",
			TypeLine:  "Token Creature — Demon",
			Power:     6,
			Toughness: 6,
			Colors:    []string{"B"},
			Keywords:  []string{"flying", "trample"},
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Demon — sacrifice another creature; if you can't, it deals 6 damage to you", oldB207DemonUpkeep),
		},
		Text: "Flying, trample. At the beginning of your upkeep, sacrifice another creature. If you can't, this token deals 6 damage to you.",
	}
}

// oldB207DemonUpkeep is the token's upkeep body: Lord of the Pit's with 6. A
// creature that can't be sacrificed (CR 101.2) is not a candidate.
func oldB207DemonUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.InstanceID != item.SourceCardID &&
			!g.CantBeSacrificedForEffect(c.InstanceID) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 6}.Apply(ctx)
	}
	return SacrificeChoice{
		Player:     item.Controller,
		Candidates: candidates,
		Question:   "Demon — sacrifice another creature",
	}.Apply(ctx)
}
