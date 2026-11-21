package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Doors of Durin — Legendary Artifact {3}{R}{G} (The Lord of the Rings:
// Tales of Middle-earth):
//
//	"Whenever you attack, scry 2, then you may reveal the top card of your
//	 library. If a creature card is revealed this way, put it onto the
//	 battlefield tapped and attacking. Until your next turn, it gains
//	 trample if you control a Dwarf and hexproof if you control an Elf."
//
// One trigger per declaration of attackers (OncePerBatch, Araña's). The
// scry's prompt comes first and the rest is its tail (CR 608.2c), then a
// "you may" confirmation, then the reveal of the top card. A creature card
// is put onto the battlefield tapped and attacking: the controller names
// which player or planeswalker it attacks (Raph & Mikey's put, with this
// card's own keys), and it was never declared as an attacker, so no attack
// trigger fires. Only after it has entered does "it" gain what the Dwarf
// and Elf conditions say, judged then (CR 608.2c) and lasting until the
// controller's next turn (CR 611.2b). A card that is not a creature stays
// on top of the library, revealed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d82c77a1-2b81-415d-9f18-47e059c15e01",
		Name:         "Doors of Durin",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Doors of Durin — scry 2, then you may reveal the top card of your library; a creature card is put onto the battlefield tapped and attacking",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.Scry{Player: ctx.Controller(), N: 2, ThenTail: ctx.Tail(doorsOfDurinScried)}.Apply(ctx)
				})),
		},
	})
}

// doorsOfDurinScried is the rest of the ability after the scry: the "you
// may reveal" question. An on-disk key: never renamed, never reused.
var doorsOfDurinScried = effects.CardTail("then/doors-of-durin-scried", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.MayChoice{
		Question: "Doors of Durin — reveal the top card of your library?",
		Then:     doorsOfDurinReveal,
	}.Apply(ctx)
})

var doorsOfDurinReveal = effects.MayYes("confirm/doors-of-durin-reveal", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	g := ctx.Game
	me := ctx.Controller()
	ids := g.RevealTopOfLibraryForEffect(me, ctx.Source(), 1, "Doors of Durin — revealed from the top of the library")
	if len(ids) == 0 {
		return nil
	}
	card := ids[0]
	if c, ok := g.LookupCardForEffect(card); !ok || !c.IsCreature() || c.IsToken() {
		return nil
	}
	refs := g.AttackTargetsForEffect(me)
	if len(refs) == 0 {
		return nil
	}
	if len(refs) == 1 {
		return doorsOfDurinPut(ctx, card, refs[0].ID)
	}
	options := make([]game.ChoiceOption, len(refs))
	carry := []uuid.UUID{card}
	for i, r := range refs {
		options[i] = game.ChoiceOption{Label: "Attack " + effects.AttackTargetName(g, r.ID)}
		if r.Kind != game.AttackTargetPlayer {
			options[i].Cards = []uuid.UUID{r.ID}
		}
		carry = append(carry, r.ID)
	}
	return effects.PickOption{
		Question: "Doors of Durin — what does it attack?",
		Options:  options,
		ThenKey:  doorsOfDurinAttacks,
		Carry:    carry,
	}.Apply(ctx)
})

var doorsOfDurinAttacks = effects.OptionPickThen("option-pick/doors-of-durin-attacks", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) < 2 {
		return nil
	}
	card, targets := r.Carry[0], r.Carry[1:]
	index := r.Index
	if index < 0 || index >= len(targets) {
		index = 0
	}
	return doorsOfDurinPut(ctx, card, targets[index])
})

// doorsOfDurinPut puts the revealed creature card onto the battlefield tapped
// and attacking `target`, if it is still on top of the library where the
// instruction found it, and gives it the conditional keywords once it is there.
func doorsOfDurinPut(ctx *effects.Context, card, target uuid.UUID) error {
	g := ctx.Game
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneLibrary {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: card, From: game.ZoneLibrary}},
		game.ZoneEntryOptions{Controller: ctx.Controller(), Tapped: true, Attacking: target},
		ctx.Tail(doorsOfDurinEntered))
}

// doorsOfDurinEntered is "until your next turn, it gains trample if you
// control a Dwarf and hexproof if you control an Elf", run once the card has
// entered. An on-disk key: never renamed, never reused.
var doorsOfDurinEntered = effects.EnteredTail("then/doors-of-durin-entered", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	me := ctx.Controller()
	var keywords []string
	dwarf, elf := ctx.TextSubtype("Dwarf"), ctx.TextSubtype("Elf")
	if effects.CountControlled(ctx.Game, me, func(c game.Card) bool { return effects.CardHasSubtype(c, dwarf) }) > 0 {
		keywords = append(keywords, "trample")
	}
	if effects.CountControlled(ctx.Game, me, func(c game.Card) bool { return effects.CardHasSubtype(c, elf) }) > 0 {
		keywords = append(keywords, "hexproof")
	}
	if len(keywords) == 0 {
		return nil
	}
	return effects.ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddKeywordsMod(keywords...)},
		Duration: effects.DurationUntilYourNextTurn(ctx, me),
		Label:    "Doors of Durin — until your next turn",
	}.Apply(ctx)
})
