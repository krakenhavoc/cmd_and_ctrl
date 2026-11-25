package cardss

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Selective Adaptation — {4}{G}{G} Sorcery (Commander 2020):
//
//	"Reveal the top seven cards of your library. Choose from among them a card
//	 with flying, a card with first strike, and so on for double strike,
//	 deathtouch, haste, hexproof, indestructible, lifelink, menace, reach,
//	 trample, and vigilance. Put one of the chosen cards onto the battlefield,
//	 the other chosen cards into your hand, and the rest into your graveyard."
//
// The seven are revealed, then the keywords are walked in the printed order
// (Kathril's keyed walk): each asks for a revealed card with that keyword (read
// where it sits, from its printed keywords), taking the lone candidate without
// asking and skipping a keyword no revealed card has. One card may be chosen
// for several keywords and is then one chosen card. The revealed cards and the
// chosen ones ride the prompt as data (the revealed ids, a nil separator, the
// chosen ids). Then the player picks which chosen permanent card enters the
// battlefield (the rest of the chosen ones go to hand; a chosen card that is not
// a permanent card cannot be the one), and every revealed card that was not
// chosen goes to the graveyard (a put, not a mill).
func init() {
	selectiveAdaptationPick = effects.AskedPick("choose-cards/selective-adaptation-keyword", func(ctx *effects.Context, r game.CardsChosen) error {
		revealed, chosen := selectiveAdaptationDecode(r.Carry)
		next := 0
		if len(r.Nums) > 0 {
			next = r.Nums[0]
		}
		for _, id := range r.Picked {
			if !slices.Contains(chosen, id) {
				chosen = append(chosen, id)
			}
		}
		return selectiveAdaptationStep(ctx, revealed, chosen, next)
	})
	selectiveAdaptationBattlefield = effects.AskedPick("choose-cards/selective-adaptation-battlefield", func(ctx *effects.Context, r game.CardsChosen) error {
		revealed, chosen := selectiveAdaptationDecode(r.Carry)
		onto := uuid.Nil
		if len(r.Picked) > 0 {
			onto = r.Picked[0]
		}
		return selectiveAdaptationFinish(ctx, revealed, chosen, onto)
	})
	effects.Register(effects.Spec{
		OracleID:     "b68df6d7-d69e-446f-9b1b-7a6ca56efbec",
		Name:         "Selective Adaptation",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			revealed := ctx.Game.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 7, "Selective Adaptation — reveal the top seven cards")
			return selectiveAdaptationStep(ctx, revealed, nil, 0)
		},
	})
}

var (
	selectiveAdaptationPick        game.ChooseCardsThen
	selectiveAdaptationBattlefield game.ChooseCardsThen
)

var selectiveAdaptationKeywords = []string{
	"flying", "first strike", "double strike", "deathtouch", "haste", "hexproof",
	"indestructible", "lifelink", "menace", "reach", "trample", "vigilance",
}

func selectiveAdaptationEncode(revealed, chosen []uuid.UUID) []uuid.UUID {
	out := append([]uuid.UUID(nil), revealed...)
	out = append(out, uuid.Nil)
	return append(out, chosen...)
}

func selectiveAdaptationDecode(carry []uuid.UUID) (revealed, chosen []uuid.UUID) {
	i := slices.Index(carry, uuid.Nil)
	if i < 0 {
		return carry, nil
	}
	return slices.Clone(carry[:i]), slices.Clone(carry[i+1:])
}

// selectiveAdaptationStep asks for the next keyword's card from `from` on.
func selectiveAdaptationStep(ctx *effects.Context, revealed, chosen []uuid.UUID, from int) error {
	g := ctx.Game
	for i := from; i < len(selectiveAdaptationKeywords); i++ {
		var cands []uuid.UUID
		for _, id := range effects.CardsStillInALibrary(g, revealed) {
			if c, ok := g.LookupCardForEffect(id); ok && game.HasKeyword(&c, selectiveAdaptationKeywords[i]) {
				cands = append(cands, id)
			}
		}
		switch len(cands) {
		case 0:
			continue
		case 1:
			if !slices.Contains(chosen, cands[0]) {
				chosen = append(chosen, cands[0])
			}
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:   ctx.Controller(),
			Source:    ctx.Source(),
			Question:  "Selective Adaptation — choose a card with " + selectiveAdaptationKeywords[i],
			Cards:     cands,
			Min:       1,
			Max:       1,
			LiveZone:  game.ZoneLibrary,
			ThenKey:   selectiveAdaptationPick,
			Carry:     selectiveAdaptationEncode(revealed, chosen),
			CarryNums: []int{i + 1},
			Asked:     game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	// Every keyword is answered: which chosen permanent card enters?
	var perms []uuid.UUID
	for _, id := range chosen {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsPermanent() {
			perms = append(perms, id)
		}
	}
	switch len(perms) {
	case 0:
		return selectiveAdaptationFinish(ctx, revealed, chosen, uuid.Nil)
	case 1:
		return selectiveAdaptationFinish(ctx, revealed, chosen, perms[0])
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Selective Adaptation — put one of the chosen cards onto the battlefield (the other chosen cards go into your hand)",
		Cards:    perms,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  selectiveAdaptationBattlefield,
		Carry:    selectiveAdaptationEncode(revealed, chosen),
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// selectiveAdaptationFinish puts `onto` onto the battlefield, the other chosen
// cards into the hand and the rest of the revealed cards into the graveyard.
func selectiveAdaptationFinish(ctx *effects.Context, revealed, chosen []uuid.UUID, onto uuid.UUID) error {
	g := ctx.Game
	player := ctx.Controller()
	if onto != uuid.Nil {
		if _, err := g.PutFromLibraryOntoBattlefieldForEffect(onto, game.LibraryEntryOptions{Controller: player}); err != nil {
			return err
		}
	}
	var hand, rest []uuid.UUID
	for _, id := range effects.CardsStillInALibrary(g, chosen) {
		if id != onto {
			hand = append(hand, id)
		}
	}
	for _, id := range revealed {
		if !slices.Contains(chosen, id) {
			rest = append(rest, id)
		}
	}
	if len(hand) > 0 {
		g.TakeFromLibraryToHandForEffect(player, hand)
	}
	return effects.RestIntoGraveyard(g, effects.CardsStillInALibrary(g, rest))
}
