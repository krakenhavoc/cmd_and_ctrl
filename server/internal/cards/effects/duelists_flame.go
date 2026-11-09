package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Duelist's Flame — Instant {X}{R}{R} (Final Fantasy, #4462):
//
//	"Until end of turn, target blocked creature you control gets +X/+0 and
//	 gains trample and "Whenever this creature deals combat damage to a
//	 player, look at that many cards from the top of your library. Exile up
//	 to one nonland card from among them and put the rest on the bottom of
//	 your library in a random order. You may cast the exiled card without
//	 paying its mana cost.""
//
// "Blocked" is CR 509.1h's record (BlockedCreature): a creature stays
// blocked after its blockers are gone. The three gains are one duration
// grant (Unnatural Moonrise's shape, ADR 0093): the boost, trample and a
// catalog bundle carrying the trigger, all gone at the cleanup step. The
// trigger is the CREATURE's, so its controller looks, and "that many" is the
// damage the creature dealt, read off the event.
//
// The look is private. The pick is a choose_cards prompt over the nonland
// cards among them, up to one; the rest, lands included, go to the bottom in
// a random order (CR 401.4). The exiled card is offered a free cast while the
// ability resolves (CR 608.2g) and is left in exile if you decline, with X at
// 0 and any additional cost still paid, as the rulings say.
//
// No simplification.
const duelistsFlameGrant = "duelists-flame/look-exile-cast"

func init() {
	Register(Spec{
		OracleID:     "bbc990e4-1601-498f-9031-0cd663139367",
		Name:         "Duelist's Flame",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetCreature("target blocked creature you control", YouControl(), BlockedCreature()),
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Grants: []AbilityGrant{{
			Key: duelistsFlameGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Duelist's Flame — look at that many cards; you may exile a nonland card and cast it free", duelistsFlameLook),
			},
			Text: "Whenever this creature deals combat damage to a player, look at that many cards from the top of your library. Exile up to one nonland card from among them and put the rest on the bottom of your library in a random order. You may cast the exiled card without paying its mana cost.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: target, Power: ctx.X(), Label: "Duelist's Flame — +X/+0"}).Apply(ctx); err != nil {
				return err
			}
			return GrantAbilitiesFor{
				Target: target,
				Keys:   []string{duelistsFlameGrant},
				Also:   []game.Mod{game.AddKeywordsMod("trample")},
				Label:  "Duelist's Flame — trample and the look on combat damage",
			}.Apply(ctx)
		},
	})
}

// duelistsFlameLook is the granted trigger: look at as many cards as the
// damage dealt, and ask which nonland card, if any, to exile.
func duelistsFlameLook(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	player := ctx.Controller()
	looked := g.LookAtTopOfLibraryForEffect(player, item.Trigger.Event.Amount)
	if len(looked) == 0 {
		return nil
	}
	var nonland []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
			nonland = append(nonland, id)
		}
	}
	if len(nonland) == 0 {
		return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Duelist's Flame — exile up to one nonland card from among them (the rest go to the bottom in a random order); you may cast it without paying its mana cost",
		Cards:    nonland,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  duelistsFlamePicked,
		Carry:    looked,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// duelistsFlamePicked puts the rest on the bottom in a random order, then
// exiles the pick and offers its free cast. An on-disk key: never renamed,
// never reused.
var duelistsFlamePicked = askedPick("choose-cards/duelists-flame-exile", func(ctx *Context, r game.CardsChosen) error {
	var picked uuid.UUID
	if len(r.Picked) > 0 {
		picked = r.Picked[0]
	}
	var rest []uuid.UUID
	for _, id := range r.Carry {
		if id != picked {
			rest = append(rest, id)
		}
	}
	if len(rest) > 0 {
		if err := ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, rest); err != nil {
			return err
		}
	}
	if picked == uuid.Nil {
		return nil
	}
	const name = "Duelist's Flame"
	return exileWithPermissionThenOffer(ctx.Game, picked, ctx.Source(),
		ctx.Game.FreeCastFromExilePermission(ctx.Controller(), ctx.Source(), name), name, game.MayCastKeywordFreeCast)
})
