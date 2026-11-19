package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB55PutTargetCreatureCardOntoTheBattlefield is "put target creature card
// from a graveyard onto the battlefield under your control" as an ability's
// whole body: the announced card, if it is still a legal pick as the ability
// resolves (CR 608.2b), enters under the activator's control.
func gapB55PutTargetCreatureCardOntoTheBattlefield(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	card, found := b16FirstLegalTargetCard(ctx)
	if !found {
		return nil
	}
	return ReturnFromGraveyard{Target: card, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
}

// gapB55BirthingRitualSacrificed is the "if you do" of Birthing Ritual: the
// sacrifice has settled, so X is 1 plus the sacrificed creature's mana value
// as it last existed (CR 608.2h; a token that is no copy is 0). A creature
// card with mana value X or less from among the looked-at cards may be put
// onto the battlefield, and the rest go to the bottom in a random order
// whether or not one was. A sacrifice that did not happen (a replacement
// kept the creature) is "if you didn't": nothing is put, and the rest still
// go to the bottom. An on-disk key.
var gapB55BirthingRitualSacrificed = CardTail("gap-b55-birthing-ritual/sacrificed", func(ctx *Context, o game.TailOutcome) error {
	g, player := ctx.Game, ctx.Controller()
	if len(o.IDs) == 0 {
		return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, o.Tail.Carry)
	}
	x := 1 + departedManaValue(g, o.IDs[0])
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    o.Tail.Carry,
		Match:    And(Creature(), ManaValueLE(x)),
		Max:      1,
		Optional: true,
		Label:    "Birthing Ritual — you may put a creature card with mana value " + strconv.Itoa(x) + " or less from among them onto the battlefield",
		ThenTail: gapB55BirthingRitualRest,
	}.Apply(ctx)
})

// gapB55BirthingRitualRest is "put the rest on the bottom of your library in
// a random order" after the put. An on-disk key.
var gapB55BirthingRitualRest = PutFromLibraryTail("gap-b55-birthing-ritual/rest",
	func(ctx *Context, res PutFromLibraryResult) error { return PutRestOnBottomInRandomOrder(ctx.Game, res) })

// gapB55BirthingRitualPick is the sacrifice question: "you may sacrifice a
// creature", with the seven looked-at cards riding the pick as data. A
// pick of nothing is "you didn't": the seven go to the bottom in a random
// order. An on-disk key.
var gapB55BirthingRitualPick = RegisterPermanentsPick("gap-b55-birthing-ritual/sacrifice", PermanentsDef{
	ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
		cards := picked.Cards()
		if len(cards) == 0 {
			return ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, r.Carry)
		}
		tail := game.KeyedTail{Ref: gapB55BirthingRitualSacrificed, Player: ctx.Controller(), Source: r.Source,
			Carry: append([]uuid.UUID(nil), r.Carry...), Asked: game.NewAskedItem(ctx.Item)}
		return ctx.Game.SacrificeAllKeyedForEffect(uuid.Nil, cards, tail)
	},
})

// gapB55BirthingRitualLook is the trigger's body: look at the top seven
// cards, then ask whether to sacrifice a creature.
func gapB55BirthingRitualLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	looked := g.LookAtTopOfLibraryForEffect(ctx.Controller(), 7)
	return ChoosePermanents{
		Question: "Birthing Ritual — you may sacrifice a creature",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			ids := permanentsControlledByMatching(g, of, Creature())
			if len(ids) == 0 {
				return nil, 0, 0
			}
			return ids, 0, 1
		},
		Keyed: gapB55BirthingRitualPick,
		Carry: looked,
	}.Apply(ctx)
}
